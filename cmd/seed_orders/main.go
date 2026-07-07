package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("=========================================")
	fmt.Println("    NUGINTERIOR ORDER SEEDER FOR VPS     ")
	fmt.Println("=========================================")

	// Load environment variables
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	// Auto-route database connection when running on the host VPS.
	// The database container exposes port 5433 on localhost of the host VPS.
	if os.Getenv("DB_HOST") == "" || os.Getenv("DB_HOST") == "db" {
		os.Setenv("DB_HOST", "127.0.0.1")
		fmt.Println("[Config] Redirecting DB_HOST to 127.0.0.1 for host execution")
	}
	if os.Getenv("DB_PORT") == "" || os.Getenv("DB_PORT") == "5432" {
		os.Setenv("DB_PORT", "5433")
		fmt.Println("[Config] Redirecting DB_PORT to 5433 for host execution")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[Error] Failed to load configuration: %v", err)
	}

	fmt.Printf("[Config] Connecting to database '%s' on %s:%s...\n", cfg.DBName, cfg.DBHost, cfg.DBPort)

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("[Error] Failed to connect to database: %v\nDSN: %s", err, cfg.DBDSN())
	}

	fmt.Println("[Database] Connected successfully!")

	// 1. Ensure company with ID = 1 exists
	var company entity.Company
	if err := db.First(&company, 1).Error; err != nil {
		fmt.Println("[Database] Seed company PT. INTERIORPRO INDONESIA...")
		company = entity.Company{
			ID:           1,
			Name:         "PT. INTERIORPRO INDONESIA",
			DirectorName: "Super Admin",
			CeoNik:       "1234567890123456",
			Email:        "admin@interiorpro.com",
			Phone:        "+62 812-0000-0000",
			Status:       "verified",
		}
		if err := db.Create(&company).Error; err != nil {
			log.Fatalf("[Error] Failed to seed default company: %v", err)
		}
	}

	// 2. Define mock orders
	now := time.Now()
	t1 := now.AddDate(0, 0, -5)
	t2 := now.AddDate(0, 0, -2)

	orders := []entity.Order{
		{
			CompanyID:            1,
			NomorOrder:           "ORD-2026-0001",
			NamaProject:          "Apartemen Sudirman Mansion Renovasi",
			JenisInterior:        constants.JenisInteriorApartment,
			NamaCustomer:         "Budi Santoso",
			TeleponCustomer:      "081234567890",
			EmailCustomer:        "budi@gmail.com",
			ProjectStatus:        constants.ProjectStatusPending,
			PriorityLevel:        constants.PriorityHigh,
			TahapanProyek:        constants.TahapanSurvey,
			PaymentStatus:        constants.PaymentNotStart,
			TanggalMasukCustomer: &t1,
		},
		{
			CompanyID:            1,
			NomorOrder:           "ORD-2026-0002",
			NamaProject:          "Kitchen Set Modern Pantai Indah Kapuk",
			JenisInterior:        constants.JenisInteriorResidential,
			NamaCustomer:         "Jessica Wijaya",
			TeleponCustomer:      "081987654321",
			EmailCustomer:        "jessica@outlook.com",
			ProjectStatus:        constants.ProjectStatusInProgress,
			PriorityLevel:        constants.PriorityMedium,
			TahapanProyek:        constants.TahapanMoodboard,
			PaymentStatus:        constants.PaymentNotStart,
			TanggalMasukCustomer: &t2,
		},
		{
			CompanyID:            1,
			NomorOrder:           "ORD-2026-0003",
			NamaProject:          "Office Lobby PT Tech Startup BSD",
			JenisInterior:        constants.JenisInteriorOffice,
			NamaCustomer:         "Ahmad Hidayat",
			TeleponCustomer:      "081122334455",
			EmailCustomer:        "ahmad@techstartup.com",
			NamaPerusahaan:       "PT. Tech Startup Indonesia",
			ProjectStatus:        constants.ProjectStatusDeal,
			PriorityLevel:        constants.PriorityMedium,
			TahapanProyek:        constants.TahapanDesainFinal,
			PaymentStatus:        constants.PaymentCmFee,
			TanggalMasukCustomer: &now,
		},
	}

	// 3. Seed orders
	for _, o := range orders {
		var existing entity.Order
		err := db.Where("nomor_order = ?", o.NomorOrder).First(&existing).Error
		if err == nil {
			fmt.Printf("[Database] Order %s already exists. Skipping.\n", o.NomorOrder)
			continue
		}

		if err := db.Create(&o).Error; err != nil {
			fmt.Printf("[Error] Failed to create order %s: %v\n", o.NomorOrder, err)
		} else {
			fmt.Printf("[Database] Seeded order %s (%s) successfully.\n", o.NomorOrder, o.NamaProject)
		}
	}

	fmt.Println("=========================================")
	fmt.Println("    SEEDING COMPLETED SUCCESSFULLY!      ")
	fmt.Println("=========================================")
}
