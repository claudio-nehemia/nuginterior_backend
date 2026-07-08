package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"github.com/claudio-nehemia/interior_backend/pkg/pdf"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MoodboardService interface {
	GetAll(ctx context.Context) ([]dto.MoodboardResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.MoodboardResponse, error)
	GetByOrderID(ctx context.Context, orderID uint) (*dto.MoodboardResponse, error)

	// Module A: Moodboard (Desain Kasar)
	ResponseMoodboard(ctx context.Context, orderID uint, designerName string) (*dto.MoodboardResponse, error)
	UploadKasar(ctx context.Context, moodboardID uint, files []*multipart.FileHeader) ([]dto.MoodboardFileResponse, error)
	AcceptDesain(ctx context.Context, moodboardID uint, req dto.AcceptDesainRequest, pmName string) (*dto.MoodboardResponse, error)
	ReviseDesain(ctx context.Context, moodboardID uint, req dto.ReviseRequest, pmName string) (*dto.MoodboardResponse, error)
	DeleteFile(ctx context.Context, fileID uint) error

	// Module B: Estimasi (RAB Kasar)
	ResponseEstimasi(ctx context.Context, moodboardID uint, estimatorName string) (*dto.MoodboardResponse, error)
	UploadEstimasi(ctx context.Context, estimasiID uint, moodboardFileID uint, fileHeader *multipart.FileHeader) (*dto.EstimasiFileResponse, error)

	// Module C: Commitment Fee
	ResponseCommitmentFee(ctx context.Context, moodboardID uint, csName string) (*dto.MoodboardResponse, error)
	UpdateTotalFee(ctx context.Context, feeID uint, req dto.UpdateTotalFeeRequest) (*dto.MoodboardResponse, error)
	UploadPaymentProof(ctx context.Context, feeID uint, fileHeader *multipart.FileHeader) (*dto.MoodboardResponse, error)
	VerifyPayment(ctx context.Context, feeID uint, csName string) (*dto.MoodboardResponse, error)
	ResetPayment(ctx context.Context, feeID uint) (*dto.MoodboardResponse, error)
	RevisePaymentFee(ctx context.Context, feeID uint, req dto.UpdateTotalFeeRequest) (*dto.MoodboardResponse, error)
	PrintInvoice(ctx context.Context, feeID uint) ([]byte, string, error) // returns PDF bytes, filename, error
}

type moodboardService struct {
	repo       repository.MoodboardRepository
	db         *gorm.DB
	cache      cache.Store
	logger     *zap.Logger
	uploadDir  string
	logTaskSvc ProjectLogTaskService
}

func NewMoodboardService(repo repository.MoodboardRepository, db *gorm.DB, cacheStore cache.Store, logger *zap.Logger, uploadDir string, logTaskSvc ProjectLogTaskService) MoodboardService {
	return &moodboardService{
		repo:       repo,
		db:         db,
		cache:      cacheStore,
		logger:     logger,
		uploadDir:  uploadDir,
		logTaskSvc: logTaskSvc,
	}
}

func (s *moodboardService) GetAll(ctx context.Context) ([]dto.MoodboardResponse, error) {
	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.MoodboardResponse, len(list))
	for i, m := range list {
		result[i] = toMoodboardResponseEntity(m, true)
	}
	return result, nil
}

func (s *moodboardService) GetByID(ctx context.Context, id uint) (*dto.MoodboardResponse, error) {
	m, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toMoodboardResponseEntity(*m, true)
	return &resp, nil
}

func (s *moodboardService) GetByOrderID(ctx context.Context, orderID uint) (*dto.MoodboardResponse, error) {
	m, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	resp := toMoodboardResponseEntity(*m, true)
	return &resp, nil
}

// ==========================================
// MODULE A: MOODBOARD (DESAIN KASAR)
// ==========================================

