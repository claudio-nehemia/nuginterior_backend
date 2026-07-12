package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/jung-kurt/gofpdf/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type InvoiceService interface {
	GetContractInvoiceList(ctx context.Context) ([]dto.ContractInvoiceListResponse, error)
	GetInvoicesByContractID(ctx context.Context, contractID uint) (*dto.ContractInvoiceListResponse, error)
	SubmitInvoiceResponse(ctx context.Context, contractID uint, email string) (*dto.ContractResponse, error)
	UpdateDeadline(ctx context.Context, id uint, deadlineStr string) (*dto.InvoiceResponse, error)
	UploadPaymentProof(ctx context.Context, id uint, fileHeader *multipart.FileHeader) (*dto.InvoiceResponse, error)
	GenerateInvoicePDF(ctx context.Context, id uint) ([]byte, string, error) // Returns PDF bytes, filename, error
	GenerateInvoicesForContract(ctx context.Context, contractID uint) error
}

type invoiceService struct {
	repo         repository.InvoiceRepository
	contractRepo repository.ContractRepository
	userRepo     repository.UserRepository
	db           *gorm.DB
	logger       *zap.Logger
	uploadDir    string
	logTaskSvc   ProjectLogTaskService
}

func NewInvoiceService(
	repo repository.InvoiceRepository,
	contractRepo repository.ContractRepository,
	userRepo repository.UserRepository,
	db *gorm.DB,
	logger *zap.Logger,
	uploadDir string,
	logTaskSvc ProjectLogTaskService,
) InvoiceService {
	return &invoiceService{
		repo:         repo,
		contractRepo: contractRepo,
		userRepo:     userRepo,
		db:           db,
		logger:       logger,
		uploadDir:    uploadDir,
		logTaskSvc:   logTaskSvc,
	}
}

func (s *invoiceService) GetContractInvoiceList(ctx context.Context) ([]dto.ContractInvoiceListResponse, error) {
	contracts, err := s.repo.FindAllContractsWithInvoices(ctx)
	if err != nil {
		return nil, err
	}

	var res []dto.ContractInvoiceListResponse
	for _, c := range contracts {
		// Resolve target company ID from the contract
		companyID := uint(1)
		if c.Order != nil {
			companyID = c.Order.CompanyID
		} else {
			companyID = database.GetContextCompanyID(ctx)
			if companyID == 0 {
				companyID = 1
			}
		}

		// Check setting finance_auto_invoice
		var autoInvoice bool = true
		var settingVal string
		if err := s.db.WithContext(ctx).Model(&entity.Setting{}).Where("company_id = ? AND key = ?", companyID, "finance_auto_invoice").Pluck("value", &settingVal).Error; err == nil {
			autoInvoice = (settingVal == "true")
		}

		// Auto-generate invoices if contract has deal status but no invoices generated yet
		if autoInvoice && len(c.Invoices) == 0 && c.Termin != nil {
			var rab entity.RAB
			if err := s.db.WithContext(ctx).First(&rab, c.RABID).Error; err == nil {
				for _, step := range c.Termin.Tahapan {
					inv := entity.Invoice{
						ContractID: c.ID,
						OrderID:    c.OrderID,
						Step:       step.Step,
						Keterangan: step.Text,
						Persentase: step.Persentase,
						Amount:     (step.Persentase / 100.0) * rab.GrandTotal,
						Status:     "belum_bayar",
					}
					_ = s.repo.Create(ctx, &inv)
				}
				// Re-fetch contract to get preloaded invoices
				updatedContract, err := s.contractRepo.FindByID(ctx, c.ID)
				if err == nil {
					c = *updatedContract
				}
			}
		}

		// Map invoices to response DTOs
		invoiceDTOs := make([]dto.InvoiceResponse, len(c.Invoices))
		paidCount := 0
		var lastPaidInv *entity.Invoice

		for i, inv := range c.Invoices {
			invoiceDTOs[i] = dto.InvoiceResponse{
				ID:           inv.ID,
				ContractID:   inv.ContractID,
				OrderID:      inv.OrderID,
				Step:         inv.Step,
				Keterangan:   inv.Keterangan,
				Persentase:   inv.Persentase,
				Amount:       inv.Amount,
				Deadline:     inv.Deadline,
				Status:       inv.Status,
				PaymentProof: inv.PaymentProof,
				PaidAt:       inv.PaidAt,
				CreatedAt:    inv.CreatedAt,
				UpdatedAt:    inv.UpdatedAt,
			}
			if inv.Status == "terbayar" {
				paidCount++
				if lastPaidInv == nil || inv.Step > lastPaidInv.Step {
					lastPaidInv = &c.Invoices[i]
				}
			}
		}

		// Determine contract status pembayaran
		statusPembayaran := "Belum Bayar"
		if paidCount == len(c.Invoices) && len(c.Invoices) > 0 {
			statusPembayaran = "Lunas (100%)"
		} else if lastPaidInv != nil {
			statusPembayaran = fmt.Sprintf("%s (%.0f%%) Terbayar", lastPaidInv.Keterangan, lastPaidInv.Persentase)
		}

		// Map termin
		var terminDTO *dto.TerminResponse
		if c.Termin != nil {
			tahapanDTOs := make([]dto.TahapanResponse, len(c.Termin.Tahapan))
			for i, t := range c.Termin.Tahapan {
				tahapanDTOs[i] = dto.TahapanResponse{
					Step:       t.Step,
					Text:       t.Text,
					Persentase: t.Persentase,
				}
			}
			terminDTO = &dto.TerminResponse{
				ID:        c.Termin.ID,
				KodeTipe:  c.Termin.KodeTipe,
				NamaTipe:  c.Termin.NamaTipe,
				Deskripsi: c.Termin.Deskripsi,
				Tahapan:   tahapanDTOs,
			}
		}

		orderNum := ""
		projName := ""
		custName := ""
		if c.Order != nil {
			orderNum = c.Order.NomorOrder
			projName = c.Order.NamaProject
			custName = c.Order.NamaCustomer
		}

		res = append(res, dto.ContractInvoiceListResponse{
			ContractID:          c.ID,
			OrderID:             c.OrderID,
			NomorOrder:          orderNum,
			NamaProject:         projName,
			NamaCustomer:        custName,
			TerminID:            c.TerminID,
			Termin:              terminDTO,
			StatusPembayaran:    statusPembayaran,
			InvoiceResponseBy:   c.InvoiceResponseBy,
			InvoiceResponseTime: c.InvoiceResponseTime,
			Invoices:            invoiceDTOs,
		})
	}

	return res, nil
}

