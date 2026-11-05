package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"wargapay-core/dto"
	"wargapay-core/entity"
	"wargapay-core/repository"
)

type TransactionService interface {
	CreateTransaction(userID uint, req dto.CreateTransactionRequest) (*entity.Transaction, error)
	HandleMidtransWebhook(payload dto.MidtransNotificationRequest) error
	SimulatePaymentSuccess(txID, userID uint, isAdmin bool) (*entity.Transaction, error)
	ListTransactions(userID uint, txType, status string) ([]entity.Transaction, error)
	GetTransactionByID(txID, userID uint, isAdmin bool) (*entity.Transaction, error)
	GetTransactionReport(startDate, endDate string) (*dto.TransactionReportResponse, error)
}

type transactionService struct {
	txRepo          repository.TransactionRepository
	userRepo        repository.UserRepository
	duesRepo        repository.DuesRepository
	wasteRepo       repository.WasteServiceRepository
	midtransService MidtransService
	emailService    EmailService
	notifSvc        NotificationService
}

func NewTransactionService(
	txRepo repository.TransactionRepository,
	userRepo repository.UserRepository,
	duesRepo repository.DuesRepository,
	wasteRepo repository.WasteServiceRepository,
	midtransService MidtransService,
	emailService EmailService,
	notifSvc ...NotificationService,
) TransactionService {
	svc := &transactionService{
		txRepo:          txRepo,
		userRepo:        userRepo,
		duesRepo:        duesRepo,
		wasteRepo:       wasteRepo,
		midtransService: midtransService,
		emailService:    emailService,
	}
	if len(notifSvc) > 0 {
		svc.notifSvc = notifSvc[0]
	}
	return svc
}

func (s *transactionService) CreateTransaction(userID uint, req dto.CreateTransactionRequest) (*entity.Transaction, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}

	reqType := strings.ToLower(strings.TrimSpace(req.Type))
	if reqType != "dues" && reqType != "waste_service" {
		return nil, errors.New("type transaksi harus 'dues' atau 'waste_service'")
	}

	var amount int64
	var itemName string
	var duesID *uint
	var wasteServiceID *uint
	var pickupDate *time.Time

	if reqType == "dues" {
		if req.DuesID == nil || *req.DuesID == 0 {
			return nil, errors.New("dues_id wajib diisi untuk transaksi pembayaran iuran")
		}
		dues, err := s.duesRepo.FindByID(*req.DuesID)
		if err != nil {
			return nil, errors.New("data tagihan iuran tidak ditemukan")
		}
		if dues.Status == "paid" {
			return nil, errors.New("tagihan iuran ini sudah lunas")
		}
		if user.Role != "admin" && dues.UserID != userID {
			return nil, errors.New("anda tidak memiliki akses untuk membayar tagihan ini")
		}

		amount = dues.Total
		itemName = fmt.Sprintf("Iuran Periode %s", dues.Period)
		duesID = &dues.ID
	} else {
		// waste_service
		if req.WasteServiceID == nil || *req.WasteServiceID == 0 {
			return nil, errors.New("waste_service_id wajib diisi untuk pemesanan penjemputan sampah")
		}
		serviceItem, err := s.wasteRepo.FindByID(*req.WasteServiceID)
		if err != nil {
			return nil, errors.New("layanan sampah tidak ditemukan")
		}
		if !serviceItem.IsActive {
			return nil, errors.New("layanan sampah ini sedang tidak aktif")
		}
		if strings.TrimSpace(req.PickupAddress) == "" {
			return nil, errors.New("alamat penjemputan sampah (pickup_address) wajib diisi")
		}

		if req.PickupDate != "" {
			parsedDate, err := time.Parse("2006-01-02", req.PickupDate)
			if err == nil {
				pickupDate = &parsedDate
			}
		}

		amount = serviceItem.Price
		itemName = fmt.Sprintf("Penjemputan: %s", serviceItem.Name)
		wasteServiceID = &serviceItem.ID
	}

	orderID := fmt.Sprintf("WP-%s-%d-%d", strings.ToUpper(reqType[:4]), userID, time.Now().Unix())

	// Request Snap Token ke Midtrans Sandbox
	snapToken, redirectURL, err := s.midtransService.CreateSnapTransaction(
		orderID,
		amount,
		user.Name,
		user.Email,
		itemName,
	)
	if err != nil {
		log.Printf("[CREATE SNAP ERROR] %v", err)
		return nil, fmt.Errorf("gagal membuat transaksi midtrans: %w", err)
	}

	tx := &entity.Transaction{
		UserID:          userID,
		OrderID:         orderID,
		Type:            reqType,
		DuesID:          duesID,
		WasteServiceID:  wasteServiceID,
		Amount:          amount,
		Status:          "pending",
		SnapToken:       snapToken,
		SnapRedirectURL: redirectURL,
		Notes:           req.Notes,
		PickupAddress:   req.PickupAddress,
		PickupDate:      pickupDate,
	}

	if err := s.txRepo.Create(tx); err != nil {
		return nil, fmt.Errorf("gagal membuat transaksi: %w", err)
	}

	return tx, nil
}