func (s *moodboardService) ResponseMoodboard(ctx context.Context, orderID uint, designerName string) (*dto.MoodboardResponse, error) {
	m, err := s.repo.FindByOrderID(ctx, orderID)
	now := time.Now()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create a new moodboard record
			m = &entity.Moodboard{
				OrderID:      orderID,
				ResponseBy:   designerName,
				ResponseTime: &now,
			}
			if err := s.repo.Create(ctx, m); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// Update response logs if not already set
		if m.ResponseBy == "" {
			m.ResponseBy = designerName
			m.ResponseTime = &now
			if err := s.repo.Update(ctx, m); err != nil {
				return nil, err
			}
		}
	}

	// Move order stage to "moodboard" and log transition
	if err := s.logTaskSvc.TransitionStage(ctx, orderID, "moodboard", designerName); err != nil {
		s.logger.Error("Failed to update order stage", zap.Error(err))
	}

	s.invalidateCache(ctx)
	return s.GetByID(ctx, m.ID)
}

func (s *moodboardService) UploadKasar(ctx context.Context, moodboardID uint, files []*multipart.FileHeader) ([]dto.MoodboardFileResponse, error) {
	m, err := s.repo.FindByID(ctx, moodboardID)
	if err != nil {
		return nil, err
	}

	targetDir := filepath.Join(s.uploadDir, "moodboards")
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return nil, err
	}

	var uploaded []dto.MoodboardFileResponse

	for _, fileHeader := range files {
		ext := filepath.Ext(fileHeader.Filename)
		filename := fmt.Sprintf("kasar_%d_%d%s", moodboardID, time.Now().UnixNano(), ext)
		dstPath := filepath.Join(targetDir, filename)

		if err := s.saveUploadedFile(fileHeader, dstPath); err != nil {
			return nil, err
		}

		dbFile := &entity.MoodboardFile{
			MoodboardID:  moodboardID,
			FilePath:     fmt.Sprintf("/uploads/moodboards/%s", filename),
			OriginalName: fileHeader.Filename,
		}

		if err := s.repo.CreateFile(ctx, dbFile); err != nil {
			return nil, err
		}

		uploaded = append(uploaded, dto.MoodboardFileResponse{
			ID:           dbFile.ID,
			FilePath:     dbFile.FilePath,
			FileType:     "kasar",
			OriginalName: dbFile.OriginalName,
			Status:       dbFile.Status,
			Revisi:       dbFile.Revisi,
			CreatedAt:    dbFile.CreatedAt,
		})
	}

	// Update moodboard to trigger GORM update (like UpdatedAt)
	_ = s.repo.Update(ctx, m)

	// Transition stage to estimasi and log transition
	if errStage := s.logTaskSvc.TransitionStage(ctx, m.OrderID, "estimasi", ""); errStage != nil {
		s.logger.Error("Failed to update order stage to estimasi", zap.Error(errStage))
	}

	s.invalidateCache(ctx)
	return uploaded, nil
}

func (s *moodboardService) AcceptDesain(ctx context.Context, moodboardID uint, req dto.AcceptDesainRequest, pmName string) (*dto.MoodboardResponse, error) {
	m, err := s.repo.FindByID(ctx, moodboardID)
	if err != nil {
		return nil, err
	}

	// Find the targeted rough file
	var targetFile *entity.MoodboardFile
	for _, f := range m.Files {
		if f.ID == req.MoodboardFileID {
			targetFile = &f
			break
		}
	}
	if targetFile == nil {
		return nil, errors.New("file desain kasar tidak ditemukan")
	}

	// CRITICAL VALIDATION: Ensure the selected rough file option has a corresponding Estimasi / RAB cost file
	if m.Estimasi == nil {
		return nil, errors.New("estimasi belum diinisialisasi oleh estimator")
	}

	var targetEstimasiFile *entity.EstimasiFile
	for _, ef := range m.Estimasi.Files {
		if ef.MoodboardFileID == targetFile.ID {
			targetEstimasiFile = &ef
			break
		}
	}
	if targetEstimasiFile == nil {
		return nil, errors.New("file RAB / Estimasi untuk opsi desain ini belum diunggah oleh estimator")
	}

	targetFile.Status = "approved"
	if err := s.repo.UpdateFile(ctx, targetFile); err != nil {
		return nil, err
	}

	now := time.Now()
	// Lock the corresponding RAB cost file in Estimasi table
	m.Estimasi.EstimatedCost = targetEstimasiFile.FilePath
	m.Estimasi.PmResponseBy = pmName
	m.Estimasi.PmResponseTime = &now

	if err := s.repo.UpdateEstimasi(ctx, m.Estimasi); err != nil {
		return nil, err
	}

	// Update moodboard marketing response fields
	m.MarketingResponseBy = pmName
	m.MarketingResponseTime = &now
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}

	// Transition stage to cm_fee and log transition
	if errStage := s.logTaskSvc.TransitionStage(ctx, m.OrderID, "cm_fee", pmName); errStage != nil {
		s.logger.Error("Failed to update order stage to cm_fee", zap.Error(errStage))
	}

	s.invalidateCache(ctx)
	return s.GetByID(ctx, moodboardID)
}

