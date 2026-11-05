package service

import (
	"fmt"
	"time"
	"wargapay-core/entity"
	"wargapay-core/repository"
)

type NotificationService interface {
	NotifyWelcome(user *entity.User)
	NotifyDuesCreated(user *entity.User, dues *entity.Dues)
	NotifyDuesReminder(user *entity.User, dues *entity.Dues)
	NotifyPaymentSuccess(user *entity.User, orderID string, amount int64, description string)
	GetNotifications(userID uint) ([]map[string]interface{}, int, error)
	MarkAllAsRead(userID uint) error
}

type notificationService struct {
	notifRepo repository.NotificationRepository
	userRepo  repository.UserRepository
	duesRepo  repository.DuesRepository
	emailSvc  EmailService
}

func NewNotificationService(
	notifRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	duesRepo repository.DuesRepository,
	emailSvc EmailService,
) NotificationService {
	return &notificationService{
		notifRepo: notifRepo,
		userRepo:  userRepo,
		duesRepo:  duesRepo,
		emailSvc:  emailSvc,
	}
}

// NotifyWelcome mengirim email selamat datang ke Mailtrap dan mencatat notifikasi di dashboard
func (s *notificationService) NotifyWelcome(user *entity.User) {
	if user == nil || user.Email == "" {
		return
	}

	title := "Selamat Datang di WargaPay"
	msg := fmt.Sprintf("Halo %s, selamat bergabung di lingkungan RT 05 Cluster Diamond. Akun warga Anda telah aktif dan siap digunakan untuk memantau iuran IPL dan layanan sanitasi.", user.Name)

	nowStr := time.Now().Format("02 Jan 2006, 15:04")
	notif := &entity.Notification{
		UserID:   user.ID,
		Type:     "welcome",
		Title:    title,
		Message:  msg,
		Date:     nowStr,
		Priority: "normal",
		IsRead:   false,
	}

	_ = s.notifRepo.Create(notif)
	go s.emailSvc.SendNotificationEmail(user.Email, title, msg)
}

// NotifyDuesCreated mengirim email tagihan baru ke Mailtrap dan mencatat notifikasi
func (s *notificationService) NotifyDuesCreated(user *entity.User, dues *entity.Dues) {
	if user == nil || user.Email == "" || dues == nil {
		return
	}

	title := fmt.Sprintf("Tagihan IPL Periode %s Diterbitkan", dues.Period)
	msg := fmt.Sprintf("Halo %s, pengurus RT telah menerbitkan tagihan iuran IPL periode %s sebesar Rp %d (Kebersihan: Rp %d, Keamanan: Rp %d, Kas: Rp %d). Jatuh tempo tanggal 10. Bayar praktis via QRIS/Midtrans.",
		user.Name, dues.Period, dues.Total, dues.WasteFee, dues.SecurityFee, dues.RTFund)

	nowStr := time.Now().Format("02 Jan 2006, 15:04")
	notif := &entity.Notification{
		UserID:   user.ID,
		Type:     "dues_reminder",
		Title:    title,
		Message:  msg,
		Date:     nowStr,
		Priority: "high",
		IsRead:   false,
		DuesID:   dues.ID,
		Amount:   dues.Total,
		Period:   dues.Period,
	}

	_ = s.notifRepo.Create(notif)
	go s.emailSvc.SendNotificationEmail(user.Email, title, msg)
}

// NotifyDuesReminder mengirim pengingat tagihan jatuh tempo ke Mailtrap dan mencatat notifikasi
func (s *notificationService) NotifyDuesReminder(user *entity.User, dues *entity.Dues) {
	if user == nil || user.Email == "" || dues == nil {
		return
	}

	title := fmt.Sprintf("Pengingat: Tagihan IPL Periode %s Belum Selesai", dues.Period)
	msg := fmt.Sprintf("Halo %s, ini adalah pengingat bahwa tagihan iuran IPL periode %s sebesar Rp %d belum diselesaikan. Mohon segera selesaikan sebelum tanggal 10 demi kelancaran operasional lingkungan.",
		user.Name, dues.Period, dues.Total)

	nowStr := time.Now().Format("02 Jan 2006, 15:04")
	notif := &entity.Notification{
		UserID:   user.ID,
		Type:     "dues_reminder",
		Title:    title,
		Message:  msg,
		Date:     nowStr,
		Priority: "high",
		IsRead:   false,
		DuesID:   dues.ID,
		Amount:   dues.Total,
		Period:   dues.Period,
	}

	_ = s.notifRepo.Create(notif)
	go s.emailSvc.SendNotificationEmail(user.Email, title, msg)
}

// NotifyPaymentSuccess mengirim konfirmasi pelunasan ke Mailtrap dan mencatat notifikasi
func (s *notificationService) NotifyPaymentSuccess(user *entity.User, orderID string, amount int64, description string) {
	if user == nil || user.Email == "" {
		return
	}

	title := "Konfirmasi Pembayaran Sukses - WargaPay"
	msg := fmt.Sprintf("Terima kasih %s. Pembayaran Anda untuk %s (Order ID: %s) sebesar Rp %d telah berhasil diverifikasi lunas secara otomatis oleh Midtrans.",
		user.Name, description, orderID, amount)

	nowStr := time.Now().Format("02 Jan 2006, 15:04")
	notif := &entity.Notification{
		UserID:   user.ID,
		Type:     "payment_success",
		Title:    title,
		Message:  msg,
		Date:     nowStr,
		Priority: "normal",
		IsRead:   false,
	}

	_ = s.notifRepo.Create(notif)
	go s.emailSvc.SendNotificationEmail(user.Email, title, msg)
}

// GetNotifications mengambil daftar notifikasi terpadu untuk warga
func (s *notificationService) GetNotifications(userID uint) ([]map[string]interface{}, int, error) {
	records, err := s.notifRepo.FindByUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	// Jika pengguna belum memiliki riwayat notifikasi di tabel,
	// buatkan otomatis notifikasi selamat datang agar tidak kosong
	if len(records) == 0 {
		user, uErr := s.userRepo.FindByID(userID)
		if uErr == nil && user != nil {
			s.NotifyWelcome(user)
			records, _ = s.notifRepo.FindByUserID(userID)
		}
	}

	var results []map[string]interface{}
	unreadCount := 0

	for _, n := range records {
		if !n.IsRead {
			unreadCount++
		}
		item := map[string]interface{}{
			"id":       fmt.Sprintf("notif-%d", n.ID),
			"raw_id":   n.ID,
			"type":     n.Type,
			"title":    n.Title,
			"message":  n.Message,
			"date":     n.Date,
			"priority": n.Priority,
			"is_read":  n.IsRead,
		}
		if n.DuesID > 0 {
			item["dues_id"] = n.DuesID
			item["amount"] = n.Amount
			item["period"] = n.Period
		}
		results = append(results, item)
	}

	// Tambahkan broadcast berkala lingkungan sanitasi
	results = append(results, map[string]interface{}{
		"id":       "notif-broadcast-sanitasi",
		"type":     "info",
		"title":    "Jadwal Armada Pengangkutan Sampah RT 05",
		"message":  "Truk kebersihan beroperasi setiap Senin, Rabu, dan Jumat (06:30 - 08:30 WITA). Mohon sampah diletakkan di depan pagar sebelum pukul 06:00 WITA.",
		"date":     "Jadwal Rutin",
		"priority": "normal",
		"is_read":  true,
	})

	return results, unreadCount, nil
}

func (s *notificationService) MarkAllAsRead(userID uint) error {
	return s.notifRepo.MarkAllAsRead(userID)
}

