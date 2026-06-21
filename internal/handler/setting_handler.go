package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SettingHandler struct {
	svc service.SettingService
}

func NewSettingHandler(svc service.SettingService) *SettingHandler {
	return &SettingHandler{svc: svc}
}

func (h *SettingHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Data settings berhasil dimuat", data)
}

func (h *SettingHandler) Update(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		helper.BadRequest(c, "Key tidak valid", nil)
		return
	}
	var req dto.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Validasi gagal", err.Error())
		return
	}
	data, err := h.svc.Update(c.Request.Context(), key, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Setting berhasil diupdate", data)
}

func (h *SettingHandler) GetSidebarConfig(c *gin.Context) {
	config, err := h.svc.GetByKey(c.Request.Context(), "sidebar_configuration")
	if err != nil {
		defaultJSON := `[{"id":"master_data","name":"Master Data","icon":"Database","items":[{"code":"divisi","name":"Divisi","icon":"Database","path":"/dashboard/divisi","permission":"divisi.index","visible":true},{"code":"roles","name":"Role & User","icon":"Users","path":"/dashboard/roles","permission":"role.index","visible":true},{"code":"produk","name":"Produk","icon":"Package","path":"/dashboard/produk","permission":"produk.index","visible":true},{"code":"item","name":"Item","icon":"Box","path":"/dashboard/item","permission":"item.index","visible":true},{"code":"pengukuran","name":"Jenis Pengukuran","icon":"Ruler","path":"/dashboard/pengukuran","permission":"jenis_pengukuran.index","visible":true},{"code":"termin","name":"Termin","icon":"Wallet","path":"/dashboard/termin","permission":"termin.index","visible":true}]},{"id":"operations","name":"Operations","icon":"ShoppingCart","items":[{"code":"order","name":"Order","icon":"ShoppingCart","path":"/dashboard/order","permission":"order.index","visible":true},{"code":"survey","name":"Survey","icon":"ClipboardCheck","path":"/dashboard/survey","permission":"survey.index","visible":true},{"code":"moodboard","name":"Moodboard","icon":"Palette","path":"/dashboard/moodboard","permission":"moodboard.index","visible":true},{"code":"estimasi","name":"Estimasi","icon":"Calculator","path":"/dashboard/estimasi","permission":"moodboard.index","visible":true},{"code":"desain_final","name":"Desain Final","icon":"Palette","path":"/dashboard/desain-final","permission":"moodboard.index","visible":true},{"code":"input_item","name":"Input Item","icon":"ClipboardCheck","path":"/dashboard/input-item","permission":"input_item.index","visible":true},{"code":"gambar_kerja","name":"Gambar Kerja","icon":"FileText","path":"/dashboard/gambar-kerja","permission":"moodboard.index","visible":true},{"code":"approval_material","name":"Approval Material","icon":"ClipboardCheck","path":"/dashboard/approval-material","permission":"moodboard.index","visible":true},{"code":"workplan","name":"Workplan","icon":"FileText","path":"/dashboard/workplan","permission":"workplan.index","visible":true}]},{"id":"finance","name":"Finance","icon":"Coins","items":[{"code":"commitment_fee","name":"Commitment Fee","icon":"Wallet","path":"/dashboard/commitment-fee","permission":"moodboard.index","visible":true},{"code":"rab","name":"RAB","icon":"Coins","path":"/dashboard/rab","permission":"rab.index","visible":true},{"code":"kontrak","name":"Kontrak","icon":"FileText","path":"/dashboard/kontrak","permission":"contract.index","visible":true},{"code":"invoice","name":"Invoice","icon":"Receipt","path":"/dashboard/invoice","permission":"invoice.index","visible":true}]}]`
		helper.OK(c, "Sidebar configuration loaded (fallback)", dto.SettingResponse{
			Key:   "sidebar_configuration",
			Value: defaultJSON,
		})
		return
	}
	helper.OK(c, "Sidebar configuration loaded", config)
}
