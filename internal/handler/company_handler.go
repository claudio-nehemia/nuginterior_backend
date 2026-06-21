package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	svc service.CompanyService
}

func NewCompanyHandler(svc service.CompanyService) *CompanyHandler {
	return &CompanyHandler{svc: svc}
}

// Index handles GET /api/companies
func (h *CompanyHandler) Index(c *gin.Context) {
	companyID, _ := c.Get("company_id")
	roleName, _ := c.Get("role_name")
	if companyID.(uint) != 1 || roleName.(string) != "Super Admin" {
		helper.Forbidden(c, "Hanya Super Admin yang dapat mengakses semua data perusahaan")
		return
	}

	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat daftar perusahaan")
		return
	}
	helper.OK(c, "Daftar perusahaan", data)
}

// Show handles GET /api/companies/:id
func (h *CompanyHandler) Show(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	companyID, _ := c.Get("company_id")
	roleName, _ := c.Get("role_name")
	if companyID.(uint) != id && (companyID.(uint) != 1 || roleName.(string) != "Super Admin") {
		helper.Forbidden(c, "Akses ditolak")
		return
	}

	data, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		helper.NotFound(c, "Perusahaan tidak ditemukan")
		return
	}
	helper.OK(c, "Detail perusahaan", data)
}

// Update handles PUT /api/companies/:id
func (h *CompanyHandler) Update(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	companyID, _ := c.Get("company_id")
	roleName, _ := c.Get("role_name")
	if companyID.(uint) != id && (companyID.(uint) != 1 || roleName.(string) != "Super Admin") {
		helper.Forbidden(c, "Akses ditolak untuk memperbarui profil perusahaan ini")
		return
	}

	var req dto.CompanyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}

	data, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		helper.BadRequest(c, err.Error(), nil)
		return
	}
	helper.OK(c, "Profil perusahaan berhasil diperbarui", data)
}

// Verify handles PUT /api/companies/:id/verify
func (h *CompanyHandler) Verify(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	companyID, _ := c.Get("company_id")
	roleName, _ := c.Get("role_name")
	if companyID.(uint) != 1 || roleName.(string) != "Super Admin" {
		helper.Forbidden(c, "Hanya Super Admin yang dapat memverifikasi perusahaan")
		return
	}

	if err := h.svc.Verify(c.Request.Context(), id); err != nil {
		helper.BadRequest(c, err.Error(), nil)
		return
	}
	helper.OK(c, "Perusahaan berhasil diverifikasi", nil)
}

// Reject handles PUT /api/companies/:id/reject
func (h *CompanyHandler) Reject(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}

	companyID, _ := c.Get("company_id")
	roleName, _ := c.Get("role_name")
	if companyID.(uint) != 1 || roleName.(string) != "Super Admin" {
		helper.Forbidden(c, "Hanya Super Admin yang dapat menolak pendaftaran perusahaan")
		return
	}

	if err := h.svc.Reject(c.Request.Context(), id); err != nil {
		helper.BadRequest(c, err.Error(), nil)
		return
	}
	helper.OK(c, "Pendaftaran perusahaan ditolak", nil)
}
