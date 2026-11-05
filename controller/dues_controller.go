package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"wargapay-core/dto"
	"wargapay-core/service"

	"github.com/labstack/echo/v4"
)

type DuesController struct {
	service  *service.DuesService
	notifSvc service.NotificationService
}

func NewDuesController(svc *service.DuesService, notifSvc ...service.NotificationService) *DuesController {
	ctl := &DuesController{service: svc}
	if len(notifSvc) > 0 {
		ctl.notifSvc = notifSvc[0]
	}
	return ctl
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
		"message": "tagihan berhasil dibuat",
		"data":    item,
	})
}

// GetByID
func (ctl *DuesController) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id harus berupa angka",
		})
	}
	item, err := ctl.service.GetByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": item,
	})
}

// Get adalah alias untuk GetByID
func (ctl *DuesController) Get(c echo.Context) error {
	return ctl.GetByID(c)
}

// List
func (ctl *DuesController) List(c echo.Context) error {
	var userID uint
	if uID, ok := currentUserID(c); ok {
		// Jika bukan admin, hanya bisa melihat data miliknya sendiri
		if role, _ := c.Get("role").(string); role != "admin" {
			userID = uID
		}
	}

	period := c.QueryParam("period")
	status := c.QueryParam("status")

	items, err := ctl.service.List(userID, period, status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": items,
	})
}

// PayDues
func (ctl *DuesController) PayDues(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id harus berupa angka",
		})
	}

	item, err := ctl.service.PayDues(uint(id))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "pembayaran iuran berhasil dicatat",
		"data":    item,
	})
}

// Pay adalah alias untuk PayDues
func (ctl *DuesController) Pay(c echo.Context) error {
	return ctl.PayDues(c)
}

// Update
func (ctl *DuesController) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id harus berupa angka",
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

// Delete
func (ctl *DuesController) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id harus berupa angka",
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

	if ctl.notifSvc != nil {
		notifications, unread, err := ctl.notifSvc.GetNotifications(userID)
		if err == nil {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"message": "Berhasil mengambil notifikasi pengingat",
				"data":    notifications,
				"unread":  unread,
			})
		}
	}

	// Fallback ke logic pengambilan dues langsung jika notifSvc tidak diaktifkan
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
			Type:     "dues_reminder",
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

// MarkNotificationsRead menandai semua notifikasi sudah dibaca
func (ctl *DuesController) MarkNotificationsRead(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "identitas token tidak valid"})
	}

	if ctl.notifSvc != nil {
		_ = ctl.notifSvc.MarkAllAsRead(userID)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Semua notifikasi telah ditandai telah dibaca",
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
