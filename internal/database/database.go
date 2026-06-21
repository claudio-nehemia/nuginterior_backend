package database

import (
	"fmt"
	"log"

	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// createDBIfNotExists connects to default postgres db and creates target db if not found
func createDBIfNotExists(cfg *config.Config) error {
	defaultDB, err := gorm.Open(postgres.Open(cfg.DBDSNDefault()), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := defaultDB.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	// Check if database exists
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s')", cfg.DBName)
	err = sqlDB.QueryRow(query).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		log.Printf("Database %s does not exist, creating it...", cfg.DBName)
		_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", cfg.DBName))
		if err != nil {
			return err
		}
		log.Printf("Database %s created successfully.", cfg.DBName)
	}
	return nil
}

func Connect(cfg *config.Config) (*gorm.DB, error) {
	if err := createDBIfNotExists(cfg); err != nil {
		log.Printf("Failed to check/create database: %v\n", err)
		// We still try to connect to the target database in case the user does not have permission to create DB
		// but the DB already exists and was accessible.
	}

	return gorm.Open(postgres.Open(cfg.DBDSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
}

// AutoMigrate runs GORM auto migration for all entity models.
func AutoMigrate(db *gorm.DB) error {
	// Drop old unique constraints to prevent PostgreSQL errors when migrating to composite indexes
	_ = db.Exec("ALTER TABLE settings DROP CONSTRAINT IF EXISTS settings_key_key")
	_ = db.Exec("DROP INDEX IF EXISTS uni_settings_key")
	_ = db.Exec("DROP INDEX IF EXISTS idx_settings_key")

	_ = db.Exec("ALTER TABLE roles DROP CONSTRAINT IF EXISTS roles_nama_role_key")
	_ = db.Exec("DROP INDEX IF EXISTS uni_roles_nama_role")
	_ = db.Exec("DROP INDEX IF EXISTS idx_roles_nama_role")

	err := db.AutoMigrate(
		&entity.Company{},
		&entity.Divisi{},
		&entity.Role{},
		&entity.Permission{},
		&entity.RolePermission{},
		&entity.User{},
		&entity.Produk{},
		&entity.ProdukImage{},
		&entity.Item{},
		&entity.BahanBaku{},
		&entity.ProdukBahanBaku{},
		&entity.JenisPengukuran{},
		&entity.Termin{},
		&entity.Order{},
		&entity.OrderTeam{},
		&entity.OrderProduk{},
		&entity.OrderItem{},
		&entity.OrderPembayaran{},
		&entity.Survey{},
		&entity.SurveyPengukuran{},
		&entity.Moodboard{},
		&entity.MoodboardFile{},
		&entity.Estimasi{},
		&entity.EstimasiFile{},
		&entity.CommitmentFee{},
		&entity.Setting{},
		&entity.DesainFinal{},
		&entity.DesainFinalFile{},
		&entity.GambarKerja{},
		&entity.GambarKerjaFile{},
		&entity.InputItem{},
		&entity.InputItemRoom{},
		&entity.InputItemRoomBahanBaku{},
		&entity.InputItemRoomFinishing{},
		&entity.InputItemRoomAksesoris{},
		&entity.RAB{},
		&entity.RABRoom{},
		&entity.RABRoomBahanBaku{},
		&entity.RABRoomFinishing{},
		&entity.RABRoomAksesoris{},
		&entity.Contract{},
		&entity.Invoice{},
		&entity.ApprovalMaterial{},
		&entity.ApprovalMaterialItem{},
		&entity.Workplan{},
		&entity.WorkplanStageMaster{},
		&entity.WorkplanStage{},
		&entity.WorkplanDefect{},
		&entity.ProjectLogTask{},
		&entity.Notification{},
	)
	if err != nil {
		return err
	}

	// Manually alter settings.value to TEXT because GORM AutoMigrate doesn't alter column types in PostgreSQL by default
	_ = db.Exec("ALTER TABLE settings ALTER COLUMN value TYPE text")
	_ = db.Exec("DELETE FROM settings WHERE key = 'sidebar_configuration' AND (value NOT LIKE '%project_management%' OR value NOT LIKE '%log_task%' OR value NOT LIKE '%companies%')")

	return nil
}
