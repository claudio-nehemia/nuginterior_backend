package service

import (
	"bytes"
	"context"
	"encoding/json"
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

type ContractService interface {
	GetContractList(ctx context.Context) ([]dto.RABContractResponse, error)
	CreateContract(ctx context.Context, req dto.CreateContractRequest) (*dto.ContractResponse, error)
	SubmitResponse(ctx context.Context, rabID uint, email string) (*dto.ContractResponse, error)
	UploadSignedContract(ctx context.Context, id uint, fileHeader *multipart.FileHeader) (*dto.ContractResponse, error)
	GenerateContractPDF(ctx context.Context, id uint) ([]byte, string, error) // Returns PDF bytes, filename, error
}

type contractService struct {
	repo       repository.ContractRepository
	rabRepo    repository.RABRepository
	terminRepo repository.TerminRepository
	userRepo   repository.UserRepository
	db         *gorm.DB
	logger     *zap.Logger
	uploadDir  string
	logTaskSvc ProjectLogTaskService
}

func NewContractService(
	repo repository.ContractRepository,
	rabRepo repository.RABRepository,
	terminRepo repository.TerminRepository,
	userRepo repository.UserRepository,
	db *gorm.DB,
	logger *zap.Logger,
	uploadDir string,
	logTaskSvc ProjectLogTaskService,
) ContractService {
	return &contractService{
		repo:       repo,
		rabRepo:    rabRepo,
		terminRepo: terminRepo,
		userRepo:   userRepo,
		db:         db,
		logger:     logger,
		uploadDir:  uploadDir,
		logTaskSvc: logTaskSvc,
	}
}

func (s *contractService) GetContractList(ctx context.Context) ([]dto.RABContractResponse, error) {
	rabs, err := s.rabRepo.FindAll(ctx) // Fetch all RABs
	if err != nil {
		return nil, err
	}

	var res []dto.RABContractResponse
	for _, rab := range rabs {
		// Only include submitted RABs
		if rab.Status != "submitted" {
			continue
		}

		// Check if contract exists (load it manually if not preloaded)
		var contract *entity.Contract
		err := s.db.WithContext(ctx).
			Preload("Termin").
			Where("rab_id = ?", rab.ID).
			First(&contract).Error

		var contractDTO *dto.ContractResponse
		status := "belum_dibuat"
		var contractID *uint

		if err == nil && contract != nil {
			contractID = &contract.ID
			status = contract.Status
			contractDTO = s.mapToContractResponse(contract)
		}

		orderNum := ""
		projName := ""
		custName := ""
		if rab.Order != nil {
			orderNum = rab.Order.NomorOrder
			projName = rab.Order.NamaProject
			custName = rab.Order.NamaCustomer
		}

		res = append(res, dto.RABContractResponse{
			RABID:        rab.ID,
			OrderID:      rab.OrderID,
			NomorOrder:   orderNum,
			NamaProject:  projName,
			NamaCustomer: custName,
			GrandTotal:   rab.GrandTotal,
			Status:       status,
			ContractID:   contractID,
			Contract:     contractDTO,
		})
	}

	return res, nil
}

func (s *contractService) CreateContract(ctx context.Context, req dto.CreateContractRequest) (*dto.ContractResponse, error) {
	// 1. Fetch RAB
	rab, err := s.rabRepo.FindByID(ctx, req.RABID)
	if err != nil {
		return nil, errors.New("RAB tidak ditemukan")
	}
	if rab.Status != "submitted" {
		return nil, errors.New("RAB belum disubmit, tidak dapat membuat kontrak")
	}

	// 2. Fetch Termin
	var termin entity.Termin
	if err := s.db.WithContext(ctx).First(&termin, req.TerminID).Error; err != nil {
		return nil, errors.New("Termin tidak ditemukan")
	}

	// 3. Check if contract already exists
	var contract entity.Contract
	err = s.db.WithContext(ctx).Where("rab_id = ?", req.RABID).First(&contract).Error
	exists := err == nil

	// Resolve target company ID from the order
	companyID := uint(1)
	if rab.Order != nil {
		companyID = rab.Order.CompanyID
	} else {
		companyID = database.GetContextCompanyID(ctx)
		if companyID == 0 {
			companyID = 1
		}
	}

	// Check setting workflow_rab_approval_required
	var settingVal string
	if err := s.db.WithContext(ctx).Model(&entity.Setting{}).Where("company_id = ? AND key = ?", companyID, "workflow_rab_approval_required").Pluck("value", &settingVal).Error; err == nil {
		if settingVal == "true" {
			// Also check if response_enabled is active; if disabled, skip response check
			var responseEnabled string
			s.db.WithContext(ctx).Model(&entity.Setting{}).Where("company_id = ? AND key = ?", companyID, "response_enabled").Pluck("value", &responseEnabled)
			if responseEnabled != "false" {
				if !exists || contract.ResponseTime == nil {
					return nil, errors.New("kontrak wajib merespons regular terlebih dahulu sebelum digenerate")
				}
			}
		}
	}

	if exists {
		contract.TerminID = &req.TerminID
		contract.LamaKontrak = req.LamaKontrak
		contract.Status = "draft"
		if err := s.repo.Update(ctx, &contract); err != nil {
			return nil, err
		}
	} else {
		contract = entity.Contract{
			RABID:       req.RABID,
			OrderID:     rab.OrderID,
			TerminID:    &req.TerminID,
			LamaKontrak: req.LamaKontrak,
			Status:      "draft",
		}
		if err := s.repo.Create(ctx, &contract); err != nil {
			return nil, err
		}
	}

	// Load updated details
	updated, err := s.repo.FindByID(ctx, contract.ID)
	if err != nil {
		return nil, err
	}

	_ = s.logTaskSvc.RecordTouch(ctx, contract.OrderID, "kontrak", "")

	return s.mapToContractResponse(updated), nil
}

func (s *contractService) SubmitResponse(ctx context.Context, rabID uint, email string) (*dto.ContractResponse, error) {
	// 1. Fetch RAB
	rab, err := s.rabRepo.FindByID(ctx, rabID)
	if err != nil {
		return nil, errors.New("RAB tidak ditemukan")
	}
	if rab.Status != "submitted" {
		return nil, errors.New("RAB belum disubmit, tidak dapat merespons kontrak")
	}

	// 2. Fetch User
	var user entity.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	// 3. Find or Create contract
	var contract entity.Contract
	err = s.db.WithContext(ctx).Where("rab_id = ?", rabID).First(&contract).Error
	now := time.Now()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			contract = entity.Contract{
				RABID:        rabID,
				OrderID:      rab.OrderID,
				Status:       "belum_dibuat",
				ResponseBy:   user.Name,
				ResponseTime: &now,
			}
			if err := s.repo.Create(ctx, &contract); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		contract.ResponseBy = user.Name
		contract.ResponseTime = &now
		if err := s.repo.Update(ctx, &contract); err != nil {
			return nil, err
		}
	}

	// Load detail
	updated, err := s.repo.FindByID(ctx, contract.ID)
	if err != nil {
		return nil, err
	}

	_ = s.logTaskSvc.RecordTouch(ctx, contract.OrderID, "kontrak", user.Name)

	return s.mapToContractResponse(updated), nil
}

