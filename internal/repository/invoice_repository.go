package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	FindAllContractsWithInvoices(ctx context.Context) ([]entity.Contract, error)
	FindByID(ctx context.Context, id uint) (*entity.Invoice, error)
	Create(ctx context.Context, invoice *entity.Invoice) error
	Update(ctx context.Context, invoice *entity.Invoice) error
	UpdateOrderStage(ctx context.Context, orderID uint, stage string) error
}

type invoiceRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewInvoiceRepository(db *gorm.DB, logger *zap.Logger) InvoiceRepository {
	return &invoiceRepository{db: db, logger: logger}
}

func (r *invoiceRepository) FindAllContractsWithInvoices(ctx context.Context) ([]entity.Contract, error) {
	var list []entity.Contract
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("Termin").
		Preload("Invoices").
		Where("status = ?", "deal").
		Order("id DESC").
		Find(&list).Error
	return list, err
}

func (r *invoiceRepository) FindByID(ctx context.Context, id uint) (*entity.Invoice, error) {
	var invoice entity.Invoice
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Contract").
		Preload("Contract.Order").
		Preload("Contract.Termin").
		First(&invoice, id).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) Create(ctx context.Context, invoice *entity.Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

func (r *invoiceRepository) Update(ctx context.Context, invoice *entity.Invoice) error {
	return r.db.WithContext(ctx).Save(invoice).Error
}

func (r *invoiceRepository) UpdateOrderStage(ctx context.Context, orderID uint, stage string) error {
	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", orderID).Update("tahapan_proyek", stage).Error
}

