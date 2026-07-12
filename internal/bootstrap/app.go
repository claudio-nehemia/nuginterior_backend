package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/middleware"
	"github.com/claudio-nehemia/interior_backend/internal/routes"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"github.com/claudio-nehemia/interior_backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	Router   *gin.Engine
	Services *service.Services
	Config   *config.Config
	Cache    cache.Store
	Logger   *zap.Logger
}

func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log, err := logger.New()
	if err != nil {
		return nil, err
	}

	db, err := database.Connect(cfg)
	if err != nil {
		return nil, err
	}

	cacheStore := cache.NewRedis(cfg.RedisAddr(), cfg.RedisPassword)

	// Test Redis connection
	if err := cacheStore.Ping(context.Background()); err != nil {
		log.Warn("Redis connection failed, running without cache", zap.Error(err))
	} else {
		log.Info("Redis connected successfully")
	}

	// Run database migrations
	if err := database.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}
	log.Info("Database migration completed")

	// Seed permissions
	database.SeedPermissions(db, log)
	database.SeedAdminUser(db, log)
	database.SeedSettings(db, log)
	database.SeedRoles(db, log)
	database.SeedUsers(db, log)
	database.SeedItems(db, log)
	database.SeedWorkplanStages(db, log)
	database.SyncAllRolePermissions(db, log)
	database.SyncPostgresSequences(db, log)

	// Flush role permission and settings cache so newly-seeded data takes effect immediately
	_ = cacheStore.DeletePattern(context.Background(), constants.KeyRolePermissions+"*")
	_ = cacheStore.Del(context.Background(), constants.KeySettingAll)
	log.Info("Role permission and settings cache cleared")

	services := service.NewServices(service.Dependencies{
		Config: cfg,
		DB:     db,
		Cache:  cacheStore,
		Logger: log,
	})

	router := gin.New()
	router.MaxMultipartMemory = 10 << 20 // 10 MB
	router.Use(gin.Recovery(), gin.Logger())
	router.Use(middleware.CORS())

	routes.Register(router, services, cfg, cacheStore)

	return &App{
		Router:   router,
		Services: services,
		Config:   cfg,
		Cache:    cacheStore,
		Logger:   log,
	}, nil
}

func (a *App) Run() error {
	a.Logger.Info("Server starting", zap.String("port", a.Config.AppPort))

	// Start project deadlines checker for H-1 warnings
	go func() {
		time.Sleep(5 * time.Second) // wait for server boot
		a.Logger.Info("Running initial deadline checker")
		_ = a.Services.Notification.CheckDeadlines(context.Background())

		ticker := time.NewTicker(30 * time.Minute)
		for range ticker.C {
			a.Logger.Info("Running scheduled deadline checker")
			_ = a.Services.Notification.CheckDeadlines(context.Background())
		}
	}()

	return a.Router.Run(a.Config.Addr())
}
