package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MoodboardRepository handles moodboard-related database operations.
type MoodboardRepository interface {
	FindAll(ctx context.Context) ([]entity.Moodboard, error)
	FindByID(ctx context.Context, id uint) (*entity.Moodboard, error)
	FindByOrderID(ctx context.Context, orderID uint) (*entity.Moodboard, error)
	Create(ctx context.Context, moodboard *entity.Moodboard) error
	Update(ctx context.Context, moodboard *entity.Moodboard) error
	Delete(ctx context.Context, id uint) error

	// Moodboard Files
	FindFileByID(ctx context.Context, fileID uint) (*entity.MoodboardFile, error)
	CreateFile(ctx context.Context, file *entity.MoodboardFile) error
	UpdateFile(ctx context.Context, file *entity.MoodboardFile) error
	DeleteFile(ctx context.Context, fileID uint) error

	// Estimasi (RAB Kasar)
	FindEstimasiByMoodboardID(ctx context.Context, moodboardID uint) (*entity.Estimasi, error)
	CreateEstimasi(ctx context.Context, estimasi *entity.Estimasi) error
	UpdateEstimasi(ctx context.Context, estimasi *entity.Estimasi) error

	// Estimasi Files
	FindEstimasiFileByID(ctx context.Context, fileID uint) (*entity.EstimasiFile, error)
	FindEstimasiFileByMoodboardFileID(ctx context.Context, moodboardFileID uint) (*entity.EstimasiFile, error)
	CreateEstimasiFile(ctx context.Context, file *entity.EstimasiFile) error
	UpdateEstimasiFile(ctx context.Context, file *entity.EstimasiFile) error
	DeleteEstimasiFile(ctx context.Context, fileID uint) error

	// Commitment Fee
	FindCommitmentFeeByMoodboardID(ctx context.Context, moodboardID uint) (*entity.CommitmentFee, error)
	FindCommitmentFeeByID(ctx context.Context, id uint) (*entity.CommitmentFee, error)
	CreateCommitmentFee(ctx context.Context, fee *entity.CommitmentFee) error
	UpdateCommitmentFee(ctx context.Context, fee *entity.CommitmentFee) error

	// Order Helper
	UpdateOrderStageAndPayment(ctx context.Context, orderID uint, stage string, paymentStatus string) error
}

type moodboardRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewMoodboardRepository(db *gorm.DB, logger *zap.Logger) MoodboardRepository {
	return &moodboardRepository{db: db, logger: logger}
}

func (r *moodboardRepository) FindAll(ctx context.Context) ([]entity.Moodboard, error) {
	var list []entity.Moodboard
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Files").
		Preload("Estimasi.Files").
		Preload("CommitmentFee").
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *moodboardRepository) FindByID(ctx context.Context, id uint) (*entity.Moodboard, error) {
	var moodboard entity.Moodboard
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Files").
		Preload("Estimasi.Files").
		Preload("CommitmentFee").
		First(&moodboard, id).Error
	if err != nil {
		return nil, err
	}
	return &moodboard, nil
}

func (r *moodboardRepository) FindByOrderID(ctx context.Context, orderID uint) (*entity.Moodboard, error) {
	var moodboard entity.Moodboard
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("Files").
		Preload("Estimasi.Files").
		Preload("CommitmentFee").
		Where("order_id = ?", orderID).
		First(&moodboard).Error
	if err != nil {
		return nil, err
	}
	return &moodboard, nil
}

func (r *moodboardRepository) Create(ctx context.Context, moodboard *entity.Moodboard) error {
	return r.db.WithContext(ctx).Create(moodboard).Error
}

func (r *moodboardRepository) Update(ctx context.Context, moodboard *entity.Moodboard) error {
	return r.db.WithContext(ctx).Save(moodboard).Error
}

func (r *moodboardRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Moodboard{}, id).Error
}

// Moodboard Files