func (s *moodboardService) ReviseDesain(ctx context.Context, moodboardID uint, req dto.ReviseRequest, pmName string) (*dto.MoodboardResponse, error) {
	m, err := s.repo.FindByID(ctx, moodboardID)
	if err != nil {
		return nil, err
	}

	var targetFile *entity.MoodboardFile
	for i := range m.Files {
		if m.Files[i].ID == req.MoodboardFileID {
			targetFile = &m.Files[i]
			break
		}
	}
	if targetFile == nil {
		return nil, errors.New("file opsi desain tidak ditemukan")
	}

	targetFile.Status = "revisi"
	targetFile.Revisi = req.Notes
	if err := s.repo.UpdateFile(ctx, targetFile); err != nil {
		return nil, err
	}

	now := time.Now()
	m.MarketingResponseBy = pmName
	m.MarketingResponseTime = &now
	if err := s.repo.Update(ctx, m); err != nil {
		return nil, err
	}

	s.invalidateCache(ctx)
	return s.GetByID(ctx, moodboardID)
}

func (s *moodboardService) DeleteFile(ctx context.Context, fileID uint) error {
	file, err := s.repo.FindFileByID(ctx, fileID)
	if err != nil {
		return err
	}

	// Delete from local storage
	localPath := filepath.Join(s.uploadDir, "moodboards", filepath.Base(file.FilePath))
	_ = os.Remove(localPath)

	// Delete also from matching EstimasiFile links
	estFile, err := s.repo.FindEstimasiFileByMoodboardFileID(ctx, file.ID)
	if err == nil && estFile != nil {
		localEstPath := filepath.Join(s.uploadDir, "estimasis", filepath.Base(estFile.FilePath))
		_ = os.Remove(localEstPath)
		_ = s.repo.DeleteEstimasiFile(ctx, estFile.ID)
	}

	if err := s.repo.DeleteFile(ctx, fileID); err != nil {
		return err
	}

	s.invalidateCache(ctx)
	return nil
}

// ==========================================
// MODULE B: ESTIMASI (RAB KASAR)
// ==========================================

func (s *moodboardService) ResponseEstimasi(ctx context.Context, moodboardID uint, estimatorName string) (*dto.MoodboardResponse, error) {
	mb, err := s.repo.FindByID(ctx, moodboardID)
	if err != nil {
		return nil, err
	}
	_ = s.logTaskSvc.RecordTouch(ctx, mb.OrderID, "estimasi", estimatorName)

	est, err := s.repo.FindEstimasiByMoodboardID(ctx, moodboardID)
	now := time.Now()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			est = &entity.Estimasi{
				MoodboardID:  moodboardID,
				ResponseBy:   estimatorName,
				ResponseTime: &now,
			}
			if err := s.repo.CreateEstimasi(ctx, est); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		if est.ResponseBy == "" {
			est.ResponseBy = estimatorName
			est.ResponseTime = &now
			if err := s.repo.UpdateEstimasi(ctx, est); err != nil {
				return nil, err
			}
		}
	}

	s.invalidateCache(ctx)
	return s.GetByID(ctx, moodboardID)
}

