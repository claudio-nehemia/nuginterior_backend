package routes

import (
	"net/http"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/handler"
	"github.com/claudio-nehemia/interior_backend/internal/middleware"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, services *service.Services, cfg *config.Config, cacheStore cache.Store) {
	handlers := handler.NewHandlers(services, cfg)
	
	// Serve static files for uploads
	router.Static("/uploads", cfg.UploadDir)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "ok"})
	})

	api := router.Group("/api")

	// ─── AUTH (public) ───
	auth := api.Group("/auth")
	auth.Use(middleware.RateLimiter(cacheStore, 20, 1*time.Minute))
	{
		auth.POST("/login", handlers.Auth.Login)
		auth.POST("/register", handlers.Auth.Register)
		auth.POST("/register-company", handlers.Auth.RegisterCompany)
		auth.POST("/refresh", handlers.Auth.RefreshToken)
	}

	// ─── AUTH (protected) ───
	authProtected := api.Group("/auth")
	authProtected.Use(middleware.JWTAuth(cfg, services.Auth))
	{
		authProtected.GET("/me", handlers.Auth.Me)
		authProtected.POST("/logout", handlers.Auth.Logout)
	}

	// ─── Protected API routes ───
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(cfg, services.Auth))
	{
		// DIVISI
		divisi := protected.Group("/divisi")
		{
			divisi.GET("", middleware.RequirePermission(services.Auth, constants.PermDivisiIndex), handlers.Divisi.Index)
			divisi.POST("", middleware.RequirePermission(services.Auth, constants.PermDivisiCreate), handlers.Divisi.Store)
			divisi.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermDivisiUpdate), handlers.Divisi.Update)
			divisi.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermDivisiDelete), handlers.Divisi.Destroy)
		}

		// ROLES
		roles := protected.Group("/roles")
		{
			roles.GET("", middleware.RequirePermission(services.Auth, constants.PermRoleIndex), handlers.Role.Index)
			roles.POST("", middleware.RequirePermission(services.Auth, constants.PermRoleCreate), handlers.Role.Store)
			roles.GET("/:id", middleware.RequirePermission(services.Auth, constants.PermRoleShow), handlers.Role.Show)
			roles.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermRoleUpdate), handlers.Role.Update)
			roles.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermRoleDelete), handlers.Role.Destroy)
			roles.POST("/:id/permissions", middleware.RequirePermission(services.Auth, constants.PermRoleSyncPermissions), handlers.Role.SyncPermissions)
		}

		// USERS
		users := protected.Group("/users")
		{
			users.GET("", middleware.RequirePermission(services.Auth, constants.PermUserIndex), handlers.User.Index)
			users.POST("", middleware.RequirePermission(services.Auth, constants.PermUserCreate), handlers.User.Store)
			users.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermUserUpdate), handlers.User.Update)
			users.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermUserDelete), handlers.User.Destroy)
		}

		// PERMISSIONS
		protected.GET("/permissions", middleware.RequirePermission(services.Auth, constants.PermPermissionIndex), handlers.Permission.Index)

		// PRODUK
		produk := protected.Group("/produk")
		{
			produk.GET("", middleware.RequireAnyPermission(services.Auth, constants.PermProdukIndex, constants.PermRABIndex, constants.PermRABCreate, constants.PermRABUpdate), handlers.Produk.Index)
			produk.POST("", middleware.RequirePermission(services.Auth, constants.PermProdukCreate), handlers.Produk.Store)
			produk.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermProdukUpdate), handlers.Produk.Update)
			produk.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermProdukDelete), handlers.Produk.Destroy)
			produk.DELETE("/:id/image/:imageId", middleware.RequirePermission(services.Auth, constants.PermProdukDeleteImage), handlers.Produk.DestroyImage)
		}

		// BAHAN BAKU
		bahanBaku := protected.Group("/bahan-baku")
		{
			bahanBaku.GET("", middleware.RequireAnyPermission(services.Auth, constants.PermBahanBakuIndex, constants.PermProdukIndex, constants.PermProdukCreate, constants.PermProdukUpdate), handlers.BahanBaku.Index)
			bahanBaku.POST("", middleware.RequirePermission(services.Auth, constants.PermBahanBakuCreate), handlers.BahanBaku.Store)
			bahanBaku.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermBahanBakuUpdate), handlers.BahanBaku.Update)
			bahanBaku.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermBahanBakuDelete), handlers.BahanBaku.Destroy)
		}

		// ITEMS
		items := protected.Group("/items")
		{
			items.GET("", middleware.RequireAnyPermission(services.Auth, constants.PermItemIndex, constants.PermRABIndex, constants.PermRABCreate, constants.PermRABUpdate), handlers.Item.Index)
			items.POST("", middleware.RequirePermission(services.Auth, constants.PermItemCreate), handlers.Item.Store)
			items.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermItemUpdate), handlers.Item.Update)
			items.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermItemDelete), handlers.Item.Destroy)
		}

		// JENIS PENGUKURAN
		jenisPengukuran := protected.Group("/jenis-pengukuran")
		{
			jenisPengukuran.GET("", middleware.RequirePermission(services.Auth, constants.PermJenisPengukuranIndex), handlers.JenisPengukuran.Index)
			jenisPengukuran.POST("", middleware.RequirePermission(services.Auth, constants.PermJenisPengukuranCreate), handlers.JenisPengukuran.Store)
			jenisPengukuran.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermJenisPengukuranUpdate), handlers.JenisPengukuran.Update)
			jenisPengukuran.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermJenisPengukuranDelete), handlers.JenisPengukuran.Destroy)
		}

		// TERMIN
		termin := protected.Group("/termin")
		{
			termin.GET("", middleware.RequirePermission(services.Auth, constants.PermTerminIndex), handlers.Termin.Index)
			termin.POST("", middleware.RequirePermission(services.Auth, constants.PermTerminCreate), handlers.Termin.Store)
			termin.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermTerminUpdate), handlers.Termin.Update)
			termin.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermTerminDelete), handlers.Termin.Destroy)
		}

		// UPLOAD
		protected.POST("/upload", middleware.RequirePermission(services.Auth, constants.PermUpload), handlers.Upload.UploadImage)

		// DASHBOARD
		protected.GET("/dashboard/stats", middleware.RequirePermission(services.Auth, constants.PermOrderIndex), handlers.Dashboard.GetStats)
		protected.GET("/dashboard/stats/ai-analysis", middleware.RequirePermission(services.Auth, constants.PermOrderIndex), handlers.AI.GetGlobalAnalysis)
		protected.POST("/ai/transcribe", handlers.AI.Transcribe)

		// ORDERS
		orders := protected.Group("/orders")
		{
			orders.GET("", middleware.RequirePermission(services.Auth, constants.PermOrderIndex), handlers.Order.Index)
			orders.GET("/export/pdf", middleware.RequirePermission(services.Auth, constants.PermOrderIndex), handlers.Order.ExportPDF)
			orders.GET("/export/excel", middleware.RequirePermission(services.Auth, constants.PermOrderIndex), handlers.Order.ExportExcel)
			orders.POST("", middleware.RequirePermission(services.Auth, constants.PermOrderCreate), handlers.Order.Store)
			orders.GET("/:id", middleware.RequirePermission(services.Auth, constants.PermOrderShow), handlers.Order.Show)
			orders.GET("/:id/ai-health", middleware.RequirePermission(services.Auth, constants.PermOrderShow), handlers.AI.GetProjectHealth)
			orders.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermOrderUpdate), handlers.Order.Update)
			orders.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermOrderDelete), handlers.Order.Destroy)
			orders.POST("/:id/teams", middleware.RequirePermission(services.Auth, constants.PermOrderUpdate), handlers.Order.SyncTeams)
			orders.GET("/:id/teams", middleware.RequirePermission(services.Auth, constants.PermOrderShow), handlers.Order.GetTeams)
			orders.POST("/:id/moodboard/response", middleware.RequirePermission(services.Auth, constants.PermMoodboardResponse), handlers.Moodboard.ResponseMoodboard)
		}

		// SURVEYS
		surveys := protected.Group("/surveys")
		{
			surveys.GET("", middleware.RequirePermission(services.Auth, constants.PermSurveyIndex), handlers.Survey.Index)
			surveys.POST("", middleware.RequirePermission(services.Auth, constants.PermSurveyCreate), handlers.Survey.Store)
			surveys.GET("/:id", middleware.RequirePermission(services.Auth, constants.PermSurveyShow), handlers.Survey.Show)
			surveys.GET("/:id/ai-summary", middleware.RequirePermission(services.Auth, constants.PermSurveyShow), handlers.AI.GetSurveySummary)
			surveys.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermSurveyUpdate), handlers.Survey.Update)
			surveys.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermSurveyDelete), handlers.Survey.Destroy)
			surveys.POST("/:id/response", middleware.RequirePermission(services.Auth, constants.PermSurveyResponse), handlers.Survey.Response)
			surveys.POST("/:id/marketing-response", middleware.RequirePermission(services.Auth, constants.PermSurveyMarketingResponse), handlers.Survey.MarketingResponse)
		}

		// MOODBOARDS
		moodboards := protected.Group("/moodboards")
		{
			moodboards.GET("", middleware.RequirePermission(services.Auth, constants.PermMoodboardIndex), handlers.Moodboard.Index)
			moodboards.POST("/upload-kasar", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.Moodboard.UploadKasar)
			moodboards.POST("/:id/accept-desain", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.Moodboard.AcceptDesain)
			moodboards.POST("/:id/revise", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.Moodboard.Revise)
			moodboards.DELETE("/files/:file_id", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.Moodboard.DeleteFile)
			
			// Estimasi Response
			moodboards.POST("/:id/estimasi/response", middleware.RequirePermission(services.Auth, constants.PermMoodboardResponse), handlers.Moodboard.ResponseEstimasi)
			
			// Commitment Fee Response
			moodboards.POST("/:id/commitment-fee/response", middleware.RequirePermission(services.Auth, constants.PermMoodboardResponse), handlers.Moodboard.ResponseCommitmentFee)
			
		}

		// DESAIN FINALS
		desainFinals := protected.Group("/desain-finals")
		{
			desainFinals.GET("", middleware.RequirePermission(services.Auth, constants.PermMoodboardIndex), handlers.DesainFinal.Index)
			desainFinals.POST("/:id/response", middleware.RequirePermission(services.Auth, constants.PermMoodboardResponse), handlers.DesainFinal.Response)
			desainFinals.POST("/upload", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.DesainFinal.Upload)
			desainFinals.POST("/:id/accept", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.DesainFinal.Accept)
			desainFinals.POST("/:id/revise", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.DesainFinal.Revise)
			desainFinals.DELETE("/files/:file_id", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.DesainFinal.DeleteFile)
		}

		// GAMBAR KERJA
		gambarKerja := protected.Group("/gambar-kerja")
		{
			gambarKerja.GET("", middleware.RequirePermission(services.Auth, constants.PermMoodboardIndex), handlers.GambarKerja.Index)
			gambarKerja.POST("/:id/response", middleware.RequirePermission(services.Auth, constants.PermMoodboardResponse), handlers.GambarKerja.Response)
			gambarKerja.POST("/upload", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.GambarKerja.Upload)
			gambarKerja.POST("/files/:file_id/revise", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.GambarKerja.ReviseFile)
			gambarKerja.POST("/:id/revise-general", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.GambarKerja.ReviseGeneral)
			gambarKerja.POST("/:id/approve", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.GambarKerja.Approve)
			gambarKerja.DELETE("/files/:file_id", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.GambarKerja.DeleteFile)
		}

		// APPROVAL MATERIAL
		approvalMaterials := protected.Group("/approval-materials")
		{
			approvalMaterials.GET("", middleware.RequirePermission(services.Auth, constants.PermMoodboardIndex), handlers.ApprovalMaterial.Index)
			approvalMaterials.POST("/:id/response", middleware.RequirePermission(services.Auth, constants.PermMoodboardResponse), handlers.ApprovalMaterial.Response)
			approvalMaterials.GET("/order/:orderId", middleware.RequirePermission(services.Auth, constants.PermMoodboardIndex), handlers.ApprovalMaterial.Show)
			approvalMaterials.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.ApprovalMaterial.Update)
			approvalMaterials.GET("/:id/pdf", middleware.RequirePermission(services.Auth, constants.PermMoodboardIndex), handlers.ApprovalMaterial.ExportPDF)
		}

		// WORKPLAN
		workplans := protected.Group("/workplans")
		{
			workplans.GET("", middleware.RequirePermission(services.Auth, constants.PermWorkplanIndex), handlers.Workplan.Index)
			workplans.POST("/:id/response", middleware.RequirePermission(services.Auth, constants.PermWorkplanResponse), handlers.Workplan.Response)
			workplans.GET("/order/:orderId", middleware.RequirePermission(services.Auth, constants.PermWorkplanIndex), handlers.Workplan.Show)
			workplans.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermWorkplanUpdate), handlers.Workplan.Update)
			workplans.GET("/:id/excel", middleware.RequirePermission(services.Auth, constants.PermWorkplanIndex), handlers.Workplan.ExportExcel)
			workplans.GET("/:id/progress/pdf", middleware.RequirePermission(services.Auth, constants.PermWorkplanIndex), handlers.Workplan.ExportProgressPDF)
			workplans.GET("/:id/progress/excel", middleware.RequirePermission(services.Auth, constants.PermWorkplanIndex), handlers.Workplan.ExportProgressExcel)
			workplans.POST("/stages/:id/complete", middleware.RequirePermission(services.Auth, constants.PermWorkplanUpdate), handlers.Workplan.CompleteStage)
			workplans.POST("/:id/request-extension", middleware.RequirePermission(services.Auth, constants.PermWorkplanUpdate), handlers.Workplan.RequestExtension)
			workplans.POST("/:id/handle-extension", middleware.RequirePermission(services.Auth, constants.PermWorkplanUpdate), handlers.Workplan.HandleExtension)
			// Defect Management
			workplans.POST("/stages/:id/defects", middleware.RequirePermission(services.Auth, constants.PermWorkplanUpdate), handlers.Workplan.ReportDefect)
			workplans.PUT("/defects/:defectId/fix", middleware.RequirePermission(services.Auth, constants.PermWorkplanUpdate), handlers.Workplan.SubmitDefectFix)
			workplans.PUT("/defects/:defectId/review", middleware.RequirePermission(services.Auth, constants.PermWorkplanUpdate), handlers.Workplan.ReviewDefect)
			workplans.GET("/:id/defects", middleware.RequirePermission(services.Auth, constants.PermWorkplanIndex), handlers.Workplan.GetDefects)
			// BAST
			workplans.POST("/:id/bast", middleware.RequirePermission(services.Auth, constants.PermWorkplanUpdate), handlers.Workplan.SubmitBast)
			workplans.GET("/:id/bast/pdf", middleware.RequirePermission(services.Auth, constants.PermWorkplanIndex), handlers.Workplan.GenerateBastPDF)
		}

		// ESTIMASI
		estimasi := protected.Group("/estimasi")
		{
			estimasi.POST("/upload", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.Moodboard.UploadEstimasi)
		}

		// COMMITMENT FEE
		commitmentFees := protected.Group("/commitment-fees")
		{
			commitmentFees.PUT("/:id/total", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.Moodboard.UpdateTotalFee)
			commitmentFees.POST("/:id/payment", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.Moodboard.UploadPaymentProof)
			commitmentFees.POST("/:id/verify", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.Moodboard.VerifyPayment)
			commitmentFees.POST("/:id/reset", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.Moodboard.ResetPayment)
			commitmentFees.POST("/:id/revise", middleware.RequirePermission(services.Auth, constants.PermMoodboardUpdate), handlers.Moodboard.RevisePaymentFee)
			commitmentFees.GET("/:id/print", middleware.RequirePermission(services.Auth, constants.PermMoodboardIndex), handlers.Moodboard.PrintInvoice)
		}

		// INPUT ITEMS
		inputItems := protected.Group("/input-items")
		{
			inputItems.GET("", middleware.RequireAnyPermission(services.Auth, constants.PermInputItemIndex, constants.PermRABIndex, constants.PermRABCreate, constants.PermRABUpdate), handlers.InputItem.Index)
			inputItems.GET("/:id", middleware.RequireAnyPermission(services.Auth, constants.PermInputItemShow, constants.PermInputItemIndex, constants.PermRABCreate, constants.PermRABUpdate, constants.PermRABIndex), handlers.InputItem.Show)
			inputItems.GET("/desain-final/:df_id", middleware.RequirePermission(services.Auth, constants.PermInputItemShow), handlers.InputItem.ShowByDesainFinal)
			inputItems.POST("", middleware.RequirePermission(services.Auth, constants.PermInputItemCreate), handlers.InputItem.Store)
			inputItems.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermInputItemUpdate), handlers.InputItem.Update)
			inputItems.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermInputItemDelete), handlers.InputItem.Destroy)
			inputItems.POST("/desain-final/:df_id/response/designer", middleware.RequirePermission(services.Auth, constants.PermMoodboardResponse), handlers.InputItem.InputItemResponseDesigner)
			inputItems.POST("/desain-final/:df_id/response/marketing", middleware.RequirePermission(services.Auth, constants.PermMoodboardResponse), handlers.InputItem.InputItemResponseMarketing)
		}

		// RAB
		rab := protected.Group("/rab")
		{
			rab.GET("", middleware.RequirePermission(services.Auth, constants.PermRABIndex), handlers.RAB.Index)
			rab.GET("/:id", middleware.RequirePermission(services.Auth, constants.PermRABShow), handlers.RAB.Show)
			rab.GET("/input-item/:input_item_id", middleware.RequirePermission(services.Auth, constants.PermRABShow), handlers.RAB.ShowByInputItemID)
			rab.POST("", middleware.RequirePermission(services.Auth, constants.PermRABCreate), handlers.RAB.Store)
			rab.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermRABUpdate), handlers.RAB.Update)
			rab.DELETE("/:id", middleware.RequirePermission(services.Auth, constants.PermRABDelete), handlers.RAB.Destroy)
			rab.POST("/:id/submit", middleware.RequirePermission(services.Auth, constants.PermRABSubmit), handlers.RAB.Submit)
			rab.GET("/:id/export", middleware.RequirePermission(services.Auth, constants.PermRABShow), handlers.RAB.Export)
		}

		// CONTRACTS
		contracts := protected.Group("/contracts")
		{
			contracts.GET("", middleware.RequirePermission(services.Auth, constants.PermContractIndex), handlers.Contract.Index)
			contracts.POST("/rab/:rabId/response", middleware.RequirePermission(services.Auth, constants.PermContractResponse), handlers.Contract.Response)
			contracts.POST("", middleware.RequirePermission(services.Auth, constants.PermContractCreate), handlers.Contract.Store)
			contracts.GET("/:id/pdf", middleware.RequirePermission(services.Auth, constants.PermContractIndex), handlers.Contract.ExportPDF)
			contracts.POST("/:id/upload-signed", middleware.RequirePermission(services.Auth, constants.PermContractUpdate), handlers.Contract.UploadSigned)
		}

		// INVOICES
		invoices := protected.Group("/invoices")
		{
			invoices.GET("", middleware.RequirePermission(services.Auth, constants.PermInvoiceIndex), handlers.Invoice.Index)
			invoices.GET("/contract/:contractId", middleware.RequirePermission(services.Auth, constants.PermInvoiceShow), handlers.Invoice.Show)
			invoices.POST("/contract/:contractId/response", middleware.RequirePermission(services.Auth, constants.PermInvoiceResponse), handlers.Invoice.Response)
			invoices.POST("/contract/:contractId/generate", middleware.RequirePermission(services.Auth, constants.PermInvoiceCreate), handlers.Invoice.Generate)
			invoices.PUT("/:id/deadline", middleware.RequirePermission(services.Auth, constants.PermInvoiceUpdateDeadline), handlers.Invoice.UpdateDeadline)
			invoices.POST("/:id/payment", middleware.RequirePermission(services.Auth, constants.PermInvoiceUploadProof), handlers.Invoice.UploadPayment)
			invoices.GET("/:id/pdf", middleware.RequirePermission(services.Auth, constants.PermInvoiceDownload), handlers.Invoice.ExportPDF)
		}

		protected.GET("/sidebar", handlers.Setting.GetSidebarConfig)

		// LOG TASKS
		protected.GET("/log-tasks", middleware.RequirePermission(services.Auth, constants.PermLogTaskIndex), handlers.ProjectLogTask.Index)

		// NOTIFICATIONS
		notifications := protected.Group("/notifications")
		{
			notifications.GET("", handlers.Notification.Index)
			notifications.PUT("/:id/read", handlers.Notification.MarkAsRead)
			notifications.PUT("/read-all", handlers.Notification.MarkAllAsRead)
			notifications.GET("/unread-count", handlers.Notification.GetUnreadCount)
		}

		// COMPANIES
		companies := protected.Group("/companies")
		{
			companies.GET("", middleware.RequirePermission(services.Auth, constants.PermCompanyIndex), handlers.Company.Index)
			companies.GET("/:id", middleware.RequirePermission(services.Auth, constants.PermCompanyIndex), handlers.Company.Show)
			companies.PUT("/:id", middleware.RequirePermission(services.Auth, constants.PermCompanyUpdate), handlers.Company.Update)
			companies.PUT("/:id/verify", middleware.RequirePermission(services.Auth, constants.PermCompanyVerify), handlers.Company.Verify)
			companies.PUT("/:id/reject", middleware.RequirePermission(services.Auth, constants.PermCompanyVerify), handlers.Company.Reject)
		}

		// SETTINGS
		settings := protected.Group("/settings")
		{
			settings.GET("", handlers.Setting.Index)
			settings.PUT("/:key", middleware.RequirePermission(services.Auth, constants.PermSettingUpdate), handlers.Setting.Update)
			settings.GET("/workplan-stages", middleware.RequirePermission(services.Auth, constants.PermSettingIndex), handlers.Workplan.GetStageMasters)
			settings.PUT("/workplan-stages", middleware.RequirePermission(services.Auth, constants.PermSettingUpdate), handlers.Workplan.UpdateStageMasters)
		}
	}
}
