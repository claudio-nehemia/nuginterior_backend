package constants

// Permission names used for RBAC authorization.
// Format: {module}.{action}
const (
	// Divisi permissions
	PermDivisiIndex  = "divisi.index"
	PermDivisiCreate = "divisi.create"
	PermDivisiUpdate = "divisi.update"
	PermDivisiDelete = "divisi.delete"

	// Role permissions
	PermRoleIndex           = "role.index"
	PermRoleCreate          = "role.create"
	PermRoleShow            = "role.show"
	PermRoleUpdate          = "role.update"
	PermRoleDelete          = "role.delete"
	PermRoleSyncPermissions = "role.sync_permissions"

	// User permissions
	PermUserIndex  = "user.index"
	PermUserCreate = "user.create"
	PermUserUpdate = "user.update"
	PermUserDelete = "user.delete"

	// Permission permissions
	PermPermissionIndex = "permission.index"

	// Produk permissions
	PermProdukIndex       = "produk.index"
	PermProdukCreate      = "produk.create"
	PermProdukUpdate      = "produk.update"
	PermProdukDelete      = "produk.delete"
	PermProdukDeleteImage = "produk.delete_image"

	// Bahan Baku permissions
	PermBahanBakuIndex  = "bahan_baku.index"
	PermBahanBakuCreate = "bahan_baku.create"
	PermBahanBakuUpdate = "bahan_baku.update"
	PermBahanBakuDelete = "bahan_baku.delete"

	// Item permissions
	PermItemIndex  = "item.index"
	PermItemCreate = "item.create"
	PermItemUpdate = "item.update"
	PermItemDelete = "item.delete"

	// Jenis Pengukuran permissions
	PermJenisPengukuranIndex  = "jenis_pengukuran.index"
	PermJenisPengukuranCreate = "jenis_pengukuran.create"
	PermJenisPengukuranUpdate = "jenis_pengukuran.update"
	PermJenisPengukuranDelete = "jenis_pengukuran.delete"

	// Termin permissions
	PermTerminIndex  = "termin.index"
	PermTerminCreate = "termin.create"
	PermTerminUpdate = "termin.update"
	PermTerminDelete = "termin.delete"
	
	// Upload permissions
	PermUpload = "upload.image"

	// Order permissions
	PermOrderIndex             = "order.index"
	PermOrderCreate            = "order.create"
	PermOrderShow              = "order.show"
	PermOrderUpdate            = "order.update"
	PermOrderDelete            = "order.delete"

	// Survey permissions
	PermSurveyIndex             = "survey.index"
	PermSurveyCreate            = "survey.create"
	PermSurveyShow              = "survey.show"
	PermSurveyUpdate            = "survey.update"
	PermSurveyDelete            = "survey.delete"
	PermSurveyResponse          = "survey.response"
	PermSurveyMarketingResponse = "survey.marketing_response"

	// Moodboard permissions
	PermMoodboardIndex             = "moodboard.index"
	PermMoodboardCreate            = "moodboard.create"
	PermMoodboardShow              = "moodboard.show"
	PermMoodboardUpdate            = "moodboard.update"
	PermMoodboardDelete            = "moodboard.delete"
	PermMoodboardDeleteImage       = "moodboard.delete_image"
	PermMoodboardResponse          = "moodboard.response"
	PermMoodboardMarketingResponse = "moodboard.marketing_response"

	// Setting permissions
	PermSettingIndex  = "setting.index"
	PermSettingUpdate = "setting.update"

	// Input Item permissions
	PermInputItemIndex  = "input_item.index"
	PermInputItemCreate = "input_item.create"
	PermInputItemShow   = "input_item.show"
	PermInputItemUpdate = "input_item.update"
	PermInputItemDelete = "input_item.delete"

	// RAB permissions
	PermRABIndex  = "rab.index"
	PermRABCreate = "rab.create"
	PermRABShow   = "rab.show"
	PermRABUpdate = "rab.update"
	PermRABDelete = "rab.delete"
	PermRABSubmit = "rab.submit"

	// Contract permissions
	PermContractIndex    = "contract.index"
	PermContractCreate   = "contract.create"
	PermContractUpdate   = "contract.update"
	PermContractResponse = "contract.response"

	// Invoice permissions
	PermInvoiceIndex          = "invoice.index"
	PermInvoiceShow           = "invoice.show"
	PermInvoiceCreate         = "invoice.create"
	PermInvoiceDownload       = "invoice.download"
	PermInvoiceUploadProof    = "invoice.upload_proof"
	PermInvoiceUpdateDeadline = "invoice.update_deadline"
	PermInvoiceResponse       = "invoice.response"

	// Workplan permissions
	PermWorkplanIndex    = "workplan.index"
	PermWorkplanCreate   = "workplan.create"
	PermWorkplanShow     = "workplan.show"
	PermWorkplanUpdate   = "workplan.update"
	PermWorkplanResponse = "workplan.response"

	// Log Task permissions
	PermLogTaskIndex = "log_task.index"

	// Company permissions
	PermCompanyIndex  = "company.index"
	PermCompanyCreate = "company.create"
	PermCompanyUpdate = "company.update"
	PermCompanyDelete = "company.delete"
	PermCompanyVerify = "company.verify"
)

