package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProdukRepository handles produk-related database operations.
type ProdukRepository interface {
	FindAll(ctx context.Context) ([]entity.Produk, error)
	FindByID(ctx context.Context, id uint) (*entity.Produk, error)
	Create(ctx context.Context, produk *entity.Produk) error
	Update(ctx context.Context, produk *entity.Produk) error
	Delete(ctx context.Context, id uint) error
	DeleteImage(ctx context.Context, imageID uint) error
	RecalculateHarga(ctx context.Context, produkID uint) error
}

type produkRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewProdukRepository(db *gorm.DB, logger *zap.Logger) ProdukRepository {
	return &produkRepository{db: db, logger: logger}
}

func (r *produkRepository) FindAll(ctx context.Context) ([]entity.Produk, error) {
	var produks []entity.Produk
	err := r.db.WithContext(ctx).
		Preload("Images").
		Preload("BahanBakus").
		Preload("BahanBakus.BahanBaku").
		Order("id ASC").
		Find(&produks).Error
	return produks, err
}

func (r *produkRepository) FindByID(ctx context.Context, id uint) (*entity.Produk, error) {
	var produk entity.Produk
	err := r.db.WithContext(ctx).
		Preload("Images").
		Preload("BahanBakus").
		Preload("BahanBakus.BahanBaku").
		First(&produk, id).Error
	if err != nil {
		return nil, err
	}
	return &produk, nil
}

func (r *produkRepository) Create(ctx context.Context, produk *entity.Produk) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(produk).Error; err != nil {
			return err
		}
		return r.recalculate(tx, ctx, produk.ID)
	})
}

func (r *produkRepository) Update(ctx context.Context, produk *entity.Produk) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update produk name
		if err := tx.Model(produk).Updates(map[string]interface{}{
			"nama_produk": produk.NamaProduk,
		}).Error; err != nil {
			return err
		}

		// Replace images
		if err := tx.Where("produk_id = ?", produk.ID).Delete(&entity.ProdukImage{}).Error; err != nil {
			return err
		}
		for _, img := range produk.Images {
			img.ProdukID = produk.ID
			if err := tx.Create(&img).Error; err != nil {
				return err
			}
		}

		// Replace bahan bakus
		if err := tx.Where("produk_id = ?", produk.ID).Delete(&entity.ProdukBahanBaku{}).Error; err != nil {
			return err
		}
		for _, bb := range produk.BahanBakus {
			bb.ProdukID = produk.ID
			if err := tx.Create(&bb).Error; err != nil {
				return err
			}
		}

		return r.recalculate(tx, ctx, produk.ID)
	})
}

func (r *produkRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("produk_id = ?", id).Delete(&entity.ProdukImage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("produk_id = ?", id).Delete(&entity.ProdukBahanBaku{}).Error; err != nil {
			return err
		}
		return tx.Delete(&entity.Produk{}, id).Error
	})
}

func (r *produkRepository) DeleteImage(ctx context.Context, imageID uint) error {
	return r.db.WithContext(ctx).Delete(&entity.ProdukImage{}, imageID).Error
}

func (r *produkRepository) RecalculateHarga(ctx context.Context, produkID uint) error {
	return r.recalculate(r.db, ctx, produkID)
}

func (r *produkRepository) recalculate(tx *gorm.DB, ctx context.Context, produkID uint) error {
	var bahanBakus []entity.ProdukBahanBaku
	if err := tx.WithContext(ctx).Where("produk_id = ?", produkID).Find(&bahanBakus).Error; err != nil {
		return err
	}

	if len(bahanBakus) == 0 {
		return tx.WithContext(ctx).Model(&entity.Produk{}).Where("id = ?", produkID).Updates(map[string]interface{}{
			"harga":      decimal.NewFromInt(0),
			"harga_jasa": decimal.NewFromInt(0),
		}).Error
	}

	var minBB entity.ProdukBahanBaku
	first := true
	for _, bb := range bahanBakus {
		if first {
			minBB = bb
			first = false
			continue
		}
		currentTotal := bb.HargaDasar.Add(bb.HargaJasa)
		minTotal := minBB.HargaDasar.Add(minBB.HargaJasa)
		if currentTotal.LessThan(minTotal) {
			minBB = bb
		}
	}

	return tx.WithContext(ctx).Model(&entity.Produk{}).Where("id = ?", produkID).Updates(map[string]interface{}{
		"harga":      minBB.HargaDasar,
		"harga_jasa": minBB.HargaJasa,
	}).Error
}
