package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ItemRepository handles item-related database operations.
type ItemRepository interface {
	FindAll(ctx context.Context, jenis string) ([]entity.Item, error)
	FindByID(ctx context.Context, id uint) (*entity.Item, error)
	Create(ctx context.Context, item *entity.Item) error
	Update(ctx context.Context, item *entity.Item) error
	Delete(ctx context.Context, id uint) error
}

type itemRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewItemRepository(db *gorm.DB, logger *zap.Logger) ItemRepository {
	return &itemRepository{db: db, logger: logger}
}

func (r *itemRepository) FindAll(ctx context.Context, jenis string) ([]entity.Item, error) {
	var items []entity.Item
	query := r.db.WithContext(ctx)
	if jenis != "" {
		query = query.Where("jenis_item = ?", jenis)
	}
	err := query.Order("id ASC").Find(&items).Error
	return items, err
}

func (r *itemRepository) FindByID(ctx context.Context, id uint) (*entity.Item, error) {
	var item entity.Item
	err := r.db.WithContext(ctx).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *itemRepository) Create(ctx context.Context, item *entity.Item) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *itemRepository) Update(ctx context.Context, item *entity.Item) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *itemRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Item{}, id).Error
}
