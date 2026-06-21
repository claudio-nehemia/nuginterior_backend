package service

import (
	"context"
	"errors"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"go.uber.org/zap"
)

type InputItemService interface {
	GetAll(ctx context.Context) ([]dto.InputItemResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.InputItemResponse, error)
	GetByDesainFinalID(ctx context.Context, dfID uint) (*dto.InputItemResponse, error)
	Create(ctx context.Context, req dto.CreateInputItemRequest) (*dto.InputItemResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateInputItemRequest) (*dto.InputItemResponse, error)
	Delete(ctx context.Context, id uint) error

	InputItemResponseDesigner(ctx context.Context, dfID uint, designerName string) (*dto.InputItemResponse, error)
	InputItemResponseMarketing(ctx context.Context, dfID uint, marketingName string) (*dto.InputItemResponse, error)
}

type inputItemService struct {
	repo       repository.InputItemRepository
	logger     *zap.Logger
	logTaskSvc ProjectLogTaskService
}

func NewInputItemService(repo repository.InputItemRepository, logger *zap.Logger, logTaskSvc ProjectLogTaskService) InputItemService {
	return &inputItemService{repo: repo, logger: logger, logTaskSvc: logTaskSvc}
}

func (s *inputItemService) GetAll(ctx context.Context) ([]dto.InputItemResponse, error) {
	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.InputItemResponse, len(list))
	for i, item := range list {
		res[i] = toInputItemResponse(item)
	}
	return res, nil
}

func (s *inputItemService) GetByID(ctx context.Context, id uint) (*dto.InputItemResponse, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toInputItemResponse(*item)
	return &resp, nil
}

func (s *inputItemService) GetByDesainFinalID(ctx context.Context, dfID uint) (*dto.InputItemResponse, error) {
	item, err := s.repo.FindByDesainFinalID(ctx, dfID)
	if err != nil {
		return nil, err
	}
	resp := toInputItemResponse(*item)
	return &resp, nil
}

