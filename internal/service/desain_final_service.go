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

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DesainFinalService interface {
	GetAll(ctx context.Context) ([]dto.DesainFinalResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.DesainFinalResponse, error)
	GetByOrderID(ctx context.Context, orderID uint) (*dto.DesainFinalResponse, error)
	Response(ctx context.Context, orderID uint, designerName string) (*dto.DesainFinalResponse, error)
	Upload(ctx context.Context, orderID uint, files []*multipart.FileHeader) ([]dto.DesainFinalFileResponse, error)
	Accept(ctx context.Context, desainFinalID uint, fileID uint, pmName string) (*dto.DesainFinalResponse, error)
	Revise(ctx context.Context, desainFinalID uint, fileID uint, notes string, pmName string) (*dto.DesainFinalResponse, error)
	DeleteFile(ctx context.Context, fileID uint) error
}

type desainFinalService struct {
	repo          repository.DesainFinalRepository
	moodboardRepo repository.MoodboardRepository
	cache         cache.Store
	logger        *zap.Logger
	uploadDir     string
	logTaskSvc    ProjectLogTaskService
	db            *gorm.DB
}

func NewDesainFinalService(
	repo repository.DesainFinalRepository,
	moodboardRepo repository.MoodboardRepository,
	cache cache.Store,
	logger *zap.Logger,
	uploadDir string,
	logTaskSvc ProjectLogTaskService,
	db *gorm.DB,
) DesainFinalService {
	return &desainFinalService{
		repo:          repo,
		moodboardRepo: moodboardRepo,
		cache:         cache,
		logger:        logger,
		uploadDir:     uploadDir,
		logTaskSvc:    logTaskSvc,
		db:            db,
	}
}

func (s *desainFinalService) checkCommitmentFee(ctx context.Context, orderID uint) error {
	var responseEnabled string
	s.db.WithContext(ctx).Model(&entity.Setting{}).Where("key = ?", "response_enabled").Pluck("value", &responseEnabled)
	if responseEnabled == "false" {
		return nil
	}

	m, err := s.moodboardRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("desain final terkunci! Selesaikan pembayaran commitment fee terlebih dahulu")
		}
		return err
	}

	if m.CommitmentFee == nil || m.CommitmentFee.PaymentStatus != "completed" {
		return errors.New("desain final terkunci! Selesaikan pembayaran commitment fee terlebih dahulu")
	}

	return nil
}

func (s *desainFinalService) GetAll(ctx context.Context) ([]dto.DesainFinalResponse, error) {
	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := []dto.DesainFinalResponse{}
	for _, df := range list {
		// Verify if commitment fee is paid for this order (only list valid ones)
		if err := s.checkCommitmentFee(ctx, df.OrderID); err == nil {
			result = append(result, *toDesainFinalResponse(df, true))
		}
	}
	return result, nil
}

func (s *desainFinalService) GetByID(ctx context.Context, id uint) (*dto.DesainFinalResponse, error) {
	df, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDesainFinalResponse(*df, true), nil
}

func (s *desainFinalService) GetByOrderID(ctx context.Context, orderID uint) (*dto.DesainFinalResponse, error) {
	df, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If not found but commitment fee is paid, let's create a pending DesainFinal record
			if errFee := s.checkCommitmentFee(ctx, orderID); errFee != nil {
				return nil, errFee
			}
			newDF := &entity.DesainFinal{
				OrderID: orderID,
				Status:  "pending",
			}
			if errCreate := s.repo.Create(ctx, newDF); errCreate != nil {
				return nil, errCreate
			}
			return s.GetByOrderID(ctx, orderID)
		}
		return nil, err
	}
	return toDesainFinalResponse(*df, true), nil
}

func (s *desainFinalService) Response(ctx context.Context, orderID uint, designerName string) (*dto.DesainFinalResponse, error) {
	if err := s.checkCommitmentFee(ctx, orderID); err != nil {
		return nil, err
	}

	df, err := s.repo.FindByOrderID(ctx, orderID)
	now := time.Now()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			df = &entity.DesainFinal{
				OrderID:      orderID,
				ResponseBy:   designerName,
				ResponseTime: &now,
				Status:       "pending",
			}
			if err := s.repo.Create(ctx, df); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		if df.ResponseBy == "" || df.ResponseBy == "CS / Designer / Estimator" {
			df.ResponseBy = designerName
			df.ResponseTime = &now
			if err := s.repo.Update(ctx, df); err != nil {
				return nil, err
			}
		}
	}

	// Update order stage to "desain_final" and log transition
	if err := s.logTaskSvc.TransitionStage(ctx, orderID, "desain_final", designerName); err != nil {
		s.logger.Error("Failed to update order stage to desain_final", zap.Error(err))
	}

	return s.GetByOrderID(ctx, orderID)
}

func (s *desainFinalService) Upload(ctx context.Context, orderID uint, files []*multipart.FileHeader) ([]dto.DesainFinalFileResponse, error) {
	if err := s.checkCommitmentFee(ctx, orderID); err != nil {
		return nil, err
	}

	df, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Record touch on desain_final stage
	_ = s.logTaskSvc.RecordTouch(ctx, orderID, "desain_final", "")

	targetDir := filepath.Join(s.uploadDir, "desain_finals")
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return nil, err
	}

	var uploaded []dto.DesainFinalFileResponse

	for _, fileHeader := range files {
		ext := filepath.Ext(fileHeader.Filename)
		filename := fmt.Sprintf("final_%d_%d%s", df.ID, time.Now().UnixNano(), ext)
		dstPath := filepath.Join(targetDir, filename)

		if err := s.saveUploadedFile(fileHeader, dstPath); err != nil {
			return nil, err
		}

		dbFile := &entity.DesainFinalFile{
			DesainFinalID: df.ID,
			FilePath:      fmt.Sprintf("/uploads/desain_finals/%s", filename),
			OriginalName:  fileHeader.Filename,
			Status:        "pending",
		}

		if err := s.repo.CreateFile(ctx, dbFile); err != nil {
			return nil, err
		}

		uploaded = append(uploaded, *toDesainFinalFileResponse(*dbFile))
	}

	// Update overall DesainFinal status to uploaded
	df.Status = "uploaded"
	if err := s.repo.Update(ctx, df); err != nil {
		s.logger.Error("Failed to update DesainFinal status to uploaded", zap.Error(err))
	}

	return uploaded, nil
}