func (s *moodboardService) UploadEstimasi(ctx context.Context, estimasiID uint, moodboardFileID uint, fileHeader *multipart.FileHeader) (*dto.EstimasiFileResponse, error) {
	// Verify moodboard file exists
	mbFile, err := s.repo.FindFileByID(ctx, moodboardFileID)
	if err != nil {
		return nil, errors.New("moodboard file option tidak valid")
	}

	// Record touch on estimasi stage
	mb, errMb := s.repo.FindByID(ctx, mbFile.MoodboardID)
	if errMb == nil && mb != nil {
		_ = s.logTaskSvc.RecordTouch(ctx, mb.OrderID, "estimasi", "")
	}

	targetDir := filepath.Join(s.uploadDir, "estimasis")
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return nil, err
	}

	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("rab_%d_%d%s", estimasiID, time.Now().UnixNano(), ext)
	dstPath := filepath.Join(targetDir, filename)

	if err := s.saveUploadedFile(fileHeader, dstPath); err != nil {
		return nil, err
	}

	// Check if estimasi file link already exists for this moodboardFileID
	dbFile, err := s.repo.FindEstimasiFileByMoodboardFileID(ctx, moodboardFileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbFile = &entity.EstimasiFile{
				EstimasiID:      estimasiID,
				MoodboardFileID: moodboardFileID,
				FilePath:        fmt.Sprintf("/uploads/estimasis/%s", filename),
				OriginalName:    fileHeader.Filename,
			}
			if err := s.repo.CreateEstimasiFile(ctx, dbFile); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// Overwrite file
		oldLocal := filepath.Join(targetDir, filepath.Base(dbFile.FilePath))
		_ = os.Remove(oldLocal)

		dbFile.FilePath = fmt.Sprintf("/uploads/estimasis/%s", filename)
		dbFile.OriginalName = fileHeader.Filename
		if err := s.repo.UpdateEstimasiFile(ctx, dbFile); err != nil {
			return nil, err
		}
	}

	s.invalidateCache(ctx)
	return &dto.EstimasiFileResponse{
		ID:              dbFile.ID,
		MoodboardFileID: mbFile.ID,
		FilePath:        dbFile.FilePath,
		OriginalName:    dbFile.OriginalName,
		CreatedAt:       dbFile.CreatedAt,
	}, nil
}

// ==========================================
// MODULE C: COMMITMENT FEE
// ==========================================

func (s *moodboardService) ResponseCommitmentFee(ctx context.Context, moodboardID uint, csName string) (*dto.MoodboardResponse, error) {
	m, err := s.repo.FindByID(ctx, moodboardID)
	if err != nil {
		return nil, err
	}

	var responseEnabled string
	s.db.WithContext(ctx).Model(&entity.Setting{}).Where("key = ?", "response_enabled").Pluck("value", &responseEnabled)

	var hasApprovedKasar bool
	for _, f := range m.Files {
		if f.Status == "approved" {
			hasApprovedKasar = true
			break
		}
	}
	if responseEnabled != "false" && !hasApprovedKasar {
		return nil, errors.New("moodboard kasar belum disetujui")
	}

	fee, err := s.repo.FindCommitmentFeeByMoodboardID(ctx, moodboardID)
	now := time.Now()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fee = &entity.CommitmentFee{
				MoodboardID:   moodboardID,
				PaymentStatus: entity.PaymentStatusPending,
				ResponseBy:    csName,
				ResponseTime:  &now,
			}
			if err := s.repo.CreateCommitmentFee(ctx, fee); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		if fee.ResponseBy == "" {
			fee.ResponseBy = csName
			fee.ResponseTime = &now
			if err := s.repo.UpdateCommitmentFee(ctx, fee); err != nil {
				return nil, err
			}
		}
	}

	// Sync Order TahapanProyek to "cm_fee" if not already set and log transition
	if err := s.logTaskSvc.TransitionStage(ctx, m.OrderID, "cm_fee", csName); err != nil {
		s.logger.Error("Failed to update order stage", zap.Error(err))
	}

	s.invalidateCache(ctx)
	return s.GetByID(ctx, moodboardID)
}

