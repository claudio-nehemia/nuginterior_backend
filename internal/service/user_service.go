package service

import (
	"context"
	"fmt"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"go.uber.org/zap"
)

type UserService interface {
	GetAll(ctx context.Context, search string, roleID uint) ([]dto.UserResponse, error)
	Create(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	Delete(ctx context.Context, id uint) error
}

type userService struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

func NewUserService(repo repository.UserRepository, logger *zap.Logger) UserService {
	return &userService{repo: repo, logger: logger}
}

func (s *userService) GetAll(ctx context.Context, search string, roleID uint) ([]dto.UserResponse, error) {
	users, err := s.repo.FindAll(ctx, search, roleID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.UserResponse, len(users))
	for i, u := range users {
		result[i] = toUserResponse(u)
	}
	return result, nil
}

func (s *userService) Create(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("email sudah digunakan")
	}

	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hash,
		RoleID:   req.RoleID,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	created, err := s.repo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	resp := toUserResponse(*created)
	return &resp, nil
}

func (s *userService) Update(ctx context.Context, id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Email != "" && req.Email != user.Email {
		exists, err := s.repo.ExistsByEmail(ctx, req.Email, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("email sudah digunakan")
		}
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Password != "" {
		hash, err := helper.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hash
	}
	if req.RoleID != nil {
		user.RoleID = req.RoleID
		// Clear preloaded association to ensure Gorm updates the foreign key
		user.Role = nil 
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	resp := toUserResponse(*updated)
	return &resp, nil
}

func (s *userService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func toUserResponse(u entity.User) dto.UserResponse {
	resp := dto.UserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		EmailVerifiedAt: u.EmailVerifiedAt,
		RoleID:          u.RoleID,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
	if u.Role != nil {
		resp.Role = &dto.RoleSimple{
			ID:       u.Role.ID,
			NamaRole: u.Role.NamaRole,
			DivisiID: u.Role.DivisiID,
		}
		if u.Role.Divisi.ID != 0 {
			resp.DivisiName = u.Role.Divisi.NamaDivisi
		}
	}
	return resp
}