func (s *desainFinalService) Accept(ctx context.Context, desainFinalID uint, fileID uint, pmName string) (*dto.DesainFinalResponse, error) {
	df, err := s.repo.FindByID(ctx, desainFinalID)
	if err != nil {
		return nil, err
	}

	if err := s.checkCommitmentFee(ctx, df.OrderID); err != nil {
		return nil, err
	}

	var targetFile *entity.DesainFinalFile
	for i := range df.Files {
		if df.Files[i].ID == fileID {
			targetFile = &df.Files[i]
			break
		}
	}
	if targetFile == nil {
		return nil, errors.New("file desain final tidak ditemukan")
	}

	// Mark all other files as pending/ignored and this one as approved
	for i := range df.Files {
		if df.Files[i].ID == fileID {
			df.Files[i].Status = "approved"
			if err := s.repo.UpdateFile(ctx, &df.Files[i]); err != nil {
				return nil, err
			}
		} else if df.Files[i].Status == "approved" {
			df.Files[i].Status = "pending"
			if err := s.repo.UpdateFile(ctx, &df.Files[i]); err != nil {
				return nil, err
			}
		}
	}

	// Update overall DesainFinal status to accepted
	df.Status = "accepted"
	now := time.Now()
	df.MarketingResponseBy = pmName
	df.MarketingResponseTime = &now
	if err := s.repo.Update(ctx, df); err != nil {
		return nil, err
	}

	// Automatically transition order stage to "input_item" and log transition
	if err := s.logTaskSvc.TransitionStage(ctx, df.OrderID, "input_item", pmName); err != nil {
		s.logger.Error("Failed to transition order stage to input_item", zap.Error(err))
	}

	return s.GetByID(ctx, desainFinalID)
}

func (s *desainFinalService) Revise(ctx context.Context, desainFinalID uint, fileID uint, notes string, pmName string) (*dto.DesainFinalResponse, error) {
	df, err := s.repo.FindByID(ctx, desainFinalID)
	if err != nil {
		return nil, err
	}

	var targetFile *entity.DesainFinalFile
	for i := range df.Files {
		if df.Files[i].ID == fileID {
			targetFile = &df.Files[i]
			break
		}
	}
	if targetFile == nil {
		return nil, errors.New("file opsi desain final tidak ditemukan")
	}

	targetFile.Status = "revisi"
	targetFile.Revisi = notes
	if err := s.repo.UpdateFile(ctx, targetFile); err != nil {
		return nil, err
	}

	// Update overall DesainFinal status to revision
	df.Status = "revision"
	now := time.Now()
	df.MarketingResponseBy = pmName
	df.MarketingResponseTime = &now
	if err := s.repo.Update(ctx, df); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, desainFinalID)
}

func (s *desainFinalService) DeleteFile(ctx context.Context, fileID uint) error {
	file, err := s.repo.FindFileByID(ctx, fileID)
	if err != nil {
		return err
	}

	// Delete from storage
	filePath := filepath.Join(s.uploadDir, "..", file.FilePath)
	if err := os.Remove(filePath); err != nil {
		s.logger.Warn("Failed to delete physical file", zap.String("path", filePath), zap.Error(err))
	}

	return s.repo.DeleteFile(ctx, fileID)
}

func (s *desainFinalService) saveUploadedFile(file *multipart.FileHeader, dst string) error {
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

func toDesainFinalFileResponse(f entity.DesainFinalFile) *dto.DesainFinalFileResponse {
	return &dto.DesainFinalFileResponse{
		ID:            f.ID,
		DesainFinalID: f.DesainFinalID,
		FilePath:      f.FilePath,
		OriginalName:  f.OriginalName,
		Status:        f.Status,
		Revisi:        f.Revisi,
		CreatedAt:     f.CreatedAt,
	}
}

func toDesainFinalResponse(df entity.DesainFinal, includeOrder bool) *dto.DesainFinalResponse {
	filesResp := make([]dto.DesainFinalFileResponse, len(df.Files))
	for i, f := range df.Files {
		filesResp[i] = *toDesainFinalFileResponse(f)
	}

	resp := &dto.DesainFinalResponse{
		ID:                    df.ID,
		OrderID:               df.OrderID,
		Status:                df.Status,
		ResponseTime:          df.ResponseTime,
		ResponseBy:            df.ResponseBy,
		MarketingResponseTime: df.MarketingResponseTime,
		MarketingResponseBy:   df.MarketingResponseBy,
		CreatedAt:             df.CreatedAt,
		UpdatedAt:             df.UpdatedAt,
		Files:                 filesResp,
	}

	if includeOrder && df.Order != nil {
		resp.Order = &dto.OrderBriefResponse{
			ID:                   df.Order.ID,
			NomorOrder:           df.Order.NomorOrder,
			NamaProject:          df.Order.NamaProject,
			NamaCustomer:         df.Order.NamaCustomer,
			NamaPerusahaan:       df.Order.NamaPerusahaan,
			JenisInterior:        df.Order.JenisInterior,
			TanggalMasukCustomer: df.Order.TanggalMasukCustomer,
		}
	}

	return resp
}

