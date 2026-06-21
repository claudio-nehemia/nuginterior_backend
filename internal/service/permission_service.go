package service

import (
	"context"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"go.uber.org/zap"
)

type PermissionService interface {
	GetAllGrouped(ctx context.Context) (*dto.PermissionGroupedResponse, error)
}

type permissionService struct {
	repo   repository.PermissionRepository
	cache  cache.Store
	logger *zap.Logger
}

func NewPermissionService(repo repository.PermissionRepository, cacheStore cache.Store, logger *zap.Logger) PermissionService {
	return &permissionService{repo: repo, cache: cacheStore, logger: logger}
}

func (s *permissionService) GetAllGrouped(ctx context.Context) (*dto.PermissionGroupedResponse, error) {
	var result dto.PermissionGroupedResponse
	err := s.cache.GetJSON(ctx, constants.KeyPermissionsAll, &result)
	if err == nil {
		return &result, nil
	}

	perms, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	groups := make(map[string][]dto.PermissionResponse)
	for _, p := range perms {
		pr := toPermissionResponse(p)
		groups[p.Group] = append(groups[p.Group], pr)
	}

	result = dto.PermissionGroupedResponse{Groups: groups}
	_ = s.cache.SetJSON(ctx, constants.KeyPermissionsAll, result, 1*time.Hour)
	return &result, nil
}
