package entity

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Produk represents the produks table.
type Produk struct {
	ID         uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	NamaProduk string          `gorm:"size:255;not null" json:"nama_produk"`
	Harga      decimal.Decimal `gorm:"type:decimal(18,2);default:0" json:"harga"`
	HargaJasa  decimal.Decimal `gorm:"type:decimal(18,2);default:0" json:"harga_jasa"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`

	// Relations
	Images    []ProdukImage    `gorm:"foreignKey:ProdukID" json:"images,omitempty"`
	BahanBakus []ProdukBahanBaku `gorm:"foreignKey:ProdukID" json:"bahan_bakus,omitempty"`
}

func (Produk) TableName() string {
	return "produks"
}

// RecalculateHarga recalculates Harga and HargaJasa from BahanBakus pivot, selecting the option with the minimum total price.
func (p *Produk) RecalculateHarga() {
	if len(p.BahanBakus) == 0 {
		p.Harga = decimal.NewFromInt(0)
		p.HargaJasa = decimal.NewFromInt(0)
		return
	}

	var minBB ProdukBahanBaku
	first := true
	for _, bb := range p.BahanBakus {
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
	p.Harga = minBB.HargaDasar
	p.HargaJasa = minBB.HargaJasa
}

// AfterFind is a GORM hook that can be used after querying.
func (p *Produk) AfterFind(tx *gorm.DB) error {
	return nil
}
