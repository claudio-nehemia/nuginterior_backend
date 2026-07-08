package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ApprovalMaterialService interface {
	GetAll(ctx context.Context) ([]dto.ApprovalMaterialResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.ApprovalMaterialResponse, error)
	GetByOrderID(ctx context.Context, orderID uint) (*dto.ApprovalMaterialResponse, error)
	Response(ctx context.Context, orderID uint, email string) (*dto.ApprovalMaterialResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateApprovalMaterialRequest) (*dto.ApprovalMaterialResponse, error)
}

type approvalMaterialService struct {
	repo       repository.ApprovalMaterialRepository
	settingSvc SettingService
	db         *gorm.DB
	logger     *zap.Logger
	logTaskSvc ProjectLogTaskService
}

func NewApprovalMaterialService(
	repo repository.ApprovalMaterialRepository,
	settingSvc SettingService,
	db *gorm.DB,
	logger *zap.Logger,
	logTaskSvc ProjectLogTaskService,
) ApprovalMaterialService {
	return &approvalMaterialService{
		repo:       repo,
		settingSvc: settingSvc,
		db:         db,
		logger:     logger,
		logTaskSvc: logTaskSvc,
	}
}

func (s *approvalMaterialService) checkGambarKerjaApproved(ctx context.Context, orderID uint) error {
	var gk entity.GambarKerja
	err := s.db.WithContext(ctx).Where("order_id = ?", orderID).First(&gk).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("approval material terkunci! Gambar kerja belum diupload/disetujui")
		}
		return err
	}

	if gk.Status != "approved" {
		return errors.New("approval material terkunci! Status gambar kerja belum disetujui")
	}

	return nil
}

func (s *approvalMaterialService) GetAll(ctx context.Context) ([]dto.ApprovalMaterialResponse, error) {
	// Query orders that have an approved GambarKerja
	var eligibleOrders []entity.Order
	err := s.db.WithContext(ctx).
		Joins("JOIN gambar_kerja gk ON gk.order_id = orders.id").
		Where("gk.status = ?", "approved").
		Find(&eligibleOrders).Error

	if err != nil {
		return nil, err
	}

	var result []dto.ApprovalMaterialResponse
	for _, order := range eligibleOrders {
		am, err := s.repo.FindByOrderID(ctx, order.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Not response'd/created yet, return preview with default info
				orderCopy := order
				brief := toOrderBriefResponse(&orderCopy)
				result = append(result, dto.ApprovalMaterialResponse{
					OrderID: order.ID,
					Status:  "pending",
					Order:   brief,
					Items:   []dto.ApprovalMaterialItemResponse{},
				})
				continue
			}
			return nil, err
		}

		// Items exist, convert to response DTO
		resp := s.toApprovalMaterialResponse(*am)
		result = append(result, *resp)
	}

	return result, nil
}

func (s *approvalMaterialService) GetByID(ctx context.Context, id uint) (*dto.ApprovalMaterialResponse, error) {
	am, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toApprovalMaterialResponse(*am), nil
}

func (s *approvalMaterialService) GetByOrderID(ctx context.Context, orderID uint) (*dto.ApprovalMaterialResponse, error) {
	if err := s.checkGambarKerjaApproved(ctx, orderID); err != nil {
		return nil, err
	}

	am, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Check if response is disabled to auto-initialize
			isEnabled, _ := s.settingSvc.IsEnabled(ctx, "response_enabled")
			if !isEnabled {
				// Auto-initialize
				newAM, errCreate := s.initApprovalMaterial(ctx, orderID, "System Auto")
				if errCreate != nil {
					return nil, errCreate
				}
				return s.toApprovalMaterialResponse(*newAM), nil
			}
			return nil, errors.New("approval material belum diresponse")
		}
		return nil, err
	}

	return s.toApprovalMaterialResponse(*am), nil
}

func (s *approvalMaterialService) Response(ctx context.Context, orderID uint, email string) (*dto.ApprovalMaterialResponse, error) {
	if err := s.checkGambarKerjaApproved(ctx, orderID); err != nil {
		return nil, err
	}

	am, err := s.repo.FindByOrderID(ctx, orderID)
	now := time.Now()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Initialize new sheet
			newAM, errCreate := s.initApprovalMaterial(ctx, orderID, email)
			if errCreate != nil {
				return nil, errCreate
			}
			am = newAM
		} else {
			return nil, err
		}
	} else {
		if am.ResponseBy == "" {
			am.ResponseBy = email
			am.ResponseTime = &now
			if err := s.repo.Update(ctx, am); err != nil {
				return nil, err
			}
		}
	}

	// Update order stage to approval_material and log transition
	if err := s.logTaskSvc.TransitionStage(ctx, orderID, "approval_material", email); err != nil {
		s.logger.Error("Failed to update order stage to approval_material", zap.Error(err))
	}

	return s.GetByOrderID(ctx, orderID)
}

