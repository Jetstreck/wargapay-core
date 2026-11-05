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
	repo     repository.DuesRepository
	userRepo repository.UserRepository
	notifSvc NotificationService
}

func NewDuesService(repo repository.DuesRepository, opts ...interface{}) *DuesService {
	s := &DuesService{repo: repo}
	for _, opt := range opts {
		if ur, ok := opt.(repository.UserRepository); ok {
			s.userRepo = ur
		}
		if ns, ok := opt.(NotificationService); ok {
			s.notifSvc = ns
		}
	}
	return s
}

// Create
func (s *DuesService) Create(req dto.DuesRequest) (*entity.Dues, error) {
	if req.UserID == 0 || !validPeriod(req.Period) {
		return nil, errors.New("user_id dan period dengan format YYYY-MM wajib valid")
	}

	if req.WasteFee < 0 || req.SecurityFee < 0 || req.RTFund < 0 {
		return nil, errors.New("nominal iuran tidak boleh negatif")
	}

	// Cek apakah tagihan untuk user dan periode ini sudah ada
	existingList, err := s.repo.FindAll(req.UserID, req.Period, "")
	if err == nil && len(existingList) > 0 {
		existing := existingList[0]
		if existing.Status == "paid" {
			return nil, fmt.Errorf("tagihan periode %s untuk warga ini sudah lunas", req.Period)
		}
		// Perbarui data tagihan yang belum lunas
		existing.WasteFee = req.WasteFee
		existing.SecurityFee = req.SecurityFee
		existing.RTFund = req.RTFund
		existing.Total = req.WasteFee + req.SecurityFee + req.RTFund
		if err := s.repo.Update(&existing); err != nil {
			return nil, err
		}

		// Kirim email tagihan ke Mailtrap & rekam ke notifikasi warga
		if s.userRepo != nil && s.notifSvc != nil {
			user, _ := s.userRepo.FindByID(existing.UserID)
			if user != nil {
				s.notifSvc.NotifyDuesCreated(user, &existing)
			}
		}

		return &existing, nil
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

	// Kirim email tagihan ke Mailtrap & rekam ke notifikasi warga
	if s.userRepo != nil && s.notifSvc != nil {
		user, _ := s.userRepo.FindByID(item.UserID)
		if user != nil {
			s.notifSvc.NotifyDuesCreated(user, item)
		}
	}

	return item, nil
}

// GetByID
func (s *DuesService) GetByID(id uint) (*entity.Dues, error) {
	return s.repo.FindByID(id)
}

// List
func (s *DuesService) List(userID uint, period, status string) ([]entity.Dues, error) {
	return s.repo.FindAll(userID, period, status)
}

// PayDues
func (s *DuesService) PayDues(id uint) (*entity.Dues, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("tagihan tidak ditemukan")
	}

	if item.Status == "paid" {
		return nil, errors.New("tagihan ini sudah dibayar lunas sebelumnya")
	}

	now := time.Now()
	item.Status = "paid"
	item.PaidAt = &now

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}

	// Kirim email konfirmasi sukses bayar ke Mailtrap & simpan notifikasi
	if s.userRepo != nil && s.notifSvc != nil {
		user, _ := s.userRepo.FindByID(item.UserID)
		if user != nil {
			s.notifSvc.NotifyPaymentSuccess(user, fmt.Sprintf("MANUAL-DUE-%d", item.ID), item.Total, fmt.Sprintf("Iuran IPL Periode %s", item.Period))
		}
	}

	return item, nil
}

// Update
func (s *DuesService) Update(id uint, req dto.DuesRequest) (*entity.Dues, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("tagihan tidak ditemukan")
	}

	if item.Status == "paid" {
		return nil, errors.New("tagihan yang sudah lunas tidak dapat diubah")
	}

	if req.Period != "" {
		if !validPeriod(req.Period) {
			return nil, errors.New("format periode tidak valid, gunakan YYYY-MM")
		}
		item.Period = req.Period
	}

	if req.WasteFee >= 0 {
		item.WasteFee = req.WasteFee
	}
	if req.SecurityFee >= 0 {
		item.SecurityFee = req.SecurityFee
	}
	if req.RTFund >= 0 {
		item.RTFund = req.RTFund
	}

	item.Total = item.WasteFee + item.SecurityFee + item.RTFund

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}

// Delete
func (s *DuesService) Delete(id uint) error {
	item, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("tagihan tidak ditemukan")
		}
		return err
	}

	if item.Status == "paid" {
		return errors.New("tagihan yang sudah berstatus lunas tidak boleh dihapus demi integritas pembukuan kas RT")
	}

	return s.repo.Delete(id)
}

func validPeriod(p string) bool {
	if len(p) != 7 {
		return false
	}
	_, err := time.Parse("2006-01", p)
	return err == nil
}