func (s *moodboardService) UpdateTotalFee(ctx context.Context, feeID uint, req dto.UpdateTotalFeeRequest) (*dto.MoodboardResponse, error) {
	fee, err := s.repo.FindCommitmentFeeByID(ctx, feeID)
	if err != nil {
		return nil, err
	}

	// Record touch on cm_fee stage
	mb, errMb := s.repo.FindByID(ctx, fee.MoodboardID)
	if errMb == nil && mb != nil {
		_ = s.logTaskSvc.RecordTouch(ctx, mb.OrderID, "cm_fee", "")
	}

	if fee.PaymentStatus == entity.PaymentStatusCompleted {
		return nil, errors.New("tidak dapat mengubah total biaya pada pembayaran yang telah selesai")
	}

	fee.TotalFee = &req.TotalFee
	if err := s.repo.UpdateCommitmentFee(ctx, fee); err != nil {
		return nil, err
	}

	s.invalidateCache(ctx)
	return s.GetByID(ctx, fee.MoodboardID)
}

func (s *moodboardService) UploadPaymentProof(ctx context.Context, feeID uint, fileHeader *multipart.FileHeader) (*dto.MoodboardResponse, error) {
	fee, err := s.repo.FindCommitmentFeeByID(ctx, feeID)
	if err != nil {
		return nil, err
	}

	// Record touch on cm_fee stage
	mb, errMb := s.repo.FindByID(ctx, fee.MoodboardID)
	if errMb == nil && mb != nil {
		_ = s.logTaskSvc.RecordTouch(ctx, mb.OrderID, "cm_fee", "")
	}

	if fee.TotalFee == nil || *fee.TotalFee <= 0 {
		return nil, errors.New("silakan tentukan total biaya commitment fee terlebih dahulu")
	}

	targetDir := filepath.Join(s.uploadDir, "payments")
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return nil, err
	}

	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("proof_%d_%d%s", feeID, time.Now().UnixNano(), ext)
	dstPath := filepath.Join(targetDir, filename)

	if err := s.saveUploadedFile(fileHeader, dstPath); err != nil {
		return nil, err
	}

	if fee.PaymentProof != "" {
		_ = os.Remove(filepath.Join(targetDir, filepath.Base(fee.PaymentProof)))
	}

	fee.PaymentProof = fmt.Sprintf("/uploads/payments/%s", filename)
	if err := s.repo.UpdateCommitmentFee(ctx, fee); err != nil {
		return nil, err
	}

	s.invalidateCache(ctx)
	return s.GetByID(ctx, fee.MoodboardID)
}

func (s *moodboardService) VerifyPayment(ctx context.Context, feeID uint, csName string) (*dto.MoodboardResponse, error) {
	fee, err := s.repo.FindCommitmentFeeByID(ctx, feeID)
	if err != nil {
		return nil, err
	}

	if fee.PaymentProof == "" {
		return nil, errors.New("bukti pembayaran belum diunggah oleh pelanggan")
	}

	now := time.Now()
	fee.PaymentStatus = entity.PaymentStatusCompleted
	fee.PmResponseBy = csName
	fee.PmResponseTime = &now

	if err := s.repo.UpdateCommitmentFee(ctx, fee); err != nil {
		return nil, err
	}

	// Fetch Moodboard & Order to sync Order table and log transition
	m, err := s.repo.FindByID(ctx, fee.MoodboardID)
	if err == nil && m != nil {
		if err := s.logTaskSvc.TransitionStage(ctx, m.OrderID, "desain_final", csName); err != nil {
			s.logger.Error("Failed to update order stage to desain_final", zap.Error(err))
		}
		if err := s.repo.UpdateOrderStageAndPayment(ctx, m.OrderID, "", "cm_fee"); err != nil {
			s.logger.Error("Failed to update order payment status to cm_fee", zap.Error(err))
		}
	}

	s.invalidateCache(ctx)
	return s.GetByID(ctx, fee.MoodboardID)
}

