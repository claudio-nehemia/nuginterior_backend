package repository

import (
	"context"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/database"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type InputItemRepository interface {
	FindAll(ctx context.Context) ([]entity.InputItem, error)
	FindByID(ctx context.Context, id uint) (*entity.InputItem, error)
	FindByDesainFinalID(ctx context.Context, dfID uint) (*entity.InputItem, error)
	Create(ctx context.Context, item *entity.InputItem) error
	Update(ctx context.Context, item *entity.InputItem) error
	Delete(ctx context.Context, id uint) error
	GetOrderIDByDesainFinalID(ctx context.Context, dfID uint) (uint, error)
	UpdateOrderStage(ctx context.Context, orderID uint, stage string) error
}

type inputItemRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewInputItemRepository(db *gorm.DB, logger *zap.Logger) InputItemRepository {
	return &inputItemRepository{db: db, logger: logger}
}

func (r *inputItemRepository) FindAll(ctx context.Context) ([]entity.InputItem, error) {
	var list []entity.InputItem
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("DesainFinal").
		Preload("Rooms").
		Preload("Rooms.Produk").
		Preload("Rooms.BahanBakus").
		Preload("Rooms.BahanBakus.BahanBaku").
		Preload("Rooms.FinishingDalams").
		Preload("Rooms.FinishingDalams.Item").
		Preload("Rooms.FinishingLuars").
		Preload("Rooms.FinishingLuars.Item").
		Preload("Rooms.Aksesoris").
		Preload("Rooms.Aksesoris.Item").
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *inputItemRepository) FindByID(ctx context.Context, id uint) (*entity.InputItem, error) {
	var item entity.InputItem
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("DesainFinal").
		Preload("Rooms").
		Preload("Rooms.Produk").
		Preload("Rooms.BahanBakus").
		Preload("Rooms.BahanBakus.BahanBaku").
		Preload("Rooms.FinishingDalams").
		Preload("Rooms.FinishingDalams.Item").
		Preload("Rooms.FinishingLuars").
		Preload("Rooms.FinishingLuars.Item").
		Preload("Rooms.Aksesoris").
		Preload("Rooms.Aksesoris.Item").
		First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *inputItemRepository) FindByDesainFinalID(ctx context.Context, dfID uint) (*entity.InputItem, error) {
	var item entity.InputItem
	err := r.db.WithContext(ctx).
		Scopes(database.OrderScope(ctx)).
		Preload("Order").
		Preload("DesainFinal").
		Preload("Rooms").
		Preload("Rooms.Produk").
		Preload("Rooms.BahanBakus").
		Preload("Rooms.BahanBakus.BahanBaku").
		Preload("Rooms.FinishingDalams").
		Preload("Rooms.FinishingDalams.Item").
		Preload("Rooms.FinishingLuars").
		Preload("Rooms.FinishingLuars.Item").
		Preload("Rooms.Aksesoris").
		Preload("Rooms.Aksesoris.Item").
		Where("desain_final_id = ?", dfID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *inputItemRepository) Create(ctx context.Context, item *entity.InputItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *inputItemRepository) Update(ctx context.Context, item *entity.InputItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Fetch old room IDs to remove child objects
		var roomIDs []uint
		tx.Model(&entity.InputItemRoom{}).Where("input_item_id = ?", item.ID).Pluck("id", &roomIDs)

		if len(roomIDs) > 0 {
			if err := tx.Where("input_item_room_id IN ?", roomIDs).Delete(&entity.InputItemRoomBahanBaku{}).Error; err != nil {
				return err
			}
			if err := tx.Where("input_item_room_id IN ?", roomIDs).Delete(&entity.InputItemRoomFinishing{}).Error; err != nil {
				return err
			}
			if err := tx.Where("input_item_room_id IN ?", roomIDs).Delete(&entity.InputItemRoomAksesoris{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", roomIDs).Delete(&entity.InputItemRoom{}).Error; err != nil {
				return err
			}
		}

		// 2. Update parent info
		if err := tx.Model(item).Select("status", "response_by", "response_time", "marketing_response_by", "marketing_response_time", "updated_at").Updates(map[string]interface{}{
			"status":                  item.Status,
			"response_by":             item.ResponseBy,
			"response_time":           item.ResponseTime,
			"marketing_response_by":   item.MarketingResponseBy,
			"marketing_response_time": item.MarketingResponseTime,
			"updated_at":              time.Now(),
		}).Error; err != nil {
			return err
		}

		// 3. Create new rooms with associations
		for i := range item.Rooms {
			room := &item.Rooms[i]
			room.InputItemID = item.ID
			if err := tx.Create(room).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *inputItemRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roomIDs []uint
		tx.Model(&entity.InputItemRoom{}).Where("input_item_id = ?", id).Pluck("id", &roomIDs)

		if len(roomIDs) > 0 {
			if err := tx.Where("input_item_room_id IN ?", roomIDs).Delete(&entity.InputItemRoomBahanBaku{}).Error; err != nil {
				return err
			}
			if err := tx.Where("input_item_room_id IN ?", roomIDs).Delete(&entity.InputItemRoomFinishing{}).Error; err != nil {
				return err
			}
			if err := tx.Where("input_item_room_id IN ?", roomIDs).Delete(&entity.InputItemRoomAksesoris{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", roomIDs).Delete(&entity.InputItemRoom{}).Error; err != nil {
				return err
			}
		}

		return tx.Delete(&entity.InputItem{}, id).Error
	})
}

func (r *inputItemRepository) GetOrderIDByDesainFinalID(ctx context.Context, dfID uint) (uint, error) {
	var df entity.DesainFinal
	err := r.db.WithContext(ctx).Select("order_id").First(&df, dfID).Error
	if err != nil {
		return 0, err
	}
	return df.OrderID, nil
}

func (r *inputItemRepository) UpdateOrderStage(ctx context.Context, orderID uint, stage string) error {
	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", orderID).Update("tahapan_proyek", stage).Error
}
