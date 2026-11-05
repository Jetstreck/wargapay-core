package repository

import (
	"wargapay-core/entity"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(item *entity.Notification) error
	FindByUserID(userID uint) ([]entity.Notification, error)
	MarkAsRead(id uint, userID uint) error
	MarkAllAsRead(userID uint) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(item *entity.Notification) error {
	return r.db.Create(item).Error
}

func (r *notificationRepository) FindByUserID(userID uint) ([]entity.Notification, error) {
	var items []entity.Notification
	err := r.db.Where("user_id = ?", userID).Order("id DESC").Find(&items).Error
	return items, err
}

func (r *notificationRepository) MarkAsRead(id uint, userID uint) error {
	return r.db.Model(&entity.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *notificationRepository) MarkAllAsRead(userID uint) error {
	return r.db.Model(&entity.Notification{}).
		Where("user_id = ?", userID).
		Update("is_read", true).Error
}