func (s *contractService) UploadSignedContract(ctx context.Context, id uint, fileHeader *multipart.FileHeader) (*dto.ContractResponse, error) {
	contract, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("kontrak tidak ditemukan")
	}

	targetDir := filepath.Join(s.uploadDir, "contracts")
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return nil, err
	}

	ext := filepath.Ext(fileHeader.Filename)
	if ext != ".pdf" {
		return nil, errors.New("file yang diunggah harus berformat PDF")
	}

	filename := fmt.Sprintf("signed_%d_%d%s", id, time.Now().UnixNano(), ext)
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

	contract.SignedContractFile = fmt.Sprintf("/uploads/contracts/%s", filename)
	contract.Status = "deal"
	now := time.Now()
	contract.ResponseTime = &now

	if err := s.repo.Update(ctx, contract); err != nil {
		return nil, err
	}

	// Update order stage to "invoice", and set contract price/date
	var hargaVal float64
	if contract.RAB != nil {
		hargaVal = contract.RAB.GrandTotal
	}

	// Transition stage to invoice and log transition
	if errStage := s.logTaskSvc.TransitionStage(ctx, contract.OrderID, "invoice", ""); errStage != nil {
		s.logger.Error("Failed to update order stage to invoice", zap.Error(errStage))
	}

	if errUpdate := s.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", contract.OrderID).Updates(map[string]interface{}{
		"harga_kontrak":   hargaVal,
		"tanggal_kontrak": &now,
	}).Error; errUpdate != nil {
		s.logger.Error("Failed to update order contract details", zap.Error(errUpdate))
	}

	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.mapToContractResponse(updated), nil
}

