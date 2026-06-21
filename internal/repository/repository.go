package repository

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repositories struct {
	Auth            AuthRepository
	Divisi          DivisiRepository
	Role            RoleRepository
	Permission      PermissionRepository
	User            UserRepository
	Produk          ProdukRepository
	BahanBaku       BahanBakuRepository
	Item            ItemRepository
	JenisPengukuran JenisPengukuranRepository
	Termin          TerminRepository
	Order           OrderRepository
	Survey          SurveyRepository
	Moodboard       MoodboardRepository
	Setting         SettingRepository
	DesainFinal     DesainFinalRepository
	InputItem       InputItemRepository
	RAB             RABRepository
	Contract        ContractRepository
	Invoice         InvoiceRepository
	GambarKerja     GambarKerjaRepository
	ApprovalMaterial ApprovalMaterialRepository
	Workplan         WorkplanRepository
	ProjectLogTask   ProjectLogTaskRepository
	Notification     NotificationRepository
}

func NewRepositories(db *gorm.DB, logger *zap.Logger) *Repositories {
	return &Repositories{
		Auth:            NewAuthRepository(db, logger),
		Divisi:          NewDivisiRepository(db, logger),
		Role:            NewRoleRepository(db, logger),
		Permission:      NewPermissionRepository(db, logger),
		User:            NewUserRepository(db, logger),
		Produk:          NewProdukRepository(db, logger),
		BahanBaku:       NewBahanBakuRepository(db, logger),
		Item:            NewItemRepository(db, logger),
		JenisPengukuran: NewJenisPengukuranRepository(db, logger),
		Termin:          NewTerminRepository(db, logger),
		Order:           NewOrderRepository(db, logger),
		Survey:          NewSurveyRepository(db, logger),
		Moodboard:       NewMoodboardRepository(db, logger),
		Setting:         NewSettingRepository(db, logger),
		DesainFinal:     NewDesainFinalRepository(db, logger),
		InputItem:       NewInputItemRepository(db, logger),
		RAB:             NewRABRepository(db, logger),
		Contract:        NewContractRepository(db, logger),
		Invoice:         NewInvoiceRepository(db, logger),
		GambarKerja:     NewGambarKerjaRepository(db, logger),
		ApprovalMaterial: NewApprovalMaterialRepository(db, logger),
		Workplan:         NewWorkplanRepository(db, logger),
		ProjectLogTask:   NewProjectLogTaskRepository(db, logger),
		Notification:     NewNotificationRepository(db, logger),
	}
}