func (s *inputItemService) Create(ctx context.Context, req dto.CreateInputItemRequest) (*dto.InputItemResponse, error) {
	// Check if already exists for this DesainFinalID to enforce unique relation
	existing, err := s.repo.FindByDesainFinalID(ctx, req.DesainFinalID)
	if err == nil && existing != nil {
		return nil, errors.New("rincian item untuk desain final ini sudah dibuat")
	}

	// Validation for "approved" status
	if req.Status == "approved" {
		if len(req.Rooms) == 0 {
			return nil, errors.New("ruangan harus diisi minimal 1 untuk disetujui")
		}
		for _, roomReq := range req.Rooms {
			if roomReq.NamaRuangan == "" {
				return nil, errors.New("nama ruangan tidak boleh kosong")
			}
			if roomReq.ProdukID == nil {
				return nil, errors.New("produk harus dipilih untuk setiap ruangan")
			}
			if roomReq.Qty <= 0 {
				return nil, errors.New("quantity produk harus lebih dari 0")
			}
			if roomReq.Panjang <= 0 || roomReq.Lebar <= 0 || roomReq.Tinggi <= 0 {
				return nil, errors.New("panjang, lebar, dan tinggi harus diisi (lebih dari 0) untuk ruangan: " + roomReq.NamaRuangan)
			}
		}
	}

	item := &entity.InputItem{
		DesainFinalID: req.DesainFinalID,
		OrderID:       req.OrderID,
		Status:        req.Status,
	}

	for _, roomReq := range req.Rooms {
		room := entity.InputItemRoom{
			NamaRuangan: roomReq.NamaRuangan,
			ProdukID:    roomReq.ProdukID,
			Qty:         roomReq.Qty,
			Panjang:     roomReq.Panjang,
			Lebar:       roomReq.Lebar,
			Tinggi:      roomReq.Tinggi,
		}

		for _, bbID := range roomReq.BahanBakus {
			room.BahanBakus = append(room.BahanBakus, entity.InputItemRoomBahanBaku{
				BahanBakuID: bbID,
			})
		}

		for _, fdReq := range roomReq.FinishingDalams {
			room.FinishingDalams = append(room.FinishingDalams, entity.InputItemRoomFinishing{
				ItemID: fdReq.ItemID,
				Type:   "dalam",
				Notes:  fdReq.Notes,
			})
		}

		for _, flReq := range roomReq.FinishingLuars {
			room.FinishingLuars = append(room.FinishingLuars, entity.InputItemRoomFinishing{
				ItemID: flReq.ItemID,
				Type:   "luar",
				Notes:  flReq.Notes,
			})
		}

		for _, aksReq := range roomReq.Aksesoris {
			room.Aksesoris = append(room.Aksesoris, entity.InputItemRoomAksesoris{
				ItemID: aksReq.ItemID,
				Qty:    aksReq.Qty,
				Notes:  aksReq.Notes,
			})
		}

		item.Rooms = append(item.Rooms, room)
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	_ = s.logTaskSvc.RecordTouch(ctx, item.OrderID, "input_item", "")

	if item.Status == "approved" {
		if errStage := s.logTaskSvc.TransitionStage(ctx, item.OrderID, "rab", ""); errStage != nil {
			s.logger.Error("Failed to update order stage to rab", zap.Error(errStage))
		}
	}

	created, err := s.repo.FindByID(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	resp := toInputItemResponse(*created)
	return &resp, nil
}

func (s *inputItemService) Update(ctx context.Context, id uint, req dto.UpdateInputItemRequest) (*dto.InputItemResponse, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validation for "approved" status
	if req.Status == "approved" {
		if len(req.Rooms) == 0 {
			return nil, errors.New("ruangan harus diisi minimal 1 untuk disetujui")
		}
		for _, roomReq := range req.Rooms {
			if roomReq.NamaRuangan == "" {
				return nil, errors.New("nama ruangan tidak boleh kosong")
			}
			if roomReq.ProdukID == nil {
				return nil, errors.New("produk harus dipilih untuk setiap ruangan")
			}
			if roomReq.Qty <= 0 {
				return nil, errors.New("quantity produk harus lebih dari 0")
			}
			if roomReq.Panjang <= 0 || roomReq.Lebar <= 0 || roomReq.Tinggi <= 0 {
				return nil, errors.New("panjang, lebar, dan tinggi harus diisi (lebih dari 0) untuk ruangan: " + roomReq.NamaRuangan)
			}
		}
	}

	item.Status = req.Status
	item.Rooms = nil // Clean relationships

	for _, roomReq := range req.Rooms {
		room := entity.InputItemRoom{
			NamaRuangan: roomReq.NamaRuangan,
			ProdukID:    roomReq.ProdukID,
			Qty:         roomReq.Qty,
			Panjang:     roomReq.Panjang,
			Lebar:       roomReq.Lebar,
			Tinggi:      roomReq.Tinggi,
		}

		for _, bbID := range roomReq.BahanBakus {
			room.BahanBakus = append(room.BahanBakus, entity.InputItemRoomBahanBaku{
				BahanBakuID: bbID,
			})
		}

		for _, fdReq := range roomReq.FinishingDalams {
			room.FinishingDalams = append(room.FinishingDalams, entity.InputItemRoomFinishing{
				ItemID: fdReq.ItemID,
				Type:   "dalam",
				Notes:  fdReq.Notes,
			})
		}

		for _, flReq := range roomReq.FinishingLuars {
			room.FinishingLuars = append(room.FinishingLuars, entity.InputItemRoomFinishing{
				ItemID: flReq.ItemID,
				Type:   "luar",
				Notes:  flReq.Notes,
			})
		}

		for _, aksReq := range roomReq.Aksesoris {
			room.Aksesoris = append(room.Aksesoris, entity.InputItemRoomAksesoris{
				ItemID: aksReq.ItemID,
				Qty:    aksReq.Qty,
				Notes:  aksReq.Notes,
			})
		}

		item.Rooms = append(item.Rooms, room)
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	_ = s.logTaskSvc.RecordTouch(ctx, item.OrderID, "input_item", "")

	if item.Status == "approved" {
		if errStage := s.logTaskSvc.TransitionStage(ctx, item.OrderID, "rab", ""); errStage != nil {
			s.logger.Error("Failed to update order stage to rab", zap.Error(errStage))
		}
	}

	updated, err := s.repo.FindByID(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	resp := toInputItemResponse(*updated)
	return &resp, nil
}

func (s *inputItemService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func toInputItemResponse(item entity.InputItem) dto.InputItemResponse {
	resp := dto.InputItemResponse{
		ID:                    item.ID,
		DesainFinalID:         item.DesainFinalID,
		OrderID:               item.OrderID,
		Status:                item.Status,
		ResponseTime:          item.ResponseTime,
		ResponseBy:            item.ResponseBy,
		MarketingResponseTime: item.MarketingResponseTime,
		MarketingResponseBy:   item.MarketingResponseBy,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
		Rooms:                 []dto.RoomResponse{},
	}

	if item.Order != nil {
		resp.Order = &dto.InputItemOrderResponse{
			ID:             item.Order.ID,
			NomorOrder:     item.Order.NomorOrder,
			NamaProject:    item.Order.NamaProject,
			NamaCustomer:   item.Order.NamaCustomer,
			NamaPerusahaan: item.Order.NamaPerusahaan,
			JenisInterior:  item.Order.JenisInterior,
		}
	}

	for _, room := range item.Rooms {
		produkName := ""
		if room.Produk != nil {
			produkName = room.Produk.NamaProduk
		}

		roomResp := dto.RoomResponse{
			ID:              room.ID,
			NamaRuangan:     room.NamaRuangan,
			ProdukID:        room.ProdukID,
			NamaProduk:      produkName,
			Qty:             room.Qty,
			Panjang:         room.Panjang,
			Lebar:           room.Lebar,
			Tinggi:          room.Tinggi,
			BahanBakus:      []dto.RoomBahanBakuResponse{},
			FinishingDalams: []dto.RoomFinishingResponse{},
			FinishingLuars:  []dto.RoomFinishingResponse{},
			Aksesoris:       []dto.RoomAksesorisResponse{},
		}

		for _, bb := range room.BahanBakus {
			namaBahan := ""
			if bb.BahanBaku != nil {
				namaBahan = bb.BahanBaku.NamaBahanBaku
			}
			roomResp.BahanBakus = append(roomResp.BahanBakus, dto.RoomBahanBakuResponse{
				ID:          bb.ID,
				BahanBakuID: bb.BahanBakuID,
				NamaBahan:   namaBahan,
			})
		}

		for _, f := range room.FinishingDalams {
			namaItem := ""
			if f.Item != nil {
				namaItem = f.Item.NamaItem
			}
			if f.Type == "dalam" {
				roomResp.FinishingDalams = append(roomResp.FinishingDalams, dto.RoomFinishingResponse{
					ID:     f.ID,
					ItemID: f.ItemID,
					Nama:   namaItem,
					Notes:  f.Notes,
				})
			}
		}

		for _, f := range room.FinishingLuars {
			namaItem := ""
			if f.Item != nil {
				namaItem = f.Item.NamaItem
			}
			if f.Type == "luar" {
				roomResp.FinishingLuars = append(roomResp.FinishingLuars, dto.RoomFinishingResponse{
					ID:     f.ID,
					ItemID: f.ItemID,
					Nama:   namaItem,
					Notes:  f.Notes,
				})
			}
		}

		for _, aks := range room.Aksesoris {
			namaItem := ""
			if aks.Item != nil {
				namaItem = aks.Item.NamaItem
			}
			roomResp.Aksesoris = append(roomResp.Aksesoris, dto.RoomAksesorisResponse{
				ID:     aks.ID,
				ItemID: aks.ItemID,
				Nama:   namaItem,
				Qty:    aks.Qty,
				Notes:  aks.Notes,
			})
		}

		resp.Rooms = append(resp.Rooms, roomResp)
	}

	return resp
}

func (s *inputItemService) InputItemResponseDesigner(ctx context.Context, dfID uint, designerName string) (*dto.InputItemResponse, error) {
	item, err := s.repo.FindByDesainFinalID(ctx, dfID)
	now := time.Now()
	if err != nil {
		orderID, errOrder := s.repo.GetOrderIDByDesainFinalID(ctx, dfID)
		if errOrder != nil {
			return nil, errors.New("desain final tidak ditemukan atau tidak valid")
		}

		item = &entity.InputItem{
			DesainFinalID: dfID,
			OrderID:       orderID,
			Status:        "draft",
			ResponseTime:  &now,
			ResponseBy:    designerName,
		}
		if errCreate := s.repo.Create(ctx, item); errCreate != nil {
			return nil, errCreate
		}
	} else {
		item.ResponseTime = &now
		item.ResponseBy = designerName
		if errUpdate := s.repo.Update(ctx, item); errUpdate != nil {
			return nil, errUpdate
		}
	}

	_ = s.logTaskSvc.RecordTouch(ctx, item.OrderID, "input_item", designerName)

	created, errFetch := s.repo.FindByID(ctx, item.ID)
	if errFetch != nil {
		return nil, errFetch
	}
	resp := toInputItemResponse(*created)
	return &resp, nil
}

func (s *inputItemService) InputItemResponseMarketing(ctx context.Context, dfID uint, marketingName string) (*dto.InputItemResponse, error) {
	item, err := s.repo.FindByDesainFinalID(ctx, dfID)
	now := time.Now()
	if err != nil {
		orderID, errOrder := s.repo.GetOrderIDByDesainFinalID(ctx, dfID)
		if errOrder != nil {
			return nil, errors.New("desain final tidak ditemukan atau tidak valid")
		}

		item = &entity.InputItem{
			DesainFinalID:         dfID,
			OrderID:               orderID,
			Status:                "draft",
			MarketingResponseTime: &now,
			MarketingResponseBy:   marketingName,
		}
		if errCreate := s.repo.Create(ctx, item); errCreate != nil {
			return nil, errCreate
		}
	} else {
		item.MarketingResponseTime = &now
		item.MarketingResponseBy = marketingName
		if errUpdate := s.repo.Update(ctx, item); errUpdate != nil {
			return nil, errUpdate
		}
	}

	_ = s.logTaskSvc.RecordTouch(ctx, item.OrderID, "input_item", marketingName)

	created, errFetch := s.repo.FindByID(ctx, item.ID)
	if errFetch != nil {
		return nil, errFetch
	}
	resp := toInputItemResponse(*created)
	return &resp, nil
}