func (s *approvalMaterialService) initApprovalMaterial(ctx context.Context, orderID uint, userEmail string) (*entity.ApprovalMaterial, error) {
	now := time.Now()
	newAM := &entity.ApprovalMaterial{
		OrderID:      orderID,
		Status:       "pending",
		ResponseBy:   userEmail,
		ResponseTime: &now,
	}

	// Save main record first
	if err := s.repo.Create(ctx, newAM); err != nil {
		return nil, err
	}

	// Consolidate materials from InputItem
	items, err := s.gatherUniqueMaterials(ctx, orderID)
	if err != nil {
		s.logger.Warn("Consolidation yielded empty or error: ", zap.Error(err))
	}

	for i := range items {
		items[i].ApprovalMaterialID = newAM.ID
		if err := s.repo.SaveItem(ctx, &items[i]); err != nil {
			s.logger.Error("Failed to save default approval item", zap.Error(err))
		}
	}

	// Reload with items
	return s.repo.FindByOrderID(ctx, orderID)
}

func (s *approvalMaterialService) gatherUniqueMaterials(ctx context.Context, orderID uint) ([]entity.ApprovalMaterialItem, error) {
	var inputItem entity.InputItem
	err := s.db.WithContext(ctx).
		Preload("Rooms.BahanBakus.BahanBaku").
		Preload("Rooms.FinishingDalams.Item").
		Preload("Rooms.FinishingLuars.Item").
		Preload("Rooms.Aksesoris.Item").
		Where("order_id = ? AND status = ?", orderID, "approved").
		First(&inputItem).Error
	if err != nil {
		// Fallback to any input item if approved is not found
		err = s.db.WithContext(ctx).
			Preload("Rooms.BahanBakus.BahanBaku").
			Preload("Rooms.FinishingDalams.Item").
			Preload("Rooms.FinishingLuars.Item").
			Preload("Rooms.Aksesoris.Item").
			Where("order_id = ?", orderID).
			First(&inputItem).Error
		if err != nil {
			return nil, err
		}
	}

	// Maps to dedup by name/id
	bahanBakuMap := make(map[uint]string)
	finishingDalamMap := make(map[uint]string)
	finishingLuarMap := make(map[uint]string)
	aksesorisMap := make(map[uint]string)

	for _, room := range inputItem.Rooms {
		for _, bb := range room.BahanBakus {
			if bb.BahanBaku != nil {
				bahanBakuMap[bb.BahanBakuID] = bb.BahanBaku.NamaBahanBaku
			}
		}
		for _, fd := range room.FinishingDalams {
			if fd.Item != nil {
				finishingDalamMap[fd.ItemID] = fd.Item.NamaItem
			}
		}
		for _, fl := range room.FinishingLuars {
			if fl.Item != nil {
				finishingLuarMap[fl.ItemID] = fl.Item.NamaItem
			}
		}
		for _, aks := range room.Aksesoris {
			if aks.Item != nil {
				aksesorisMap[aks.ItemID] = aks.Item.NamaItem
			}
		}
	}

	var items []entity.ApprovalMaterialItem
	emptyJSON := datatypes.JSON("[]")

	// Map to model items
	for id, name := range bahanBakuMap {
		items = append(items, entity.ApprovalMaterialItem{
			Category:     "bahan_baku",
			SourceID:     id,
			ItemName:     name,
			KodeMaterial: emptyJSON,
			BrandSpek:    emptyJSON,
		})
	}
	for id, name := range finishingLuarMap {
		items = append(items, entity.ApprovalMaterialItem{
			Category:     "finishing_luar",
			SourceID:     id,
			ItemName:     name,
			KodeMaterial: emptyJSON,
			BrandSpek:    emptyJSON,
		})
	}
	for id, name := range finishingDalamMap {
		items = append(items, entity.ApprovalMaterialItem{
			Category:     "finishing_dalam",
			SourceID:     id,
			ItemName:     name,
			KodeMaterial: emptyJSON,
			BrandSpek:    emptyJSON,
		})
	}
	for id, name := range aksesorisMap {
		items = append(items, entity.ApprovalMaterialItem{
			Category:     "aksesoris",
			SourceID:     id,
			ItemName:     name,
			KodeMaterial: emptyJSON,
			BrandSpek:    emptyJSON,
		})
	}

	return items, nil
}

