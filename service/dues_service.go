package service

import (
	"errors"
	"fmt"
	"time"
	"wargapay-core/dto"
	"wargapay-core/entity"
	"wargapay-core/repository"

	"gorm.io/gorm"
)

type DuesService struct {
	repo repository.DuesRepository
}

func NewDuesService(repo repository.DuesRepository) *DuesService {
	return &DuesService{repo: repo}
}

// Create
func (s *DuesService) Create(req dto.DuesRequest) (*entity.Dues, error) {
	if req.UserID == 0 || !validPeriod(req.Period) {
		return nil, errors.New("user_id dan period dengan format YYYY-MM wajib valid")
	}

	if req.WasteFee < 0 || req.SecurityFee < 0 || req.RTFund < 0 {
		return nil, errors.New("nominal iuran tidak boleh negatif")
	}

	item := &entity.Dues{
		UserID:      req.UserID,
		Period:      req.Period,
		WasteFee:    req.WasteFee,
		SecurityFee: req.SecurityFee,
		RTFund:      req.RTFund,
		Total:       req.WasteFee + req.SecurityFee + req.RTFund,
		Status:      "unpaid",
	}
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}

	return item, nil
}

// List
func (s *DuesService) List(userID uint, period, status string) ([]entity.Dues, error) {
	if status != "" && status != "paid" && status != "unpaid" {
		return nil, errors.New("status harus paid atau unpaid")
	}

	return s.repo.FindAll(userID, period, status)
}

// Get
func (s *DuesService) Get(id uint) (*entity.Dues, error) {
	item, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("tagihan tidak ditemukan")
	}

	return item, err
}

// Update
func (s *DuesService) Update(id uint, req dto.DuesRequest) (*entity.Dues, error) {
	item, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if req.UserID == 0 || !validPeriod(req.Period) {
		return nil, errors.New("user_id dan period dengan format YYYY-MM wajib valid")
	}

	if req.WasteFee < 0 || req.SecurityFee < 0 || req.RTFund < 0 {
		return nil, errors.New("nominal iuran tidak boleh negatif")
	}

	if item.Status == "paid" {
		return nil, errors.New("tagihan lunas tidak dapat diubah")
	}

	item.UserID, item.Period = req.UserID, req.Period
	item.WasteFee, item.SecurityFee, item.RTFund = req.WasteFee, req.SecurityFee, req.RTFund
	item.Total = req.WasteFee + req.SecurityFee + req.RTFund
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	return item, nil
}

// MarkPaid menandai tagihan sebagai lunas dan mencatat waktu pembayarannya
func (s *DuesService) MarkPaid(id, userID uint, isAdmin bool) (*entity.Dues, error) {
	item, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if !isAdmin && item.UserID != userID {
		return nil, errors.New("anda tidak berhak membayar tagihan ini")
	}

	if item.Status == "paid" {
		return nil, errors.New("tagihan sudah lunas")
	}

	now := time.Now()
	item.Status, item.PaidAt = "paid", &now
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}

// Delete
func (s *DuesService) Delete(id uint) error {
	item, err := s.Get(id)
	if err != nil {
		return err
	}
	if item.Status == "paid" {
		return errors.New("tagihan lunas tidak dapat dihapus")
	}
	return s.repo.Delete(id)
}

// Memastikan periode menggunakan format YYYY-MM.
func validPeriod(value string) bool {
	parsed, err := time.Parse("2006-01", value)
	return err == nil && parsed.Format("2006-01") == value
}

// Format rupiah
func FormatRupiah(value int64) string {
	return fmt.Sprintf("Rp%d", value)
}
