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

type BahanBakuService interface {
	GetAll(ctx context.Context) ([]dto.BahanBakuResponse, error)
	Create(ctx context.Context, req dto.CreateBahanBakuRequest) (*dto.BahanBakuResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateBahanBakuRequest) (*dto.BahanBakuResponse, error)
	Delete(ctx context.Context, id uint) error
}

type bahanBakuService struct {
	repo   repository.BahanBakuRepository
	cache  cache.Store
	logger *zap.Logger
}

func NewBahanBakuService(repo repository.BahanBakuRepository, cacheStore cache.Store, logger *zap.Logger) BahanBakuService {
	return &bahanBakuService{repo: repo, cache: cacheStore, logger: logger}
}

func (s *bahanBakuService) GetAll(ctx context.Context) ([]dto.BahanBakuResponse, error) {
	var result []dto.BahanBakuResponse
	err := s.cache.GetJSON(ctx, constants.KeyBahanBakuAll, &result)
	if err == nil {
		return result, nil
	}

	list, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result = make([]dto.BahanBakuResponse, len(list))
	for i, bb := range list {
		result[i] = toBahanBakuResponse(bb)
	}

	_ = s.cache.SetJSON(ctx, constants.KeyBahanBakuAll, result, 15*time.Minute)
	return result, nil
}

func (s *bahanBakuService) Create(ctx context.Context, req dto.CreateBahanBakuRequest) (*dto.BahanBakuResponse, error) {
	bb := &entity.BahanBaku{NamaBahanBaku: req.NamaBahanBaku}
	if err := s.repo.Create(ctx, bb); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, constants.KeyBahanBakuAll)
	resp := toBahanBakuResponse(*bb)
	return &resp, nil
}

func (s *bahanBakuService) Update(ctx context.Context, id uint, req dto.UpdateBahanBakuRequest) (*dto.BahanBakuResponse, error) {
	bb, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	bb.NamaBahanBaku = req.NamaBahanBaku
	if err := s.repo.Update(ctx, bb); err != nil {
		return nil, err
	}
	_ = s.cache.Del(ctx, constants.KeyBahanBakuAll)
	resp := toBahanBakuResponse(*bb)
	return &resp, nil
}

func (s *bahanBakuService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.cache.Del(ctx, constants.KeyBahanBakuAll)
	return nil
}

func toBahanBakuResponse(bb entity.BahanBaku) dto.BahanBakuResponse {
	return dto.BahanBakuResponse{
		ID:            bb.ID,
		NamaBahanBaku: bb.NamaBahanBaku,
		CreatedAt:     bb.CreatedAt,
		UpdatedAt:     bb.UpdatedAt,
	}
}