// AllPermissions returns all permission definitions for seeding.
func AllPermissions() []PermissionDef {
	return []PermissionDef{
		// Divisi
		{Name: PermDivisiIndex, DisplayName: "Lihat Divisi", Group: "Divisi"},
		{Name: PermDivisiCreate, DisplayName: "Tambah Divisi", Group: "Divisi"},
		{Name: PermDivisiUpdate, DisplayName: "Update Divisi", Group: "Divisi"},
		{Name: PermDivisiDelete, DisplayName: "Hapus Divisi", Group: "Divisi"},

		// Role
		{Name: PermRoleIndex, DisplayName: "Lihat Role", Group: "Role"},
		{Name: PermRoleCreate, DisplayName: "Tambah Role", Group: "Role"},
		{Name: PermRoleShow, DisplayName: "Detail Role", Group: "Role"},
		{Name: PermRoleUpdate, DisplayName: "Update Role", Group: "Role"},
		{Name: PermRoleDelete, DisplayName: "Hapus Role", Group: "Role"},
		{Name: PermRoleSyncPermissions, DisplayName: "Sync Permission Role", Group: "Role"},

		// User
		{Name: PermUserIndex, DisplayName: "Lihat User", Group: "User"},
		{Name: PermUserCreate, DisplayName: "Tambah User", Group: "User"},
		{Name: PermUserUpdate, DisplayName: "Update User", Group: "User"},
		{Name: PermUserDelete, DisplayName: "Hapus User", Group: "User"},

		// Permission
		{Name: PermPermissionIndex, DisplayName: "Lihat Permission", Group: "Permission"},

		// Produk
		{Name: PermProdukIndex, DisplayName: "Lihat Produk", Group: "Produk"},
		{Name: PermProdukCreate, DisplayName: "Tambah Produk", Group: "Produk"},
		{Name: PermProdukUpdate, DisplayName: "Update Produk", Group: "Produk"},
		{Name: PermProdukDelete, DisplayName: "Hapus Produk", Group: "Produk"},
		{Name: PermProdukDeleteImage, DisplayName: "Hapus Gambar Produk", Group: "Produk"},

		// Bahan Baku
		{Name: PermBahanBakuIndex, DisplayName: "Lihat Bahan Baku", Group: "Bahan Baku"},
		{Name: PermBahanBakuCreate, DisplayName: "Tambah Bahan Baku", Group: "Bahan Baku"},
		{Name: PermBahanBakuUpdate, DisplayName: "Update Bahan Baku", Group: "Bahan Baku"},
		{Name: PermBahanBakuDelete, DisplayName: "Hapus Bahan Baku", Group: "Bahan Baku"},

		// Item
		{Name: PermItemIndex, DisplayName: "Lihat Item", Group: "Item"},
		{Name: PermItemCreate, DisplayName: "Tambah Item", Group: "Item"},
		{Name: PermItemUpdate, DisplayName: "Update Item", Group: "Item"},
		{Name: PermItemDelete, DisplayName: "Hapus Item", Group: "Item"},

		// Jenis Pengukuran
		{Name: PermJenisPengukuranIndex, DisplayName: "Lihat Jenis Pengukuran", Group: "Jenis Pengukuran"},
		{Name: PermJenisPengukuranCreate, DisplayName: "Tambah Jenis Pengukuran", Group: "Jenis Pengukuran"},
		{Name: PermJenisPengukuranUpdate, DisplayName: "Update Jenis Pengukuran", Group: "Jenis Pengukuran"},
		{Name: PermJenisPengukuranDelete, DisplayName: "Hapus Jenis Pengukuran", Group: "Jenis Pengukuran"},

		// Termin
		{Name: PermTerminIndex, DisplayName: "Lihat Termin", Group: "Termin"},
		{Name: PermTerminCreate, DisplayName: "Tambah Termin", Group: "Termin"},
		{Name: PermTerminUpdate, DisplayName: "Update Termin", Group: "Termin"},
		{Name: PermTerminDelete, DisplayName: "Hapus Termin", Group: "Termin"},

		// Upload
		{Name: PermUpload, DisplayName: "Upload Gambar", Group: "Utilities"},

		// Order
		{Name: PermOrderIndex, DisplayName: "Lihat Order", Group: "Order"},
		{Name: PermOrderCreate, DisplayName: "Tambah Order", Group: "Order"},
		{Name: PermOrderShow, DisplayName: "Detail Order", Group: "Order"},
		{Name: PermOrderUpdate, DisplayName: "Update Order", Group: "Order"},
		{Name: PermOrderDelete, DisplayName: "Hapus Order", Group: "Order"},

		// Survey
		{Name: PermSurveyIndex, DisplayName: "Lihat Survey", Group: "Survey"},
		{Name: PermSurveyCreate, DisplayName: "Tambah Survey", Group: "Survey"},
		{Name: PermSurveyShow, DisplayName: "Detail Survey", Group: "Survey"},
		{Name: PermSurveyUpdate, DisplayName: "Update Survey", Group: "Survey"},
		{Name: PermSurveyDelete, DisplayName: "Hapus Survey", Group: "Survey"},
		{Name: PermSurveyResponse, DisplayName: "Response Survey", Group: "Survey"},
		{Name: PermSurveyMarketingResponse, DisplayName: "Marketing Response Survey", Group: "Survey"},

		// Moodboard
		{Name: PermMoodboardIndex, DisplayName: "Lihat Moodboard", Group: "Moodboard"},
		{Name: PermMoodboardCreate, DisplayName: "Tambah Moodboard", Group: "Moodboard"},
		{Name: PermMoodboardShow, DisplayName: "Detail Moodboard", Group: "Moodboard"},
		{Name: PermMoodboardUpdate, DisplayName: "Update Moodboard", Group: "Moodboard"},
		{Name: PermMoodboardDelete, DisplayName: "Hapus Moodboard", Group: "Moodboard"},
		{Name: PermMoodboardDeleteImage, DisplayName: "Hapus Gambar Moodboard", Group: "Moodboard"},
		{Name: PermMoodboardResponse, DisplayName: "Response Moodboard", Group: "Moodboard"},
		{Name: PermMoodboardMarketingResponse, DisplayName: "Marketing Response Moodboard", Group: "Moodboard"},

		// Setting
		{Name: PermSettingIndex, DisplayName: "Lihat Setting", Group: "Setting"},
		{Name: PermSettingUpdate, DisplayName: "Update Setting", Group: "Setting"},

		// Input Item
		{Name: PermInputItemIndex, DisplayName: "Lihat Input Item", Group: "Input Item"},
		{Name: PermInputItemCreate, DisplayName: "Tambah Input Item", Group: "Input Item"},
		{Name: PermInputItemShow, DisplayName: "Detail Input Item", Group: "Input Item"},
		{Name: PermInputItemUpdate, DisplayName: "Update Input Item", Group: "Input Item"},
		{Name: PermInputItemDelete, DisplayName: "Hapus Input Item", Group: "Input Item"},

		// RAB
		{Name: PermRABIndex, DisplayName: "Lihat RAB", Group: "RAB"},
		{Name: PermRABCreate, DisplayName: "Tambah RAB", Group: "RAB"},
		{Name: PermRABShow, DisplayName: "Detail RAB", Group: "RAB"},
		{Name: PermRABUpdate, DisplayName: "Update RAB", Group: "RAB"},
		{Name: PermRABDelete, DisplayName: "Hapus RAB", Group: "RAB"},
		{Name: PermRABSubmit, DisplayName: "Submit RAB", Group: "RAB"},

		// Contract
		{Name: PermContractIndex, DisplayName: "Lihat Kontrak", Group: "Kontrak"},
		{Name: PermContractCreate, DisplayName: "Tambah Kontrak", Group: "Kontrak"},
		{Name: PermContractUpdate, DisplayName: "Update Kontrak", Group: "Kontrak"},
		{Name: PermContractResponse, DisplayName: "Response Kontrak", Group: "Kontrak"},

		// Invoice
		{Name: PermInvoiceIndex, DisplayName: "Lihat Invoice List", Group: "Invoice"},
		{Name: PermInvoiceShow, DisplayName: "Lihat Invoice Detail", Group: "Invoice"},
		{Name: PermInvoiceCreate, DisplayName: "Buat/Terbitkan Invoice", Group: "Invoice"},
		{Name: PermInvoiceDownload, DisplayName: "Download PDF Invoice", Group: "Invoice"},
		{Name: PermInvoiceUploadProof, DisplayName: "Upload Bukti Pembayaran", Group: "Invoice"},
		{Name: PermInvoiceUpdateDeadline, DisplayName: "Update Deadline Invoice", Group: "Invoice"},
		{Name: PermInvoiceResponse, DisplayName: "Response Regular Invoice", Group: "Invoice"},

		// Workplan
		{Name: PermWorkplanIndex, DisplayName: "Lihat Workplan", Group: "Workplan"},
		{Name: PermWorkplanCreate, DisplayName: "Tambah Workplan", Group: "Workplan"},
		{Name: PermWorkplanShow, DisplayName: "Detail Workplan", Group: "Workplan"},
		{Name: PermWorkplanUpdate, DisplayName: "Update Workplan", Group: "Workplan"},
		{Name: PermWorkplanResponse, DisplayName: "Response Workplan", Group: "Workplan"},

		// Log Task
		{Name: PermLogTaskIndex, DisplayName: "Lihat Log Task", Group: "Log Task"},

		// Company
		{Name: PermCompanyIndex, DisplayName: "Lihat Company", Group: "Company"},
		{Name: PermCompanyCreate, DisplayName: "Tambah Company", Group: "Company"},
		{Name: PermCompanyUpdate, DisplayName: "Update Company", Group: "Company"},
		{Name: PermCompanyDelete, DisplayName: "Hapus Company", Group: "Company"},
		{Name: PermCompanyVerify, DisplayName: "Verifikasi Company", Group: "Company"},
	}
}

// PermissionDef represents a permission definition for seeding.
type PermissionDef struct {
	Name        string
	DisplayName string
	Group       string
}
