package controller

import (
	"net/http"
	"strconv"
	"wargapay-core/dto"
	"wargapay-core/service"

	"github.com/labstack/echo/v4"
)

type ExpenseController struct {
	expenseService service.ExpenseService
}

func NewExpenseController(expenseService service.ExpenseService) *ExpenseController {
	return &ExpenseController{expenseService: expenseService}
}

func (ctl *ExpenseController) List(c echo.Context) error {
	category := c.QueryParam("category")
	expenses, err := ctl.expenseService.List(category)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Berhasil mengambil data pengeluaran kas RT",
		"data":    expenses,
	})
}

func (ctl *ExpenseController) Create(c echo.Context) error {
	var req dto.ExpenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	expense, err := ctl.expenseService.Create(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Pengeluaran kas berhasil dicatat",
		"data":    expense,
	})
}

func (ctl *ExpenseController) Delete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ID tidak valid"})
	}

	if err := ctl.expenseService.Delete(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Pengeluaran kas berhasil dihapus"})
}

func (ctl *ExpenseController) Total(c echo.Context) error {
	total, err := ctl.expenseService.GetTotalExpense()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"total_expense": total,
	})
}

