package handler

import (
	"net/http"

	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/claudio-nehemia/interior_backend/pkg/pdf"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApprovalMaterialHandler struct {
	svc        service.ApprovalMaterialService
	settingSvc service.SettingService
	db         *gorm.DB
	cfg        *config.Config
}

func NewApprovalMaterialHandler(
	svc service.ApprovalMaterialService,
	settingSvc service.SettingService,
	db *gorm.DB,
	cfg *config.Config,
) *ApprovalMaterialHandler {
	return &ApprovalMaterialHandler{
		svc:        svc,
		settingSvc: settingSvc,
		db:         db,
		cfg:        cfg,
	}
}

func (h *ApprovalMaterialHandler) getUserEmail(c *gin.Context) string {
	email, exists := c.Get("email")
	if !exists {
		return "CS / Designer / Estimator"
	}
	return email.(string)
}

// GET /api/approval-materials
func (h *ApprovalMaterialHandler) Index(c *gin.Context) {
	data, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		helper.InternalError(c, "Gagal memuat daftar approval material: "+err.Error())
		return
	}
	helper.OK(c, "Daftar approval material berhasil dimuat", data)
}

// GET /api/approval-materials/order/:orderId
func (h *ApprovalMaterialHandler) Show(c *gin.Context) {
	orderID, err := helper.ParseIDParam(c, "orderId")
	if err != nil {
		helper.BadRequest(c, "Order ID tidak valid", nil)
		return
	}

	data, err := h.svc.GetByOrderID(c.Request.Context(), orderID)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Detail approval material order berhasil dimuat", data)
}

// POST /api/approval-materials/:id/response
// Di sini :id adalah Order ID
func (h *ApprovalMaterialHandler) Response(c *gin.Context) {
	orderID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Order ID tidak valid", nil)
		return
	}

	enabled, _ := h.settingSvc.IsEnabled(c.Request.Context(), "response_enabled")
	if !enabled {
		helper.Forbidden(c, "Fitur response dinonaktifkan")
		return
	}

	userEmail := h.getUserEmail(c)
	data, err := h.svc.Response(c.Request.Context(), orderID, userEmail)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Berhasil response approval material order", data)
}

// PUT /api/approval-materials/:id
// Di sini :id adalah ApprovalMaterial ID
func (h *ApprovalMaterialHandler) Update(c *gin.Context) {
	amID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Approval Material ID tidak valid", nil)
		return
	}

	var req dto.UpdateApprovalMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.BadRequest(c, "Format data tidak valid: "+err.Error(), nil)
		return
	}

	data, err := h.svc.Update(c.Request.Context(), amID, req)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}
	helper.OK(c, "Approval material berhasil diperbarui", data)
}

// GET /api/approval-materials/:id/pdf
// Di sini :id adalah ApprovalMaterial ID
func (h *ApprovalMaterialHandler) ExportPDF(c *gin.Context) {
	amID, err := helper.ParseIDParam(c, "id")
	if err != nil {
		helper.BadRequest(c, "Approval Material ID tidak valid", nil)
		return
	}

	// Fetch approval material data
	am, err := h.svc.GetByID(c.Request.Context(), amID)
	if err != nil {
		helper.NotFound(c, "Approval material tidak ditemukan: "+err.Error())
		return
	}

	// Fetch dynamic company profile from DB
	var companyID uint = 1
	if am.OrderID > 0 {
		var order entity.Order
		if errOrder := h.db.WithContext(c.Request.Context()).First(&order, am.OrderID).Error; errOrder == nil {
			companyID = order.CompanyID
		}
	}
	entityProfile := entity.GetCompanyProfile(h.db, companyID)
	pdfProfile := pdf.CompanyProfile{
		Name:        entityProfile.Name,
		Director:    entityProfile.Director,
		Logo:        entityProfile.Logo,
		Address:     entityProfile.Address,
		BankName:    entityProfile.BankName,
		BankAccount: entityProfile.BankAccount,
		BankHolder:  entityProfile.BankHolder,
		Email:       entityProfile.Email,
		Phone:       entityProfile.Phone,
	}

	pdfBytes, err := pdf.GenerateApprovalMaterialPDF(am, pdfProfile, h.cfg.UploadDir)
	if err != nil {
		helper.InternalError(c, "Gagal mengekspor PDF approval material: "+err.Error())
		return
	}

	filename := "approval_material_"
	if am.Order != nil {
		filename += am.Order.NomorOrder
	} else {
		filename += string(rune(amID))
	}
	filename += ".pdf"

	c.Header("Content-Disposition", "inline; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
