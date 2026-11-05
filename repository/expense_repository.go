package repository

import (
	"wargapay-core/entity"

	"gorm.io/gorm"
)

type ExpenseRepository interface {
	Create(expense *entity.Expense) error
	FindAll(category string) ([]entity.Expense, error)
	FindByID(id uint) (*entity.Expense, error)
	Delete(id uint) error
	GetTotalExpense() (int64, error)
}

type expenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) Create(expense *entity.Expense) error {
	return r.db.Create(expense).Error
}

func (r *expenseRepository) FindAll(category string) ([]entity.Expense, error) {
	var expenses []entity.Expense
	query := r.db.Order("id desc")
	if category != "" && category != "Semua" {
		query = query.Where("category = ?", category)
	}
	err := query.Find(&expenses).Error
	return expenses, err
}

func (r *expenseRepository) FindByID(id uint) (*entity.Expense, error) {
	var expense entity.Expense
	if err := r.db.First(&expense, id).Error; err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Expense{}, id).Error
}

func (r *expenseRepository) GetTotalExpense() (int64, error) {
	var total int64
	err := r.db.Model(&entity.Expense{}).Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	return total, err
}

