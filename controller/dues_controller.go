package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"wargapay-core/dto"
	"wargapay-core/service"

	"github.com/labstack/echo/v4"
)

type DuesController struct{ service *service.DuesService }

func NewDuesController(svc *service.DuesService) *DuesController {
	return &DuesController{service: svc}
}

// Create
func (ctl *DuesController) Create(c echo.Context) error {
	var req dto.DuesRequest
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
		"message": "tagihan berhasil dibuat", "data": item,
	})
}

// List
func (ctl *DuesController) List(c echo.Context) error {
	role := c.Get("role")
	userID, ok := currentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "identitas token tidak valid",
		})
	}
	filterUserID := userID
	if role == "admin" {
		filterUserID = 0
		if c.QueryParam("user_id") != "" {
			parsed, err := strconv.ParseUint(c.QueryParam("user_id"), 10, 64)
			if err != nil || parsed == 0 {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "user_id tidak valid",
				})
			}
			filterUserID = uint(parsed)
		}
	}
	items, err := ctl.service.List(filterUserID, c.QueryParam("period"), c.QueryParam("status"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": items, "total": len(items),
	})
}

// Get
func (ctl *DuesController) Get(c echo.Context) error {
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
	userID, ok := currentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "identitas token tidak valid",
		})
	}
	if c.Get("role") != "admin" && item.UserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "anda tidak berhak melihat tagihan ini",
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": item,
	})
}

// Update
func (ctl *DuesController) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id tidak valid",
		})
	}
	var req dto.DuesRequest
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
		"message": "tagihan berhasil diperbarui",
		"data":    item,
	})
}

// Pay
func (ctl *DuesController) Pay(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id tidak valid",
		})
	}

	item, err := ctl.service.MarkPaid(uint(id), 0, true)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "pembayaran berhasil dicatat",
		"data":    item,
	})
}

// Delete
func (ctl *DuesController) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id tidak valid",
		})
	}
	if err := ctl.service.Delete(uint(id)); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"message": "tagihan berhasil dihapus",
	})
}

// Notifications - Endpoint pengingat tagihan iuran & broadcast jadwal sanitasi
func (ctl *DuesController) Notifications(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "identitas token tidak valid",
		})
	}

	items, err := ctl.service.List(userID, "", "unpaid")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	type NotificationItem struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Title    string `json:"title"`
		Message  string `json:"message"`
		Date     string `json:"date"`
		Priority string `json:"priority"`
		DuesID   uint   `json:"dues_id,omitempty"`
		Amount   int64  `json:"amount,omitempty"`
		Period   string `json:"period,omitempty"`
	}

	var notifications []NotificationItem

	for _, d := range items {
		notifications = append(notifications, NotificationItem{
			ID:       fmt.Sprintf("notif-due-%d", d.ID),
			Type:     "payment_reminder",
			Title:    fmt.Sprintf("Pengingat Tagihan IPL Periode %s", d.Period),
			Message:  fmt.Sprintf("Tagihan iuran IPL periode %s sebesar Rp %d belum diselesaikan. Jatuh tempo tanggal 10. Bayar mudah via Midtrans QRIS.", d.Period, d.Total),
			Date:     "Pengingat Aktif",
			Priority: "high",
			DuesID:   d.ID,
			Amount:   d.Total,
			Period:   d.Period,
		})
	}

	// Notifikasi broadcast sistem sanitasi
	notifications = append(notifications, NotificationItem{
		ID:       "notif-broadcast-sanitasi",
		Type:     "info",
		Title:    "Jadwal Armada Pengangkutan Sampah RT 05",
		Message:  "Truk kebersihan beroperasi setiap Senin, Rabu, dan Jumat (06:30 - 08:30 WITA). Mohon sampah diletakkan di depan pagar sebelum pukul 06:00 WITA.",
		Date:     "Jadwal Rutin",
		Priority: "normal",
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Berhasil mengambil notifikasi pengingat",
		"data":    notifications,
		"unread":  len(items),
	})
}

func currentUserID(c echo.Context) (uint, bool) {
	switch value := c.Get("user_id").(type) {
	case float64:
		return uint(value), value > 0
	case int:
		return uint(value), value > 0
	case uint:
		return value, value > 0
	case string:
		id, err := strconv.ParseUint(value, 10, 64)
		return uint(id), err == nil && id > 0
	default:
		return 0, false
	}
}
