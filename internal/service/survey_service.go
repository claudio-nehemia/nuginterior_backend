package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type SurveyService interface {
	GetAll(ctx context.Context) ([]dto.SurveyResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.SurveyResponse, error)
	Create(ctx context.Context, req dto.CreateSurveyRequest) (*dto.SurveyResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateSurveyRequest) (*dto.SurveyResponse, error)
	Delete(ctx context.Context, id uint) error
	Response(ctx context.Context, id uint, userName string) (*dto.SurveyResponse, error)
	MarketingResponse(ctx context.Context, id uint, userName string) (*dto.SurveyResponse, error)
}

type surveyService struct {
	repo            repository.SurveyRepository
	logger          *zap.Logger
	logTaskSvc      ProjectLogTaskService
	notificationSvc NotificationService
}

func NewSurveyService(repo repository.SurveyRepository, logger *zap.Logger, logTaskSvc ProjectLogTaskService, notificationSvc NotificationService) SurveyService {
	return &surveyService{
		repo:            repo,
		logger:          logger,
		logTaskSvc:      logTaskSvc,
		notificationSvc: notificationSvc,
	}
}

func (s *surveyService) GetAll(ctx context.Context) ([]dto.SurveyResponse, error) {
	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.SurveyResponse, len(list))
	for i, survey := range list {
		result[i] = s.toSurveyResponseEntity(ctx, survey, true)
	}
	return result, nil
}

func (s *surveyService) GetByID(ctx context.Context, id uint) (*dto.SurveyResponse, error) {
	survey, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := s.toSurveyResponseEntity(ctx, *survey, true)
	return &resp, nil
}

func (s *surveyService) Create(ctx context.Context, req dto.CreateSurveyRequest) (*dto.SurveyResponse, error) {
	tanggal, err := parseTanggalSurvey(req.TanggalSurvey)
	if err != nil {
		return nil, err
	}
	tanggalUlang, err := parseTanggalSurvey(req.TanggalSurveyUlang)
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = "pending"
	}

	survey := &entity.Survey{
		OrderID:            req.OrderID,
		TanggalSurvey:      tanggal,
		Lokasi:             req.Lokasi,
		Catatan:            req.Catatan,
		Status:             status,
		SurveyorID:         req.SurveyorID,
		LayoutFiles:        datatypes.JSON(req.LayoutFiles),
		FotoLokasi:         datatypes.JSON(req.FotoLokasi),
		MoMFile:            req.MoMFile,
		MomFiles:           datatypes.JSON(req.MomFiles),
		TanggalSurveyUlang: tanggalUlang,
		SurveyUlangTeamIDs: datatypes.JSON(req.SurveyUlangTeamIDs),
		CatatanUlang:       req.CatatanUlang,
		TemuanLapangan:     datatypes.JSON(req.TemuanLapangan),
		FotoVideoUlang:     datatypes.JSON(req.FotoVideoUlang),
	}
	if err := s.repo.Create(ctx, survey); err != nil {
		return nil, err
	}

	// Record touch on survey stage
	_ = s.logTaskSvc.RecordTouch(ctx, survey.OrderID, "survey", "")

	if err := s.repo.SyncPengukuran(ctx, survey.ID, toSurveyPengukuranEntities(req.Pengukuran)); err != nil {
		return nil, err
	}

	created, err := s.repo.FindByID(ctx, survey.ID)
	if err != nil {
		return nil, err
	}
	resp := s.toSurveyResponseEntity(ctx, *created, true)
	return &resp, nil
}