func (s *invoiceService) GetInvoicesByContractID(ctx context.Context, contractID uint) (*dto.ContractInvoiceListResponse, error) {
	c, err := s.contractRepo.FindByID(ctx, contractID)
	if err != nil {
		return nil, errors.New("kontrak tidak ditemukan")
	}

	// Fetch invoices manually
	var invoices []entity.Invoice
	if err := s.db.WithContext(ctx).Where("contract_id = ?", contractID).Order("step ASC").Find(&invoices).Error; err != nil {
		return nil, err
	}
	c.Invoices = invoices

	// Resolve target company ID from the contract
	companyID := uint(1)
	if c.Order != nil {
		companyID = c.Order.CompanyID
	} else {
		companyID = database.GetContextCompanyID(ctx)
		if companyID == 0 {
			companyID = 1
		}
	}

	// Check setting finance_auto_invoice
	var autoInvoice bool = true
	var settingVal string
	if err := s.db.WithContext(ctx).Model(&entity.Setting{}).Where("company_id = ? AND key = ?", companyID, "finance_auto_invoice").Pluck("value", &settingVal).Error; err == nil {
		autoInvoice = (settingVal == "true")
	}

	// Auto-generate invoices if deal but empty invoices
	if autoInvoice && len(c.Invoices) == 0 && c.Termin != nil {
		var rab entity.RAB
		if err := s.db.WithContext(ctx).First(&rab, c.RABID).Error; err == nil {
			for _, step := range c.Termin.Tahapan {
				inv := entity.Invoice{
					ContractID: c.ID,
					OrderID:    c.OrderID,
					Step:       step.Step,
					Keterangan: step.Text,
					Persentase: step.Persentase,
					Amount:     (step.Persentase / 100.0) * rab.GrandTotal,
					Status:     "belum_bayar",
				}
				_ = s.repo.Create(ctx, &inv)
			}
			// Re-fetch invoices
			_ = s.db.WithContext(ctx).Where("contract_id = ?", contractID).Order("step ASC").Find(&invoices)
			c.Invoices = invoices
		}
	}

	// Map invoices to response DTOs
	invoiceDTOs := make([]dto.InvoiceResponse, len(c.Invoices))
	paidCount := 0
	var lastPaidInv *entity.Invoice

	for i, inv := range c.Invoices {
		invoiceDTOs[i] = dto.InvoiceResponse{
			ID:           inv.ID,
			ContractID:   inv.ContractID,
			OrderID:      inv.OrderID,
			Step:         inv.Step,
			Keterangan:   inv.Keterangan,
			Persentase:   inv.Persentase,
			Amount:       inv.Amount,
			Deadline:     inv.Deadline,
			Status:       inv.Status,
			PaymentProof: inv.PaymentProof,
			PaidAt:       inv.PaidAt,
			CreatedAt:    inv.CreatedAt,
			UpdatedAt:    inv.UpdatedAt,
		}
		if inv.Status == "terbayar" {
			paidCount++
			if lastPaidInv == nil || inv.Step > lastPaidInv.Step {
				lastPaidInv = &c.Invoices[i]
			}
		}
	}

	// Determine contract status pembayaran
	statusPembayaran := "Belum Bayar"
	if paidCount == len(c.Invoices) && len(c.Invoices) > 0 {
		statusPembayaran = "Lunas (100%)"
	} else if lastPaidInv != nil {
		statusPembayaran = fmt.Sprintf("%s (%.0f%%) Terbayar", lastPaidInv.Keterangan, lastPaidInv.Persentase)
	}

	// Map termin
	var terminDTO *dto.TerminResponse
	if c.Termin != nil {
		tahapanDTOs := make([]dto.TahapanResponse, len(c.Termin.Tahapan))
		for i, t := range c.Termin.Tahapan {
			tahapanDTOs[i] = dto.TahapanResponse{
				Step:       t.Step,
				Text:       t.Text,
				Persentase: t.Persentase,
			}
		}
		terminDTO = &dto.TerminResponse{
			ID:        c.Termin.ID,
			KodeTipe:  c.Termin.KodeTipe,
			NamaTipe:  c.Termin.NamaTipe,
			Deskripsi: c.Termin.Deskripsi,
			Tahapan:   tahapanDTOs,
		}
	}

	orderNum := ""
	projName := ""
	custName := ""
	if c.Order != nil {
		orderNum = c.Order.NomorOrder
		projName = c.Order.NamaProject
		custName = c.Order.NamaCustomer
	}

	return &dto.ContractInvoiceListResponse{
		ContractID:          c.ID,
		OrderID:             c.OrderID,
		NomorOrder:          orderNum,
		NamaProject:         projName,
		NamaCustomer:        custName,
		TerminID:            c.TerminID,
		Termin:              terminDTO,
		StatusPembayaran:    statusPembayaran,
		InvoiceResponseBy:   c.InvoiceResponseBy,
		InvoiceResponseTime: c.InvoiceResponseTime,
		Invoices:            invoiceDTOs,
	}, nil
}

