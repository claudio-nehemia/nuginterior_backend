package service

import (
	"context"
	"encoding/json"
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
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GambarKerjaService interface {
	GetAll(ctx context.Context) ([]dto.GambarKerjaResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.GambarKerjaResponse, error)
	GetByOrderID(ctx context.Context, orderID uint) (*dto.GambarKerjaResponse, error)
	Response(ctx context.Context, orderID uint, email string) (*dto.GambarKerjaResponse, error)
	Upload(ctx context.Context, orderID uint, files []*multipart.FileHeader, userID uint) ([]dto.GambarKerjaFileResponse, error)
	ReviseFile(ctx context.Context, fileID uint, notes string) error
	ReviseGeneral(ctx context.Context, gkID uint, notes string, email string) (*dto.GambarKerjaResponse, error)
	Approve(ctx context.Context, gkID uint, email string) (*dto.GambarKerjaResponse, error)
	DeleteFile(ctx context.Context, fileID uint) error
}

type gambarKerjaService struct {
	repo       repository.GambarKerjaRepository
	userRepo   repository.UserRepository
	db         *gorm.DB
	logger     *zap.Logger
	uploadDir  string
	logTaskSvc ProjectLogTaskService
}

func NewGambarKerjaService(
	repo repository.GambarKerjaRepository,
	userRepo repository.UserRepository,
	db *gorm.DB,
	logger *zap.Logger,
	uploadDir string,
	logTaskSvc ProjectLogTaskService,
) GambarKerjaService {
	return &gambarKerjaService{
		repo:       repo,
		userRepo:   userRepo,
		db:         db,
		logger:     logger,
		uploadDir:  uploadDir,
		logTaskSvc: logTaskSvc,
	}
}

func (s *gambarKerjaService) checkResurveyUploaded(ctx context.Context, orderID uint) error {
	var survey entity.Survey
	err := s.db.WithContext(ctx).Where("order_id = ?", orderID).First(&survey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("gambar kerja terkunci! Selesaikan survey ulang terlebih dahulu")
		}
		return err
	}

	if survey.TanggalSurveyUlang == nil {
		return errors.New("gambar kerja terkunci! Jadwal survey ulang belum ditentukan")
	}

	hasPhotos := false
	if len(survey.FotoVideoUlang) > 0 {
		var photos []string
		if err := json.Unmarshal(survey.FotoVideoUlang, &photos); err == nil && len(photos) > 0 {
			hasPhotos = true
		}
	}
	hasTemuan := false
	if len(survey.TemuanLapangan) > 0 {
		var temuan []interface{}
		if err := json.Unmarshal(survey.TemuanLapangan, &temuan); err == nil && len(temuan) > 0 {
			hasTemuan = true
		}
	}
	hasCatatan := survey.CatatanUlang != ""

	if !(hasPhotos || hasTemuan || hasCatatan) {
		return errors.New("gambar kerja terkunci! Hasil survey ulang belum diunggah")
	}

	return nil
}

func (s *gambarKerjaService) verifyUploadAccess(ctx context.Context, orderID uint, currentUserID uint) error {
	user, err := s.userRepo.FindByID(ctx, currentUserID)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan: %w", err)
	}

	if user.Role != nil && user.Role.NamaRole == "Super Admin" {
		return nil
	}

	var survey entity.Survey
	err = s.db.WithContext(ctx).Where("order_id = ?", orderID).First(&survey).Error
	if err != nil {
		return errors.New("data survey tidak ditemukan")
	}

	var teamIDs []uint
	if len(survey.SurveyUlangTeamIDs) > 0 {
		_ = json.Unmarshal(survey.SurveyUlangTeamIDs, &teamIDs)
	}

	for _, id := range teamIDs {
		if id == currentUserID {
			return nil
		}
	}

	return errors.New("akses ditolak! Hanya drafter, desainer, PM di tim survey ulang, atau Super Admin yang dapat mengunggah")
}