func (s *transactionService) HandleMidtransWebhook(payload dto.MidtransNotificationRequest) error {
	tx, err := s.txRepo.FindByOrderID(payload.OrderID)
	if err != nil {
		return fmt.Errorf("transaksi dengan order_id %s tidak ditemukan: %w", payload.OrderID, err)
	}

	// Verifikasi signature key
	if payload.SignatureKey != "" {
		isValid := s.midtransService.VerifySignature(
			payload.OrderID,
			payload.StatusCode,
			payload.GrossAmount,
			payload.SignatureKey,
		)
		if !isValid {
			log.Printf("[MIDTRANS WEBHOOK] Peringatan: Signature key tidak cocok untuk order %s (kemungkinan pengujian sandbox lokal)", payload.OrderID)
		}
	}

	status := strings.ToLower(payload.TransactionStatus)
	fraud := strings.ToLower(payload.FraudStatus)

	now := time.Now()
	switch status {
	case "capture":
		if fraud == "challenge" {
			tx.Status = "challenge"
		} else if fraud == "accept" {
			tx.Status = "paid"
			tx.PaidAt = &now
		}
	case "settlement":
		tx.Status = "paid"
		tx.PaidAt = &now
	case "cancel", "deny", "expire":
		tx.Status = "failed"
	case "pending":
		tx.Status = "pending"
	}

	if payload.PaymentType != "" {
		tx.PaymentType = payload.PaymentType
	}

	// Jika transaksi lunas, sinkronkan status tagihan iuran & kirim email konfirmasi
	if tx.Status == "paid" {
		if tx.Type == "dues" && tx.DuesID != nil {
			dues, err := s.duesRepo.FindByID(*tx.DuesID)
			if err == nil && dues != nil {
				dues.Status = "paid"
				dues.PaidAt = &now
				_ = s.duesRepo.Update(dues)
			}
		}

		// Kirim email notifikasi sukses bayar & catat ke notifikasi
		if s.notifSvc != nil {
			desc := "Layanan Sampah Warga"
			if tx.Type == "dues" {
				desc = "Iuran IPL Lingkungan"
			}
			s.notifSvc.NotifyPaymentSuccess(&tx.User, tx.OrderID, tx.Amount, desc)
		} else if tx.User.Email != "" {
			body := fmt.Sprintf("Halo %s,\n\nPembayaran transaksi Anda berhasil!\nOrder ID: %s\nTotal: Rp %d\nMetode: %s\nWaktu: %s\n\nTerima kasih atas partisipasi Anda di WargaPay.",
				tx.User.Name, tx.OrderID, tx.Amount, tx.PaymentType, now.Format("02-01-2006 15:04:05"))
			go s.emailService.SendNotificationEmail(tx.User.Email, "Konfirmasi Pembayaran Sukses - WargaPay", body)
		}
	}

	return s.txRepo.Update(tx)
}

func (s *transactionService) SimulatePaymentSuccess(txID, userID uint, isAdmin bool) (*entity.Transaction, error) {
	tx, err := s.txRepo.FindByID(txID)
	if err != nil {
		return nil, errors.New("transaksi tidak ditemukan")
	}

	if !isAdmin && tx.UserID != userID {
		return nil, errors.New("akses ditolak")
	}

	if tx.Status == "paid" {
		return tx, nil
	}

	now := time.Now()
	tx.Status = "paid"
	tx.PaidAt = &now
	tx.PaymentType = "sandbox_simulated"

	if tx.Type == "dues" && tx.DuesID != nil {
		dues, err := s.duesRepo.FindByID(*tx.DuesID)
		if err == nil && dues != nil {
			dues.Status = "paid"
			dues.PaidAt = &now
			_ = s.duesRepo.Update(dues)
		}
	}

	if err := s.txRepo.Update(tx); err != nil {
		return nil, err
	}

	if s.notifSvc != nil {
		desc := "Layanan Sampah Warga (Sandbox)"
		if tx.Type == "dues" {
			desc = "Iuran IPL Lingkungan (Sandbox)"
		}
		s.notifSvc.NotifyPaymentSuccess(&tx.User, tx.OrderID, tx.Amount, desc)
	} else if tx.User.Email != "" {
		body := fmt.Sprintf("Halo %s,\n\nPembayaran Anda via Sandbox berhasil dicatat.\nOrder ID: %s\nTotal: Rp %d\nWaktu: %s\n\nTerima kasih.",
			tx.User.Name, tx.OrderID, tx.Amount, now.Format("02-01-2006 15:04:05"))
		go s.emailService.SendNotificationEmail(tx.User.Email, "Konfirmasi Pembayaran Sandbox - WargaPay", body)
	}

	return tx, nil
}

func (s *transactionService) ListTransactions(userID uint, txType, status string) ([]entity.Transaction, error) {
	return s.txRepo.FindAll(userID, txType, status)
}

func (s *transactionService) GetTransactionByID(txID, userID uint, isAdmin bool) (*entity.Transaction, error) {
	tx, err := s.txRepo.FindByID(txID)
	if err != nil {
		return nil, errors.New("transaksi tidak ditemukan")
	}
	if !isAdmin && tx.UserID != userID {
		return nil, errors.New("akses ditolak: anda bukan pemilik transaksi ini")
	}
	return tx, nil
}

func (s *transactionService) GetTransactionReport(startDate, endDate string) (*dto.TransactionReportResponse, error) {
	return s.txRepo.GetReport(startDate, endDate)
}