func (s *invoiceService) SubmitInvoiceResponse(ctx context.Context, contractID uint, email string) (*dto.ContractResponse, error) {
	c, err := s.contractRepo.FindByID(ctx, contractID)
	if err != nil {
		return nil, errors.New("kontrak tidak ditemukan")
	}

	// Fetch User
	var user entity.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	now := time.Now()
	c.InvoiceResponseBy = user.Name
	c.InvoiceResponseTime = &now

	if err := s.contractRepo.Update(ctx, c); err != nil {
		return nil, err
	}

	_ = s.logTaskSvc.RecordTouch(ctx, c.OrderID, "invoice", user.Name)

	// Load detail
	updated, err := s.contractRepo.FindByID(ctx, c.ID)
	if err != nil {
		return nil, err
	}

	// Helper function inside contract service maps this, but let's map it manually or retrieve it
	// To keep it simple, we can just return a dto.ContractResponse
	var terminID uint
	if updated.TerminID != nil {
		terminID = *updated.TerminID
	}
	var orderDTO *dto.OrderBriefResponse
	if updated.Order != nil {
		orderDTO = &dto.OrderBriefResponse{
			ID:             updated.Order.ID,
			NomorOrder:     updated.Order.NomorOrder,
			NamaProject:    updated.Order.NamaProject,
			NamaCustomer:   updated.Order.NamaCustomer,
			NamaPerusahaan: updated.Order.NamaPerusahaan,
			JenisInterior:  updated.Order.JenisInterior,
		}
	}
	return &dto.ContractResponse{
		ID:                 updated.ID,
		RABID:              updated.RABID,
		OrderID:            updated.OrderID,
		TerminID:           terminID,
		LamaKontrak:        updated.LamaKontrak,
		Status:             updated.Status,
		SignedContractFile: updated.SignedContractFile,
		ResponseBy:         updated.ResponseBy,
		ResponseTime:       updated.ResponseTime,
		CreatedAt:          updated.CreatedAt,
		UpdatedAt:          updated.UpdatedAt,
		Order:              orderDTO,
	}, nil
}

