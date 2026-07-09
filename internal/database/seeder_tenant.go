package database

import (
	"fmt"
	"strings"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"gorm.io/gorm"
)

// SeedCompanyDefaults seeds divisions, roles, and settings for a verified company.
func SeedCompanyDefaults(db *gorm.DB, companyID uint) error {
	// 1. Create Divisi
	divisions := []string{"Management", "Legal", "Marketing", "Operasional", "Desain", "Estimasi"}
	divMap := make(map[string]uint)

	for _, dName := range divisions {
		var divisi entity.Divisi
		err := db.Where("company_id = ? AND nama_divisi = ?", companyID, dName).
			FirstOrCreate(&divisi, entity.Divisi{
				CompanyID:  companyID,
				NamaDivisi: dName,
			}).Error
		if err != nil {
			return fmt.Errorf("failed to seed division %s: %w", dName, err)
		}
		divMap[dName] = divisi.ID
	}

	// 2. Create Roles
	roles := map[string][]string{
		"Management":  {"Admin", "Owner"},
		"Legal":       {"Legal Admin"},
		"Marketing":   {"Customer Service", "Kepala Marketing"},
		"Operasional": {"Surveyor", "Supervisor", "Project Manager"},
		"Desain":      {"Drafter", "Desainer"},
		"Estimasi":    {"Estimator"},
	}

	roleMap := make(map[string]uint)
	for dName, rNames := range roles {
		divID := divMap[dName]
		for _, rName := range rNames {
			var role entity.Role
			err := db.Where("company_id = ? AND nama_role = ?", companyID, rName).
				FirstOrCreate(&role, entity.Role{
					CompanyID: companyID,
					NamaRole:  rName,
					DivisiID:  divID,
				}).Error
			if err != nil {
				return fmt.Errorf("failed to seed role %s: %w", rName, err)
			}
			roleMap[rName] = role.ID
		}
	}

	// 3. Assign Permissions
	var allPerms []entity.Permission
	if err := db.Find(&allPerms).Error; err != nil {
		return fmt.Errorf("failed to load permissions: %w", err)
	}

	// Admin role gets all permissions
	adminRoleID := roleMap["Admin"]
	for _, p := range allPerms {
		db.FirstOrCreate(&entity.RolePermission{}, entity.RolePermission{
			RoleID:       adminRoleID,
			PermissionID: p.ID,
		})
	}

	// Legal Admin gets invoice.* permissions
	legalRoleID := roleMap["Legal Admin"]
	for _, p := range allPerms {
		if strings.HasPrefix(p.Name, "invoice.") {
			db.FirstOrCreate(&entity.RolePermission{}, entity.RolePermission{
				RoleID:       legalRoleID,
				PermissionID: p.ID,
			})
		}
	}

	// 4. Create Settings
	defaults := []entity.Setting{
		{Key: "response_enabled", Value: "true", Description: "Aktifkan fitur Response (regular) di seluruh sistem"},
		{Key: "marketing_response_enabled", Value: "true", Description: "Aktifkan fitur Marketing Response di seluruh sistem"},
		{Key: "invoice_deadline_enabled", Value: "false", Description: "Aktifkan fitur deadline invoice"},
		{Key: "company_name", Value: "Nama Perusahaan Baru", Description: "Nama Perusahaan"},
		{Key: "company_director", Value: "Direktur Utama", Description: "Nama Direktur Perusahaan"},
		{Key: "company_logo", Value: "", Description: "URL Logo Perusahaan"},
		{Key: "company_address", Value: "Alamat Kantor Perusahaan", Description: "Alamat Kantor Perusahaan"},
		{Key: "company_bank_name", Value: "BANK CENTRAL ASIA (BCA)", Description: "Nama Bank Perusahaan"},
		{Key: "company_bank_account", Value: "000-000-0000", Description: "Nomor Rekening Perusahaan"},
		{Key: "company_bank_holder", Value: "Nama Pemegang Rekening Perusahaan", Description: "Nama Pemegang Rekening Perusahaan"},
		{Key: "company_email", Value: "office@company.com", Description: "Email Perusahaan"},
		{Key: "company_phone", Value: "+62 812-0000-0000", Description: "Nomor Telepon Perusahaan"},
		{Key: "default_active_days", Value: "4", Description: "Masa berlaku akun default untuk perusahaan baru (hari)"},
		{Key: "workplan_stage_deletion_policy", Value: "split_equally", Description: "Kebijakan pembagian persentase saat tahapan dihapus dari workplan (split_equally, transfer_to_next, transfer_to_previous)"},
		{Key: "deadline_stage_survey", Value: "3", Description: "Batas waktu penyelesaian tahap Survey (hari)"},
		{Key: "deadline_stage_moodboard", Value: "5", Description: "Batas waktu penyelesaian tahap Moodboard (hari)"},
		{Key: "deadline_stage_estimasi", Value: "3", Description: "Batas waktu penyelesaian tahap Estimasi (hari)"},
		{Key: "deadline_stage_cm_fee", Value: "3", Description: "Batas waktu penyelesaian tahap CM Fee (hari)"},
		{Key: "deadline_stage_desain_final", Value: "5", Description: "Batas waktu penyelesaian tahap Desain Final (hari)"},
		{Key: "deadline_stage_input_item", Value: "3", Description: "Batas waktu penyelesaian tahap Input Item (hari)"},
		{Key: "deadline_stage_rab", Value: "4", Description: "Batas waktu penyelesaian tahap RAB (hari)"},
		{Key: "deadline_stage_kontrak", Value: "5", Description: "Batas waktu penyelesaian tahap Kontrak (hari)"},
		{Key: "deadline_stage_invoice", Value: "3", Description: "Batas waktu penyelesaian tahap Invoice (hari)"},
		{Key: "deadline_stage_survey_ulang", Value: "3", Description: "Batas waktu penyelesaian tahap Survey Ulang (hari)"},
		{Key: "deadline_stage_gambar_kerja", Value: "7", Description: "Batas waktu penyelesaian tahap Gambar Kerja (hari)"},
		{Key: "deadline_stage_approval_material", Value: "5", Description: "Batas waktu penyelesaian tahap Approval Material (hari)"},
		{Key: "deadline_stage_workplan", Value: "5", Description: "Batas waktu penyelesaian tahap Workplan (hari)"},
		{Key: "deadline_stage_operations", Value: "14", Description: "Batas waktu penyelesaian tahap Operations (hari)"},
		{Key: "sidebar_configuration", Value: `[{"id":"master_data","name":"Master Data","icon":"Database","items":[{"code":"divisi","name":"Divisi","icon":"Database","path":"/dashboard/divisi","permission":"divisi.index","visible":true},{"code":"roles","name":"Role & User","icon":"Users","path":"/dashboard/roles","permission":"role.index","visible":true},{"code":"produk","name":"Produk","icon":"Package","path":"/dashboard/produk","permission":"produk.index","visible":true},{"code":"item","name":"Item","icon":"Box","path":"/dashboard/item","permission":"item.index","visible":true},{"code":"pengukuran","name":"Jenis Pengukuran","icon":"Ruler","path":"/dashboard/pengukuran","permission":"jenis_pengukuran.index","visible":true},{"code":"termin","name":"Termin","icon":"Wallet","path":"/dashboard/termin","permission":"termin.index","visible":true},{"code":"companies","name":"Company","icon":"Building2","path":"/dashboard/companies","permission":"company.index","visible":true}]},{"id":"operations","name":"Operations","icon":"ShoppingCart","items":[{"code":"order","name":"Order","icon":"ShoppingCart","path":"/dashboard/order","permission":"order.index","visible":true},{"code":"survey","name":"Survey","icon":"ClipboardCheck","path":"/dashboard/survey","permission":"survey.index","visible":true},{"code":"moodboard","name":"Moodboard","icon":"Palette","path":"/dashboard/moodboard","permission":"moodboard.index","visible":true},{"code":"estimasi","name":"Estimasi","icon":"Calculator","path":"/dashboard/estimasi","permission":"moodboard.index","visible":true},{"code":"desain_final","name":"Desain Final","icon":"Palette","path":"/dashboard/desain-final","permission":"moodboard.index","visible":true},{"code":"input_item","name":"Input Item","icon":"ClipboardCheck","path":"/dashboard/input-item","permission":"input_item.index","visible":true},{"code":"gambar_kerja","name":"Gambar Kerja","icon":"FileText","path":"/dashboard/gambar-kerja","permission":"moodboard.index","visible":true},{"code":"approval_material","name":"Approval Material","icon":"ClipboardCheck","path":"/dashboard/approval-material","permission":"moodboard.index","visible":true},{"code":"workplan","name":"Workplan","icon":"FileText","path":"/dashboard/workplan","permission":"workplan.index","visible":true},{"code":"project_management","name":"Project Management","icon":"ClipboardCheck","path":"/dashboard/project-management","permission":"workplan.index","visible":true},{"code":"log_task","name":"Log Task","icon":"ClipboardCheck","path":"/dashboard/log-tasks","permission":"log_task.index","visible":true}]},{"id":"finance","name":"Finance","icon":"Coins","items":[{"code":"commitment_fee","name":"Commitment Fee","icon":"Wallet","path":"/dashboard/commitment-fee","permission":"moodboard.index","visible":true},{"code":"rab","name":"RAB","icon":"Coins","path":"/dashboard/rab","permission":"rab.index","visible":true},{"code":"kontrak","name":"Kontrak","icon":"FileText","path":"/dashboard/kontrak","permission":"contract.index","visible":true},{"code":"invoice","name":"Invoice","icon":"Receipt","path":"/dashboard/invoice","permission":"invoice.index","visible":true}]}]`, Description: "Konfigurasi navigasi sidebar dinamis"},
		{Key: "notification_settings", Value: `{"always_notified_roles":["Kepala Marketing"],"rules":{"assign_order":{"name":"Assign Order","roles":[{"role_name":"Kepala Marketing","team_only":true},{"role_name":"Drafter","team_only":true},{"role_name":"Surveyor","team_only":true},{"role_name":"Project Manager","team_only":true}]},"moodboard":{"name":"Tahap Moodboard (Survey Selesai)","roles":[{"role_name":"Desainer","team_only":true},{"role_name":"Estimator","team_only":false}]},"estimasi":{"name":"Tahap Estimasi","roles":[{"role_name":"Estimator","team_only":false}]},"commitment_fee":{"name":"Tahap Commitment Fee","roles":[{"role_name":"Legal Admin","team_only":false}]},"desain_final":{"name":"Tahap Desain Final","roles":[{"role_name":"Desainer","team_only":true}]},"input_item":{"name":"Tahap Input Item","roles":[{"role_name":"Estimator","team_only":false},{"role_name":"Drafter","team_only":true}]},"rab":{"name":"Tahap RAB","roles":[{"role_name":"Estimator","team_only":false}]},"kontrak":{"name":"Tahap Kontrak","roles":[{"role_name":"Legal Admin","team_only":false}]},"invoice":{"name":"Tahap Invoice","roles":[{"role_name":"Legal Admin","team_only":false}]},"setup_survey_ulang":{"name":"Setup Survey Ulang","roles":[{"role_name":"Project Manager","team_only":true}]},"upload_survey_ulang":{"name":"Upload Hasil Survey Ulang","roles":[{"role_name":"Drafter","team_only":true},{"role_name":"Surveyor","team_only":true}]},"gambar_kerja":{"name":"Tahap Gambar Kerja","roles":[{"role_name":"Drafter","team_only":true},{"role_name":"Desainer","team_only":true}]},"approval_material":{"name":"Tahap Approval Material & Workplan","roles":[{"role_name":"Project Manager","team_only":true}]},"project_management":{"name":"Tahap Project Management","roles":[{"role_name":"Supervisor","team_only":true},{"role_name":"Project Manager","team_only":true}]}}}`, Description: "Konfigurasi sistem notifikasi dinamis per event"},
	}

	// Fetch actual company from DB to fill in Company Name, Phone, Email and Address
	var comp entity.Company
	if err := db.First(&comp, companyID).Error; err == nil {
		for i, s := range defaults {
			if s.Key == "company_name" {
				defaults[i].Value = comp.Name
			} else if s.Key == "company_director" {
				defaults[i].Value = comp.DirectorName
			} else if s.Key == "company_address" {
				defaults[i].Value = comp.Address
			} else if s.Key == "company_email" {
				defaults[i].Value = comp.Email
			} else if s.Key == "company_phone" {
				defaults[i].Value = comp.Phone
			}
		}
	}

	for _, s := range defaults {
		db.FirstOrCreate(&entity.Setting{}, entity.Setting{
			CompanyID:   companyID,
			Key:         s.Key,
			Value:       s.Value,
			Description: s.Description,
		})
	}

	return nil
}
