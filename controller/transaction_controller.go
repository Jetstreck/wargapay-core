package controller

import (
	"net/http"
	"strconv"
	"wargapay-core/dto"
	"wargapay-core/service"

	"github.com/labstack/echo/v4"
)

type TransactionController struct {
	service service.TransactionService
}

func NewTransactionController(svc service.TransactionService) *TransactionController {
	return &TransactionController{service: svc}
}

// Create membuat pesanan iuran / penjemputan sampah dan menghasilkan Snap Token Midtrans
func (ctl *TransactionController) Create(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "identitas token tidak valid",
		})
	}

	var req dto.CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "format request tidak valid: " + err.Error(),
		})
	}

	tx, err := ctl.service.CreateTransaction(userID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "pesanan berhasil dibuat, silakan lanjutkan pembayaran",
		"data":    tx,
	})
}

// MidtransWebhook menerima notifikasi callback HTTP dari server Midtrans
func (ctl *TransactionController) MidtransWebhook(c echo.Context) error {
	var payload dto.MidtransNotificationRequest
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "payload notifikasi tidak valid: " + err.Error(),
		})
	}

	if err := ctl.service.HandleMidtransWebhook(payload); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "notifikasi pembayaran berhasil diproses",
	})
}

// SimulatePayment simulasi pembayaran sandbox / local testing tanpa perlu webhook publik
func (ctl *TransactionController) SimulatePayment(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id transaksi tidak valid",
		})
	}

	userID, ok := currentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "identitas token tidak valid",
		})
	}

	isAdmin := c.Get("role") == "admin"
	tx, err := ctl.service.SimulatePaymentSuccess(uint(id), userID, isAdmin)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "simulasi pembayaran berhasil, tagihan/layanan lunas",
		"data":    tx,
	})
}

// List menampilkan riwayat transaksi warga (atau semua bagi admin)
func (ctl *TransactionController) List(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "identitas token tidak valid",
		})
	}

	filterUserID := userID
	if c.Get("role") == "admin" {
		filterUserID = 0
		if c.QueryParam("user_id") != "" {
			parsed, err := strconv.ParseUint(c.QueryParam("user_id"), 10, 64)
			if err == nil && parsed > 0 {
				filterUserID = uint(parsed)
			}
		}
	}

	txType := c.QueryParam("type")
	status := c.QueryParam("status")

	items, err := ctl.service.ListTransactions(filterUserID, txType, status)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  items,
		"total": len(items),
	})
}

// Get mengambil detail transaksi berdasarkan ID
func (ctl *TransactionController) Get(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "id transaksi tidak valid",
		})
	}

	userID, ok := currentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "identitas token tidak valid",
		})
	}

	isAdmin := c.Get("role") == "admin"
	tx, err := ctl.service.GetTransactionByID(uint(id), userID, isAdmin)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": tx,
	})
}

// Report laporan rekapitulasi keuangan dan transaksi untuk admin
func (ctl *TransactionController) Report(c echo.Context) error {
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	report, err := ctl.service.GetTransactionReport(startDate, endDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "laporan transaksi berhasil diambil",
		"data":    report,
	})
}