func (s *invoiceService) UpdateDeadline(ctx context.Context, id uint, deadlineStr string) (*dto.InvoiceResponse, error) {
	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("invoice tidak ditemukan")
	}

	if inv.Status == "terbayar" {
		return nil, errors.New("invoice sudah terbayar, deadline tidak dapat diubah")
	}

	deadlineTime, err := time.Parse("2006-01-02", deadlineStr)
	if err != nil {
		return nil, errors.New("format tanggal deadline tidak valid (harus YYYY-MM-DD)")
	}

	inv.Deadline = &deadlineTime
	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, err
	}

	_ = s.logTaskSvc.RecordTouch(ctx, inv.OrderID, "invoice", "")

	return &dto.InvoiceResponse{
		ID:           inv.ID,
		ContractID:   inv.ContractID,
		OrderID:      inv.OrderID,
		Step:         inv.Step,
		Keterangan:   inv.Keterangan,
		Persentase:   inv.Persentase,
		Amount:       inv.Amount,
		Deadline:     inv.Deadline,
		Status:       inv.Status,
		PaymentProof: inv.PaymentProof,
		PaidAt:       inv.PaidAt,
		CreatedAt:    inv.CreatedAt,
		UpdatedAt:    inv.UpdatedAt,
	}, nil
}

func (s *invoiceService) UploadPaymentProof(ctx context.Context, id uint, fileHeader *multipart.FileHeader) (*dto.InvoiceResponse, error) {
	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("invoice tidak ditemukan")
	}

	targetDir := filepath.Join(s.uploadDir, "payments")
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return nil, err
	}

	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("pay_%d_%d%s", id, time.Now().UnixNano(), ext)
	dstPath := filepath.Join(targetDir, filename)

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	now := time.Now()
	inv.PaymentProof = fmt.Sprintf("/uploads/payments/%s", filename)
	inv.Status = "terbayar"
	inv.PaidAt = &now

	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, err
	}

	_ = s.logTaskSvc.RecordTouch(ctx, inv.OrderID, "invoice", "")

	// Update order stage to "survey_ulang" if step is 1 (Down Payment) and log transition
	if inv.Step == 1 {
		if errStage := s.logTaskSvc.TransitionStage(ctx, inv.OrderID, "survey_ulang", ""); errStage != nil {
			s.logger.Error("Failed to update order stage to survey_ulang", zap.Error(errStage))
		}
	}

	return &dto.InvoiceResponse{
		ID:           inv.ID,
		ContractID:   inv.ContractID,
		OrderID:      inv.OrderID,
		Step:         inv.Step,
		Keterangan:   inv.Keterangan,
		Persentase:   inv.Persentase,
		Amount:       inv.Amount,
		Deadline:     inv.Deadline,
		Status:       inv.Status,
		PaymentProof: inv.PaymentProof,
		PaidAt:       inv.PaidAt,
		CreatedAt:    inv.CreatedAt,
		UpdatedAt:    inv.UpdatedAt,
	}, nil
}

