package repository

import (
	"wargapay-core/entity"

	"gorm.io/gorm"
)

type WasteServiceRepository interface {
	Create(item *entity.WasteService) error
	FindByID(id uint) (*entity.WasteService, error)
	FindAll(filters map[string]interface{}) ([]entity.WasteService, error)
	Update(item *entity.WasteService) error
	Delete(id uint) error
}

type wasteServiceRepository struct{ db *gorm.DB }

func NewWasteServiceRepository(db *gorm.DB) WasteServiceRepository {
	return &wasteServiceRepository{db: db}
}

// Create
func (r *wasteServiceRepository) Create(item *entity.WasteService) error {
	return r.db.Create(item).Error
}

// FindByID
func (r *wasteServiceRepository) FindByID(id uint) (*entity.WasteService, error) {
	var item entity.WasteService
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// FindAll
func (r *wasteServiceRepository) FindAll(filters map[string]interface{}) ([]entity.WasteService, error) {
	var items []entity.WasteService
	query := r.db.Model(&entity.WasteService{})
	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if min, ok := filters["min_price"].(int64); ok {
		query = query.Where("price >= ?", min)
	}
	if max, ok := filters["max_price"].(int64); ok {
		query = query.Where("price <= ?", max)
	}
	if active, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", active)
	}
	err := query.Order("id DESC").Find(&items).Error
	return items, err
}

// Update
func (r *wasteServiceRepository) Update(item *entity.WasteService) error {
	return r.db.Save(item).Error
}

// Delete
func (r *wasteServiceRepository) Delete(id uint) error {
	return r.db.Delete(&entity.WasteService{}, id).Error
}