func (s *gambarKerjaService) GetAll(ctx context.Context) ([]dto.GambarKerjaResponse, error) {
	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := []dto.GambarKerjaResponse{}
	for _, gk := range list {
		if err := s.checkResurveyUploaded(ctx, gk.OrderID); err == nil {
			result = append(result, *toGambarKerjaResponse(gk, true))
		}
	}
	return result, nil
}

func (s *gambarKerjaService) GetByID(ctx context.Context, id uint) (*dto.GambarKerjaResponse, error) {
	gk, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toGambarKerjaResponse(*gk, true), nil
}

func (s *gambarKerjaService) GetByOrderID(ctx context.Context, orderID uint) (*dto.GambarKerjaResponse, error) {
	gk, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if errRes := s.checkResurveyUploaded(ctx, orderID); errRes != nil {
				return nil, errRes
			}
			newGK := &entity.GambarKerja{
				OrderID: orderID,
				Status:  "pending",
			}
			if errCreate := s.repo.Create(ctx, newGK); errCreate != nil {
				return nil, errCreate
			}
			return s.GetByOrderID(ctx, orderID)
		}
		return nil, err
	}
	return toGambarKerjaResponse(*gk, true), nil
}

func (s *gambarKerjaService) Response(ctx context.Context, orderID uint, email string) (*dto.GambarKerjaResponse, error) {
	if err := s.checkResurveyUploaded(ctx, orderID); err != nil {
		return nil, err
	}

	gk, err := s.repo.FindByOrderID(ctx, orderID)
	now := time.Now()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			gk = &entity.GambarKerja{
				OrderID:      orderID,
				ResponseBy:   email,
				ResponseTime: &now,
				Status:       "pending",
			}
			if err := s.repo.Create(ctx, gk); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		if gk.ResponseBy == "" {
			gk.ResponseBy = email
			gk.ResponseTime = &now
			if err := s.repo.Update(ctx, gk); err != nil {
				return nil, err
			}
		}
	}

	if err := s.logTaskSvc.TransitionStage(ctx, orderID, "gambar_kerja", email); err != nil {
		s.logger.Error("Failed to update order stage to gambar_kerja", zap.Error(err))
	}

	return s.GetByOrderID(ctx, orderID)
}

func (s *gambarKerjaService) Upload(ctx context.Context, orderID uint, files []*multipart.FileHeader, userID uint) ([]dto.GambarKerjaFileResponse, error) {
	if err := s.checkResurveyUploaded(ctx, orderID); err != nil {
		return nil, err
	}

	if err := s.verifyUploadAccess(ctx, orderID, userID); err != nil {
		return nil, err
	}

	gk, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Record touch on gambar_kerja stage
	_ = s.logTaskSvc.RecordTouch(ctx, orderID, "gambar_kerja", "")

	targetDir := filepath.Join(s.uploadDir, "gambar_kerja")
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return nil, err
	}

	var uploaded []dto.GambarKerjaFileResponse

	for _, fileHeader := range files {
		ext := filepath.Ext(fileHeader.Filename)
		filename := fmt.Sprintf("working_%d_%d%s", gk.ID, time.Now().UnixNano(), ext)
		dstPath := filepath.Join(targetDir, filename)

		if err := s.saveUploadedFile(fileHeader, dstPath); err != nil {
			return nil, err
		}

		dbFile := &entity.GambarKerjaFile{
			GambarKerjaID: gk.ID,
			FilePath:      fmt.Sprintf("/uploads/gambar_kerja/%s", filename),
			OriginalName:  fileHeader.Filename,
			Status:        "pending",
		}

		if err := s.repo.CreateFile(ctx, dbFile); err != nil {
			return nil, err
		}

		uploaded = append(uploaded, *toGambarKerjaFileResponse(*dbFile))
	}

	gk.Status = "uploaded"
	if err := s.repo.Update(ctx, gk); err != nil {
		s.logger.Error("Failed to update GambarKerja status to uploaded", zap.Error(err))
	}

	return uploaded, nil
}

func (s *gambarKerjaService) ReviseFile(ctx context.Context, fileID uint, notes string) error {
	file, err := s.repo.FindFileByID(ctx, fileID)
	if err != nil {
		return err
	}

	file.Status = "revisi"
	file.Revisi = notes
	return s.repo.UpdateFile(ctx, file)
}

