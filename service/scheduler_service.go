package service

import (
	"log"

	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	cron         *cron.Cron
	emailService EmailService
}

func NewSchedulerService(emailService EmailService) *SchedulerService {
	return &SchedulerService{
		cron:         cron.New(cron.WithLocation(cron.Location("Asia/Jakarta"))),
		emailService: emailService,
	}
}

func (s *SchedulerService) Start() {
	s.cron.AddFunc("0 8 1 * *", func() {
		log.Println("[CRONJOB] Menjalankan pengingat tagihan iuran bulanan...")
	})

	s.cron.AddFunc("0 0 * * *", func() {
		log.Println("[CRONJOB] Menjalankan tugas rutin backup database sistem...")
	})

	s.cron.Start()
	log.Println("Background job scheduler started successfully")
}

func (s *SchedulerService) Stop() {
	s.cron.Stop()
}
