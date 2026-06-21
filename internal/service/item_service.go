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

type ItemService interface {
	GetAll(ctx context.Context, jenis string) ([]dto.ItemResponse, error)
	Create(ctx context.Context, req dto.CreateItemRequest) (*dto.ItemResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateItemRequest) (*dto.ItemResponse, error)
	Delete(ctx context.Context, id uint) error
}

type itemService struct {
	repo   repository.ItemRepository
	cache  cache.Store
	logger *zap.Logger
}

func NewItemService(repo repository.ItemRepository, cacheStore cache.Store, logger *zap.Logger) ItemService {
	return &itemService{repo: repo, cache: cacheStore, logger: logger}
}

func (s *itemService) GetAll(ctx context.Context, jenis string) ([]dto.ItemResponse, error) {
	if jenis != "" {
		cacheKey := constants.KeyItemsJenis + jenis
		var result []dto.ItemResponse
		err := s.cache.GetJSON(ctx, cacheKey, &result)
		if err == nil {
			return result, nil
		}
	}

	items, err := s.repo.FindAll(ctx, jenis)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ItemResponse, len(items))
	for i, item := range items {
		result[i] = toItemResponse(item)
	}

	if jenis != "" {
		cacheKey := constants.KeyItemsJenis + jenis
		_ = s.cache.SetJSON(ctx, cacheKey, result, 15*time.Minute)
	}

	return result, nil
}

func (s *itemService) Create(ctx context.Context, req dto.CreateItemRequest) (*dto.ItemResponse, error) {
	item := &entity.Item{
		NamaItem:  req.NamaItem,
		JenisItem: req.JenisItem,
		Harga:     req.Harga,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	s.invalidateItemCache(ctx)
	resp := toItemResponse(*item)
	return &resp, nil
}

func (s *itemService) Update(ctx context.Context, id uint, req dto.UpdateItemRequest) (*dto.ItemResponse, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	item.NamaItem = req.NamaItem
	item.JenisItem = req.JenisItem
	item.Harga = req.Harga
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	s.invalidateItemCache(ctx)
	resp := toItemResponse(*item)
	return &resp, nil
}

func (s *itemService) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.invalidateItemCache(ctx)
	return nil
}

func (s *itemService) invalidateItemCache(ctx context.Context) {
	_ = s.cache.Del(ctx,
		constants.KeyItemsJenis+entity.JenisFinishingDalam,
		constants.KeyItemsJenis+entity.JenisFinishingLuar,
		constants.KeyItemsJenis+entity.JenisAksesoris,
	)
}

func toItemResponse(item entity.Item) dto.ItemResponse {
	return dto.ItemResponse{
		ID:        item.ID,
		NamaItem:  item.NamaItem,
		JenisItem: item.JenisItem,
		Harga:     item.Harga,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