func (r *moodboardRepository) FindFileByID(ctx context.Context, fileID uint) (*entity.MoodboardFile, error) {
	var file entity.MoodboardFile
	err := r.db.WithContext(ctx).First(&file, fileID).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *moodboardRepository) CreateFile(ctx context.Context, file *entity.MoodboardFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *moodboardRepository) UpdateFile(ctx context.Context, file *entity.MoodboardFile) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *moodboardRepository) DeleteFile(ctx context.Context, fileID uint) error {
	return r.db.WithContext(ctx).Delete(&entity.MoodboardFile{}, fileID).Error
}

// Estimasi (RAB Kasar)

func (r *moodboardRepository) FindEstimasiByMoodboardID(ctx context.Context, moodboardID uint) (*entity.Estimasi, error) {
	var estimasi entity.Estimasi
	err := r.db.WithContext(ctx).
		Preload("Files").
		Where("moodboard_id = ?", moodboardID).
		First(&estimasi).Error
	if err != nil {
		return nil, err
	}
	return &estimasi, nil
}

func (r *moodboardRepository) CreateEstimasi(ctx context.Context, estimasi *entity.Estimasi) error {
	return r.db.WithContext(ctx).Create(estimasi).Error
}

func (r *moodboardRepository) UpdateEstimasi(ctx context.Context, estimasi *entity.Estimasi) error {
	return r.db.WithContext(ctx).Save(estimasi).Error
}

// Estimasi Files

func (r *moodboardRepository) FindEstimasiFileByID(ctx context.Context, fileID uint) (*entity.EstimasiFile, error) {
	var file entity.EstimasiFile
	err := r.db.WithContext(ctx).First(&file, fileID).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *moodboardRepository) FindEstimasiFileByMoodboardFileID(ctx context.Context, moodboardFileID uint) (*entity.EstimasiFile, error) {
	var file entity.EstimasiFile
	err := r.db.WithContext(ctx).Where("moodboard_file_id = ?", moodboardFileID).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *moodboardRepository) CreateEstimasiFile(ctx context.Context, file *entity.EstimasiFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *moodboardRepository) UpdateEstimasiFile(ctx context.Context, file *entity.EstimasiFile) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *moodboardRepository) DeleteEstimasiFile(ctx context.Context, fileID uint) error {
	return r.db.WithContext(ctx).Delete(&entity.EstimasiFile{}, fileID).Error
}

// Commitment Fee

func (r *moodboardRepository) FindCommitmentFeeByMoodboardID(ctx context.Context, moodboardID uint) (*entity.CommitmentFee, error) {
	var fee entity.CommitmentFee
	err := r.db.WithContext(ctx).Where("moodboard_id = ?", moodboardID).First(&fee).Error
	if err != nil {
		return nil, err
	}
	return &fee, nil
}

func (r *moodboardRepository) FindCommitmentFeeByID(ctx context.Context, id uint) (*entity.CommitmentFee, error) {
	var fee entity.CommitmentFee
	err := r.db.WithContext(ctx).First(&fee, id).Error
	if err != nil {
		return nil, err
	}
	return &fee, nil
}

func (r *moodboardRepository) CreateCommitmentFee(ctx context.Context, fee *entity.CommitmentFee) error {
	return r.db.WithContext(ctx).Create(fee).Error
}

func (r *moodboardRepository) UpdateCommitmentFee(ctx context.Context, fee *entity.CommitmentFee) error {
	return r.db.WithContext(ctx).Save(fee).Error
}

// Order Helper

func (r *moodboardRepository) UpdateOrderStageAndPayment(ctx context.Context, orderID uint, stage string, paymentStatus string) error {
	updates := map[string]interface{}{}
	if stage != "" {
		updates["tahapan_proyek"] = stage
	}
	if paymentStatus != "" {
		updates["payment_status"] = paymentStatus
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&entity.Order{}).Where("id = ?", orderID).Updates(updates).Error
}
