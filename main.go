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
	// Jika file .env tidak ditemukan (misal di production), program tidak akan error (diabaikan)
	_ = godotenv.Load()

	// Menghubungkan ke database PostgreSQL
	config.ConnectDB()

	// Melakukan auto-migrate untuk entity User
	err := config.DB.AutoMigrate(&entity.User{}, &entity.WasteService{}, &entity.Dues{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// Inisialisasi Service, Repository, dan Controller
	emailSvc := service.NewEmailService()
	userRepo := repository.NewUserRepository(config.DB)
	userSvc := service.NewUserService(userRepo, emailSvc)
	userCtl := controller.NewUserController(userSvc)

	// Inisialisasi fitur layanan sampah, tagihan iuran, dan informasi cuaca.
	wasteRepo := repository.NewWasteServiceRepository(config.DB)
	wasteSvc := service.NewWasteServiceService(wasteRepo)
	wasteCtl := controller.NewWasteServiceController(wasteSvc)

	duesRepo := repository.NewDuesRepository(config.DB)
	duesSvc := service.NewDuesService(duesRepo)
	duesCtl := controller.NewDuesController(duesSvc)

	weatherSvc := service.NewWeatherService()
	weatherCtl := controller.NewWeatherController(weatherSvc)

	// Inisialisasi dan jalankan Scheduler Service
	scheduler := service.NewSchedulerService(emailSvc)
	scheduler.Start()

	// Inisialisasi framework Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Menyiapkan routing aplikasi
	routes.SetupRoutes(e, userCtl, wasteCtl, duesCtl, weatherCtl)

	// Menentukan port server (default ke 8080 jika tidak ada di env)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
