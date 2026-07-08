package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RABService interface {
	GetAll(ctx context.Context) ([]dto.RABResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.RABResponse, error)
	GetByInputItemID(ctx context.Context, inputItemID uint) (*dto.RABResponse, error)
	Create(ctx context.Context, req dto.CreateRABRequest) (*dto.RABResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateRABRequest) (*dto.RABResponse, error)
	Delete(ctx context.Context, id uint) error
	Submit(ctx context.Context, id uint, submittedBy string) (*dto.RABResponse, error)
	ExportPDF(ctx context.Context, id uint, mode string) ([]byte, string, error)
	ExportExcel(ctx context.Context, id uint, mode string) ([]byte, string, error)
}

type rabService struct {
	repo          repository.RABRepository
	inputItemRepo repository.InputItemRepository
	db            *gorm.DB // we need DB access for pivot queries and settings
	logger        *zap.Logger
	uploadDir     string
	logTaskSvc    ProjectLogTaskService
}

func NewRABService(repo repository.RABRepository, inputItemRepo repository.InputItemRepository, db *gorm.DB, logger *zap.Logger, uploadDir string, logTaskSvc ProjectLogTaskService) RABService {
	return &rabService{
		repo:          repo,
		inputItemRepo: inputItemRepo,
		db:            db,
		logger:        logger,
		uploadDir:     uploadDir,
		logTaskSvc:    logTaskSvc,
	}
}

func (s *rabService) GetAll(ctx context.Context) ([]dto.RABResponse, error) {
	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.RABResponse, len(list))
	for i, r := range list {
		res[i] = s.mapToRABResponse(r, "internal")
	}
	return res, nil
}

func (s *rabService) GetByID(ctx context.Context, id uint) (*dto.RABResponse, error) {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := s.mapToRABResponse(*r, "internal")
	return &resp, nil
}

func (s *rabService) GetByInputItemID(ctx context.Context, inputItemID uint) (*dto.RABResponse, error) {
	r, err := s.repo.FindByInputItemID(ctx, inputItemID)
	if err != nil {
		return nil, err
	}
	resp := s.mapToRABResponse(*r, "internal")
	return &resp, nil
}

func (s *rabService) Create(ctx context.Context, req dto.CreateRABRequest) (*dto.RABResponse, error) {
	// Check if already exists for this InputItemID
	existing, err := s.repo.FindByInputItemID(ctx, req.InputItemID)
	if err == nil && existing != nil {
		return nil, errors.New("RAB untuk rincian item ini sudah dibuat")
	}

	// Fetch input item to verify status is approved
	ii, err := s.inputItemRepo.FindByID(ctx, req.InputItemID)
	if err != nil {
		return nil, errors.New("rincian item tidak ditemukan")
	}
	if ii.Status != "approved" {
		return nil, errors.New("rincian item harus disetujui (approved) terlebih dahulu")
	}

	// Build GORM entity
	rab := &entity.RAB{
		InputItemID:   req.InputItemID,
		OrderID:       req.OrderID,
		MarkupGeneral: req.MarkupGeneral,
		Status:        "draft",
	}

	var grandTotal float64 = 0

	for _, roomReq := range req.Rooms {
		// Calculate costs
		var totalBahanBaku float64 = 0
		var totalFinishing float64 = 0
		var totalAksesoris float64 = 0

		room := entity.RABRoom{
			NamaRuangan: roomReq.NamaRuangan,
			ProdukID:    roomReq.ProdukID,
			Qty:         roomReq.Qty,
			Panjang:     roomReq.Panjang,
			Lebar:       roomReq.Lebar,
			Tinggi:      roomReq.Tinggi,
			Markup:      roomReq.Markup,
		}

		// 1. Bahan Baku
		for _, bbReq := range roomReq.BahanBakus {
			var hargaDasar float64 = 0
			var hargaJasa float64 = 0

			// Fetch prices from produk_bahan_bakus pivot
			if roomReq.ProdukID != nil {
				var pbb entity.ProdukBahanBaku
				errPbb := s.db.WithContext(ctx).
					Where("produk_id = ? AND bahan_baku_id = ?", *roomReq.ProdukID, bbReq.BahanBakuID).
					First(&pbb).Error
				if errPbb == nil {
					hargaDasar, _ = pbb.HargaDasar.Float64()
					hargaJasa, _ = pbb.HargaJasa.Float64()
				}
			}

			room.BahanBakus = append(room.BahanBakus, entity.RABRoomBahanBaku{
				BahanBakuID: bbReq.BahanBakuID,
				HargaDasar:  hargaDasar,
				HargaJasa:   hargaJasa,
				Markup:      bbReq.Markup,
			})
			totalBahanBaku += hargaDasar * (1 + bbReq.Markup/100)
		}

		// Helper to fetch item price
		fetchItemPrice := func(itemID uint) float64 {
			var it entity.Item
			if errIt := s.db.WithContext(ctx).First(&it, itemID).Error; errIt == nil {
				return it.Harga
			}
			return 0
		}

		// 2. Finishing Dalams
		for _, fdReq := range roomReq.FinishingDalams {
			price := fetchItemPrice(fdReq.ItemID)
			room.FinishingDalams = append(room.FinishingDalams, entity.RABRoomFinishing{
				ItemID: fdReq.ItemID,
				Type:   "dalam",
				Harga:  price,
				Markup: fdReq.Markup,
			})
			totalFinishing += price * (1 + fdReq.Markup/100)
		}

		// 3. Finishing Luars
		for _, flReq := range roomReq.FinishingLuars {
			price := fetchItemPrice(flReq.ItemID)
			room.FinishingLuars = append(room.FinishingLuars, entity.RABRoomFinishing{
				ItemID: flReq.ItemID,
				Type:   "luar",
				Harga:  price,
				Markup: flReq.Markup,
			})
			totalFinishing += price * (1 + flReq.Markup/100)
		}

		// 4. Aksesoris
		for _, aksReq := range roomReq.Aksesoris {
			price := fetchItemPrice(aksReq.ItemID)
			// Formula: harga aksesoris x (1 + markup) x qty
			aksTotal := price * (1 + aksReq.Markup/100) * float64(aksReq.Qty)

			room.Aksesoris = append(room.Aksesoris, entity.RABRoomAksesoris{
				ItemID:     aksReq.ItemID,
				Qty:        aksReq.Qty,
				Harga:      price,
				Markup:     aksReq.Markup,
				HargaTotal: aksTotal,
			})
			totalAksesoris += aksTotal
		}

		// Pricing calculation:
		// Harga Dasar = Total Bahan Baku (marked up) + Total Finishing (marked up)
		hargaDasarRoom := totalBahanBaku + totalFinishing
		// Harga Satuan = Harga Dasar * (Panjang * Lebar * Tinggi) * Qty
		volume := room.Panjang * room.Lebar * room.Tinggi
		hargaSatuanRoom := hargaDasarRoom * volume * float64(room.Qty)
		// Harga Total = Harga Satuan + total aksesoris
		hargaTotalRoom := hargaSatuanRoom + totalAksesoris

		room.HargaDasar = hargaDasarRoom
		room.HargaSatuan = hargaSatuanRoom
		room.HargaTotal = hargaTotalRoom

		grandTotal += hargaTotalRoom
		rab.Rooms = append(rab.Rooms, room)
	}

	var taxEnabled bool
	var settingVal string
	if err := s.db.WithContext(ctx).Model(&entity.Setting{}).Where("key = ?", "finance_tax_enabled").Pluck("value", &settingVal).Error; err == nil {
		taxEnabled = (settingVal == "true")
	}

	if taxEnabled {
		rab.GrandTotal = grandTotal * 1.11
	} else {
		rab.GrandTotal = grandTotal
	}

	if errCreate := s.repo.Create(ctx, rab); errCreate != nil {
		return nil, errCreate
	}

	_ = s.logTaskSvc.RecordTouch(ctx, rab.OrderID, "rab", "")

	created, errFetch := s.repo.FindByID(ctx, rab.ID)
	if errFetch != nil {
		return nil, errFetch
	}
	resp := s.mapToRABResponse(*created, "internal")
	return &resp, nil
}

