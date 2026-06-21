package handler

import (
	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/service"
)

type Handlers struct {
	Auth            *AuthHandler
	Divisi          *DivisiHandler
	Role            *RoleHandler
	Permission      *PermissionHandler
	User            *UserHandler
	Produk          *ProdukHandler
	BahanBaku       *BahanBakuHandler
	Item            *ItemHandler
	JenisPengukuran *JenisPengukuranHandler
	Termin          *TerminHandler
	Upload          *UploadHandler
	Order           *OrderHandler
	Survey          *SurveyHandler
	Moodboard       *MoodboardHandler
	Setting         *SettingHandler
	DesainFinal     *DesainFinalHandler
	InputItem       *InputItemHandler
	RAB             *RABHandler
	Contract        *ContractHandler
	Invoice         *InvoiceHandler
	GambarKerja     *GambarKerjaHandler
	ApprovalMaterial *ApprovalMaterialHandler
	Workplan         *WorkplanHandler
	Dashboard        *DashboardHandler
	ProjectLogTask   *ProjectLogTaskHandler
	Notification     *NotificationHandler
	AI               *AIHandler
	Company          *CompanyHandler
}

func NewHandlers(services *service.Services, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth:            NewAuthHandler(services.Auth),
		Divisi:          NewDivisiHandler(services.Divisi),
		Role:            NewRoleHandler(services.Role),
		Permission:      NewPermissionHandler(services.Permission),
		User:            NewUserHandler(services.User),
		Produk:          NewProdukHandler(services.Produk),
		BahanBaku:       NewBahanBakuHandler(services.BahanBaku),
		Item:            NewItemHandler(services.Item),
		JenisPengukuran: NewJenisPengukuranHandler(services.JenisPengukuran),
		Termin:          NewTerminHandler(services.Termin),
		Upload:          NewUploadHandler(cfg),
		Order:           NewOrderHandler(services.Order, services.Setting),
		Survey:          NewSurveyHandler(services.Survey, services.Setting),
		Moodboard:       NewMoodboardHandler(services.Moodboard, services.Setting),
		Setting:         NewSettingHandler(services.Setting),
		DesainFinal:     NewDesainFinalHandler(services.DesainFinal),
		InputItem:       NewInputItemHandler(services.InputItem),
		RAB:             NewRABHandler(services.RAB),
		Contract:        NewContractHandler(services.Contract, services.Setting),
		Invoice:         NewInvoiceHandler(services.Invoice, services.Setting),
		GambarKerja:     NewGambarKerjaHandler(services.GambarKerja),
		ApprovalMaterial: NewApprovalMaterialHandler(services.ApprovalMaterial, services.Setting, services.DB, cfg),
		Workplan:         NewWorkplanHandler(services.Workplan, services.Setting, services.DB, cfg),
		Dashboard:        NewDashboardHandler(services.Dashboard),
		ProjectLogTask:   NewProjectLogTaskHandler(services.ProjectLogTask),
		Notification:     NewNotificationHandler(services.Notification),
		AI:               NewAIHandler(services.AI, services.Dashboard),
		Company:          NewCompanyHandler(services.Company),
	}
}