func (s *contractService) GenerateContractPDF(ctx context.Context, id uint) ([]byte, string, error) {
	contract, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", errors.New("kontrak tidak ditemukan")
	}

	// Fetch detail RAB
	rab, err := s.rabRepo.FindByID(ctx, contract.RABID)
	if err != nil {
		return nil, "", errors.New("RAB tidak ditemukan")
	}

	// Initialize A4 Portrait PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	// 1. Header (Dynamic)
	companyID := uint(1)
	if contract.Order != nil {
		companyID = contract.Order.CompanyID
	} else if rab.Order != nil {
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
		pdf.Image(logoFile, 20, 15, 0, 15, false, "", 0, "")
	}
	pdf.SetY(15)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 7, cp.Name, "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(0, 4, "Premium Interior Design & Architecture Services", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 4, fmt.Sprintf("Email: %s | Phone: %s", cp.Email, cp.Phone), "", 1, "C", false, 0, "")

	pdf.SetLeftMargin(20)
	pdf.SetX(20)
	pdf.SetY(32)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.8)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(6)

	// 2. Title
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 6, "SURAT PERJANJIAN KERJASAMA DESAIN & PENGERJAAN INTERIOR", "", 1, "C", false, 0, "")
	
	docNo := fmt.Sprintf("No. Kontrak: KONTRAK/NUG/%d/%02d/%04d", contract.CreatedAt.Year(), contract.CreatedAt.Month(), contract.ID)
	pdf.SetFont("Arial", "I", 9)
	pdf.CellFormat(0, 5, docNo, "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Opening statement
	nowStr := time.Now().Format("02 January 2006")
	pdf.SetFont("Arial", "", 9.5)
	pdf.SetTextColor(70, 70, 70)
	opening := fmt.Sprintf("Pada hari ini, %s, telah disepakati Perjanjian Kerjasama Pengerjaan Interior antara pihak-pihak di bawah ini:", nowStr)
	pdf.MultiCell(0, 5, opening, "", "L", false)
	pdf.Ln(4)

	// Party 1 info
	pdf.SetFont("Arial", "B", 9.5)
	pdf.CellFormat(0, 5, "PIHAK PERTAMA (Penyedia Jasa):", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9.5)
	pdf.CellFormat(40, 5, "  Nama Perusahaan", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, ": "+cp.Name, "", 1, "L", false, 0, "")
	pdf.CellFormat(40, 5, "  Perwakilan", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, ": "+cp.Director, "", 1, "L", false, 0, "")
	pdf.CellFormat(40, 5, "  Alamat Kantor", "", 0, "L", false, 0, "")
	pdf.MultiCell(0, 5, ": "+cp.Address, "", "L", false)
	pdf.Ln(3)

	// Party 2 info
	custName := "-"
	custPhone := "-"
	custEmail := "-"
	custAddress := "-"
	projName := "-"
	interiorType := "-"
	if contract.Order != nil {
		custName = contract.Order.NamaCustomer
		custPhone = contract.Order.TeleponCustomer
		custEmail = contract.Order.EmailCustomer
		custAddress = contract.Order.Alamat
		projName = contract.Order.NamaProject
		interiorType = contract.Order.JenisInterior
	}

	pdf.SetFont("Arial", "B", 9.5)
	pdf.CellFormat(0, 5, "PIHAK KEDUA (Customer):", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9.5)
	pdf.CellFormat(40, 5, "  Nama Lengkap", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, ": "+custName, "", 1, "L", false, 0, "")
	pdf.CellFormat(40, 5, "  No. Telepon", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, ": "+custPhone, "", 1, "L", false, 0, "")
	pdf.CellFormat(40, 5, "  Email", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, ": "+custEmail, "", 1, "L", false, 0, "")
	pdf.CellFormat(40, 5, "  Alamat Customer", "", 0, "L", false, 0, "")
	pdf.MultiCell(0, 5, ": "+custAddress, "", "L", false)
	pdf.Ln(4)

	agreementIntro := "Kedua belah pihak telah sepakat untuk mengadakan kerjasama pengerjaan interior dengan ketentuan yang diatur dalam pasal-pasal berikut:"
	pdf.MultiCell(0, 5, agreementIntro, "", "L", false)
	pdf.Ln(5)

	// Resolve termin name before dynamic clauses
	terminName := "Custom Term"
	if contract.Termin != nil {
		terminName = contract.Termin.NamaTipe
	}

	// Fetch dynamic contract clauses from settings
	var clausesJSON string
	if err := s.db.WithContext(ctx).Scopes(database.CompanyScope(ctx)).Model(&entity.Setting{}).Where("company_id = ? AND key = ?", companyID, "contract_clauses").Pluck("value", &clausesJSON).Error; err != nil || clausesJSON == "" {
		// use default if not found
		clausesJSON = `[{"title":"PASAL 1: LINGKUP PEKERJAAN DAN OBJEK PROYEK","content":"Pihak Pertama berkewajiban untuk menyelesaikan pekerjaan interior untuk Pihak Kedua pada proyek bernama \"{nama_project}\" dengan kategori jenis interior \"{jenis_interior}\" beralamat di \"{alamat_customer}\". Rincian pengerjaan mengacu sepenuhnya pada Rencana Anggaran Biaya (RAB) nomor {rab_id}."},{"title":"PASAL 2: NILAI KONTRAK","content":"Nilai kontrak keseluruhan yang disepakati oleh kedua belah pihak adalah sebesar Rp {nilai_kontrak} (rupiah) sesuai dengan total perhitungan RAB yang telah disetujui bersama."},{"title":"PASAL 3: KETENTUAN DAN TERMIN PEMBAYARAN","content":"Metode pembayaran disepakati menggunakan skema termin \"{skema_termin}\", dengan rincian pembagian sebagai berikut:\n\n{tabel_termin}"},{"title":"PASAL 4: JANGKA WAKTU PENGERJAAN","content":"Seluruh lingkup pengerjaan disepakati akan diselesaikan oleh Pihak Pertama dalam jangka waktu pengerjaan selama {lama_kontrak} terhitung sejak ditandatanganinya perjanjian ini dan pembayaran termin pertama (DP) telah diterima oleh Pihak Pertama."}]`
	}

	var clauses []struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(clausesJSON), &clauses); err != nil {
		return nil, "", fmt.Errorf("failed to parse contract clauses: %w", err)
	}

	// Build placeholder replacements
	replacements := map[string]string{
		"{nama_project}":        projName,
		"{jenis_interior}":      interiorType,
		"{alamat_customer}":     custAddress,
		"{nama_customer}":       custName,
		"{telepon_customer}":    custPhone,
		"{email_customer}":      custEmail,
		"{nama_perusahaan}":     cp.Name,
		"{direktur_perusahaan}": cp.Director,
		"{alamat_perusahaan}":   cp.Address,
		"{nilai_kontrak}":       formatRupiah(rab.GrandTotal),
		"{skema_termin}":        terminName,
		"{lama_kontrak}":        contract.LamaKontrak,
		"{rab_id}":              fmt.Sprintf("%d", rab.ID),
	}

	// Render each clause dynamically
	for _, clause := range clauses {
		// Print clause title
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(0, 5, clause.Title, "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 9.5)

		// Replace all placeholders in content
		content := clause.Content
		for placeholder, value := range replacements {
			content = strings.ReplaceAll(content, placeholder, value)
		}

		// Check if content contains the termin table placeholder
		if strings.Contains(content, "{tabel_termin}") {
			parts := strings.SplitN(content, "{tabel_termin}", 2)
			// Render text before the table
			if strings.TrimSpace(parts[0]) != "" {
				pdf.MultiCell(0, 5, parts[0], "", "L", false)
			}
			// Render termin details table
			if contract.Termin != nil {
				pdf.SetFont("Arial", "B", 9)
				pdf.SetFillColor(240, 240, 240)
				pdf.CellFormat(15, 6, "Step", "1", 0, "C", true, 0, "")
				pdf.CellFormat(85, 6, "Tahapan Pekerjaan / Keterangan", "1", 0, "L", true, 0, "")
				pdf.CellFormat(25, 6, "Persentase", "1", 0, "C", true, 0, "")
				pdf.CellFormat(45, 6, "Jumlah Pembayaran (Rupiah)", "1", 1, "R", true, 0, "")

				pdf.SetFont("Arial", "", 9)
				for _, item := range contract.Termin.Tahapan {
					amount := (item.Persentase / 100.0) * rab.GrandTotal
					pdf.CellFormat(15, 6, fmt.Sprintf("%d", item.Step), "1", 0, "C", false, 0, "")
					pdf.CellFormat(85, 6, " "+item.Text, "1", 0, "L", false, 0, "")
					pdf.CellFormat(25, 6, fmt.Sprintf("%.1f%%", item.Persentase), "1", 0, "C", false, 0, "")
					pdf.CellFormat(45, 6, fmt.Sprintf("Rp %s ", formatRupiah(amount)), "1", 1, "R", false, 0, "")
				}
			}
			// Render text after the table
			if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
				pdf.SetFont("Arial", "", 9.5)
				pdf.MultiCell(0, 5, parts[1], "", "L", false)
			}
		} else {
			pdf.MultiCell(0, 5, content, "", "L", false)
		}
		pdf.Ln(4)
	}
	pdf.Ln(4)

	// Signatures section
	pdf.SetFont("Arial", "", 9.5)
	pdf.CellFormat(85, 5, "Pihak Pertama (Penyedia Jasa)", "", 0, "C", false, 0, "")
	pdf.CellFormat(85, 5, "Pihak Kedua (Customer)", "", 1, "C", false, 0, "")
	pdf.Ln(18) // Spacing for signature

	pdf.SetFont("Arial", "B", 9.5)
	pdf.CellFormat(85, 5, fmt.Sprintf("( %s )", cp.Name), "", 0, "C", false, 0, "")
	pdf.CellFormat(85, 5, fmt.Sprintf("( %s )", custName), "", 1, "C", false, 0, "")

	// Output PDF to buffer
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", err
	}

	safeProjName := strings.ReplaceAll(projName, " ", "_")
	filename := fmt.Sprintf("Kontrak_%s_%d.pdf", safeProjName, contract.ID)

	return buf.Bytes(), filename, nil
}

