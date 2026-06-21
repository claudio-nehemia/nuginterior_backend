package repository

import (
	"context"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *entity.Notification) error
	FindAllByUserID(ctx context.Context, userID uint) ([]entity.Notification, error)
	FindByID(ctx context.Context, id uint) (*entity.Notification, error)
	Update(ctx context.Context, notification *entity.Notification) error
	MarkAllAsRead(ctx context.Context, userID uint) error
	CountUnread(ctx context.Context, userID uint) (int64, error)
}

type notificationRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewNotificationRepository(db *gorm.DB, logger *zap.Logger) NotificationRepository {
	return &notificationRepository{db: db, logger: logger}
}

func (r *notificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepository) FindAllByUserID(ctx context.Context, userID uint) ([]entity.Notification, error) {
	var notifications []entity.Notification
	err := r.db.WithContext(ctx).
		Preload("Order").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) FindByID(ctx context.Context, id uint) (*entity.Notification, error) {
	var notification entity.Notification
	err := r.db.WithContext(ctx).
		Preload("Order").
		First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) Update(ctx context.Context, notification *entity.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}
