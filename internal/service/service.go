package service

import (
	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config *config.Config
	DB     *gorm.DB
	Cache  cache.Store
	Logger *zap.Logger
}

type Services struct {
	DB              *gorm.DB
	Auth            AuthService
	Divisi          DivisiService
	Role            RoleService
	Permission      PermissionService
	User            UserService
	Produk          ProdukService
	BahanBaku       BahanBakuService
	Item            ItemService
	JenisPengukuran JenisPengukuranService
	Termin          TerminService
	Order           OrderService
	Survey          SurveyService
	Moodboard       MoodboardService
	Setting         SettingService
	DesainFinal     DesainFinalService
	InputItem       InputItemService
	RAB             RABService
	Contract        ContractService
	Invoice         InvoiceService
	GambarKerja     GambarKerjaService
	ApprovalMaterial ApprovalMaterialService
	Workplan         WorkplanService
	Dashboard        DashboardService
	ProjectLogTask   ProjectLogTaskService
	Notification     NotificationService
	AI               AIService
	Company          CompanyService
}

func NewServices(deps Dependencies) *Services {
	repos := repository.NewRepositories(deps.DB, deps.Logger)
	settingSvc := NewSettingService(repos.Setting, deps.Cache, deps.Logger)
	notificationSvc := NewNotificationService(repos.Notification, repos.Setting, deps.DB, deps.Logger)
	projectLogTaskSvc := NewProjectLogTaskService(repos.ProjectLogTask, repos.Order, repos.Setting, deps.Logger, notificationSvc)

	return &Services{
		DB:              deps.DB,
		Auth:            NewAuthService(deps.Config, repos.Auth, repos.Role, deps.Cache, deps.Logger),
		Divisi:          NewDivisiService(repos.Divisi, deps.Cache, deps.Logger),
		Role:            NewRoleService(repos.Role, deps.Cache, deps.Logger),
		Permission:      NewPermissionService(repos.Permission, deps.Cache, deps.Logger),
		User:            NewUserService(repos.User, deps.Logger),
		Produk:          NewProdukService(repos.Produk, deps.Logger),
		BahanBaku:       NewBahanBakuService(repos.BahanBaku, deps.Cache, deps.Logger),
		Item:            NewItemService(repos.Item, deps.Cache, deps.Logger),
		JenisPengukuran: NewJenisPengukuranService(repos.JenisPengukuran, deps.Cache, deps.Logger),
		Termin:          NewTerminService(repos.Termin, deps.Cache, deps.Logger),
		Order:           NewOrderService(repos.Order, deps.DB, deps.Logger, deps.Config.UploadDir, projectLogTaskSvc, notificationSvc),
		Survey:          NewSurveyService(repos.Survey, deps.Logger, projectLogTaskSvc, notificationSvc),
		Moodboard:       NewMoodboardService(repos.Moodboard, deps.DB, deps.Cache, deps.Logger, deps.Config.UploadDir, projectLogTaskSvc),
		Setting:         settingSvc,
		DesainFinal:     NewDesainFinalService(repos.DesainFinal, repos.Moodboard, deps.Cache, deps.Logger, deps.Config.UploadDir, projectLogTaskSvc, deps.DB),
		InputItem:       NewInputItemService(repos.InputItem, deps.Logger, projectLogTaskSvc),
		RAB:             NewRABService(repos.RAB, repos.InputItem, deps.DB, deps.Logger, deps.Config.UploadDir, projectLogTaskSvc),
		Contract:        NewContractService(repos.Contract, repos.RAB, repos.Termin, repos.User, deps.DB, deps.Logger, deps.Config.UploadDir, projectLogTaskSvc),
		Invoice:         NewInvoiceService(repos.Invoice, repos.Contract, repos.User, deps.DB, deps.Logger, deps.Config.UploadDir, projectLogTaskSvc),
		GambarKerja:     NewGambarKerjaService(repos.GambarKerja, repos.User, deps.DB, deps.Logger, deps.Config.UploadDir, projectLogTaskSvc),
		ApprovalMaterial: NewApprovalMaterialService(repos.ApprovalMaterial, settingSvc, deps.DB, deps.Logger, projectLogTaskSvc),
		Workplan:         NewWorkplanService(repos.Workplan, settingSvc, deps.DB, deps.Logger, deps.Config.UploadDir, projectLogTaskSvc),
		Dashboard:        NewDashboardService(deps.DB, deps.Logger),
		ProjectLogTask:   projectLogTaskSvc,
		Notification:     notificationSvc,
		AI:               NewAIService(deps.DB, deps.Config.OpenAIKey, deps.Logger),
		Company:          NewCompanyService(deps.DB, deps.Logger),
	}
}
