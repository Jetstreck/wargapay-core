package service

import (
	"errors"
	"fmt"
	"time"
	"wargapay-core/dto"
	"wargapay-core/entity"
	"wargapay-core/repository"
)

type ExpenseService interface {
	Create(req dto.ExpenseRequest) (*entity.Expense, error)
	List(category string) ([]entity.Expense, error)
	Delete(id uint) error
	GetTotalExpense() (int64, error)
}

type expenseService struct {
	repo repository.ExpenseRepository
}

func NewExpenseService(repo repository.ExpenseRepository) ExpenseService {
	return &expenseService{repo: repo}
}

func (s *expenseService) Create(req dto.ExpenseRequest) (*entity.Expense, error) {
	if req.Title == "" || req.Amount <= 0 {
		return nil, errors.New("judul dan nominal pengeluaran harus diisi")
	}

	dateStr := req.Date
	if dateStr == "" {
		dateStr = time.Now().Format("02 Jan 2006")
	}

	receiptRef := req.ReceiptRef
	if receiptRef == "" {
		receiptRef = fmt.Sprintf("KW-%d-%04d", time.Now().Year(), time.Now().Unix()%10000)
	}

	approvedBy := req.ApprovedBy
	if approvedBy == "" {
		approvedBy = "Pengurus RT 05"
	}

	expense := &entity.Expense{
		Title:       req.Title,
		Category:    req.Category,
		Amount:      req.Amount,
		Vendor:      req.Vendor,
		Date:        dateStr,
		Description: req.Description,
		ReceiptRef:  receiptRef,
		ApprovedBy:  approvedBy,
	}

	if err := s.repo.Create(expense); err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *expenseService) List(category string) ([]entity.Expense, error) {
	return s.repo.FindAll(category)
}

func (s *expenseService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *expenseService) GetTotalExpense() (int64, error) {
	return s.repo.GetTotalExpense()
}