func (s *rabService) Update(ctx context.Context, id uint, req dto.UpdateRABRequest) (*dto.RABResponse, error) {
	rab, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("RAB tidak ditemukan")
	}

	if rab.Status == "submitted" {
		return nil, errors.New("tidak dapat memperbarui RAB yang sudah disubmit")
	}

	rab.MarkupGeneral = req.MarkupGeneral
	rab.Rooms = nil // Clear child associations to reconstruct

	var grandTotal float64 = 0

	for _, roomReq := range req.Rooms {
		var totalBahanBaku float64 = 0
		var totalFinishing float64 = 0
		var totalAksesoris float64 = 0

		room := entity.RABRoom{
			NamaRuangan: roomReq.NamaRuangan,
			ProdukID:    roomReq.ProdukID,
			Qty:         roomReq.Qty,
			Panjang:     roomReq.Panjang,
			Lebar:       roomReq.Lebar,
			Tinggi:      roomReq.Tinggi,
			Markup:      roomReq.Markup,
		}

		// 1. Bahan Baku
		for _, bbReq := range roomReq.BahanBakus {
			var hargaDasar float64 = 0
			var hargaJasa float64 = 0

			if roomReq.ProdukID != nil {
				var pbb entity.ProdukBahanBaku
				errPbb := s.db.WithContext(ctx).
					Where("produk_id = ? AND bahan_baku_id = ?", *roomReq.ProdukID, bbReq.BahanBakuID).
					First(&pbb).Error
				if errPbb == nil {
					hargaDasar, _ = pbb.HargaDasar.Float64()
					hargaJasa, _ = pbb.HargaJasa.Float64()
				}
			}

			room.BahanBakus = append(room.BahanBakus, entity.RABRoomBahanBaku{
				BahanBakuID: bbReq.BahanBakuID,
				HargaDasar:  hargaDasar,
				HargaJasa:   hargaJasa,
				Markup:      bbReq.Markup,
			})
			totalBahanBaku += hargaDasar * (1 + bbReq.Markup/100)
		}

		fetchItemPrice := func(itemID uint) float64 {
			var it entity.Item
			if errIt := s.db.WithContext(ctx).First(&it, itemID).Error; errIt == nil {
				return it.Harga
			}
			return 0
		}

		// 2. Finishing Dalams
		for _, fdReq := range roomReq.FinishingDalams {
			price := fetchItemPrice(fdReq.ItemID)
			room.FinishingDalams = append(room.FinishingDalams, entity.RABRoomFinishing{
				ItemID: fdReq.ItemID,
				Type:   "dalam",
				Harga:  price,
				Markup: fdReq.Markup,
			})
			totalFinishing += price * (1 + fdReq.Markup/100)
		}

		// 3. Finishing Luars
		for _, flReq := range roomReq.FinishingLuars {
			price := fetchItemPrice(flReq.ItemID)
			room.FinishingLuars = append(room.FinishingLuars, entity.RABRoomFinishing{
				ItemID: flReq.ItemID,
				Type:   "luar",
				Harga:  price,
				Markup: flReq.Markup,
			})
			totalFinishing += price * (1 + flReq.Markup/100)
		}

		// 4. Aksesoris
		for _, aksReq := range roomReq.Aksesoris {
			price := fetchItemPrice(aksReq.ItemID)
			aksTotal := price * (1 + aksReq.Markup/100) * float64(aksReq.Qty)

			room.Aksesoris = append(room.Aksesoris, entity.RABRoomAksesoris{
				ItemID:     aksReq.ItemID,
				Qty:        aksReq.Qty,
				Harga:      price,
				Markup:     aksReq.Markup,
				HargaTotal: aksTotal,
			})
			totalAksesoris += aksTotal
		}

		// Pricing calculation:
		// Harga Dasar = Total Bahan Baku (marked up) + Total Finishing (marked up)
		hargaDasarRoom := totalBahanBaku + totalFinishing
		volume := room.Panjang * room.Lebar * room.Tinggi
		hargaSatuanRoom := hargaDasarRoom * volume * float64(room.Qty)
		hargaTotalRoom := hargaSatuanRoom + totalAksesoris

		room.HargaDasar = hargaDasarRoom
		room.HargaSatuan = hargaSatuanRoom
		room.HargaTotal = hargaTotalRoom

		grandTotal += hargaTotalRoom
		rab.Rooms = append(rab.Rooms, room)
	}

	var taxEnabled bool
	var settingVal string
	if err := s.db.WithContext(ctx).Model(&entity.Setting{}).Where("key = ?", "finance_tax_enabled").Pluck("value", &settingVal).Error; err == nil {
		taxEnabled = (settingVal == "true")
	}

	if taxEnabled {
		rab.GrandTotal = grandTotal * 1.11
	} else {
		rab.GrandTotal = grandTotal
	}

	if errUpdate := s.repo.Update(ctx, rab); errUpdate != nil {
		return nil, errUpdate
	}

	_ = s.logTaskSvc.RecordTouch(ctx, rab.OrderID, "rab", "")

	updated, errFetch := s.repo.FindByID(ctx, rab.ID)
	if errFetch != nil {
		return nil, errFetch
	}
	resp := s.mapToRABResponse(*updated, "internal")
	return &resp, nil
}

func (s *rabService) Delete(ctx context.Context, id uint) error {
	rab, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("RAB tidak ditemukan")
	}
	if rab.Status == "submitted" {
		return errors.New("tidak dapat menghapus RAB yang sudah disubmit")
	}
	return s.repo.Delete(ctx, id)
}

func (s *rabService) Submit(ctx context.Context, id uint, submittedBy string) (*dto.RABResponse, error) {
	rab, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("RAB tidak ditemukan")
	}
	if rab.Status == "submitted" {
		return nil, errors.New("RAB sudah berstatus submitted")
	}

	now := time.Now()
	rab.Status = "submitted"
	rab.SubmittedAt = &now
	rab.SubmittedBy = submittedBy

	// Save
	if errUpdate := s.repo.Update(ctx, rab); errUpdate != nil {
		return nil, errUpdate
	}

	_ = s.logTaskSvc.RecordTouch(ctx, rab.OrderID, "rab", submittedBy)

	// Update order stage to "kontrak" and log transition
	if errStage := s.logTaskSvc.TransitionStage(ctx, rab.OrderID, "kontrak", submittedBy); errStage != nil {
		s.logger.Error("Failed to update order stage to kontrak", zap.Error(errStage))
	}

	updated, errFetch := s.repo.FindByID(ctx, rab.ID)
	if errFetch != nil {
		return nil, errFetch
	}
	resp := s.mapToRABResponse(*updated, "internal")
	return &resp, nil
}