func (s *surveyService) Update(ctx context.Context, id uint, req dto.UpdateSurveyRequest) (*dto.SurveyResponse, error) {
	survey, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tanggal, err := parseTanggalSurvey(req.TanggalSurvey)
	if err != nil {
		return nil, err
	}
	tanggalUlang, err := parseTanggalSurvey(req.TanggalSurveyUlang)
	if err != nil {
		return nil, err
	}

	var oldTeamIDs []uint
	if len(survey.SurveyUlangTeamIDs) > 0 {
		_ = json.Unmarshal(survey.SurveyUlangTeamIDs, &oldTeamIDs)
	}

	if req.Status != "" {
		survey.Status = req.Status
	}
	survey.OrderID = req.OrderID
	survey.TanggalSurvey = tanggal
	survey.Lokasi = req.Lokasi
	survey.Catatan = req.Catatan
	survey.SurveyorID = req.SurveyorID
	survey.LayoutFiles = datatypes.JSON(req.LayoutFiles)
	survey.FotoLokasi = datatypes.JSON(req.FotoLokasi)
	survey.MoMFile = req.MoMFile
	survey.MomFiles = datatypes.JSON(req.MomFiles)
	survey.TanggalSurveyUlang = tanggalUlang
	survey.SurveyUlangTeamIDs = datatypes.JSON(req.SurveyUlangTeamIDs)
	survey.CatatanUlang = req.CatatanUlang
	survey.TemuanLapangan = datatypes.JSON(req.TemuanLapangan)
	survey.FotoVideoUlang = datatypes.JSON(req.FotoVideoUlang)

	// Detect if survey ulang team changed or was newly set up
	var newTeamIDs []uint
	if len(req.SurveyUlangTeamIDs) > 0 {
		_ = json.Unmarshal(req.SurveyUlangTeamIDs, &newTeamIDs)
	}

	isTeamSetupOrUpdated := false
	if len(newTeamIDs) > 0 {
		if len(oldTeamIDs) != len(newTeamIDs) {
			isTeamSetupOrUpdated = true
		} else {
			oldMap := make(map[uint]bool)
			for _, id := range oldTeamIDs {
				oldMap[id] = true
			}
			for _, id := range newTeamIDs {
				if !oldMap[id] {
					isTeamSetupOrUpdated = true
					break
				}
			}
		}
	}

	if isTeamSetupOrUpdated {
		_ = s.notificationSvc.SendNotification(
			ctx,
			survey.OrderID,
			"upload_survey_ulang",
			"Tugas Upload Hasil Survey Ulang",
			"Anda ditugaskan dalam tim survey ulang. Mohon segera lakukan survey ulang dan upload hasil survey lapangan.",
			fmt.Sprintf("/dashboard/survey/%d", id),
		)
	}

	currentStage := "survey"
	if survey.Order != nil && survey.Order.TahapanProyek != "" {
		currentStage = survey.Order.TahapanProyek
	}
	_ = s.logTaskSvc.RecordTouch(ctx, survey.OrderID, currentStage, "")

	if err := s.repo.Update(ctx, survey); err != nil {
		return nil, err
	}

	if err := s.repo.SyncPengukuran(ctx, survey.ID, toSurveyPengukuranEntities(req.Pengukuran)); err != nil {
		return nil, err
	}

	// Transition to gambar_kerja if survey ulang is filled
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

	if hasPhotos || hasTemuan || hasCatatan {
		if errStage := s.logTaskSvc.TransitionStage(ctx, survey.OrderID, "gambar_kerja", ""); errStage != nil {
			s.logger.Error("Failed to update order stage to gambar_kerja", zap.Error(errStage))
		}
	}

	updated, err := s.repo.FindByID(ctx, survey.ID)
	if err != nil {
		return nil, err
	}
	resp := s.toSurveyResponseEntity(ctx, *updated, true)
	return &resp, nil
}

func (s *surveyService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *surveyService) Response(ctx context.Context, id uint, userName string) (*dto.SurveyResponse, error) {
	survey, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = s.logTaskSvc.RecordTouch(ctx, survey.OrderID, "survey", userName)

	now := time.Now()
	name := userName
	survey.ResponseBy = &name
	survey.ResponseTime = &now
	if err := s.repo.Update(ctx, survey); err != nil {
		return nil, err
	}
	companyID := uint(1)
	if survey.Order != nil {
		companyID = survey.Order.CompanyID
	}
	// Transition to moodboard if marketing response is disabled
	if enabled, errSetting := s.repo.IsMarketingResponseEnabled(ctx, companyID); errSetting == nil && !enabled {
		if errStage := s.logTaskSvc.TransitionStage(ctx, survey.OrderID, "moodboard", userName); errStage != nil {
			s.logger.Error("Failed to update order stage to moodboard", zap.Error(errStage))
		}
	}
	resp := s.toSurveyResponseEntity(ctx, *survey, true)
	return &resp, nil
}

