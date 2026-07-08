package database

import (
	"fmt"
	"strings"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SeedPermissions creates or updates all defined permissions in the database.
func SeedPermissions(db *gorm.DB, logger *zap.Logger) {
	allPerms := constants.AllPermissions()
	for _, def := range allPerms {
		perm := entity.Permission{
			Name:        def.Name,
			DisplayName: def.DisplayName,
			Group:       def.Group,
		}
		result := db.Where("name = ?", perm.Name).
			Assign(entity.Permission{
				DisplayName: perm.DisplayName,
				Group:       perm.Group,
			}).
			FirstOrCreate(&perm)
		if result.Error != nil {
			logger.Error("Failed to seed permission", zap.String("name", def.Name), zap.Error(result.Error))
		}
	}
	logger.Info("Permissions seeded", zap.Int("count", len(allPerms)))
}

// SeedAdminUser creates a default division, superadmin role, and the first admin user.
func SeedAdminUser(db *gorm.DB, logger *zap.Logger) {
	// 0. Create Company 1
	var company entity.Company
	db.Where("id = ?", 1).FirstOrCreate(&company, entity.Company{
		ID:           1,
		Name:         "PT. INTERIORPRO INDONESIA",
		DirectorName: "Super Admin",
		CeoNik:       "1234567890123456",
		Email:        "admin@interiorpro.com",
		Phone:        "+62 812-0000-0000",
		Status:       "verified",
	})

	// 1. Create Divisi
	var divisi entity.Divisi
	db.Where("company_id = ? AND nama_divisi = ?", 1, "Management").FirstOrCreate(&divisi, entity.Divisi{
		CompanyID:  1,
		NamaDivisi: "Management",
	})

	// 2. Create Role Super Admin
	var role entity.Role
	db.Where("company_id = ? AND nama_role = ?", 1, "Super Admin").FirstOrCreate(&role, entity.Role{
		CompanyID: 1,
		NamaRole:  "Super Admin",
		DivisiID:  divisi.ID,
	})

	// 3. Assign all permissions to Super Admin
	var perms []entity.Permission
	db.Find(&perms)
	
	for _, p := range perms {
		db.FirstOrCreate(&entity.RolePermission{}, entity.RolePermission{
			RoleID:       role.ID,
			PermissionID: p.ID,
		})
	}

	// 4. Create User
	var user entity.User
	result := db.Where("email = ?", "admin@interiorpro.com").First(&user)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		hashedPassword, _ := helper.HashPassword("password")
		user = entity.User{
			CompanyID: 1,
			Name:      "Admin User",
			Email:     "admin@interiorpro.com",
			Password:  hashedPassword,
			RoleID:    &role.ID,
		}
		db.Create(&user)
		logger.Info("Default Admin user created", zap.String("email", user.Email))
	} else {
		logger.Info("Default Admin user already exists")
	}
}

