package controller

import (
	"net/http"
	"strconv"
	"strings"
	"wargapay-core/dto"
	"wargapay-core/service"

	"github.com/labstack/echo/v4"
)

type WasteServiceController struct{ service *service.WasteServiceService }

func NewWasteServiceController(svc *service.WasteServiceService) *WasteServiceController {
	return &WasteServiceController{service: svc}
}

// Create
func (ctl *WasteServiceController) Create(c echo.Context) error {
	var req dto.WasteServiceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "format request tidak valid",
		})
	}
	item, err := ctl.service.Create(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "layanan berhasil dibuat",
		"data":    item,
	})
}

// List
func (ctl *WasteServiceController) List(c echo.Context) error {
	filters := map[string]interface{}{"is_active": true}
	if name := strings.TrimSpace(c.QueryParam("name")); name != "" {
		filters["name"] = name
	}
	if value := c.QueryParam("min_price"); value != "" {
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil || n < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "min_price harus berupa angka >= 0",
			})
		}
		filters["min_price"] = n
	}
	if value := c.QueryParam("max_price"); value != "" {
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil || n < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "max_price harus berupa angka >= 0",
			})
		}
		filters["max_price"] = n
	}
	if value := c.QueryParam("is_active"); value != "" {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "is_active harus true atau false",
			})
		}
		filters["is_active"] = b
	}
	items, err := ctl.service.List(filters)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "gagal mengambil layanan",
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  items,
		"total": len(items),
	})
}

// Get
func (ctl *WasteServiceController) Get(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id tidak valid",
		})
	}
	item, err := ctl.service.Get(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": item,
	})
}

// Update
func (ctl *WasteServiceController) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id tidak valid",
		})
	}
	var req dto.WasteServiceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "format request tidak valid",
		})
	}
	item, err := ctl.service.Update(uint(id), req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "layanan berhasil diperbarui",
		"data":    item,
	})
}

// Delete
func (ctl *WasteServiceController) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id tidak valid",
		})
	}
	if err := ctl.service.Delete(uint(id)); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"message": "layanan berhasil dihapus",
	})
}
