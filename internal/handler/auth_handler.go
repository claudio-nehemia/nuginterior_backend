package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// translateValidationErrors converts Gin/validator errors to Indonesian messages.
func translateValidationErrors(err error) string {
	fieldNames := map[string]string{
		"UserName":       "Nama Admin",
		"UserEmail":      "Email Login",
		"UserPassword":   "Password",
		"CompanyName":    "Nama Perusahaan",
		"DirectorName":   "Nama Direktur",
		"CeoNik":         "NIK CEO",
		"Nib":            "NIB",
		"CompanyEmail":   "Email Perusahaan",
		"CompanyPhone":   "Telepon Perusahaan",
		"CompanyAddress": "Alamat Perusahaan",
	}

	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}

	var msgs []string
	for _, fe := range ve {
		name := fe.Field()
		if friendly, exists := fieldNames[name]; exists {
			name = friendly
		}
		switch fe.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s wajib diisi", name))
		case "email":
			msgs = append(msgs, fmt.Sprintf("%s harus berupa email yang valid", name))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s minimal %s karakter", name, fe.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s maksimal %s karakter", name, fe.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s tidak valid", name))
		}
	}
	return strings.Join(msgs, "; ")
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		helper.Unauthorized(c, err.Error())
		return
	}

	helper.OK(c, "Login berhasil", token)
}

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	user, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		helper.BadRequest(c, err.Error(), nil)
		return
	}

	helper.Created(c, "Registrasi berhasil", user)
}

// RegisterCompany handles POST /api/auth/register-company
func (h *AuthHandler) RegisterCompany(c *gin.Context) {
	var req dto.RegisterCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		msg := translateValidationErrors(err)
		helper.BadRequest(c, msg, nil)
		return
	}

	user, err := h.svc.RegisterCompany(c.Request.Context(), req)
	if err != nil {
		helper.BadRequest(c, err.Error(), nil)
		return
	}

	helper.Created(c, "Pendaftaran perusahaan berhasil, silakan tunggu verifikasi oleh Super Admin", user)
}

// Me handles GET /api/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.svc.Me(c.Request.Context(), userID.(uint))
	if err != nil {
		helper.NotFound(c, "User tidak ditemukan")
		return
	}
	helper.OK(c, "Data user", user)
}

// Logout handles POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	jti, _ := c.Get("jti")
	tokenExp, _ := c.Get("token_exp")
	expTime := time.Unix(tokenExp.(int64), 0)

	if err := h.svc.Logout(c.Request.Context(), jti.(string), expTime); err != nil {
		helper.InternalError(c, "Gagal logout")
		return
	}

	helper.OK(c, "Logout berhasil", nil)
}

// RefreshToken handles POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	token, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		helper.Unauthorized(c, err.Error())
		return
	}

	helper.OK(c, "Token berhasil di-refresh", token)
}
