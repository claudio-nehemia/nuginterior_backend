package service

import (
	"context"
	"fmt"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"go.uber.org/zap"
)

type RoleService interface {
	GetAll(ctx context.Context) ([]dto.RoleResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.RoleDetailResponse, error)
	Create(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateRoleRequest) (*dto.RoleResponse, error)
	Delete(ctx context.Context, id uint) error
	SyncPermissions(ctx context.Context, roleID uint, req dto.SyncPermissionsRequest) error
}

type roleService struct {
	repo   repository.RoleRepository
	cache  cache.Store
	logger *zap.Logger
}

func NewRoleService(repo repository.RoleRepository, cacheStore cache.Store, logger *zap.Logger) RoleService {
	return &roleService{repo: repo, cache: cacheStore, logger: logger}
}

func (s *roleService) GetAll(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.RoleResponse, len(roles))
	for i, r := range roles {
		usersCount, _ := s.repo.CountUsersByRoleID(ctx, r.ID)
		permsCount, _ := s.repo.CountPermissionsByRoleID(ctx, r.ID)
		result[i] = dto.RoleResponse{
			ID:               r.ID,
			NamaRole:         r.NamaRole,
			DivisiID:         r.DivisiID,
			UsersCount:       usersCount,
			PermissionsCount: permsCount,
			CreatedAt:        r.CreatedAt,
			UpdatedAt:        r.UpdatedAt,
		}
		if r.Divisi.ID != 0 {
			d := toDivisiResponse(r.Divisi)
			result[i].Divisi = &d
		}
	}
	return result, nil
}

func (s *roleService) GetByID(ctx context.Context, id uint) (*dto.RoleDetailResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &dto.RoleDetailResponse{
		ID:        role.ID,
		NamaRole:  role.NamaRole,
		DivisiID:  role.DivisiID,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}

	if role.Divisi.ID != 0 {
		d := toDivisiResponse(role.Divisi)
		resp.Divisi = &d
	}

	// Map users
	users := make([]dto.UserResponse, len(role.Users))
	for i, u := range role.Users {
		users[i] = toUserResponse(u)
	}
	resp.Users = users

	// Map permissions and group them
	perms := make([]dto.PermissionResponse, len(role.Permissions))
	grouped := make(map[string][]dto.PermissionResponse)
	for i, p := range role.Permissions {
		pr := toPermissionResponse(p)
		perms[i] = pr
		grouped[p.Group] = append(grouped[p.Group], pr)
	}
	resp.Permissions = perms
	resp.PermissionsGrouped = grouped

	return resp, nil
}

func (s *roleService) Create(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	role := &entity.Role{
		NamaRole: req.NamaRole,
		DivisiID: req.DivisiID,
	}
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}

	if len(req.Permissions) > 0 {
		if err := s.repo.SyncPermissions(ctx, role.ID, req.Permissions); err != nil {
			return nil, err
		}
		s.invalidatePermCache(ctx, role.ID)
	}

	usersCount, _ := s.repo.CountUsersByRoleID(ctx, role.ID)
	permsCount, _ := s.repo.CountPermissionsByRoleID(ctx, role.ID)

	resp := dto.RoleResponse{
		ID:               role.ID,
		NamaRole:         role.NamaRole,
		DivisiID:         role.DivisiID,
		UsersCount:       usersCount,
		PermissionsCount: permsCount,
		CreatedAt:        role.CreatedAt,
		UpdatedAt:        role.UpdatedAt,
	}
	return &resp, nil
}

func (s *roleService) Update(ctx context.Context, id uint, req dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	role.NamaRole = req.NamaRole
	role.DivisiID = req.DivisiID
	// Clear preloaded association to ensure Gorm updates the foreign key
	role.Divisi = entity.Divisi{} 
	
	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}

	if req.Permissions != nil {
		if err := s.repo.SyncPermissions(ctx, role.ID, req.Permissions); err != nil {
			return nil, err
		}
		s.invalidatePermCache(ctx, role.ID)
	}

	usersCount, _ := s.repo.CountUsersByRoleID(ctx, role.ID)
	permsCount, _ := s.repo.CountPermissionsByRoleID(ctx, role.ID)

	resp := dto.RoleResponse{
		ID:               role.ID,
		NamaRole:         role.NamaRole,
		DivisiID:         role.DivisiID,
		UsersCount:       usersCount,
		PermissionsCount: permsCount,
		CreatedAt:        role.CreatedAt,
		UpdatedAt:        role.UpdatedAt,
	}
	return &resp, nil
}

func (s *roleService) Delete(ctx context.Context, id uint) error {
	count, err := s.repo.CountUsersByRoleID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("tidak bisa menghapus role yang masih memiliki %d user", count)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.invalidatePermCache(ctx, id)
	return nil
}

func (s *roleService) SyncPermissions(ctx context.Context, roleID uint, req dto.SyncPermissionsRequest) error {
	if err := s.repo.SyncPermissions(ctx, roleID, req.Permissions); err != nil {
		return err
	}
	s.invalidatePermCache(ctx, roleID)
	return nil
}

func (s *roleService) invalidatePermCache(ctx context.Context, roleID uint) {
	key := fmt.Sprintf("%s%d", constants.KeyRolePermissions, roleID)
	_ = s.cache.Del(ctx, key)
	_ = s.cache.Del(ctx, constants.KeyPermissionsAll)
}

func toPermissionResponse(p entity.Permission) dto.PermissionResponse {
	return dto.PermissionResponse{
		ID:          p.ID,
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Group:       p.Group,
	}
}