// SeedSettings creates default settings if they don't exist.
func SeedSettings(db *gorm.DB, logger *zap.Logger) {
	defaults := []entity.Setting{
		{Key: "response_enabled", Value: "true", Description: "Aktifkan fitur Response (regular) di seluruh sistem"},
		{Key: "marketing_response_enabled", Value: "true", Description: "Aktifkan fitur Marketing Response di seluruh sistem"},
		{Key: "invoice_deadline_enabled", Value: "false", Description: "Aktifkan fitur deadline invoice"},
		{Key: "workflow_rab_approval_required", Value: "true", Description: "Menentukan apakah pembuatan Kontrak harus menunggu persetujuan (approval) RAB dari klien di sistem"},
		{Key: "finance_tax_enabled", Value: "false", Description: "Menentukan apakah sistem otomatis menambahkan PPN (11%) pada setiap RAB dan Invoice yang terbit"},
		{Key: "finance_auto_invoice", Value: "true", Description: "Jika diaktifkan, sistem otomatis men-generate invoice baru begitu kontrak disetujui (deal)"},
		{Key: "company_name", Value: "PT. INTERIORPRO INDONESIA", Description: "Nama Perusahaan"},
		{Key: "company_director", Value: "Super Admin", Description: "Nama Direktur Perusahaan"},
		{Key: "company_logo", Value: "", Description: "URL Logo Perusahaan"},
		{Key: "company_address", Value: "Alamat Perusahaan", Description: "Alamat Kantor Perusahaan"},
		{Key: "company_bank_name", Value: "BANK CENTRAL ASIA (BCA)", Description: "Nama Bank Perusahaan"},
		{Key: "company_bank_account", Value: "000-000-0000", Description: "Nomor Rekening Perusahaan"},
		{Key: "company_bank_holder", Value: "PT. INTERIORPRO INDONESIA", Description: "Nama Pemegang Rekening Perusahaan"},
		{Key: "company_email", Value: "finance@interiorpro.com", Description: "Email Perusahaan"},
		{Key: "company_phone", Value: "+62 812-0000-0000", Description: "Nomor Telepon Perusahaan"},
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
	for _, s := range defaults {
		var existing entity.Setting
		db.Where("company_id = ? AND key = ?", 1, s.Key).FirstOrCreate(&existing, entity.Setting{
			CompanyID:   1,
			Key:         s.Key,
			Value:       s.Value,
			Description: s.Description,
		})
	}
	logger.Info("Settings seeded")
}

// SeedRoles creates predefined roles grouped by divisi.
func SeedRoles(db *gorm.DB, logger *zap.Logger) {
	type roleDef struct {
		Divisi string
		Roles  []string
	}

	definitions := []roleDef{
		{Divisi: "Management", Roles: []string{"Super Admin", "Owner"}},
		{Divisi: "Legal", Roles: []string{"Legal Admin"}},
		{Divisi: "Marketing", Roles: []string{"Customer Service", "Kepala Marketing"}},
		{Divisi: "Operasional", Roles: []string{"Surveyor", "Supervisor", "Project Manager"}},
		{Divisi: "Desain", Roles: []string{"Drafter", "Desainer"}},
		{Divisi: "Estimasi", Roles: []string{"Estimator"}},
	}

	var count int
	for _, def := range definitions {
		var divisi entity.Divisi
		db.Where("company_id = ? AND nama_divisi = ?", 1, def.Divisi).FirstOrCreate(&divisi, entity.Divisi{
			CompanyID:  1,
			NamaDivisi: def.Divisi,
		})

		for _, roleName := range def.Roles {
			var role entity.Role
			result := db.Where("company_id = ? AND nama_role = ?", 1, roleName).FirstOrCreate(&role, entity.Role{
				CompanyID: 1,
				NamaRole:  roleName,
				DivisiID:  divisi.ID,
			})
			if result.RowsAffected > 0 {
				count++
			}
		}
	}
	logger.Info("Roles seeded", zap.Int("new", count))

	// Assign invoice.* permissions to Legal Admin
	var legalRole entity.Role
	if err := db.Where("company_id = ? AND nama_role = ?", 1, "Legal Admin").First(&legalRole).Error; err == nil {
		var invoicePerms []entity.Permission
		if err := db.Where("name LIKE ?", "invoice.%").Find(&invoicePerms).Error; err == nil {
			for _, p := range invoicePerms {
				db.FirstOrCreate(&entity.RolePermission{}, entity.RolePermission{
					RoleID:       legalRole.ID,
					PermissionID: p.ID,
				})
			}
			logger.Info("Invoice permissions assigned to Legal Admin", zap.Int("count", len(invoicePerms)))
		}
	}
}

// SeedUsers ensures that every role in the database has at least 3 users.
func SeedUsers(db *gorm.DB, logger *zap.Logger) {
	var roles []entity.Role
	if err := db.Where("company_id = ?", 1).Find(&roles).Error; err != nil {
		logger.Error("Failed to fetch roles for seeding users", zap.Error(err))
		return
	}

	hashedPassword, _ := helper.HashPassword("password")

	var seededCount int
	for _, role := range roles {
		// Count existing users with this role
		var count int64
		if err := db.Model(&entity.User{}).Where("company_id = ? AND role_id = ?", 1, role.ID).Count(&count).Error; err != nil {
			logger.Error("Failed to count users for role", zap.String("role", role.NamaRole), zap.Error(err))
			continue
		}

		// If less than 3, seed more users
		for i := count; i < 3; i++ {
			// Format role name to be a clean email prefix
			roleClean := strings.ToLower(strings.ReplaceAll(role.NamaRole, " ", ""))
			email := fmt.Sprintf("%s%d@interiorpro.com", roleClean, i+1)
			name := fmt.Sprintf("%s User %d", role.NamaRole, i+1)

			user := entity.User{
				CompanyID: 1,
				Name:      name,
				Email:     email,
				Password:  hashedPassword,
				RoleID:    &role.ID,
			}

			// We use FirstOrCreate to avoid duplicates
			if err := db.Where("email = ?", email).FirstOrCreate(&user).Error; err != nil {
				logger.Error("Failed to seed user", zap.String("email", email), zap.Error(err))
			} else {
				seededCount++
			}
		}
	}
	logger.Info("Users seeding completed", zap.Int("seeded_new_users", seededCount))
}

// SeedItems seeds default master items for finishing dalam, luar, and aksesoris.
func SeedItems(db *gorm.DB, logger *zap.Logger) {
	defaultItems := []entity.Item{
		// Finishing Dalam
		{NamaItem: "Melamine Clear", JenisItem: entity.JenisFinishingDalam, Harga: 50000},
		{NamaItem: "HPL Putih Glossy", JenisItem: entity.JenisFinishingDalam, Harga: 85000},
		{NamaItem: "Cat Duco Putih", JenisItem: entity.JenisFinishingDalam, Harga: 120000},
		{NamaItem: "Melamine Doff", JenisItem: entity.JenisFinishingDalam, Harga: 55000},
		{NamaItem: "HPL Putih Doff", JenisItem: entity.JenisFinishingDalam, Harga: 75000},

		// Finishing Luar
		{NamaItem: "HPL Motif Kayu", JenisItem: entity.JenisFinishingLuar, Harga: 95000},
		{NamaItem: "HPL Abu-Abu Matte", JenisItem: entity.JenisFinishingLuar, Harga: 90000},
		{NamaItem: "Cat Duco Custom Color", JenisItem: entity.JenisFinishingLuar, Harga: 150000},
		{NamaItem: "HPL Hitam Tekstur", JenisItem: entity.JenisFinishingLuar, Harga: 100000},
		{NamaItem: "Melamine Semi Glossy", JenisItem: entity.JenisFinishingLuar, Harga: 60000},

		// Aksesoris
		{NamaItem: "Engsel Soft-Close (Pair)", JenisItem: entity.JenisAksesoris, Harga: 25000},
		{NamaItem: "Rel Laci Tandem (Set)", JenisItem: entity.JenisAksesoris, Harga: 75000},
		{NamaItem: "Tarikan Laci Minimalis Hitam", JenisItem: entity.JenisAksesoris, Harga: 15000},
		{NamaItem: "Kunci Laci", JenisItem: entity.JenisAksesoris, Harga: 20000},
		{NamaItem: "Hidrolik Pintu Atas (Gas Spring)", JenisItem: entity.JenisAksesoris, Harga: 35000},
		{NamaItem: "Lampu LED Strip (Meter)", JenisItem: entity.JenisAksesoris, Harga: 45000},
	}

	var count int
	for _, item := range defaultItems {
		var existing entity.Item
		result := db.Where("nama_item = ? AND jenis_item = ?", item.NamaItem, item.JenisItem).
			FirstOrCreate(&existing, item)
		if result.Error != nil {
			logger.Error("Failed to seed item", zap.String("nama", item.NamaItem), zap.Error(result.Error))
		} else if result.RowsAffected > 0 {
			count++
		}
	}
	logger.Info("Items seeded", zap.Int("new_items", count))
}

// SeedWorkplanStages seeds the 9 default workplan stages if they don't exist.
func SeedWorkplanStages(db *gorm.DB, logger *zap.Logger) {
	defaultStages := []entity.WorkplanStageMaster{
		{Code: "potong", Name: "Potong", Percentage: 10.0, SortOrder: 1},
		{Code: "rangkai", Name: "Rangkai", Percentage: 15.0, SortOrder: 2},
		{Code: "finishing", Name: "Finishing", Percentage: 20.0, SortOrder: 3},
		{Code: "finishing_qc", Name: "Finishing QC", Percentage: 10.0, SortOrder: 4},
		{Code: "packing", Name: "Packing", Percentage: 10.0, SortOrder: 5},
		{Code: "pengiriman", Name: "Pengiriman", Percentage: 10.0, SortOrder: 6},
		{Code: "trap", Name: "Trap", Percentage: 5.0, SortOrder: 7},
		{Code: "install", Name: "Install", Percentage: 15.0, SortOrder: 8},
		{Code: "install_qc", Name: "Install QC", Percentage: 5.0, SortOrder: 9},
	}

	var count int
	for _, stg := range defaultStages {
		var existing entity.WorkplanStageMaster
		result := db.Where("code = ?", stg.Code).
			Assign(entity.WorkplanStageMaster{
				Code:       stg.Code,
				Name:       stg.Name,
				Percentage: stg.Percentage,
				SortOrder:  stg.SortOrder,
			}).
			FirstOrCreate(&existing)
		if result.Error != nil {
			logger.Error("Failed to seed workplan stage master", zap.String("code", stg.Code), zap.Error(result.Error))
		} else if result.RowsAffected > 0 {
			count++
		}
	}
	logger.Info("Workplan stage masters seeded", zap.Int("new_or_updated_stages", count))
}

