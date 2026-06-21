package service

import (
	"context"
	"fmt"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/constants"
	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DashboardService interface {
	GetStats(ctx context.Context) (*dto.DashboardStatsResponse, error)
}

type dashboardService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewDashboardService(db *gorm.DB, logger *zap.Logger) DashboardService {
	return &dashboardService{db: db, logger: logger}
}

func (s *dashboardService) GetStats(ctx context.Context) (*dto.DashboardStatsResponse, error) {
	companyID, _ := ctx.Value(constants.ContextKeyCompanyID).(uint)
	role, _ := ctx.Value(constants.ContextKeyUserRole).(string)

	applyScope := func(query *gorm.DB) *gorm.DB {
		if companyID == 1 && role == "Super Admin" {
			if filterVal := ctx.Value(constants.ContextKeyFilterCompanyID); filterVal != nil {
				if filterID, ok := filterVal.(uint); ok && filterID > 0 {
					return query.Where("orders.company_id = ?", filterID)
				}
				return query
			}
			return query
		}
		return query.Where("orders.company_id = ?", companyID)
	}

	var totalOrders int64
	qTotal := s.db.WithContext(ctx).Table("orders")
	qTotal = applyScope(qTotal)
	if err := qTotal.Count(&totalOrders).Error; err != nil {
		s.logger.Error("failed to get total orders", zap.Error(err))
		return nil, err
	}

	var activeOrders int64
	qActive := s.db.WithContext(ctx).Table("orders").
		Where("orders.tahapan_proyek NOT IN (?) AND orders.project_status != ?", []string{"selesai", "batal"}, "cancel")
	qActive = applyScope(qActive)
	if err := qActive.Count(&activeOrders).Error; err != nil {
		s.logger.Error("failed to get active orders", zap.Error(err))
		return nil, err
	}

	var completedProjects int64
	qCompleted := s.db.WithContext(ctx).Table("orders").
		Where("orders.tahapan_proyek = ?", "selesai")
	qCompleted = applyScope(qCompleted)
	if err := qCompleted.Count(&completedProjects).Error; err != nil {
		s.logger.Error("failed to get completed projects", zap.Error(err))
		return nil, err
	}

	var successRate float64
	if totalOrders > 0 {
		successRate = (float64(completedProjects) * 100.0) / float64(totalOrders)
	}

	var totalContractsDeal int64
	qContracts := s.db.WithContext(ctx).Table("contracts").
		Joins("JOIN orders ON orders.id = contracts.order_id").
		Where("contracts.status = ?", "deal")
	qContracts = applyScope(qContracts)
	if err := qContracts.Count(&totalContractsDeal).Error; err != nil {
		s.logger.Error("failed to get total contracts deal", zap.Error(err))
		return nil, err
	}

	var lunasCount int64
	subQuery := s.db.Table("invoices").
		Select("DISTINCT invoices.order_id").
		Joins("JOIN orders ON orders.id = invoices.order_id").
		Where("invoices.status = ?", "belum_bayar")
	subQuery = applyScope(subQuery)

	qLunasCount := s.db.WithContext(ctx).Table("invoices").
		Select("COUNT(DISTINCT invoices.order_id)").
		Joins("JOIN orders ON orders.id = invoices.order_id")
	qLunasCount = applyScope(qLunasCount).
		Where("invoices.order_id NOT IN (?)", subQuery)

	if err := qLunasCount.Scan(&lunasCount).Error; err != nil {
		s.logger.Error("failed to get lunas count", zap.Error(err))
		return nil, err
	}

	var lunasAmount float64
	qLunasAmount := s.db.WithContext(ctx).Table("invoices").
		Joins("JOIN orders ON orders.id = invoices.order_id").
		Where("invoices.status = ?", "terbayar").
		Select("COALESCE(SUM(invoices.amount), 0)")
	qLunasAmount = applyScope(qLunasAmount)
	if err := qLunasAmount.Row().Scan(&lunasAmount); err != nil {
		s.logger.Error("failed to get lunas amount", zap.Error(err))
		return nil, err
	}

	var belumBayarCount int64
	qBelumBayarCount := s.db.WithContext(ctx).Table("invoices").
		Joins("JOIN orders ON orders.id = invoices.order_id").
		Where("invoices.status = ?", "belum_bayar").
		Select("COUNT(DISTINCT invoices.order_id)")
	qBelumBayarCount = applyScope(qBelumBayarCount)
	if err := qBelumBayarCount.Scan(&belumBayarCount).Error; err != nil {
		s.logger.Error("failed to get belum bayar count", zap.Error(err))
		return nil, err
	}

	var belumBayarAmount float64
	qBelumBayarAmount := s.db.WithContext(ctx).Table("invoices").
		Joins("JOIN orders ON orders.id = invoices.order_id").
		Where("invoices.status = ?", "belum_bayar").
		Select("COALESCE(SUM(invoices.amount), 0)")
	qBelumBayarAmount = applyScope(qBelumBayarAmount)
	if err := qBelumBayarAmount.Row().Scan(&belumBayarAmount); err != nil {
		s.logger.Error("failed to get belum bayar amount", zap.Error(err))
		return nil, err
	}

	var totalOmset float64
	qTotalOmset := s.db.WithContext(ctx).Table("invoices").
		Joins("JOIN orders ON orders.id = invoices.order_id").
		Select("COALESCE(SUM(invoices.amount), 0)")
	qTotalOmset = applyScope(qTotalOmset)
	if err := qTotalOmset.Row().Scan(&totalOmset); err != nil {
		s.logger.Error("failed to get total omset", zap.Error(err))
		return nil, err
	}

	// Fetch top 5 recent orders
	type orderResult struct {
		ID            uint
		NomorOrder    string
		NamaProject   string
		NamaCustomer  string
		ProjectStatus string
		TahapanProyek string
		HargaKontrak  float64
		CreatedAt     time.Time
	}
	var recentOrdersRaw []orderResult

	whereClause := ""
	var args []interface{}
	if companyID == 1 && role == "Super Admin" {
		if filterVal := ctx.Value(constants.ContextKeyFilterCompanyID); filterVal != nil {
			if filterID, ok := filterVal.(uint); ok && filterID > 0 {
				whereClause = "WHERE o.company_id = ?"
				args = append(args, filterID)
			}
		}
	} else {
		whereClause = "WHERE o.company_id = ?"
		args = append(args, companyID)
	}

	rawSQL := fmt.Sprintf(`
		SELECT 
			o.id, 
			o.nomor_order, 
			o.nama_project, 
			o.nama_customer, 
			o.project_status, 
			o.tahapan_proyek, 
			COALESCE(
				(SELECT grand_total FROM rabs WHERE order_id = o.id AND status = 'submitted' LIMIT 1), 
				COALESCE(
					(SELECT SUM(amount) FROM invoices WHERE order_id = o.id), 
					o.harga_kontrak
				)
			) as harga_kontrak,
			o.created_at
		FROM orders o
		%s
		ORDER BY o.id DESC
		LIMIT 5
	`, whereClause)

	if err := s.db.WithContext(ctx).Raw(rawSQL, args...).Scan(&recentOrdersRaw).Error; err != nil {
		s.logger.Error("failed to get recent orders", zap.Error(err))
		return nil, err
	}

	recentOrders := make([]dto.RecentOrderResponse, len(recentOrdersRaw))
	for i, o := range recentOrdersRaw {
		recentOrders[i] = dto.RecentOrderResponse{
			ID:            o.ID,
			NomorOrder:    o.NomorOrder,
			NamaProject:   o.NamaProject,
			NamaCustomer:  o.NamaCustomer,
			ProjectStatus: o.ProjectStatus,
			TahapanProyek: o.TahapanProyek,
			HargaKontrak:  o.HargaKontrak,
			CreatedAt:     o.CreatedAt,
		}
	}

	return &dto.DashboardStatsResponse{
		TotalOrders:        totalOrders,
		ActiveOrders:       activeOrders,
		CompletedProjects:  completedProjects,
		SuccessRate:        successRate,
		TotalContractsDeal: totalContractsDeal,
		LunasCount:         lunasCount,
		LunasAmount:        lunasAmount,
		BelumBayarCount:    belumBayarCount,
		BelumBayarAmount:   belumBayarAmount,
		TotalOmset:         totalOmset,
		RecentOrders:       recentOrders,
	}, nil
}
