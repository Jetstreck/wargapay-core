package routes

import (
	"wargapay-core/controller"
	"wargapay-core/middleware"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, userController *controller.UserController) {
	api := e.Group("/api/v1")

	api.POST("/register", userController.Register)
	api.POST("/login", userController.Login)

	adminGroup := api.Group("/admin", middleware.JWTAuthMiddleware(), middleware.RoleMiddleware("admin"))
	adminGroup.GET("/dashboard", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "Welcome to Admin Dashboard"})
	})
}
