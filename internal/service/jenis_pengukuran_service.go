package service

import (
	"context"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"go.uber.org/zap"
)

type JenisPengukuranService interface {
	GetAll(ctx context.Context) ([]dto.JenisPengukuranResponse, error)
	Create(ctx context.Context, req dto.CreateJenisPengukuranRequest) (*dto.JenisPengukuranResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateJenisPengukuranRequest) (*dto.JenisPengukuranResponse, error)
	Delete(ctx context.Context, id uint) error
}

type jenisPengukuranService struct {
	repo   repository.JenisPengukuranRepository
	cache  cache.Store
	logger *zap.Logger
}

func NewJenisPengukuranService(repo repository.JenisPengukuranRepository, cacheStore cache.Store, logger *zap.Logger) JenisPengukuranService {
	return &jenisPengukuranService{repo: repo, cache: cacheStore, logger: logger}
}

func (s *jenisPengukuranService) GetAll(ctx context.Context) ([]dto.JenisPengukuranResponse, error) {
	var result []dto.JenisPengukuranResponse
	err := s.cache.GetJSON(ctx, constants.KeyJenisPengukuran, &result)
	if err == nil {
		return result, nil
	}

	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result = make([]dto.JenisPengukuranResponse, len(list))
	for i, jp := range list {
		result[i] = toJenisPengukuranResponse(jp)
	}

	_ = s.cache.SetJSON(ctx, constants.KeyJenisPengukuran, result, 30*time.Minute)
	return result, nil
}

func (s *jenisPengukuranService) Create(ctx context.Context, req dto.CreateJenisPengukuranRequest) (*dto.JenisPengukuranResponse, error) {
	jp := &entity.JenisPengukuran{NamaPengukuran: req.NamaPengukuran}
	if err := s.repo.Create(ctx, jp); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, constants.KeyJenisPengukuran)
	resp := toJenisPengukuranResponse(*jp)
	return &resp, nil
}

func (s *jenisPengukuranService) Update(ctx context.Context, id uint, req dto.UpdateJenisPengukuranRequest) (*dto.JenisPengukuranResponse, error) {
	jp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	jp.NamaPengukuran = req.NamaPengukuran
	if err := s.repo.Update(ctx, jp); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, constants.KeyJenisPengukuran)
	resp := toJenisPengukuranResponse(*jp)
	return &resp, nil
}

func (s *jenisPengukuranService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return err
	}
	_ = s.cache.Del(ctx, constants.KeyJenisPengukuran)
	return nil
}

func toJenisPengukuranResponse(jp entity.JenisPengukuran) dto.JenisPengukuranResponse {
	return dto.JenisPengukuranResponse{
		ID:             jp.ID,
		NamaPengukuran: jp.NamaPengukuran,
		CreatedAt:      jp.CreatedAt,
		UpdatedAt:      jp.UpdatedAt,
	}
}
