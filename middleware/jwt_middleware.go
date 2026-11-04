package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"message": "Missing or malformed JWT token",
				})
			}

			tokenString := strings.Split(authHeader, " ")[1]
			secret := os.Getenv("JWT_SECRET")

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"message": "Invalid or expired JWT token",
				})
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"message": "Invalid token claims",
				})
			}

			c.Set("user_id", claims["id"])
			c.Set("email", claims["email"])
			c.Set("role", claims["role"])

			return next(c)
		}
	}
}

func RoleMiddleware(requiredRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := c.Get("role")
			if role != requiredRole {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"message": "Access denied: insufficient permissions",
				})
			}
			return next(c)
		}
	}
}