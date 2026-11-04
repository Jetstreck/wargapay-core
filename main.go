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

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	config.ConnectDB()

	err := config.DB.AutoMigrate(&entity.User{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	emailSvc := service.NewEmailService()
	userRepo := repository.NewUserRepository(config.DB)
	userSvc := service.NewUserService(userRepo, emailSvc)
	userCtl := controller.NewUserController(userSvc)

	scheduler := service.NewSchedulerService(emailSvc)
	scheduler.Start()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	routes.SetupRoutes(e, userCtl)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
