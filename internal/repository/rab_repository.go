package repository

import (
	"context"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RABRepository interface {
	FindAll(ctx context.Context) ([]entity.RAB, error)
	FindByID(ctx context.Context, id uint) (*entity.RAB, error)
	FindByInputItemID(ctx context.Context, inputItemID uint) (*entity.RAB, error)
	Create(ctx context.Context, rab *entity.RAB) error
	Update(ctx context.Context, rab *entity.RAB) error
	Delete(ctx context.Context, id uint) error
	UpdateOrderStage(ctx context.Context, orderID uint, stage string) error
}

type rabRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewRABRepository(db *gorm.DB, logger *zap.Logger) RABRepository {
	return &rabRepository{db: db, logger: logger}
}

func (r *rabRepository) FindAll(ctx context.Context) ([]entity.RAB, error) {
	var list []entity.RAB
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("InputItem").
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
		Order("id DESC").
		Find(&list).Error
	return list, err
}

func (r *rabRepository) FindByID(ctx context.Context, id uint) (*entity.RAB, error) {
	var rab entity.RAB
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("InputItem").
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
		First(&rab, id).Error
	if err != nil {
		return nil, err
	}
	return &rab, nil
}

func (r *rabRepository) FindByInputItemID(ctx context.Context, inputItemID uint) (*entity.RAB, error) {
	var rab entity.RAB
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("InputItem").
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
		Where("input_item_id = ?", inputItemID).
		First(&rab).Error
	if err != nil {
		return nil, err
	}
	return &rab, nil
}

func (r *rabRepository) Create(ctx context.Context, rab *entity.RAB) error {
	return r.db.WithContext(ctx).Create(rab).Error
}

func (r *rabRepository) Update(ctx context.Context, rab *entity.RAB) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Fetch old room IDs to remove child objects
		var roomIDs []uint
		tx.Model(&entity.RABRoom{}).Where("rab_id = ?", rab.ID).Pluck("id", &roomIDs)

		if len(roomIDs) > 0 {
			if err := tx.Where("rab_room_id IN ?", roomIDs).Delete(&entity.RABRoomBahanBaku{}).Error; err != nil {
				return err
			}
			if err := tx.Where("rab_room_id IN ?", roomIDs).Delete(&entity.RABRoomFinishing{}).Error; err != nil {
				return err
			}
			if err := tx.Where("rab_room_id IN ?", roomIDs).Delete(&entity.RABRoomAksesoris{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", roomIDs).Delete(&entity.RABRoom{}).Error; err != nil {
				return err
			}
		}

		// 2. Update parent info
		if err := tx.Model(rab).Select("markup_general", "grand_total", "status", "submitted_at", "submitted_by", "updated_at").Updates(map[string]interface{}{
			"markup_general": rab.MarkupGeneral,
			"grand_total":    rab.GrandTotal,
			"status":         rab.Status,
			"submitted_at":   rab.SubmittedAt,
			"submitted_by":   rab.SubmittedBy,
			"updated_at":     time.Now(),
		}).Error; err != nil {
			return err
		}

		// 3. Create new rooms with associations
		for i := range rab.Rooms {
			room := &rab.Rooms[i]
			room.RABID = rab.ID
			if err := tx.Create(room).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *rabRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roomIDs []uint
		tx.Model(&entity.RABRoom{}).Where("rab_id = ?", id).Pluck("id", &roomIDs)

		if len(roomIDs) > 0 {
			if err := tx.Where("rab_room_id IN ?", roomIDs).Delete(&entity.RABRoomBahanBaku{}).Error; err != nil {
				return err
			}
			if err := tx.Where("rab_room_id IN ?", roomIDs).Delete(&entity.RABRoomFinishing{}).Error; err != nil {
				return err
			}
			if err := tx.Where("rab_room_id IN ?", roomIDs).Delete(&entity.RABRoomAksesoris{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", roomIDs).Delete(&entity.RABRoom{}).Error; err != nil {
				return err
			}
		}

		return tx.Delete(&entity.RAB{}, id).Error
	})
}

func (r *rabRepository) UpdateOrderStage(ctx context.Context, orderID uint, stage string) error {
	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", orderID).Update("tahapan_proyek", stage).Error
}

