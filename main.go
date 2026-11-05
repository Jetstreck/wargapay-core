package main

import (
	"log"
	"os"

	"wargapay-core/config"
	"wargapay-core/controller"
	"wargapay-core/entity"
	"wargapay-core/repository"
	"wargapay-core/routes"
	"wargapay-core/service"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Memuat variabel dari file .env (untuk development lokal)
	_ = godotenv.Load()

	// Menghubungkan ke database PostgreSQL
	config.ConnectDB()

	// Melakukan auto-migrate untuk entity User, WasteService, Dues, Transaction, Expense, dan Notification
	err := config.DB.AutoMigrate(
		&entity.User{},
		&entity.WasteService{},
		&entity.Dues{},
		&entity.Transaction{},
		&entity.Expense{},
		&entity.Notification{},
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// Inisialisasi Repository Inti & Layanan Notifikasi Terpadu (Mailtrap + Dashboard)
	emailSvc := service.NewEmailService()
	userRepo := repository.NewUserRepository(config.DB)
	duesRepo := repository.NewDuesRepository(config.DB)
	notifRepo := repository.NewNotificationRepository(config.DB)

	notifSvc := service.NewNotificationService(notifRepo, userRepo, duesRepo, emailSvc)

	// Inisialisasi Service dan Controller Pengguna
	userSvc := service.NewUserService(userRepo, emailSvc, notifSvc)
	userCtl := controller.NewUserController(userSvc)

	// Inisialisasi fitur layanan sampah, tagihan iuran, dan informasi cuaca
	wasteRepo := repository.NewWasteServiceRepository(config.DB)
	wasteSvc := service.NewWasteServiceService(wasteRepo)
	wasteCtl := controller.NewWasteServiceController(wasteSvc)

	duesSvc := service.NewDuesService(duesRepo, userRepo, notifSvc)
	duesCtl := controller.NewDuesController(duesSvc, notifSvc)

	weatherSvc := service.NewWeatherService()
	weatherCtl := controller.NewWeatherController(weatherSvc)

	// Inisialisasi Payment Gateway (Midtrans) & Transaksi
	midtransSvc := service.NewMidtransService()
	txRepo := repository.NewTransactionRepository(config.DB)
	txSvc := service.NewTransactionService(txRepo, userRepo, duesRepo, wasteRepo, midtransSvc, emailSvc, notifSvc)
	txCtl := controller.NewTransactionController(txSvc)

	// Inisialisasi Pengeluaran Kas RT (Expenses)
	expenseRepo := repository.NewExpenseRepository(config.DB)
	expenseSvc := service.NewExpenseService(expenseRepo)
	expenseCtl := controller.NewExpenseController(expenseSvc)

	// Inisialisasi dan jalankan Scheduler Service (Pengingat Tagihan Otomatis)
	scheduler := service.NewSchedulerService(emailSvc, duesRepo, userRepo, notifSvc)
	scheduler.Start()

	// Inisialisasi framework Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Middleware CORS agar siap terintegrasi dengan frontend
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// Registrasi semua rute API
	routes.SetupRoutes(e, userCtl, wasteCtl, duesCtl, weatherCtl, txCtl, expenseCtl)

	// Menentukan port server (default 8080 jika tidak ada di env)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Menjalankan server Echo
	log.Printf("Server WargaPay berjalan di port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
