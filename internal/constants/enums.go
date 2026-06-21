package constants

// Jenis Interior
const (
	JenisInteriorResidential = "residential"
	JenisInteriorApartment   = "apartment"
	JenisInteriorOffice      = "office"
	JenisInteriorRetail      = "retail"
	JenisInteriorHospitality = "hospitality"
	JenisInteriorCustom      = "custom"
)

var JenisInteriorList = []string{
	JenisInteriorResidential, JenisInteriorApartment,
	JenisInteriorOffice, JenisInteriorRetail,
	JenisInteriorHospitality, JenisInteriorCustom,
}

// Project Status
const (
	ProjectStatusPending    = "pending"
	ProjectStatusInProgress = "in_progress"
	ProjectStatusDeal       = "deal"
	ProjectStatusCancel     = "cancel"
)

// Priority Level
const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

// Tahapan Proyek
const (
	TahapanNotStart     = "not_start"
	TahapanSurvey       = "survey"
	TahapanMoodboard    = "moodboard"
	TahapanEstimasi     = "estimasi"
	TahapanCmFee        = "cm_fee"
	TahapanDesainFinal  = "desain_final"
	TahapanInputItem    = "input_item"
	TahapanRAB          = "rab"
	TahapanKontrak      = "kontrak"
	TahapanInvoice      = "invoice"
	TahapanSurveyUlang  = "survey_ulang"
	TahapanGambarKerja  = "gambar_kerja"
	TahapanApprovalMaterial = "approval_material"
	TahapanWorkplan     = "workplan"
	TahapanOperations   = "operations"
	TahapanSelesai      = "selesai"
	TahapanBatal        = "batal"
)

// Payment Status
const (
	PaymentNotStart = "not_start"
	PaymentCmFee    = "cm_fee"
	PaymentDP       = "dp"
	PaymentTermin   = "termin"
	PaymentLunas    = "lunas"
)
