package service

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	cron         *cron.Cron
	emailService EmailService
}

func NewSchedulerService(emailService EmailService) *SchedulerService {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("Failed to load timezone Asia/Jakarta, fallback to UTC: %v", err)
		loc = time.UTC
	}

	return &SchedulerService{
		cron:         cron.New(cron.WithLocation(loc)),
		emailService: emailService,
	}
}

func (s *SchedulerService) Start() {
	// Job 1: Pengingat tagihan iuran sampah (berjalan setiap tanggal 1 jam 08:00 pagi)
	s.cron.AddFunc("0 8 1 * *", func() {
		log.Println("[CRONJOB] Menjalankan pengingat tagihan iuran bulanan...")
	})

	// Job 2: Background backup / maintenance rutin database (berjalan setiap tengah malam)
	s.cron.AddFunc("0 0 * * *", func() {
		log.Println("[CRONJOB] Menjalankan tugas rutin backup database sistem...")
	})

	s.cron.Start()
	log.Println("Background job scheduler started successfully")
}

func (s *SchedulerService) Stop() {
	s.cron.Stop()
}