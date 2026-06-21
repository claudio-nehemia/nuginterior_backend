package service

import (
	"context"
	"strings"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"go.uber.org/zap"
)

type SettingService interface {
	GetAll(ctx context.Context) ([]dto.SettingResponse, error)
	GetByKey(ctx context.Context, key string) (*dto.SettingResponse, error)
	Update(ctx context.Context, key string, req dto.UpdateSettingRequest) (*dto.SettingResponse, error)
	IsEnabled(ctx context.Context, key string) (bool, error)
}

type settingService struct {
	repo   repository.SettingRepository
	cache  cache.Store
	logger *zap.Logger
}

func NewSettingService(repo repository.SettingRepository, cacheStore cache.Store, logger *zap.Logger) SettingService {
	return &settingService{repo: repo, cache: cacheStore, logger: logger}
}

func (s *settingService) GetAll(ctx context.Context) ([]dto.SettingResponse, error) {
	var result []dto.SettingResponse
	err := s.cache.GetJSON(ctx, constants.KeySettingAll, &result)
	if err == nil {
		return result, nil
	}

	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result = make([]dto.SettingResponse, len(list))
	for i, setting := range list {
		result[i] = toSettingResponse(setting)
	}

	_ = s.cache.SetJSON(ctx, constants.KeySettingAll, result, 30*time.Minute)
	return result, nil
}

func (s *settingService) GetByKey(ctx context.Context, key string) (*dto.SettingResponse, error) {
	setting, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	resp := toSettingResponse(*setting)
	return &resp, nil
}

func (s *settingService) Update(ctx context.Context, key string, req dto.UpdateSettingRequest) (*dto.SettingResponse, error) {
	setting := &entity.Setting{
		Key:   key,
		Value: req.Value,
	}
	if err := s.repo.Upsert(ctx, setting); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, constants.KeySettingAll)
	resp := toSettingResponse(*setting)
	return &resp, nil
}

func (s *settingService) IsEnabled(ctx context.Context, key string) (bool, error) {
	setting, err := s.repo.FindByKey(ctx, key)
	if err != nil {
		return false, err
	}
	return strings.ToLower(setting.Value) == "true", nil
}

func toSettingResponse(setting entity.Setting) dto.SettingResponse {
	return dto.SettingResponse{
		ID:          setting.ID,
		Key:         setting.Key,
		Value:       setting.Value,
		Description: setting.Description,
		CreatedAt:   setting.CreatedAt,
		UpdatedAt:   setting.UpdatedAt,
	}
}
