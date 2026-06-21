package service

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/claudio-nehemia/interior_backend/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type ProdukService interface {
	GetAll(ctx context.Context) ([]dto.ProdukResponse, error)
	Create(ctx context.Context, req dto.CreateProdukRequest) (*dto.ProdukResponse, error)
	Update(ctx context.Context, id uint, req dto.UpdateProdukRequest) (*dto.ProdukResponse, error)
	Delete(ctx context.Context, id uint) error
	DeleteImage(ctx context.Context, produkID, imageID uint) error
}

type produkService struct {
	repo   repository.ProdukRepository
	logger *zap.Logger
}

func NewProdukService(repo repository.ProdukRepository, logger *zap.Logger) ProdukService {
	return &produkService{repo: repo, logger: logger}
}

func (s *produkService) GetAll(ctx context.Context) ([]dto.ProdukResponse, error) {
	produks, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ProdukResponse, len(produks))
	for i, p := range produks {
		result[i] = toProdukResponse(p)
	}
	return result, nil
}

func (s *produkService) Create(ctx context.Context, req dto.CreateProdukRequest) (*dto.ProdukResponse, error) {
	produk := &entity.Produk{NamaProduk: req.NamaProduk}

	for _, img := range req.Images {
		produk.Images = append(produk.Images, entity.ProdukImage{Image: img})
	}
	for _, bb := range req.BahanBaku {
		produk.BahanBakus = append(produk.BahanBakus, entity.ProdukBahanBaku{
			BahanBakuID: bb.BahanBakuID,
			HargaDasar:  decimal.NewFromFloat(bb.HargaDasar),
			HargaJasa:   decimal.NewFromFloat(bb.HargaJasa),
		})
	}

	if err := s.repo.Create(ctx, produk); err != nil {
		return nil, err
	}

	created, err := s.repo.FindByID(ctx, produk.ID)
	if err != nil {
		return nil, err
	}
	resp := toProdukResponse(*created)
	return &resp, nil
}

func (s *produkService) Update(ctx context.Context, id uint, req dto.UpdateProdukRequest) (*dto.ProdukResponse, error) {
	produk, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	produk.NamaProduk = req.NamaProduk
	produk.Images = nil
	produk.BahanBakus = nil

	for _, img := range req.Images {
		produk.Images = append(produk.Images, entity.ProdukImage{Image: img})
	}
	for _, bb := range req.BahanBaku {
		produk.BahanBakus = append(produk.BahanBakus, entity.ProdukBahanBaku{
			BahanBakuID: bb.BahanBakuID,
			HargaDasar:  decimal.NewFromFloat(bb.HargaDasar),
			HargaJasa:   decimal.NewFromFloat(bb.HargaJasa),
		})
	}

	if err := s.repo.Update(ctx, produk); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, produk.ID)
	if err != nil {
		return nil, err
	}
	resp := toProdukResponse(*updated)
	return &resp, nil
}

func (s *produkService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *produkService) DeleteImage(ctx context.Context, produkID, imageID uint) error {
	if err := s.repo.DeleteImage(ctx, imageID); err != nil {
		return err
	}
	return s.repo.RecalculateHarga(ctx, produkID)
}

func toProdukResponse(p entity.Produk) dto.ProdukResponse {
	harga, _ := p.Harga.Float64()
	hargaJasa, _ := p.HargaJasa.Float64()

	resp := dto.ProdukResponse{
		ID:         p.ID,
		NamaProduk: p.NamaProduk,
		Harga:      harga,
		HargaJasa:  hargaJasa,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}

	images := make([]dto.ProdukImageResponse, len(p.Images))
	for i, img := range p.Images {
		images[i] = dto.ProdukImageResponse{ID: img.ID, Image: img.Image}
	}
	resp.Images = images

	bbs := make([]dto.ProdukBahanBakuResponse, len(p.BahanBakus))
	for i, bb := range p.BahanBakus {
		hd, _ := bb.HargaDasar.Float64()
		hj, _ := bb.HargaJasa.Float64()
		bbs[i] = dto.ProdukBahanBakuResponse{
			ID:          bb.ID,
			BahanBakuID: bb.BahanBakuID,
			NamaBahan:   bb.BahanBaku.NamaBahanBaku,
			HargaDasar:  hd,
			HargaJasa:   hj,
		}
	}
	resp.BahanBakus = bbs

	return resp
}
