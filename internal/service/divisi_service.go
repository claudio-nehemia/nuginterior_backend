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

type DivisiService interface {
	GetAll(ctx context.Context) ([]dto.DivisiResponse, error)
	Create(ctx context.Context, req dto.CreateDivisiRequest) (*dto.DivisiResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateDivisiRequest) (*dto.DivisiResponse, error)
	Delete(ctx context.Context, id uint) error
}

type divisiService struct {
	repo   repository.DivisiRepository
	cache  cache.Store
	logger *zap.Logger
}

func NewDivisiService(repo repository.DivisiRepository, cacheStore cache.Store, logger *zap.Logger) DivisiService {
	return &divisiService{repo: repo, cache: cacheStore, logger: logger}
}

func (s *divisiService) GetAll(ctx context.Context) ([]dto.DivisiResponse, error) {
	var result []dto.DivisiResponse
	err := s.cache.GetJSON(ctx, constants.KeyDivisiAll, &result)
	if err == nil {
		return result, nil
	}

	divisis, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result = make([]dto.DivisiResponse, len(divisis))
	for i, d := range divisis {
		result[i] = toDivisiResponse(d)
	}

	_ = s.cache.SetJSON(ctx, constants.KeyDivisiAll, result, 30*time.Minute)
	return result, nil
}

func (s *divisiService) Create(ctx context.Context, req dto.CreateDivisiRequest) (*dto.DivisiResponse, error) {
	divisi := &entity.Divisi{NamaDivisi: req.NamaDivisi}
	if err := s.repo.Create(ctx, divisi); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, constants.KeyDivisiAll)
	resp := toDivisiResponse(*divisi)
	return &resp, nil
}

func (s *divisiService) Update(ctx context.Context, id uint, req dto.UpdateDivisiRequest) (*dto.DivisiResponse, error) {
	divisi, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	divisi.NamaDivisi = req.NamaDivisi
	if err := s.repo.Update(ctx, divisi); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, constants.KeyDivisiAll)
	resp := toDivisiResponse(*divisi)
	return &resp, nil
}

func (s *divisiService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.cache.Del(ctx, constants.KeyDivisiAll)
	return nil
}

func toDivisiResponse(d entity.Divisi) dto.DivisiResponse {
	return dto.DivisiResponse{
		ID:         d.ID,
		NamaDivisi: d.NamaDivisi,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
}