func (s *invoiceService) GenerateInvoicePDF(ctx context.Context, id uint) ([]byte, string, error) {
	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", errors.New("invoice tidak ditemukan")
	}

	// Initialize A4 Portrait PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	// Header
	companyID := uint(1)
	if inv.Contract != nil && inv.Contract.Order != nil {
		companyID = inv.Contract.Order.CompanyID
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
		pdf.Image(logoFile, 20, 15, 0, 15, false, "", 0, "")
	}
	pdf.SetY(15)

	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0) // Black instead of Teal
	pdf.CellFormat(0, 7, cp.Name, "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(0, 4, "Premium Interior Design & Architecture Services", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 4, fmt.Sprintf("Email: %s | Phone: %s", cp.Email, cp.Phone), "", 1, "C", false, 0, "")

	// Reset margins
	pdf.SetLeftMargin(20)
	pdf.SetX(20)
	pdf.SetY(32)

	pdf.SetDrawColor(0, 0, 0) // Black instead of Teal
	pdf.SetLineWidth(0.8)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(6)

	// Title
	pdf.SetFont("Arial", "B", 13)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 6, "INVOICE TAGIHAN KESEPAKATAN INTERIOR", "", 1, "C", false, 0, "")
	
	invNo := fmt.Sprintf("No. Invoice: INV/NUG/%d/%02d/%04d", inv.CreatedAt.Year(), inv.CreatedAt.Month(), inv.ID)
	pdf.SetFont("Arial", "I", 9.5)
	pdf.CellFormat(0, 5, invNo, "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Information table
	custName := "-"
	custPhone := "-"
	custEmail := "-"
	custAddress := "-"
	projName := "-"
	orderNum := "-"
	if inv.Contract != nil && inv.Contract.Order != nil {
		custName = inv.Contract.Order.NamaCustomer
		custPhone = inv.Contract.Order.TeleponCustomer
		custEmail = inv.Contract.Order.EmailCustomer
		custAddress = inv.Contract.Order.Alamat
		projName = inv.Contract.Order.NamaProject
		orderNum = inv.Contract.Order.NomorOrder
	}

	pdf.SetFont("Arial", "B", 9.5)
	pdf.SetTextColor(0, 128, 128)
	pdf.CellFormat(85, 5, "TAGIHAN KEPADA:", "", 0, "L", false, 0, "")
	pdf.CellFormat(85, 5, "RINCIAN PEMBAYARAN:", "", 1, "L", false, 0, "")
	
	pdf.SetLineWidth(0.2)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(2)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(70, 70, 70)
	
	// Print left column (Client info) and right column (Order info) side by side
	pdf.CellFormat(85, 5, "Nama: "+custName, "", 0, "L", false, 0, "")
	pdf.CellFormat(85, 5, "Nomor Order: "+orderNum, "", 1, "L", false, 0, "")
	
	pdf.CellFormat(85, 5, "Telepon: "+custPhone, "", 0, "L", false, 0, "")
	pdf.CellFormat(85, 5, "Nama Project: "+projName, "", 1, "L", false, 0, "")
	
	pdf.CellFormat(85, 5, "Email: "+custEmail, "", 0, "L", false, 0, "")
	dateStr := inv.CreatedAt.Format("02 Jan 2006")
	pdf.CellFormat(85, 5, "Tanggal Invoice: "+dateStr, "", 1, "L", false, 0, "")

	deadlineStr := "-"
	if inv.Deadline != nil {
		deadlineStr = inv.Deadline.Format("02 Jan 2006")
	}
	pdf.CellFormat(85, 5, "", "", 0, "L", false, 0, "")
	pdf.CellFormat(85, 5, "Jatuh Tempo: "+deadlineStr, "", 1, "L", false, 0, "")
	
	// Print address with MultiCell so it wraps dynamically and does not overlap
	pdf.MultiCell(170, 5, "Alamat: "+custAddress, "", "L", false)
	pdf.Ln(3)

	// Detail Tagihan Table
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(15, 6, "Step", "1", 0, "C", true, 0, "")
	pdf.CellFormat(95, 6, "Tahapan Termin / Keterangan Tagihan", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 6, "Persentase", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 6, "Jumlah Tagihan", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(15, 7, fmt.Sprintf("%d", inv.Step), "1", 0, "C", false, 0, "")
	pdf.CellFormat(95, 7, " Tagihan "+inv.Keterangan, "1", 0, "L", false, 0, "")
	pdf.CellFormat(25, 7, fmt.Sprintf("%.1f%%", inv.Persentase), "1", 0, "C", false, 0, "")
	pdf.CellFormat(35, 7, fmt.Sprintf("Rp %s ", formatRupiah(inv.Amount)), "1", 1, "R", false, 0, "")

	var taxEnabled bool
	var settingVal string
	if err := s.db.WithContext(ctx).Model(&entity.Setting{}).Where("company_id = ? AND key = ?", companyID, "finance_tax_enabled").Pluck("value", &settingVal).Error; err == nil {
		taxEnabled = (settingVal == "true")
	}

	pdf.SetFont("Arial", "B", 9)
	if taxEnabled {
		dpp := inv.Amount / 1.11
		tax := inv.Amount - dpp

		// DPP Row
		pdf.CellFormat(135, 6, "Dasar Pengenaan Pajak (DPP) ", "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("Rp %s ", formatRupiah(dpp)), "1", 1, "R", false, 0, "")

		// PPN Row
		pdf.CellFormat(135, 6, "PPN (11%) ", "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("Rp %s ", formatRupiah(tax)), "1", 1, "R", false, 0, "")
	}

	// Total Row
	pdf.CellFormat(135, 7, "TOTAL TAGIHAN ", "1", 0, "R", false, 0, "")
	pdf.CellFormat(35, 7, fmt.Sprintf("Rp %s ", formatRupiah(inv.Amount)), "1", 1, "R", false, 0, "")
	pdf.Ln(6)

	// Payment Instructions
	pdf.SetFont("Arial", "B", 9.5)
	pdf.SetTextColor(0, 128, 128)
	pdf.CellFormat(0, 5, "INSTRUKSI PEMBAYARAN:", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(0, 5, "Pembayaran dapat dilakukan melalui transfer bank ke rekening resmi berikut:", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(0, 5, fmt.Sprintf("Bank: %s", cp.BankName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Nomor Rekening: %s", cp.BankAccount), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Atas Nama: %s", cp.BankHolder), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "I", 8.5)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 4, "*Harap unggah bukti transfer ke sistem Arsiflow untuk verifikasi pelunasan.", "", 1, "L", false, 0, "")
	pdf.Ln(8)

	// Signatures
	pdf.SetFont("Arial", "", 9.5)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(85, 5, "Dibuat Oleh,", "", 0, "C", false, 0, "")
	pdf.CellFormat(85, 5, "Diterima Oleh,", "", 1, "C", false, 0, "")
	pdf.Ln(18)

	pdf.SetFont("Arial", "B", 9.5)
	pdf.CellFormat(85, 5, fmt.Sprintf("( %s )", cp.Name), "", 0, "C", false, 0, "")
	pdf.CellFormat(85, 5, fmt.Sprintf("( %s )", custName), "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", err
	}

	safeProjName := "Project"
	if inv.Contract != nil && inv.Contract.Order != nil {
		safeProjName = inv.Contract.Order.NamaProject
	}
	safeProjName = strings.ReplaceAll(safeProjName, " ", "_")
	filename := fmt.Sprintf("Invoice_%s_Step%d_%d.pdf", safeProjName, inv.Step, inv.ID)

	return buf.Bytes(), filename, nil
}

func (s *invoiceService) GenerateInvoicesForContract(ctx context.Context, contractID uint) error {
	c, err := s.contractRepo.FindByID(ctx, contractID)
	if err != nil {
		return errors.New("kontrak tidak ditemukan")
	}

	// Check if invoices already exist
	var count int64
	if err := s.db.WithContext(ctx).Model(&entity.Invoice{}).Where("contract_id = ?", contractID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("tagihan invoice untuk kontrak ini sudah diterbitkan")
	}

	if c.Termin == nil {
		return errors.New("termin tipe belum dipilih pada kontrak ini")
	}

	var rab entity.RAB
	if err := s.db.WithContext(ctx).First(&rab, c.RABID).Error; err != nil {
		return errors.New("RAB tidak ditemukan")
	}

	for _, step := range c.Termin.Tahapan {
		inv := entity.Invoice{
			ContractID: c.ID,
			OrderID:    c.OrderID,
			Step:       step.Step,
			Keterangan: step.Text,
			Persentase: step.Persentase,
			Amount:     (step.Persentase / 100.0) * rab.GrandTotal,
			Status:     "belum_bayar",
		}
		if err := s.repo.Create(ctx, &inv); err != nil {
			return err
		}
	}

	_ = s.logTaskSvc.RecordTouch(ctx, c.OrderID, "invoice", "Sistem Manual")
	return nil
}
