package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ProdukHandler struct {
	svc service.ProdukService
}

func NewProdukHandler(svc service.ProdukService) *ProdukHandler {
	return &ProdukHandler{svc: svc}
}

func (h *ProdukHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat produk")
		return
	}
	helper.OK(c, "Daftar produk", data)
}

func (h *ProdukHandler) Store(c *gin.Context) {
	var req dto.CreateProdukRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.Created(c, "Produk berhasil dibuat", data)
}

func (h *ProdukHandler) Update(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	var req dto.UpdateProdukRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Produk berhasil diupdate", data)
}

func (h *ProdukHandler) Destroy(c *gin.Context) {
	id, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "ID tidak valid", nil)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Produk berhasil dihapus", nil)
}

func (h *ProdukHandler) DestroyImage(c *gin.Context) {
	produkID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Produk ID tidak valid", nil)
		return
	}
	imageID, err := helper.ParseIDParam(c, "imageId")
	if err != nil {
		helper.BadRequest(c, "Image ID tidak valid", nil)
		return
	}
	if err := h.svc.DeleteImage(c.Request.Context(), produkID, imageID); err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Image produk berhasil dihapus", nil)
}