func (s *surveyService) MarketingResponse(ctx context.Context, id uint, userName string) (*dto.SurveyResponse, error) {
	survey, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = s.logTaskSvc.RecordTouch(ctx, survey.OrderID, "survey", userName)

	now := time.Now()
	name := userName
	survey.MarketingResponseBy = &name
	survey.MarketingResponseTime = &now
	if err := s.repo.Update(ctx, survey); err != nil {
		return nil, err
	}
	// Transition order stage to moodboard
	if errStage := s.logTaskSvc.TransitionStage(ctx, survey.OrderID, "moodboard", userName); errStage != nil {
		s.logger.Error("Failed to update order stage to moodboard", zap.Error(errStage))
	}
	resp := s.toSurveyResponseEntity(ctx, *survey, true)
	return &resp, nil
}
func parseTanggalSurvey(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func toSurveyPengukuranEntities(inputs []dto.PengukuranInput) []entity.SurveyPengukuran {
	result := make([]entity.SurveyPengukuran, len(inputs))
	for i, input := range inputs {
		result[i] = entity.SurveyPengukuran{
			JenisPengukuranID: input.JenisPengukuranID,
			NamaCustom:        input.NamaCustom,
			Checked:           input.Checked,
			Notes:             input.Notes,
			Panjang:           input.Panjang,
			Lebar:             input.Lebar,
			Tinggi:            input.Tinggi,
			HasLebar:          input.HasLebar,
			HasTinggi:         input.HasTinggi,
		}
	}
	return result
}

func (s *surveyService) toSurveyResponseEntity(ctx context.Context, survey entity.Survey, includeOrder bool) dto.SurveyResponse {
	resp := dto.SurveyResponse{
		ID:                    survey.ID,
		OrderID:               survey.OrderID,
		TanggalSurvey:         survey.TanggalSurvey,
		Lokasi:                survey.Lokasi,
		Catatan:               survey.Catatan,
		Status:                survey.Status,
		SurveyorID:            survey.SurveyorID,
		ResponseBy:            survey.ResponseBy,
		ResponseTime:          survey.ResponseTime,
		MarketingResponseBy:   survey.MarketingResponseBy,
		MarketingResponseTime: survey.MarketingResponseTime,
		LayoutFiles:           json.RawMessage(survey.LayoutFiles),
		FotoLokasi:            json.RawMessage(survey.FotoLokasi),
		MoMFile:               survey.MoMFile,
		MomFiles:              json.RawMessage(survey.MomFiles),
		TanggalSurveyUlang:    survey.TanggalSurveyUlang,
		SurveyUlangTeamIDs:    json.RawMessage(survey.SurveyUlangTeamIDs),
		CatatanUlang:          survey.CatatanUlang,
		TemuanLapangan:        json.RawMessage(survey.TemuanLapangan),
		FotoVideoUlang:        json.RawMessage(survey.FotoVideoUlang),
		CreatedAt:             survey.CreatedAt,
		UpdatedAt:             survey.UpdatedAt,
	}

	isContractDeal := false
	if includeOrder && survey.Order != nil {
		for _, c := range survey.Order.Contracts {
			if c.Status == "deal" {
				isContractDeal = true
				break
			}
		}
	}
	resp.IsContractDeal = isContractDeal

	if survey.Surveyor != nil {
		roleName := ""
		if survey.Surveyor.Role != nil {
			roleName = survey.Surveyor.Role.NamaRole
		}
		resp.Surveyor = &dto.SurveyUserResponse{
			ID:    survey.Surveyor.ID,
			Name:  survey.Surveyor.Name,
			Email: survey.Surveyor.Email,
			Role:  roleName,
		}
	}

	var teamIDs []uint
	if len(survey.SurveyUlangTeamIDs) > 0 {
		_ = json.Unmarshal(survey.SurveyUlangTeamIDs, &teamIDs)
	}
	if len(teamIDs) > 0 {
		users, err := s.repo.FindUsersByIDs(ctx, teamIDs)
		if err == nil && len(users) > 0 {
			resp.SurveyUlangTeam = make([]dto.SurveyUserResponse, len(users))
			for idx, u := range users {
				roleName := ""
				if u.Role != nil {
					roleName = u.Role.NamaRole
				}
				resp.SurveyUlangTeam[idx] = dto.SurveyUserResponse{
					ID:    u.ID,
					Name:  u.Name,
					Email: u.Email,
					Role:  roleName,
				}
			}
		}
	}

	if includeOrder && survey.Order != nil {
		resp.Order = &dto.OrderBriefResponse{
			ID:                   survey.Order.ID,
			NomorOrder:           survey.Order.NomorOrder,
			NamaProject:          survey.Order.NamaProject,
			NamaCustomer:         survey.Order.NamaCustomer,
			NamaPerusahaan:       survey.Order.NamaPerusahaan,
			JenisInterior:        survey.Order.JenisInterior,
			TanggalMasukCustomer: survey.Order.TanggalMasukCustomer,
		}
		if len(survey.Order.Teams) > 0 {
			resp.Order.Teams = make([]dto.OrderTeamResponse, len(survey.Order.Teams))
			for i, t := range survey.Order.Teams {
				resp.Order.Teams[i] = dto.OrderTeamResponse{
					ID:     t.ID,
					UserID: t.UserID,
				}
				if t.User != nil {
					resp.Order.Teams[i].Name = t.User.Name
					resp.Order.Teams[i].Email = t.User.Email
					if t.User.Role != nil {
						resp.Order.Teams[i].Role = t.User.Role.NamaRole
					}
				}
			}
		}
	}

	if len(survey.SurveyPengukuran) > 0 {
		resp.Pengukuran = make([]dto.PengukuranResponse, len(survey.SurveyPengukuran))
		for i, p := range survey.SurveyPengukuran {
			resp.Pengukuran[i] = dto.PengukuranResponse{
				ID:                p.ID,
				JenisPengukuranID: p.JenisPengukuranID,
				NamaPengukuran:    getSurveyPengukuranName(p.JenisPengukuran, p.NamaCustom),
				NamaCustom:        p.NamaCustom,
				Checked:           p.Checked,
				Notes:             p.Notes,
				Panjang:           p.Panjang,
				Lebar:             p.Lebar,
				Tinggi:            p.Tinggi,
				HasLebar:          p.HasLebar,
				HasTinggi:         p.HasTinggi,
			}
		}
	}

	return resp
}

func getSurveyPengukuranName(jp *entity.JenisPengukuran, customName string) string {
	if jp == nil {
		return customName
	}
	return jp.NamaPengukuran
}
