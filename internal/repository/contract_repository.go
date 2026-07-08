package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ContractRepository interface {
	FindAllRABsWithContract(ctx context.Context) ([]entity.RAB, error)
	FindByID(ctx context.Context, id uint) (*entity.Contract, error)
	FindByRABID(ctx context.Context, rabID uint) (*entity.Contract, error)
	Create(ctx context.Context, contract *entity.Contract) error
	Update(ctx context.Context, contract *entity.Contract) error
	UpdateOrderStage(ctx context.Context, orderID uint, stage string) error
}

type contractRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewContractRepository(db *gorm.DB, logger *zap.Logger) ContractRepository {
	return &contractRepository{db: db, logger: logger}
}

func (r *contractRepository) FindAllRABsWithContract(ctx context.Context) ([]entity.RAB, error) {
	var list []entity.RAB
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("Contract").
		Preload("Contract.Termin").
		Where("status = ?", "submitted").
		Order("id DESC").
		Find(&list).Error
	return list, err
}

func (r *contractRepository) FindByID(ctx context.Context, id uint) (*entity.Contract, error) {
	var contract entity.Contract
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("Termin").
		Preload("RAB").
		First(&contract, id).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (r *contractRepository) FindByRABID(ctx context.Context, rabID uint) (*entity.Contract, error) {
	var contract entity.Contract
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("Termin").
		Preload("RAB").
		Where("rab_id = ?", rabID).
		First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (r *contractRepository) Create(ctx context.Context, contract *entity.Contract) error {
	return r.db.WithContext(ctx).Create(contract).Error
}

func (r *contractRepository) Update(ctx context.Context, contract *entity.Contract) error {
	return r.db.WithContext(ctx).Save(contract).Error
}

func (r *contractRepository) UpdateOrderStage(ctx context.Context, orderID uint, stage string) error {
	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", orderID).Update("tahapan_proyek", stage).Error
}

