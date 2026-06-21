package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error)
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserProfileResponse, error)
	RegisterCompany(ctx context.Context, req dto.RegisterCompanyRequest) (*dto.UserProfileResponse, error)
	Me(ctx context.Context, userID uint) (*dto.UserProfileResponse, error)
	Logout(ctx context.Context, tokenJTI string, expiresAt time.Time) error
	RefreshToken(ctx context.Context, refreshTokenStr string) (*dto.TokenResponse, error)
	IsTokenBlacklisted(ctx context.Context, jti string) bool
	GetPermissionsByRoleID(ctx context.Context, roleID uint) ([]string, error)
}

type authService struct {
	config   *config.Config
	authRepo repository.AuthRepository
	roleRepo repository.RoleRepository
	cache    cache.Store
	logger   *zap.Logger
}

func NewAuthService(cfg *config.Config, authRepo repository.AuthRepository, roleRepo repository.RoleRepository, cacheStore cache.Store, logger *zap.Logger) AuthService {
	return &authService{
		config:   cfg,
		authRepo: authRepo,
		roleRepo: roleRepo,
		cache:    cacheStore,
		logger:   logger,
	}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	user, err := s.authRepo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("email atau password salah")
		}
		return nil, err
	}

	if !helper.CheckPassword(req.Password, user.Password) {
		return nil, fmt.Errorf("email atau password salah")
	}

	// Check Company Status
	company, err := s.authRepo.FindCompanyByID(ctx, user.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat data perusahaan")
	}

	if company.Status == "pending" {
		return nil, fmt.Errorf("Pendaftaran perusahaan Anda sedang menunggu verifikasi oleh Super Admin")
	} else if company.Status == "rejected" {
		return nil, fmt.Errorf("Pendaftaran perusahaan Anda ditolak oleh Super Admin")
	}

	return s.generateTokens(user)
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.UserProfileResponse, error) {
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

	if err := s.authRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Reload with relations
	created, err := s.authRepo.FindUserByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return s.toUserProfile(ctx, created), nil
}

func (s *authService) RegisterCompany(ctx context.Context, req dto.RegisterCompanyRequest) (*dto.UserProfileResponse, error) {
	_, err := s.authRepo.FindUserByEmail(ctx, req.UserEmail)
	if err == nil {
		return nil, fmt.Errorf("email pendaftar sudah terdaftar")
	}

	hash, err := helper.HashPassword(req.UserPassword)
	if err != nil {
		return nil, err
	}

	company := &entity.Company{
		Name:         req.CompanyName,
		DirectorName: req.DirectorName,
		CeoNik:       req.CeoNik,
		Nib:          req.Nib,
		Email:        req.CompanyEmail,
		Phone:        req.CompanyPhone,
		Address:      req.CompanyAddress,
		Status:       "pending",
	}

	user := &entity.User{
		Name:     req.UserName,
		Email:    req.UserEmail,
		Password: hash,
	}

	if err := s.authRepo.CreateCompanyAndUser(ctx, company, user); err != nil {
		return nil, err
	}

	// Reload with relations
	created, err := s.authRepo.FindUserByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return s.toUserProfile(ctx, created), nil
}

func (s *authService) Me(ctx context.Context, userID uint) (*dto.UserProfileResponse, error) {
	user, err := s.authRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.toUserProfile(ctx, user), nil
}

func (s *authService) Logout(ctx context.Context, tokenJTI string, expiresAt time.Time) error {
	remaining := time.Until(expiresAt)
	if remaining <= 0 {
		return nil
	}
	key := constants.KeyJWTBlacklist + tokenJTI
	return s.cache.Set(ctx, key, "1", remaining)
}

func (s *authService) RefreshToken(ctx context.Context, refreshTokenStr string) (*dto.TokenResponse, error) {
	claims, err := s.parseToken(refreshTokenStr)
	if err != nil {
		return nil, fmt.Errorf("refresh token tidak valid")
	}

	if claims["type"] != "refresh" {
		return nil, fmt.Errorf("token bukan refresh token")
	}

	jti, _ := claims["jti"].(string)
	if s.IsTokenBlacklisted(ctx, jti) {
		return nil, fmt.Errorf("token sudah di-blacklist")
	}

	userIDFloat, _ := claims["user_id"].(float64)
	userID := uint(userIDFloat)

	user, err := s.authRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Blacklist old refresh token
	exp, _ := claims["exp"].(float64)
	expTime := time.Unix(int64(exp), 0)
	_ = s.Logout(ctx, jti, expTime)

	return s.generateTokens(user)
}

func (s *authService) IsTokenBlacklisted(ctx context.Context, jti string) bool {
	key := constants.KeyJWTBlacklist + jti
	exists, _ := s.cache.Exists(ctx, key)
	return exists
}

func (s *authService) GetPermissionsByRoleID(ctx context.Context, roleID uint) ([]string, error) {
	key := fmt.Sprintf("%s%d", constants.KeyRolePermissions, roleID)

	var perms []string
	err := s.cache.GetJSON(ctx, key, &perms)
	if err == nil {
		return perms, nil
	}

	// Cache miss — fetch from DB
	permissions, err := s.roleRepo.GetPermissionsByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	perms = make([]string, len(permissions))
	for i, p := range permissions {
		perms[i] = p.Name
	}

	_ = s.cache.SetJSON(ctx, key, perms, 30*time.Minute)
	return perms, nil
}

func (s *authService) generateTokens(user *entity.User) (*dto.TokenResponse, error) {
	now := time.Now()
	accessExp := now.Add(s.config.AccessTokenDuration())
	refreshExp := now.Add(s.config.RefreshTokenDuration())

	accessJTI := uuid.New().String()
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.NamaRole
	}

	accessClaims := jwt.MapClaims{
		"user_id":    user.ID,
		"email":      user.Email,
		"company_id": user.CompanyID,
		"role_name":  roleName,
		"jti":        accessJTI,
		"type":       "access",
		"iat":        now.Unix(),
		"exp":        accessExp.Unix(),
	}
	if user.RoleID != nil {
		accessClaims["role_id"] = *user.RoleID
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshJTI := uuid.New().String()
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"jti":     refreshJTI,
		"type":    "refresh",
		"iat":     now.Unix(),
		"exp":     refreshExp.Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		TokenType:    "Bearer",
		ExpiresAt:    accessExp,
	}, nil
}

func (s *authService) parseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (s *authService) toUserProfile(ctx context.Context, user *entity.User) *dto.UserProfileResponse {
	resp := &dto.UserProfileResponse{
		ID:              user.ID,
		CompanyID:       user.CompanyID,
		Name:            user.Name,
		Email:           user.Email,
		EmailVerifiedAt: user.EmailVerifiedAt,
		RoleID:          user.RoleID,
		Permissions:     []string{},
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
	if user.Company != nil {
		resp.Company = &dto.CompanySimple{
			ID:     user.Company.ID,
			Name:   user.Company.Name,
			Status: user.Company.Status,
		}
	}
	if user.Role != nil {
		resp.Role = &dto.RoleSimple{
			ID:       user.Role.ID,
			NamaRole: user.Role.NamaRole,
			DivisiID: user.Role.DivisiID,
		}
		perms, err := s.GetPermissionsByRoleID(ctx, user.Role.ID)
		if err == nil {
			resp.Permissions = perms
		}
	}
	return resp
}
