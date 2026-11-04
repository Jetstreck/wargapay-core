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
) {
	api := e.Group("/api/v1")

	api.POST("/register", userController.Register)
	api.POST("/login", userController.Login)

	userGroup := api.Group("", middleware.JWTAuthMiddleware())
	userGroup.GET("/services", wasteController.List)
	userGroup.GET("/services/:id", wasteController.Get)
	userGroup.GET("/dues", duesController.List)
	userGroup.GET("/dues/:id", duesController.Get)
	userGroup.GET("/weather", weatherController.Current)

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
}