func (s *moodboardService) ResetPayment(ctx context.Context, feeID uint) (*dto.MoodboardResponse, error) {
	fee, err := s.repo.FindCommitmentFeeByID(ctx, feeID)
	if err != nil {
		return nil, err
	}

	if fee.PaymentProof != "" {
		_ = os.Remove(filepath.Join(s.uploadDir, "payments", filepath.Base(fee.PaymentProof)))
	}

	fee.PaymentProof = ""
	fee.TotalFee = nil
	fee.PaymentStatus = entity.PaymentStatusPending
	fee.PmResponseBy = ""
	fee.PmResponseTime = nil

	if err := s.repo.UpdateCommitmentFee(ctx, fee); err != nil {
		return nil, err
	}

	// Reset also payment status on Order
	m, err := s.repo.FindByID(ctx, fee.MoodboardID)
	if err == nil && m != nil {
		_ = s.repo.UpdateOrderStageAndPayment(ctx, m.OrderID, "", "not_start")
	}

	s.invalidateCache(ctx)
	return s.GetByID(ctx, fee.MoodboardID)
}

func (s *moodboardService) RevisePaymentFee(ctx context.Context, feeID uint, req dto.UpdateTotalFeeRequest) (*dto.MoodboardResponse, error) {
	fee, err := s.repo.FindCommitmentFeeByID(ctx, feeID)
	if err != nil {
		return nil, err
	}

	if fee.PaymentStatus == entity.PaymentStatusCompleted {
		return nil, errors.New("tidak dapat merevisi total biaya pada pembayaran yang telah selesai")
	}

	// Reset payment proof if exists when CS decides to revise fee amount
	if fee.PaymentProof != "" {
		_ = os.Remove(filepath.Join(s.uploadDir, "payments", filepath.Base(fee.PaymentProof)))
		fee.PaymentProof = ""
	}

	fee.TotalFee = &req.TotalFee
	if err := s.repo.UpdateCommitmentFee(ctx, fee); err != nil {
		return nil, err
	}

	s.invalidateCache(ctx)
	return s.GetByID(ctx, fee.MoodboardID)
}

func (s *moodboardService) PrintInvoice(ctx context.Context, feeID uint) ([]byte, string, error) {
	fee, err := s.repo.FindCommitmentFeeByID(ctx, feeID)
	if err != nil {
		return nil, "", err
	}

	m, err := s.repo.FindByID(ctx, fee.MoodboardID)
	if err != nil {
		return nil, "", err
	}

	if m.Order == nil {
		return nil, "", errors.New("data order terkait tidak ditemukan")
	}

	cpProfile := entity.GetCompanyProfile(s.db, m.Order.CompanyID)
	pdfCP := pdf.CompanyProfile{
		Name:        cpProfile.Name,
		Director:    cpProfile.Director,
		Logo:        cpProfile.Logo,
		Address:     cpProfile.Address,
		BankName:    cpProfile.BankName,
		BankAccount: cpProfile.BankAccount,
		BankHolder:  cpProfile.BankHolder,
		Email:       cpProfile.Email,
		Phone:       cpProfile.Phone,
	}

	pdfBytes, err := pdf.GenerateCommitmentFeeInvoice(m.Order, m, fee, pdfCP, s.uploadDir)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("invoice_cf_%s.pdf", m.Order.NomorOrder)
	return pdfBytes, filename, nil
}

// ==========================================
// PRIVATE SERVICE HELPERS
// ==========================================

func (s *moodboardService) saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func (s *moodboardService) invalidateCache(ctx context.Context) {
	_ = s.cache.Del(ctx, constants.KeyMoodboardAll)
}

func toMoodboardFileResponse(f entity.MoodboardFile) dto.MoodboardFileResponse {
	return dto.MoodboardFileResponse{
		ID:           f.ID,
		FilePath:     f.FilePath,
		FileType:     "kasar",
		OriginalName: f.OriginalName,
		Status:       f.Status,
		Revisi:       f.Revisi,
		CreatedAt:    f.CreatedAt,
	}
}

