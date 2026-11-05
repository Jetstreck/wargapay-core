package controller

import (
	"net/http"
	"strconv"
	"wargapay-core/service"

	"github.com/labstack/echo/v4"
)

type WeatherController struct{ service *service.WeatherService }

func NewWeatherController(svc *service.WeatherService) *WeatherController {
	return &WeatherController{service: svc}
}

// Mengambil cuaca berdasarkan latitude dan longitude
func (ctl *WeatherController) Current(c echo.Context) error {
	lat, errLat := strconv.ParseFloat(c.QueryParam("lat"), 64)
	lon, errLon := strconv.ParseFloat(c.QueryParam("lon"), 64)
	if errLat != nil || errLon != nil || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "lat dan lon wajib diisi dengan koordinat yang valid",
		})
	}
	result, err := ctl.service.GetCurrentWeather(c.Request().Context(), lat, lon)
	if err != nil {
		return c.JSON(http.StatusBadGateway, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "data cuaca berhasil diambil", "data": result,
	})
}