// mapToRABResponse formats calculations dynamically according to visual mode:
// "internal", "kontrak", "vendor", "jasa"
func (s *rabService) mapToRABResponse(rab entity.RAB, mode string) dto.RABResponse {
	resp := dto.RABResponse{
		ID:            rab.ID,
		InputItemID:   rab.InputItemID,
		OrderID:       rab.OrderID,
		MarkupGeneral: rab.MarkupGeneral,
		Status:        rab.Status,
		SubmittedAt:   rab.SubmittedAt,
		SubmittedBy:   rab.SubmittedBy,
		CreatedAt:     rab.CreatedAt,
		UpdatedAt:     rab.UpdatedAt,
		Rooms:         []dto.RABRoomResponse{},
	}

	if rab.Order != nil {
		resp.Order = &dto.InputItemOrderResponse{
			ID:             rab.Order.ID,
			NomorOrder:     rab.Order.NomorOrder,
			NamaProject:    rab.Order.NamaProject,
			NamaCustomer:   rab.Order.NamaCustomer,
			NamaPerusahaan: rab.Order.NamaPerusahaan,
			JenisInterior:  rab.Order.JenisInterior,
		}
	}

	var computedGrandTotal float64 = 0

	for _, room := range rab.Rooms {
		produkName := ""
		if room.Produk != nil {
			produkName = room.Produk.NamaProduk
		}

		roomResp := dto.RABRoomResponse{
			ID:              room.ID,
			NamaRuangan:     room.NamaRuangan,
			ProdukID:        room.ProdukID,
			NamaProduk:      produkName,
			Qty:             room.Qty,
			Panjang:         room.Panjang,
			Lebar:           room.Lebar,
			Tinggi:          room.Tinggi,
			BahanBakus:      []dto.RABRoomBahanBakuResponse{},
			FinishingDalams: []dto.RABRoomFinishingResponse{},
			FinishingLuars:  []dto.RABRoomFinishingResponse{},
			Aksesoris:       []dto.RABRoomAksesorisResponse{},
		}

		// Pull elements
		var totalBahanBaku float64 = 0
		var totalFinishing float64 = 0
		var totalAksesoris float64 = 0

		// 1. Bahan Baku
		for _, bb := range room.BahanBakus {
			namaBahan := ""
			if bb.BahanBaku != nil {
				namaBahan = bb.BahanBaku.NamaBahanBaku
			}

			bbPrice := bb.HargaDasar
			if mode == "jasa" {
				bbPrice = bb.HargaJasa
			}

			bbMarkup := bb.Markup
			if bbMarkup == 0 {
				bbMarkup = room.Markup
			}
			if mode == "vendor" || mode == "jasa" {
				bbMarkup = 0
			}

			markedUpPrice := bbPrice * (1 + bbMarkup/100)
			totalBahanBaku += markedUpPrice

			dtoPrice := bbPrice
			dtoMarkup := bbMarkup
			if mode == "kontrak" {
				dtoPrice = markedUpPrice
				dtoMarkup = 0
			}

			roomResp.BahanBakus = append(roomResp.BahanBakus, dto.RABRoomBahanBakuResponse{
				ID:          bb.ID,
				BahanBakuID: bb.BahanBakuID,
				NamaBahan:   namaBahan,
				HargaDasar:  dtoPrice,
				HargaJasa:   bb.HargaJasa,
				Markup:      dtoMarkup,
			})
		}

		// 2. Finishings (Dalam)
		for _, f := range room.FinishingDalams {
			namaItem := ""
			if f.Item != nil {
				namaItem = f.Item.NamaItem
			}
			if f.Type == "dalam" {
				fMarkup := f.Markup
				if fMarkup == 0 {
					fMarkup = room.Markup
				}
				if mode == "vendor" || mode == "jasa" {
					fMarkup = 0
				}
				markedUpPrice := f.Harga * (1 + fMarkup/100)
				totalFinishing += markedUpPrice

				dtoPrice := f.Harga
				dtoMarkup := fMarkup
				if mode == "kontrak" {
					dtoPrice = markedUpPrice
					dtoMarkup = 0
				}

				roomResp.FinishingDalams = append(roomResp.FinishingDalams, dto.RABRoomFinishingResponse{
					ID:     f.ID,
					ItemID: f.ItemID,
					Nama:   namaItem,
					Harga:  dtoPrice,
					Type:   "dalam",
					Markup: dtoMarkup,
				})
			}
		}

		// 3. Finishings (Luar)
		for _, f := range room.FinishingLuars {
			namaItem := ""
			if f.Item != nil {
				namaItem = f.Item.NamaItem
			}
			if f.Type == "luar" {
				fMarkup := f.Markup
				if fMarkup == 0 {
					fMarkup = room.Markup
				}
				if mode == "vendor" || mode == "jasa" {
					fMarkup = 0
				}
				markedUpPrice := f.Harga * (1 + fMarkup/100)
				totalFinishing += markedUpPrice

				dtoPrice := f.Harga
				dtoMarkup := fMarkup
				if mode == "kontrak" {
					dtoPrice = markedUpPrice
					dtoMarkup = 0
				}

				roomResp.FinishingLuars = append(roomResp.FinishingLuars, dto.RABRoomFinishingResponse{
					ID:     f.ID,
					ItemID: f.ItemID,
					Nama:   namaItem,
					Harga:  dtoPrice,
					Type:   "luar",
					Markup: dtoMarkup,
				})
			}
		}

		// 4. Aksesoris
		for _, aks := range room.Aksesoris {
			namaItem := ""
			if aks.Item != nil {
				namaItem = aks.Item.NamaItem
			}

			aksPrice := aks.Harga
			aksMarkup := aks.Markup
			if aksMarkup == 0 {
				aksMarkup = room.Markup
			}
			if mode == "vendor" || mode == "jasa" {
				aksMarkup = 0
			}

			if mode == "jasa" {
				// Jasa removes accessories completely
				aksPrice = 0
				aksMarkup = 0
			}

			aksTotal := aksPrice * (1 + aksMarkup/100) * float64(aks.Qty)
			totalAksesoris += aksTotal

			dtoPrice := aksPrice
			dtoMarkup := aksMarkup
			if mode == "kontrak" {
				dtoPrice = aksPrice * (1 + aksMarkup/100)
				dtoMarkup = 0
			}

			roomResp.Aksesoris = append(roomResp.Aksesoris, dto.RABRoomAksesorisResponse{
				ID:         aks.ID,
				ItemID:     aks.ItemID,
				Nama:       namaItem,
				Qty:        aks.Qty,
				Harga:      dtoPrice,
				Markup:     dtoMarkup,
				HargaTotal: aksTotal,
			})
		}

		// Apply mathematics
		var computedHargaDasar float64 = 0
		var computedHargaSatuan float64 = 0
		var computedHargaTotal float64 = 0

		computedHargaDasar = totalBahanBaku + totalFinishing
		volume := room.Panjang * room.Lebar * room.Tinggi
		computedHargaSatuan = computedHargaDasar * volume * float64(room.Qty)
		computedHargaTotal = computedHargaSatuan + totalAksesoris

		// contract visual hides markup columns by passing 0 to client
		dtoMarkupVal := room.Markup
		if mode == "kontrak" || mode == "vendor" || mode == "jasa" {
			dtoMarkupVal = 0
		}

		roomResp.Markup = dtoMarkupVal
		roomResp.HargaDasar = computedHargaDasar
		roomResp.HargaSatuan = computedHargaSatuan
		roomResp.HargaTotal = computedHargaTotal

		computedGrandTotal += computedHargaTotal
		resp.Rooms = append(resp.Rooms, roomResp)
	}

	resp.GrandTotal = computedGrandTotal
	return resp
}