func toEstimasiFileResponse(ef entity.EstimasiFile) dto.EstimasiFileResponse {
	return dto.EstimasiFileResponse{
		ID:              ef.ID,
		MoodboardFileID: ef.MoodboardFileID,
		FilePath:        ef.FilePath,
		OriginalName:    ef.OriginalName,
		CreatedAt:       ef.CreatedAt,
	}
}

func toMoodboardResponseEntity(m entity.Moodboard, includeOrder bool) dto.MoodboardResponse {
	overallStatus := "pending"
	var approvedKasarPath string
	var lastKasarRevisionNotes string

	for _, f := range m.Files {
		if f.Status == "approved" {
			overallStatus = "approved"
			approvedKasarPath = f.FilePath
		} else if f.Status == "revisi" {
			if overallStatus != "approved" {
				overallStatus = "revisi"
			}
			lastKasarRevisionNotes = f.Revisi
		}
	}

	resp := dto.MoodboardResponse{
		ID:                  m.ID,
		OrderID:             m.OrderID,
		MoodboardKasar:      approvedKasarPath,
		MoodboardFinal:      "",
		Status:              overallStatus,
		Notes:               lastKasarRevisionNotes,
		RevisiFinal:         "",
		ResponseTime:          m.ResponseTime,
		ResponseBy:            m.ResponseBy,
		MarketingResponse:     m.MarketingResponse,
		MarketingResponseBy:   m.MarketingResponseBy,
		MarketingResponseTime: m.MarketingResponseTime,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}

	if includeOrder && m.Order != nil {
		resp.Order = &dto.OrderBriefResponse{
			ID:                   m.Order.ID,
			NomorOrder:           m.Order.NomorOrder,
			NamaProject:          m.Order.NamaProject,
			NamaCustomer:         m.Order.NamaCustomer,
			NamaPerusahaan:       m.Order.NamaPerusahaan,
			JenisInterior:        m.Order.JenisInterior,
			TanggalMasukCustomer: m.Order.TanggalMasukCustomer,
		}
	}

	if len(m.Files) > 0 {
		resp.Files = make([]dto.MoodboardFileResponse, len(m.Files))
		for i, f := range m.Files {
			resp.Files[i] = toMoodboardFileResponse(f)
		}
	} else {
		resp.Files = []dto.MoodboardFileResponse{}
	}

	if m.Estimasi != nil {
		estRes := dto.EstimasiResponse{
			ID:             m.Estimasi.ID,
			EstimatedCost:  m.Estimasi.EstimatedCost,
			ResponseBy:     m.Estimasi.ResponseBy,
			ResponseTime:   m.Estimasi.ResponseTime,
			PmResponseBy:   m.Estimasi.PmResponseBy,
			PmResponseTime: m.Estimasi.PmResponseTime,
		}

		if len(m.Estimasi.Files) > 0 {
			estRes.Files = make([]dto.EstimasiFileResponse, len(m.Estimasi.Files))
			for i, ef := range m.Estimasi.Files {
				estRes.Files[i] = toEstimasiFileResponse(ef)
			}
		} else {
			estRes.Files = []dto.EstimasiFileResponse{}
		}

		resp.Estimasi = &estRes
	}

	if m.CommitmentFee != nil {
		var totFee *float64
		if m.CommitmentFee.TotalFee != nil {
			val := *m.CommitmentFee.TotalFee
			totFee = &val
		}
		resp.CommitmentFee = &dto.CommitmentFeeResponse{
			ID:             m.CommitmentFee.ID,
			TotalFee:       totFee,
			PaymentProof:   m.CommitmentFee.PaymentProof,
			PaymentStatus:  string(m.CommitmentFee.PaymentStatus),
			ResponseBy:     m.CommitmentFee.ResponseBy,
			ResponseTime:   m.CommitmentFee.ResponseTime,
			PmResponseBy:   m.CommitmentFee.PmResponseBy,
			PmResponseTime: m.CommitmentFee.PmResponseTime,
		}
	}

	return resp
}
