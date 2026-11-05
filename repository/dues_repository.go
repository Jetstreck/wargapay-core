package repository

import (
	"wargapay-core/entity"

	"gorm.io/gorm"
)

type DuesRepository interface {
	Create(item *entity.Dues) error
	FindByID(id uint) (*entity.Dues, error)
	FindAll(userID uint, period, status string) ([]entity.Dues, error)
	Update(item *entity.Dues) error
	Delete(id uint) error
}

type duesRepository struct {
	db *gorm.DB
}

func NewDuesRepository(db *gorm.DB) DuesRepository {
	return &duesRepository{db: db}
}

func (r *duesRepository) Create(item *entity.Dues) error {
	return r.db.Create(item).Error
}

func (r *duesRepository) FindByID(id uint) (*entity.Dues, error) {
	var item entity.Dues
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// FindAll
func (r *duesRepository) FindAll(userID uint, period, status string) ([]entity.Dues, error) {
	var items []entity.Dues
	query := r.db.Model(&entity.Dues{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if period != "" {
		query = query.Where("period = ?", period)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("period DESC, id DESC").Find(&items).Error
	return items, err
}

// Update
func (r *duesRepository) Update(item *entity.Dues) error {
	return r.db.Save(item).Error
}

// Delete
func (r *duesRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Dues{}, id).Error
}
