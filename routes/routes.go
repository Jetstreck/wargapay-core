package routes

import (
	"wargapay-core/controller"
	"wargapay-core/middleware"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo,
	userController *controller.UserController,
	wasteController *controller.WasteServiceController,
	duesController *controller.DuesController,
	weatherController *controller.WeatherController,
	txController *controller.TransactionController,
) {
	api := e.Group("/api/v1")

	// Auth Public Endpoints
	api.POST("/register", userController.Register)
	api.POST("/login", userController.Login)

	// Payment Gateway Webhook (Midtrans Notif Callback) - Tanpa Auth JWT
	api.POST("/payment/webhook", txController.MidtransWebhook)
	api.POST("/transactions/notification", txController.MidtransWebhook)

	// Protected Endpoints untuk Warga & Admin
	userGroup := api.Group("", middleware.JWTAuthMiddleware())
	userGroup.GET("/services", wasteController.List)
	userGroup.GET("/services/:id", wasteController.Get)
	userGroup.GET("/dues", duesController.List)
	userGroup.GET("/dues/:id", duesController.Get)
	userGroup.GET("/weather", weatherController.Current)

	// Transaksi & Payment (Warga & Admin)
	userGroup.POST("/transactions", txController.Create)
	userGroup.GET("/transactions", txController.List)
	userGroup.GET("/transactions/:id", txController.Get)
	userGroup.POST("/transactions/:id/simulate-paid", txController.SimulatePayment)

	// Admin Only Endpoints
	adminGroup := api.Group("/admin", middleware.JWTAuthMiddleware(), middleware.RoleMiddleware("admin"))
	adminGroup.GET("/dashboard", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "Welcome to Admin Dashboard"})
	})

	adminGroup.POST("/services", wasteController.Create)
	adminGroup.PUT("/services/:id", wasteController.Update)
	adminGroup.DELETE("/services/:id", wasteController.Delete)

	adminGroup.POST("/dues", duesController.Create)
	adminGroup.PUT("/dues/:id", duesController.Update)
	adminGroup.DELETE("/dues/:id", duesController.Delete)
	adminGroup.POST("/dues/:id/pay", duesController.Pay)

	// Laporan Transaksi untuk Admin
	adminGroup.GET("/transactions/report", txController.Report)
}