// Currency format helper matching invoice.go
func formatRupiah(val float64) string {
	s := fmt.Sprintf("%.0f", val)
	n := len(s)
	if n <= 3 {
		return s
	}
	var buf bytes.Buffer
	for i, c := range s {
		buf.WriteRune(c)
		if (n-i-1)%3 == 0 && i != n-1 {
			buf.WriteByte('.')
		}
	}
	return buf.String()
}

func limitString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return s
}

func (s *rabService) ExportPDF(ctx context.Context, id uint, mode string) ([]byte, string, error) {
	rab, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	// Fetch calculated response for visual consistency
	rData := s.mapToRABResponse(*rab, mode)

	// PDF Layout initialization: A4 Landscape ("L")
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Title / Brand header (Dynamic)
	companyID := uint(1)
	if rab.Order != nil {
		companyID = rab.Order.CompanyID
	}
	cp := entity.GetCompanyProfile(s.db, companyID)

	logoFile := ""
	if cp.Logo != "" {
		logoFile = filepath.Join(s.uploadDir, filepath.Base(cp.Logo))
		if _, err := os.Stat(logoFile); os.IsNotExist(err) {
			logoFile = ""
		}
	}

	if logoFile != "" {
		pdf.Image(logoFile, 15, 12, 0, 15, false, "", 0, "")
		pdf.SetLeftMargin(33)
		pdf.SetX(33)
		pdf.SetY(12)
	} else {
		pdf.SetY(12)
	}

	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(0, 128, 128) // Teal Primary Color
	pdf.CellFormat(0, 7, cp.Name, "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 4, "Premium Interior Design & Architecture Services", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 4, fmt.Sprintf("Email: %s | Phone: %s", cp.Email, cp.Phone), "", 1, "L", false, 0, "")

	// Reset margins
	pdf.SetLeftMargin(15)
	pdf.SetX(15)
	pdf.SetY(30)

	pdf.SetDrawColor(0, 128, 128)
	pdf.SetLineWidth(0.8)
	pdf.Line(15, pdf.GetY(), 282, pdf.GetY()) // 15 + 267 = 282
	pdf.Ln(6)

	// Document Title & Mode
	pdf.SetFont("Arial", "B", 13)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 6, "RENCANA ANGGARAN BIAYA (RAB) - " + strings.ToUpper(mode), "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(0, 5, fmt.Sprintf("No. Dokumen: DOC/RAB/%s/%04d", rData.Order.NomorOrder, rData.ID), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Info section
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(0, 128, 128)
	pdf.CellFormat(133.5, 5, "INFORMASI PROYEK", "", 0, "L", false, 0, "")
	pdf.CellFormat(133.5, 5, "RINCIAN DOKUMEN", "", 1, "L", false, 0, "")
	pdf.SetLineWidth(0.2)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(15, pdf.GetY(), 282, pdf.GetY())
	pdf.Ln(2)

	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(30, 4.5, "Nama Project", "", 0, "L", false, 0, "")
	pdf.CellFormat(103.5, 4.5, ": "+rData.Order.NamaProject, "", 0, "L", false, 0, "")
	pdf.CellFormat(30, 4.5, "Tanggal Dibuat", "", 0, "L", false, 0, "")
	pdf.CellFormat(103.5, 4.5, ": "+rData.CreatedAt.Format("02 Jan 2006"), "", 1, "L", false, 0, "")

	pdf.CellFormat(30, 4.5, "Nama Klien", "", 0, "L", false, 0, "")
	pdf.CellFormat(103.5, 4.5, ": "+rData.Order.NamaCustomer, "", 0, "L", false, 0, "")
	pdf.CellFormat(30, 4.5, "Status Dokumen", "", 0, "L", false, 0, "")
	pdf.CellFormat(103.5, 4.5, ": "+strings.ToUpper(rData.Status), "", 1, "L", false, 0, "")

	pdf.CellFormat(30, 4.5, "Tipe Interior", "", 0, "L", false, 0, "")
	pdf.CellFormat(237, 4.5, ": "+strings.ToUpper(rData.Order.JenisInterior), "", 1, "L", false, 0, "")
	pdf.Ln(6)

	// Column Widths dynamically defined (Landscape 267mm printable width)
	var (
		colWidthProd        = 0.0
		colWidthBB          = 0.0
		colWidthHargaBB     = 0.0
		colWidthFin         = 0.0
		colWidthHargaFin    = 0.0
		colWidthMarkup      = 0.0
		colWidthQtyVol      = 0.0
		colWidthHargaSatuan = 0.0
		colWidthAks         = 0.0
		colWidthHargaAks    = 0.0
		colWidthMarkupAks   = 0.0
		colWidthTotalAks    = 0.0
		colWidthGrandTotal  = 0.0
	)

	if mode == "internal" {
		colWidthProd        = 32.0
		colWidthBB          = 22.0
		colWidthHargaBB     = 18.0
		colWidthFin         = 26.0
		colWidthHargaFin    = 18.0
		colWidthMarkup      = 10.0
		colWidthQtyVol      = 16.0
		colWidthHargaSatuan = 20.0
		colWidthAks         = 22.0
		colWidthHargaAks    = 18.0
		colWidthMarkupAks   = 15.0
		colWidthTotalAks    = 25.0
		colWidthGrandTotal  = 25.0
	} else if mode == "jasa" {
		colWidthProd        = 45.0
		colWidthBB          = 35.0
		colWidthHargaBB     = 30.0
		colWidthFin         = 40.0
		colWidthHargaFin    = 30.0
		colWidthQtyVol      = 22.0
		colWidthHargaSatuan = 35.0
		colWidthTotalAks    = 0.0
		colWidthGrandTotal  = 30.0
	} else {
		// kontrak, vendor (11 columns)
		colWidthProd        = 35.0
		colWidthBB          = 26.0
		colWidthHargaBB     = 20.0
		colWidthFin         = 30.0
		colWidthHargaFin    = 20.0
		colWidthQtyVol      = 16.0
		colWidthHargaSatuan = 22.0
		colWidthAks         = 26.0
		colWidthHargaAks    = 22.0
		colWidthTotalAks    = 25.0
		colWidthGrandTotal  = 25.0
	}

	totalSpanWidth := 267.0 - colWidthGrandTotal

	// Table Headers
	pdf.SetFillColor(0, 128, 128) // Teal header background
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 6.5)

	pdf.CellFormat(colWidthProd, 7, "Produk", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colWidthBB, 7, "Bahan Baku", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colWidthHargaBB, 7, "Hrg Bahan", "1", 0, "R", true, 0, "")
	pdf.CellFormat(colWidthFin, 7, "Finishing", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colWidthHargaFin, 7, "Hrg Finish", "1", 0, "R", true, 0, "")
	if mode == "internal" {
		pdf.CellFormat(colWidthMarkup, 7, "Mkp", "1", 0, "C", true, 0, "")
	}
	pdf.CellFormat(colWidthQtyVol, 7, "Qty/Vol", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidthHargaSatuan, 7, "Hrg Satuan", "1", 0, "R", true, 0, "")
	if mode != "jasa" {
		pdf.CellFormat(colWidthAks, 7, "Aksesoris", "1", 0, "L", true, 0, "")
		pdf.CellFormat(colWidthHargaAks, 7, "Hrg Aks", "1", 0, "R", true, 0, "")
	}
	if mode == "internal" {
		pdf.CellFormat(colWidthMarkupAks, 7, "Mkp Aks", "1", 0, "C", true, 0, "")
	}
	if mode != "jasa" {
		pdf.CellFormat(colWidthTotalAks, 7, "Tot Aks", "1", 0, "R", true, 0, "")
	}
	pdf.CellFormat(colWidthGrandTotal, 7, "Total", "1", 1, "R", true, 0, "")

	// Group rooms by NamaRuangan
	uniqueRoomNames := []string{}
	roomGroups := make(map[string][]dto.RABRoomResponse)
	for _, room := range rData.Rooms {
		if _, exists := roomGroups[room.NamaRuangan]; !exists {
			uniqueRoomNames = append(uniqueRoomNames, room.NamaRuangan)
		}
		roomGroups[room.NamaRuangan] = append(roomGroups[room.NamaRuangan], room)
	}

	// Table Body
	for _, roomName := range uniqueRoomNames {
		productsList := roomGroups[roomName]
		roomSubtotal := 0.0

		// Print Room Group Header Row
		pdf.SetFillColor(0, 128, 128) // Teal Primary
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Arial", "B", 8)
		roomLabel := fmt.Sprintf(" %s (%d produk)", strings.ToUpper(roomName), len(productsList))
		pdf.CellFormat(267, 6, roomLabel, "1", 1, "L", true, 0, "")

		// Reset text color and font size for body
		pdf.SetTextColor(60, 60, 60)
		pdf.SetFont("Arial", "", 7.5)

		for pIdx, room := range productsList {
			roomSubtotal += room.HargaTotal

			// Prepare finishings
			var finishings []dto.RABRoomFinishingResponse
			finishings = append(finishings, room.FinishingDalams...)
			finishings = append(finishings, room.FinishingLuars...)

			// Determine maxLines
			maxLines := len(room.BahanBakus)
			if len(finishings) > maxLines {
				maxLines = len(finishings)
			}
			if mode != "jasa" && len(room.Aksesoris) > maxLines {
				maxLines = len(room.Aksesoris)
			}
			if maxLines == 0 {
				maxLines = 1
			}

			maxLenProd := int(colWidthProd / 1.3)
			maxLenBB := int(colWidthBB / 1.3)
			maxLenFin := int(colWidthFin / 1.3)
			maxLenAks := int(colWidthAks / 1.3)

			// Render lines
			for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
				// Column 1: Produk
				prodText := ""
				if lineIdx == 0 {
					prodText = fmt.Sprintf("%d. %s", pIdx+1, room.NamaProduk)
					if room.NamaProduk == "" {
						prodText = fmt.Sprintf("%d. Kustom", pIdx+1)
					}
					prodText = limitString(prodText, maxLenProd)
				}
				pdf.CellFormat(colWidthProd, 5, prodText, "1", 0, "L", false, 0, "")

				// Column 2: Bahan Baku (Nama)
				bbText := ""
				if lineIdx < len(room.BahanBakus) {
					bbText = limitString(room.BahanBakus[lineIdx].NamaBahan, maxLenBB)
				} else if lineIdx == 0 && len(room.BahanBakus) == 0 {
					bbText = "-"
				}
				pdf.CellFormat(colWidthBB, 5, bbText, "1", 0, "L", false, 0, "")

				// Column 3: Harga Bahan
				bbPriceText := ""
				if lineIdx < len(room.BahanBakus) {
					bbPriceText = "Rp " + formatRupiah(room.BahanBakus[lineIdx].HargaDasar)
				} else if lineIdx == 0 && len(room.BahanBakus) == 0 {
					bbPriceText = "-"
				}
				pdf.CellFormat(colWidthHargaBB, 5, bbPriceText, "1", 0, "R", false, 0, "")

				// Column 4: Finishing (Nama)
				finText := ""
				if lineIdx < len(finishings) {
					f := finishings[lineIdx]
					typeSuffix := "(Dalam)"
					if f.Type == "luar" {
						typeSuffix = "(Luar)"
					}
					finText = limitString(fmt.Sprintf("%s %s", f.Nama, typeSuffix), maxLenFin)
				} else if lineIdx == 0 && len(finishings) == 0 {
					finText = "-"
				}
				pdf.CellFormat(colWidthFin, 5, finText, "1", 0, "L", false, 0, "")

				// Column 5: Harga Finishing
				finPriceText := ""
				if lineIdx < len(finishings) {
					finPriceText = "Rp " + formatRupiah(finishings[lineIdx].Harga)
				} else if lineIdx == 0 && len(finishings) == 0 {
					finPriceText = "-"
				}
				pdf.CellFormat(colWidthHargaFin, 5, finPriceText, "1", 0, "R", false, 0, "")

				// Column 6: Markup (Internal only)
				if mode == "internal" {
					markupText := ""
					if lineIdx == 0 {
						markupText = fmt.Sprintf("%.0f%%", room.Markup)
					}
					pdf.CellFormat(colWidthMarkup, 5, markupText, "1", 0, "C", false, 0, "")
				}

				// Column 7: Qty / Volume
				qtyVolText := ""
				if lineIdx == 0 {
					volume := room.Panjang * room.Lebar * room.Tinggi
					qtyVolText = fmt.Sprintf("%.2fm³ (x%d)", volume, room.Qty)
				}
				pdf.CellFormat(colWidthQtyVol, 5, qtyVolText, "1", 0, "C", false, 0, "")

				// Column 8: Harga Satuan
				hargaSatuanText := ""
				if lineIdx == 0 {
					hargaSatuanText = "Rp " + formatRupiah(room.HargaSatuan)
				}
				pdf.CellFormat(colWidthHargaSatuan, 5, hargaSatuanText, "1", 0, "R", false, 0, "")

				// Columns for Aksesoris (if not jasa)
				if mode != "jasa" {
					// Column 9: Aksesoris (Nama)
					aksText := ""
					if lineIdx < len(room.Aksesoris) {
						aksText = limitString(fmt.Sprintf("%s (x%d)", room.Aksesoris[lineIdx].Nama, room.Aksesoris[lineIdx].Qty), maxLenAks)
					} else if lineIdx == 0 && len(room.Aksesoris) == 0 {
						aksText = "-"
					}
					pdf.CellFormat(colWidthAks, 5, aksText, "1", 0, "L", false, 0, "")

					// Column 10: Harga Aksesoris
					aksPriceText := ""
					if lineIdx < len(room.Aksesoris) {
						aksPriceText = "Rp " + formatRupiah(room.Aksesoris[lineIdx].HargaTotal)
					} else if lineIdx == 0 && len(room.Aksesoris) == 0 {
						aksPriceText = "-"
					}
					pdf.CellFormat(colWidthHargaAks, 5, aksPriceText, "1", 0, "R", false, 0, "")
				}

				// Column 11: Markup Aksesoris (Internal only)
				if mode == "internal" {
					markupAksText := ""
					if lineIdx < len(room.Aksesoris) {
						markupAksText = fmt.Sprintf("%.0f%%", room.Aksesoris[lineIdx].Markup)
					} else if lineIdx == 0 && len(room.Aksesoris) == 0 {
						markupAksText = "-"
					}
					pdf.CellFormat(colWidthMarkupAks, 5, markupAksText, "1", 0, "C", false, 0, "")
				}

				// Column 11.5: Total Aksesoris (if not jasa)
				if mode != "jasa" {
					totalAksText := ""
					if lineIdx == 0 {
						var totalAks float64 = 0
						for _, aks := range room.Aksesoris {
							totalAks += aks.HargaTotal
						}
						totalAksText = "Rp " + formatRupiah(totalAks)
					}
					pdf.CellFormat(colWidthTotalAks, 5, totalAksText, "1", 0, "R", false, 0, "")
				}

				// Column 12: Grand Total
				grandTotalText := ""
				if lineIdx == 0 {
					grandTotalText = "Rp " + formatRupiah(room.HargaTotal)
				}
				pdf.CellFormat(colWidthGrandTotal, 5, grandTotalText, "1", 1, "R", false, 0, "")
			}
		}

		// Print Room Subtotal Row
		pdf.SetFillColor(245, 247, 248)
		pdf.SetTextColor(80, 80, 80)
		pdf.SetFont("Arial", "B", 8)
		pdf.CellFormat(totalSpanWidth, 6, "Subtotal "+roomName+" ", "1", 0, "R", true, 0, "")
		pdf.SetTextColor(0, 128, 128)
		pdf.CellFormat(colWidthGrandTotal, 6, "Rp "+formatRupiah(roomSubtotal), "1", 1, "R", true, 0, "")
	}

	var taxEnabled bool
	var settingVal string
	if err := s.db.WithContext(ctx).Model(&entity.Setting{}).Where("key = ?", "finance_tax_enabled").Pluck("value", &settingVal).Error; err == nil {
		taxEnabled = (settingVal == "true")
	}

	if taxEnabled {
		subtotal := rData.GrandTotal / 1.11
		tax := rData.GrandTotal - subtotal

		// Subtotal Row
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(245, 245, 245)
		pdf.SetTextColor(100, 100, 100)
		pdf.CellFormat(totalSpanWidth, 6, "SUBTOTAL RAB ", "1", 0, "R", true, 0, "")
		pdf.CellFormat(colWidthGrandTotal, 6, "Rp "+formatRupiah(subtotal), "1", 1, "R", true, 0, "")

		// Tax Row
		pdf.CellFormat(totalSpanWidth, 6, "PPN (11%) ", "1", 0, "R", true, 0, "")
		pdf.CellFormat(colWidthGrandTotal, 6, "Rp "+formatRupiah(tax), "1", 1, "R", true, 0, "")
	}

	// Grand Total Row
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(230, 242, 242)
	pdf.SetTextColor(0, 102, 102)

	pdf.CellFormat(totalSpanWidth, 7, "GRAND TOTAL RAB ", "1", 0, "R", true, 0, "")
	pdf.CellFormat(colWidthGrandTotal, 7, "Rp "+formatRupiah(rData.GrandTotal), "1", 1, "R", true, 0, "")

	pdf.Ln(12)

	// Signatures
	pdf.SetTextColor(80, 80, 80)
	pdf.SetFont("Arial", "", 8.5)
	pdf.CellFormat(133.5, 4, "Disiapkan Oleh,", "", 0, "C", false, 0, "")
	pdf.CellFormat(133.5, 4, "Menyetujui Klien,", "", 1, "C", false, 0, "")
	pdf.Ln(18)

	pdf.SetFont("Arial", "B", 8.5)
	pdf.CellFormat(133.5, 4, "ESTIMATOR & FINANCE", "", 0, "C", false, 0, "")
	pdf.CellFormat(133.5, 4, rData.Order.NamaCustomer, "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 7.5)
	pdf.CellFormat(133.5, 3.5, fmt.Sprintf("(%s)", cp.Name), "", 0, "C", false, 0, "")
	pdf.CellFormat(133.5, 3.5, "(Tanda Tangan & Nama Terang)", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if errPdf := pdf.Output(&buf); errPdf != nil {
		return nil, "", errPdf
	}

	filename := fmt.Sprintf("RAB_%s_%s.pdf", rData.Order.NomorOrder, mode)
	return buf.Bytes(), filename, nil
}

func (s *rabService) ExportExcel(ctx context.Context, id uint, mode string) ([]byte, string, error) {
	rab, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	rData := s.mapToRABResponse(*rab, mode)

	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	sheetName := "RAB"
	_ = f.SetSheetName("Sheet1", sheetName)

	// Styling definitions
	styleTitle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "008080"},
	})
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"008080"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})
	styleRoomHeader, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"008080"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	styleSubtotal, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "333333", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"F5F7F8"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})
	
	currencyFormat := "\"Rp\"#,##0"
	
	styleSubtotalVal, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true, Color: "008080", Size: 10},
		NumFmt:       0,
		CustomNumFmt: &currencyFormat,
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"F5F7F8"}, Pattern: 1},
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})
	styleBody, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Size: 9},
		Border: []excelize.Border{
			{Type: "left", Color: "E0E0E0", Style: 1},
			{Type: "right", Color: "E0E0E0", Style: 1},
			{Type: "top", Color: "E0E0E0", Style: 1},
			{Type: "bottom", Color: "E0E0E0", Style: 1},
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	styleBodyR, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Size: 9},
		NumFmt:       0,
		CustomNumFmt: &currencyFormat,
		Border: []excelize.Border{
			{Type: "left", Color: "E0E0E0", Style: 1},
			{Type: "right", Color: "E0E0E0", Style: 1},
			{Type: "top", Color: "E0E0E0", Style: 1},
			{Type: "bottom", Color: "E0E0E0", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	styleBodyC, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Size: 9},
		Border: []excelize.Border{
			{Type: "left", Color: "E0E0E0", Style: 1},
			{Type: "right", Color: "E0E0E0", Style: 1},
			{Type: "top", Color: "E0E0E0", Style: 1},
			{Type: "bottom", Color: "E0E0E0", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleGrandTotal, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Bold: true, Color: "005555", Size: 10},
		NumFmt:       0,
		CustomNumFmt: &currencyFormat,
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"E6F2F2"}, Pattern: 1},
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "008080", Style: 2},
			{Type: "bottom", Color: "008080", Style: 2},
		},
	})
	styleBold, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})

	// Header Info
	companyIDExcel := uint(1)
	if rab.Order != nil {
		companyIDExcel = rab.Order.CompanyID
	}
	cp := entity.GetCompanyProfile(s.db, companyIDExcel)
	_ = f.SetCellValue(sheetName, "A1", cp.Name)
	_ = f.SetCellStyle(sheetName, "A1", "A1", styleTitle)

	_ = f.SetCellValue(sheetName, "A2", "RENCANA ANGGARAN BIAYA (RAB) - "+strings.ToUpper(mode))
	_ = f.SetCellValue(sheetName, "A4", "Nama Project:")
	_ = f.SetCellValue(sheetName, "B4", rData.Order.NamaProject)
	_ = f.SetCellValue(sheetName, "A5", "Nama Klien:")
	_ = f.SetCellValue(sheetName, "B5", rData.Order.NamaCustomer)
	_ = f.SetCellValue(sheetName, "A6", "Nomor Order:")
	_ = f.SetCellValue(sheetName, "B6", rData.Order.NomorOrder)

	_ = f.SetCellStyle(sheetName, "A4", "A6", styleBold)

	var lastColLetter, secondToLastColLetter string
	// Define column headers
	var headers []string
	if mode == "internal" {
		headers = []string{"Produk", "Bahan Baku", "Harga Bahan", "Finishing", "Harga Finishing", "Markup", "Qty / Volume", "Harga Satuan", "Aksesoris", "Harga Aksesoris", "Markup Aksesoris", "Total Aksesoris", "Grand Total"}
	} else if mode == "jasa" {
		headers = []string{"Produk", "Bahan Baku", "Harga Bahan", "Finishing", "Harga Finishing", "Qty / Volume", "Harga Satuan", "Grand Total"}
	} else {
		headers = []string{"Produk", "Bahan Baku", "Harga Bahan", "Finishing", "Harga Finishing", "Qty / Volume", "Harga Satuan", "Aksesoris", "Harga Aksesoris", "Total Aksesoris", "Grand Total"}
	}

	startRow := 8
	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, startRow)
		_ = f.SetCellValue(sheetName, cell, header)
		_ = f.SetCellStyle(sheetName, cell, cell, styleHeader)
	}

	// Group rooms by NamaRuangan
	uniqueRoomNames := []string{}
	roomGroups := make(map[string][]dto.RABRoomResponse)
	for _, room := range rData.Rooms {
		if _, exists := roomGroups[room.NamaRuangan]; !exists {
			uniqueRoomNames = append(uniqueRoomNames, room.NamaRuangan)
		}
		roomGroups[room.NamaRuangan] = append(roomGroups[room.NamaRuangan], room)
	}

	colLetter := func(col int) string {
		return string(rune('A' + col - 1))
	}

	curRow := startRow + 1
	var roomSubtotalRows []int

	for _, roomName := range uniqueRoomNames {
		productsList := roomGroups[roomName]

		// Write Room Header Row
		lastColLetter = colLetter(len(headers))
		_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", lastColLetter, curRow))
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf(" %s (%d produk)", strings.ToUpper(roomName), len(productsList)))
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", lastColLetter, curRow), styleRoomHeader)
		curRow++

		roomStartRow := curRow

		for pIdx, room := range productsList {
			var finishings []dto.RABRoomFinishingResponse
			finishings = append(finishings, room.FinishingDalams...)
			finishings = append(finishings, room.FinishingLuars...)

			maxLines := len(room.BahanBakus)
			if len(finishings) > maxLines {
				maxLines = len(finishings)
			}
			if mode != "jasa" && len(room.Aksesoris) > maxLines {
				maxLines = len(room.Aksesoris)
			}
			if maxLines == 0 {
				maxLines = 1
			}

			productStartRow := curRow

			for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
				colIdx := 1

				// 1. Produk
				if lineIdx == 0 {
					prodName := room.NamaProduk
					if prodName == "" {
						prodName = "Kustom"
					}
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%d. %s (%.2fx%.2fx%.2f)", pIdx+1, prodName, room.Panjang, room.Lebar, room.Tinggi))
				}
				_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBody)
				colIdx++

				// 2. Bahan Baku (Nama)
				if lineIdx < len(room.BahanBakus) {
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), room.BahanBakus[lineIdx].NamaBahan)
				} else if lineIdx == 0 && len(room.BahanBakus) == 0 {
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), "-")
				}
				_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBody)
				colIdx++

				// 3. Harga Bahan
				if lineIdx < len(room.BahanBakus) {
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), room.BahanBakus[lineIdx].HargaDasar)
				} else if lineIdx == 0 && len(room.BahanBakus) == 0 {
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), "-")
				}
				_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBodyR)
				colIdx++

				// 4. Finishing (Nama)
				if lineIdx < len(finishings) {
					fItem := finishings[lineIdx]
					typeSuffix := "Dalam"
					if fItem.Type == "luar" {
						typeSuffix = "Luar"
					}
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s (%s)", fItem.Nama, typeSuffix))
				} else if lineIdx == 0 && len(finishings) == 0 {
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), "-")
				}
				_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBody)
				colIdx++

				// 5. Harga Finishing
				if lineIdx < len(finishings) {
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), finishings[lineIdx].Harga)
				} else if lineIdx == 0 && len(finishings) == 0 {
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), "-")
				}
				_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBodyR)
				colIdx++

				// 6. Markup (Internal only)
				if mode == "internal" {
					if lineIdx == 0 {
						_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%.0f%%", room.Markup))
					}
					_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBodyC)
					colIdx++
				}

				// 7. Qty / Volume
				if lineIdx == 0 {
					volume := room.Panjang * room.Lebar * room.Tinggi
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%.2fm³ (x%d)", volume, room.Qty))
				}
				_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBodyC)
				colIdx++

				// 8. Harga Satuan
				if lineIdx == 0 {
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), room.HargaSatuan)
				}
				_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBodyR)
				colIdx++

				// Columns for Aksesoris
				if mode != "jasa" {
					// 9. Aksesoris (Nama)
					if lineIdx < len(room.Aksesoris) {
						_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s (x%d)", room.Aksesoris[lineIdx].Nama, room.Aksesoris[lineIdx].Qty))
					} else if lineIdx == 0 && len(room.Aksesoris) == 0 {
						_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), "-")
					}
					_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBody)
					colIdx++

					// 10. Harga Aksesoris
					if lineIdx < len(room.Aksesoris) {
						_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), room.Aksesoris[lineIdx].HargaTotal)
					} else if lineIdx == 0 && len(room.Aksesoris) == 0 {
						_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), "-")
					}
					_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBodyR)
					colIdx++
				}

				// Markup Aksesoris (Internal only)
				if mode == "internal" {
					if lineIdx < len(room.Aksesoris) {
						_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%.0f%%", room.Aksesoris[lineIdx].Markup))
					} else if lineIdx == 0 && len(room.Aksesoris) == 0 {
						_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), "-")
					}
					_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBodyC)
					colIdx++
				}

				// Total Aksesoris (if not jasa)
				if mode != "jasa" {
					if lineIdx == 0 {
						var totalAks float64 = 0
						for _, aks := range room.Aksesoris {
							totalAks += aks.HargaTotal
						}
						_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), totalAks)
					}
					_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBodyR)
					colIdx++
				}

				// 12. Grand Total
				if lineIdx == 0 {
					_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), room.HargaTotal)
				}
				_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", colLetter(colIdx), curRow), fmt.Sprintf("%s%d", colLetter(colIdx), curRow), styleBodyR)
				colIdx++

				curRow++
			}

			// Merge cells vertically for non-list columns if maxLines > 1
			if maxLines > 1 {
				_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", productStartRow), fmt.Sprintf("A%d", curRow-1))
				
				qtyVolColIdx := 7
				hsColIdx := 8
				if mode == "jasa" {
					qtyVolColIdx = 6
					hsColIdx = 7
				} else if mode != "internal" { // kontrak, vendor
					qtyVolColIdx = 6
					hsColIdx = 7
				}
				
				_ = f.MergeCell(sheetName, fmt.Sprintf("%s%d", colLetter(qtyVolColIdx), productStartRow), fmt.Sprintf("%s%d", colLetter(qtyVolColIdx), curRow-1))
				_ = f.MergeCell(sheetName, fmt.Sprintf("%s%d", colLetter(hsColIdx), productStartRow), fmt.Sprintf("%s%d", colLetter(hsColIdx), curRow-1))
				_ = f.MergeCell(sheetName, fmt.Sprintf("%s%d", colLetter(len(headers)), productStartRow), fmt.Sprintf("%s%d", colLetter(len(headers)), curRow-1))

				if mode == "internal" {
					_ = f.MergeCell(sheetName, fmt.Sprintf("F%d", productStartRow), fmt.Sprintf("F%d", curRow-1))
				}

				if mode != "jasa" {
					totalAksColIdx := 12
					if mode != "internal" {
						totalAksColIdx = 10
					}
					_ = f.MergeCell(sheetName, fmt.Sprintf("%s%d", colLetter(totalAksColIdx), productStartRow), fmt.Sprintf("%s%d", colLetter(totalAksColIdx), curRow-1))
				}
			}
		}

		// Write Room Subtotal Row
		lastColIdx := len(headers)
		lastColLetter = colLetter(lastColIdx)
		secondToLastColLetter = colLetter(lastColIdx - 1)

		_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", secondToLastColLetter, curRow))
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("Subtotal %s", roomName))
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", secondToLastColLetter, curRow), styleSubtotal)

		formula := fmt.Sprintf("=SUM(%s%d:%s%d)", lastColLetter, roomStartRow, lastColLetter, curRow-1)
		_ = f.SetCellFormula(sheetName, fmt.Sprintf("%s%d", lastColLetter, curRow), formula)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", lastColLetter, curRow), fmt.Sprintf("%s%d", lastColLetter, curRow), styleSubtotalVal)

		roomSubtotalRows = append(roomSubtotalRows, curRow)
		curRow++
	}

	// Check setting finance_tax_enabled
	var taxEnabled bool
	var settingVal string
	if err := s.db.WithContext(ctx).Model(&entity.Setting{}).Where("key = ?", "finance_tax_enabled").Pluck("value", &settingVal).Error; err == nil {
		taxEnabled = (settingVal == "true")
	}

	// Grand Total Row
	lastColLetter = colLetter(len(headers))
	secondToLastColLetter = colLetter(len(headers) - 1)

	var parts []string
	for _, rNum := range roomSubtotalRows {
		parts = append(parts, fmt.Sprintf("%s%d", lastColLetter, rNum))
	}
	finalFormula := "=" + strings.Join(parts, "+")
	if len(parts) == 0 {
		finalFormula = "=0"
	}

	if taxEnabled {
		// 1. Subtotal Row
		_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", secondToLastColLetter, curRow))
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), "SUBTOTAL")
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", secondToLastColLetter, curRow), styleSubtotal)
		_ = f.SetCellFormula(sheetName, fmt.Sprintf("%s%d", lastColLetter, curRow), finalFormula)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", lastColLetter, curRow), fmt.Sprintf("%s%d", lastColLetter, curRow), styleSubtotalVal)

		subtotalRow := curRow
		curRow++

		// 2. PPN 11% Row
		_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", secondToLastColLetter, curRow))
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), "PPN (11%)")
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", secondToLastColLetter, curRow), styleSubtotal)
		_ = f.SetCellFormula(sheetName, fmt.Sprintf("%s%d", lastColLetter, curRow), fmt.Sprintf("=%s%d*0.11", lastColLetter, subtotalRow))
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", lastColLetter, curRow), fmt.Sprintf("%s%d", lastColLetter, curRow), styleSubtotalVal)

		taxRow := curRow
		curRow++

		// 3. Grand Total Row
		_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", secondToLastColLetter, curRow))
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), "GRAND TOTAL RAB")
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", secondToLastColLetter, curRow), styleGrandTotal)
		_ = f.SetCellFormula(sheetName, fmt.Sprintf("%s%d", lastColLetter, curRow), fmt.Sprintf("=%s%d+%s%d", lastColLetter, subtotalRow, lastColLetter, taxRow))
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", lastColLetter, curRow), fmt.Sprintf("%s%d", lastColLetter, curRow), styleGrandTotal)
	} else {
		_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", secondToLastColLetter, curRow))
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", curRow), "GRAND TOTAL RAB")
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", curRow), fmt.Sprintf("%s%d", secondToLastColLetter, curRow), styleGrandTotal)
		_ = f.SetCellFormula(sheetName, fmt.Sprintf("%s%d", lastColLetter, curRow), finalFormula)
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", lastColLetter, curRow), fmt.Sprintf("%s%d", lastColLetter, curRow), styleGrandTotal)
	}

	// Column Widths set dynamically based on header names
	for colIdx, header := range headers {
		width := 16.0
		switch header {
		case "Produk":
			width = 32.0
		case "Bahan Baku", "Finishing", "Aksesoris":
			width = 24.0
		case "Harga Bahan", "Harga Finishing", "Harga Aksesoris":
			width = 16.0
		case "Markup", "Markup Aksesoris":
			width = 12.0
		case "Qty / Volume", "Harga Satuan":
			width = 18.0
		case "Total Aksesoris", "Grand Total":
			width = 20.0
		}
		_ = f.SetColWidth(sheetName, colLetter(colIdx+1), colLetter(colIdx+1), width)
	}

	// Write to bytes buffer
	buf, errXls := f.WriteToBuffer()
	if errXls != nil {
		return nil, "", errXls
	}

	filename := fmt.Sprintf("RAB_%s_%s.xlsx", rData.Order.NomorOrder, mode)
	return buf.Bytes(), filename, nil
}
