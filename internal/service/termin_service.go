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

type TerminService interface {
	GetAll(ctx context.Context) ([]dto.TerminResponse, error)
	Create(ctx context.Context, req dto.CreateTerminRequest) (*dto.TerminResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateTerminRequest) (*dto.TerminResponse, error)
	Delete(ctx context.Context, id uint) error
}

type terminService struct {
	repo   repository.TerminRepository
	cache  cache.Store
	logger *zap.Logger
}

func NewTerminService(repo repository.TerminRepository, cacheStore cache.Store, logger *zap.Logger) TerminService {
	return &terminService{repo: repo, cache: cacheStore, logger: logger}
}

func (s *terminService) GetAll(ctx context.Context) ([]dto.TerminResponse, error) {
	var result []dto.TerminResponse
	err := s.cache.GetJSON(ctx, constants.KeyTerminAll, &result)
	if err == nil {
		return result, nil
	}

	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result = make([]dto.TerminResponse, len(list))
	for i, t := range list {
		result[i] = toTerminResponse(t)
	}

	_ = s.cache.SetJSON(ctx, constants.KeyTerminAll, result, 30*time.Minute)
	return result, nil
}

func (s *terminService) Create(ctx context.Context, req dto.CreateTerminRequest) (*dto.TerminResponse, error) {
	termin := &entity.Termin{
		KodeTipe:  req.KodeTipe,
		NamaTipe:  req.NamaTipe,
		Deskripsi: req.Deskripsi,
		Tahapan:   dto.ToTahapanEntity(req.Tahapan),
	}
	if err := s.repo.Create(ctx, termin); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, constants.KeyTerminAll)
	resp := toTerminResponse(*termin)
	return &resp, nil
}

func (s *terminService) Update(ctx context.Context, id uint, req dto.UpdateTerminRequest) (*dto.TerminResponse, error) {
	termin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	termin.KodeTipe = req.KodeTipe
	termin.NamaTipe = req.NamaTipe
	termin.Deskripsi = req.Deskripsi
	termin.Tahapan = dto.ToTahapanEntity(req.Tahapan)
	if err := s.repo.Update(ctx, termin); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, constants.KeyTerminAll)
	resp := toTerminResponse(*termin)
	return &resp, nil
}

func (s *terminService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.cache.Del(ctx, constants.KeyTerminAll)
	return nil
}

func toTerminResponse(t entity.Termin) dto.TerminResponse {
	tahapan := make([]dto.TahapanResponse, len(t.Tahapan))
	for i, th := range t.Tahapan {
		tahapan[i] = dto.TahapanResponse{
			Step:       th.Step,
			Text:       th.Text,
			Persentase: th.Persentase,
		}
	}
	return dto.TerminResponse{
		ID:        t.ID,
		KodeTipe:  t.KodeTipe,
		NamaTipe:  t.NamaTipe,
		Deskripsi: t.Deskripsi,
		Tahapan:   tahapan,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