func (s *approvalMaterialService) Update(ctx context.Context, id uint, req dto.UpdateApprovalMaterialRequest) (*dto.ApprovalMaterialResponse, error) {
	am, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update items
	for _, reqItem := range req.Items {
		dbItem, errItem := s.repo.FindItemByID(ctx, reqItem.ID)
		if errItem != nil {
			s.logger.Warn("Failed to find item for update", zap.Uint("itemID", reqItem.ID))
			continue
		}

		dbItem.Area = reqItem.Area
		dbItem.Foto = reqItem.Foto
		dbItem.Notes = reqItem.Notes

		// Serialize array of strings to JSON
		if reqItem.KodeMaterial == nil {
			reqItem.KodeMaterial = []string{}
		}
		if reqItem.BrandSpek == nil {
			reqItem.BrandSpek = []string{}
		}

		kodeJSON, _ := json.Marshal(reqItem.KodeMaterial)
		brandJSON, _ := json.Marshal(reqItem.BrandSpek)

		dbItem.KodeMaterial = datatypes.JSON(kodeJSON)
		dbItem.BrandSpek = datatypes.JSON(brandJSON)

		if errSave := s.repo.SaveItem(ctx, dbItem); errSave != nil {
			return nil, fmt.Errorf("gagal menyimpan item %d: %w", reqItem.ID, errSave)
		}
	}

	_ = s.logTaskSvc.RecordTouch(ctx, am.OrderID, "approval_material", "")

	// Check status update
	if req.Status == "completed" {
		am.Status = "completed"
		// Change tahapan_proyek to workplan and log transition
		if errStage := s.logTaskSvc.TransitionStage(ctx, am.OrderID, "workplan", ""); errStage != nil {
			s.logger.Error("Failed to transition order to workplan stage", zap.Error(errStage))
		}
	} else if req.Status == "pending" {
		am.Status = "pending"
	}

	if errSave := s.repo.Update(ctx, am); errSave != nil {
		return nil, errSave
	}

	return s.GetByID(ctx, id)
}

func (s *approvalMaterialService) toApprovalMaterialResponse(am entity.ApprovalMaterial) *dto.ApprovalMaterialResponse {
	var brief *dto.OrderBriefResponse
	if am.Order != nil {
		brief = toOrderBriefResponse(am.Order)
	}

	items := make([]dto.ApprovalMaterialItemResponse, len(am.Items))
	for i, item := range am.Items {
		var codes []string
		var specs []string

		if len(item.KodeMaterial) > 0 {
			_ = json.Unmarshal(item.KodeMaterial, &codes)
		}
		if len(item.BrandSpek) > 0 {
			_ = json.Unmarshal(item.BrandSpek, &specs)
		}

		items[i] = dto.ApprovalMaterialItemResponse{
			ID:                 item.ID,
			ApprovalMaterialID: item.ApprovalMaterialID,
			Category:           item.Category,
			SourceID:           item.SourceID,
			ItemName:           item.ItemName,
			Area:               item.Area,
			Foto:               item.Foto,
			KodeMaterial:       codes,
			BrandSpek:          specs,
			Notes:              item.Notes,
			CreatedAt:          item.CreatedAt,
			UpdatedAt:          item.UpdatedAt,
		}
	}

	return &dto.ApprovalMaterialResponse{
		ID:                    am.ID,
		OrderID:               am.OrderID,
		Status:                am.Status,
		ResponseBy:            am.ResponseBy,
		ResponseTime:          am.ResponseTime,
		MarketingResponseBy:   am.MarketingResponseBy,
		MarketingResponseTime: am.MarketingResponseTime,
		CreatedAt:             am.CreatedAt,
		UpdatedAt:             am.UpdatedAt,
		Order:                 brief,
		Items:                 items,
	}
}

// Helpers
func toOrderBriefResponse(o *entity.Order) *dto.OrderBriefResponse {
	lamaKontrak := ""
	if len(o.Contracts) > 0 {
		lamaKontrak = o.Contracts[0].LamaKontrak
	}
	return &dto.OrderBriefResponse{
		ID:            o.ID,
		NomorOrder:    o.NomorOrder,
		NamaProject:   o.NamaProject,
		JenisInterior: o.JenisInterior,
		NamaCustomer:  o.NamaCustomer,
		Alamat:        o.Alamat,
		TahapanProyek: o.TahapanProyek,
		PaymentStatus: o.PaymentStatus,
		LamaKontrak:   lamaKontrak,
	}
}
