package service

import (
	"errors"
	"strings"
	"wargapay-core/dto"
	"wargapay-core/entity"
	"wargapay-core/repository"

	"gorm.io/gorm"
)

type WasteServiceService struct {
	repo repository.WasteServiceRepository
}

func NewWasteServiceService(repo repository.WasteServiceRepository) *WasteServiceService {
	return &WasteServiceService{repo: repo}
}

// Create
func (s *WasteServiceService) Create(req dto.WasteServiceRequest) (*entity.WasteService, error) {
	if strings.TrimSpace(req.Name) == "" || req.Price < 0 {
		return nil, errors.New("nama wajib diisi dan harga tidak boleh negatif")
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	item := &entity.WasteService{
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Price:       req.Price,
		IsActive:    active,
	}
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}

	return item, nil
}

// List
func (s *WasteServiceService) List(filters map[string]interface{}) ([]entity.WasteService, error) {
	return s.repo.FindAll(filters)
}

// Get
func (s *WasteServiceService) Get(id uint) (*entity.WasteService, error) {
	item, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("layanan tidak ditemukan")
	}
	return item, err
}

// Update
func (s *WasteServiceService) Update(id uint, req dto.WasteServiceRequest) (*entity.WasteService, error) {
	item, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.Name) == "" || req.Price < 0 {
		return nil, errors.New("nama wajib diisi dan harga tidak boleh negatif")
	}

	item.Name, item.Description, item.Price = strings.TrimSpace(req.Name), req.Description, req.Price
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	return item, nil
}

// Delete
func (s *WasteServiceService) Delete(id uint) error {
	if _, err := s.Get(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