func (s *gambarKerjaService) ReviseGeneral(ctx context.Context, gkID uint, notes string, email string) (*dto.GambarKerjaResponse, error) {
	gk, err := s.repo.FindByID(ctx, gkID)
	if err != nil {
		return nil, err
	}

	gk.Status = "revisi"
	gk.RevisiGeneral = notes
	now := time.Now()
	gk.MarketingResponseBy = email
	gk.MarketingResponseTime = &now

	if err := s.repo.Update(ctx, gk); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, gkID)
}

func (s *gambarKerjaService) Approve(ctx context.Context, gkID uint, email string) (*dto.GambarKerjaResponse, error) {
	gk, err := s.repo.FindByID(ctx, gkID)
	if err != nil {
		return nil, err
	}

	for i := range gk.Files {
		if gk.Files[i].Status != "approved" {
			gk.Files[i].Status = "approved"
			if err := s.repo.UpdateFile(ctx, &gk.Files[i]); err != nil {
				return nil, err
			}
		}
	}

	gk.Status = "approved"
	now := time.Now()
	gk.MarketingResponseBy = email
	gk.MarketingResponseTime = &now

	if err := s.repo.Update(ctx, gk); err != nil {
		return nil, err
	}

	// Transition stage to approval_material and log transition
	if errStage := s.logTaskSvc.TransitionStage(ctx, gk.OrderID, "approval_material", email); errStage != nil {
		s.logger.Error("Failed to transition order stage to approval_material", zap.Error(errStage))
	}

	return s.GetByID(ctx, gkID)
}

func (s *gambarKerjaService) DeleteFile(ctx context.Context, fileID uint) error {
	file, err := s.repo.FindFileByID(ctx, fileID)
	if err != nil {
		return err
	}

	filePath := filepath.Join(s.uploadDir, "..", file.FilePath)
	if err := os.Remove(filePath); err != nil {
		s.logger.Warn("Failed to delete physical file", zap.String("path", filePath), zap.Error(err))
	}

	return s.repo.DeleteFile(ctx, fileID)
}

func (s *gambarKerjaService) saveUploadedFile(file *multipart.FileHeader, dst string) error {
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

func toGambarKerjaFileResponse(f entity.GambarKerjaFile) *dto.GambarKerjaFileResponse {
	return &dto.GambarKerjaFileResponse{
		ID:            f.ID,
		GambarKerjaID: f.GambarKerjaID,
		FilePath:      f.FilePath,
		OriginalName:  f.OriginalName,
		Status:        f.Status,
		Revisi:        f.Revisi,
		CreatedAt:     f.CreatedAt,
	}
}

func toGambarKerjaResponse(gk entity.GambarKerja, includeOrder bool) *dto.GambarKerjaResponse {
	filesResp := make([]dto.GambarKerjaFileResponse, len(gk.Files))
	for i, f := range gk.Files {
		filesResp[i] = *toGambarKerjaFileResponse(f)
	}

	resp := &dto.GambarKerjaResponse{
		ID:                    gk.ID,
		OrderID:               gk.OrderID,
		Status:                gk.Status,
		ResponseBy:            gk.ResponseBy,
		ResponseTime:          gk.ResponseTime,
		MarketingResponseBy:   gk.MarketingResponseBy,
		MarketingResponseTime: gk.MarketingResponseTime,
		RevisiGeneral:         gk.RevisiGeneral,
		CreatedAt:             gk.CreatedAt,
		UpdatedAt:             gk.UpdatedAt,
		Files:                 filesResp,
	}

	if includeOrder && gk.Order != nil {
		resp.Order = &dto.OrderBriefResponse{
			ID:                   gk.Order.ID,
			NomorOrder:           gk.Order.NomorOrder,
			NamaProject:          gk.Order.NamaProject,
			NamaCustomer:         gk.Order.NamaCustomer,
			NamaPerusahaan:       gk.Order.NamaPerusahaan,
			JenisInterior:        gk.Order.JenisInterior,
			TanggalMasukCustomer: gk.Order.TanggalMasukCustomer,
		}
	}

	return resp
}
