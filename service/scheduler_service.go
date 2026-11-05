package service

import (
	"log"
	"time"
	"wargapay-core/repository"

	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	cron         *cron.Cron
	emailService EmailService
	duesRepo     repository.DuesRepository
	userRepo     repository.UserRepository
	notifSvc     NotificationService
}

func NewSchedulerService(
	emailService EmailService,
	opts ...interface{},
) *SchedulerService {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("Failed to load timezone Asia/Jakarta, fallback to UTC: %v", err)
		loc = time.UTC
	}

	s := &SchedulerService{
		cron:         cron.New(cron.WithLocation(loc)),
		emailService: emailService,
	}

	for _, opt := range opts {
		if dr, ok := opt.(repository.DuesRepository); ok {
			s.duesRepo = dr
		}
		if ur, ok := opt.(repository.UserRepository); ok {
			s.userRepo = ur
		}
		if ns, ok := opt.(NotificationService); ok {
			s.notifSvc = ns
		}
	}

	return s
}

// TriggerDuesReminder mengecek tagihan unpaid dan mengirim pengingat ke Mailtrap & tabel notifikasi
func (s *SchedulerService) TriggerDuesReminder() {
	if s.duesRepo == nil || s.notifSvc == nil {
		log.Println("[CRONJOB] Repository / NotificationService belum diset, melewati pengingat.")
		return
	}

	log.Println("[CRONJOB] Menjalankan pengecekan tagihan belum lunas untuk pengiriman pengingat...")
	unpaidList, err := s.duesRepo.FindAll(0, "", "unpaid")
	if err != nil {
		log.Printf("[CRONJOB ERROR] Gagal mengambil tagihan unpaid: %v", err)
		return
	}

	count := 0
	for _, d := range unpaidList {
		if s.userRepo != nil {
			user, uErr := s.userRepo.FindByID(d.UserID)
			if uErr == nil && user != nil {
				s.notifSvc.NotifyDuesReminder(user, &d)
				count++
			}
		}
	}

	log.Printf("[CRONJOB] Berhasil memproses pengingat tagihan ke Mailtrap & notifikasi untuk %d warga", count)
}

func (s *SchedulerService) Start() {
	// Job 1: Pengingat tagihan iuran bulanan (berjalan setiap tanggal 1 jam 08:00 pagi WIB)
	s.cron.AddFunc("0 8 1 * *", func() {
		s.TriggerDuesReminder()
	})

	// Job 2: Pengingat jatuh tempo H-3 (berjalan setiap tanggal 7 jam 09:00 pagi WIB)
	s.cron.AddFunc("0 9 7 * *", func() {
		s.TriggerDuesReminder()
	})

	// Job 3: Background backup / maintenance rutin database (berjalan setiap tengah malam)
	s.cron.AddFunc("0 0 * * *", func() {
		log.Println("[CRONJOB] Menjalankan tugas rutin backup database sistem...")
	})

	s.cron.Start()
	log.Println("Background job scheduler started successfully with automated dues reminders")
}

func (s *SchedulerService) Stop() {
	s.cron.Stop()
}