// Map helper to DTO response
func (s *contractService) mapToContractResponse(c *entity.Contract) *dto.ContractResponse {
	if c == nil {
		return nil
	}

	var orderDTO *dto.OrderBriefResponse
	if c.Order != nil {
		orderDTO = &dto.OrderBriefResponse{
			ID:             c.Order.ID,
			NomorOrder:     c.Order.NomorOrder,
			NamaProject:    c.Order.NamaProject,
			NamaCustomer:   c.Order.NamaCustomer,
			NamaPerusahaan: c.Order.NamaPerusahaan,
			JenisInterior:  c.Order.JenisInterior,
		}
	}

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
			CreatedAt: c.Termin.CreatedAt,
			UpdatedAt: c.Termin.UpdatedAt,
		}
	}

	var rabDTO *dto.RABResponse
	if c.RAB != nil {
		rabDTO = &dto.RABResponse{
			ID:          c.RAB.ID,
			InputItemID: c.RAB.InputItemID,
			OrderID:     c.RAB.OrderID,
			GrandTotal:  c.RAB.GrandTotal,
			Status:      c.RAB.Status,
			SubmittedAt: c.RAB.SubmittedAt,
			SubmittedBy: c.RAB.SubmittedBy,
		}
	}

	var terminID uint
	if c.TerminID != nil {
		terminID = *c.TerminID
	}

	return &dto.ContractResponse{
		ID:                 c.ID,
		RABID:              c.RABID,
		OrderID:            c.OrderID,
		TerminID:           terminID,
		LamaKontrak:        c.LamaKontrak,
		Status:             c.Status,
		SignedContractFile: c.SignedContractFile,
		ResponseBy:         c.ResponseBy,
		ResponseTime:       c.ResponseTime,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
		Order:              orderDTO,
		Termin:             terminDTO,
		RAB:                rabDTO,
	}
}
