package repository

import (
	"wargapay-core/dto"
	"wargapay-core/entity"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(tx *entity.Transaction) error
	FindByID(id uint) (*entity.Transaction, error)
	FindByOrderID(orderID string) (*entity.Transaction, error)
	FindAll(userID uint, txType, status string) ([]entity.Transaction, error)
	Update(tx *entity.Transaction) error
	GetReport(startDate, endDate string) (*dto.TransactionReportResponse, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(tx *entity.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *transactionRepository) FindByID(id uint) (*entity.Transaction, error) {
	var tx entity.Transaction
	err := r.db.Preload("User").Preload("Dues").Preload("WasteService").First(&tx, id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) FindByOrderID(orderID string) (*entity.Transaction, error) {
	var tx entity.Transaction
	err := r.db.Preload("User").Preload("Dues").Preload("WasteService").Where("order_id = ?", orderID).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) FindAll(userID uint, txType, status string) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	query := r.db.Model(&entity.Transaction{}).
		Preload("User").
		Preload("Dues").
		Preload("WasteService")

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if txType != "" {
		query = query.Where("type = ?", txType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) Update(tx *entity.Transaction) error {
	return r.db.Save(tx).Error
}

func (r *transactionRepository) GetReport(startDate, endDate string) (*dto.TransactionReportResponse, error) {
	var transactions []entity.Transaction
	query := r.db.Model(&entity.Transaction{}).
		Preload("User").
		Preload("Dues").
		Preload("WasteService")

	if startDate != "" {
		query = query.Where("created_at >= ?", startDate+" 00:00:00")
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	if err := query.Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}

	report := &dto.TransactionReportResponse{
		Transactions: transactions,
	}

	for _, tx := range transactions {
		report.TotalTransactions++
		switch tx.Status {
		case "paid":
			report.TotalPaid++
			report.TotalRevenue += tx.Amount
			if tx.Type == "dues" {
				report.DuesRevenue += tx.Amount
			} else if tx.Type == "waste_service" {
				report.WasteServiceRevenue += tx.Amount
			}
		case "pending":
			report.TotalPending++
		case "failed", "expired", "cancelled":
			report.TotalFailed++
		}
	}

	return report, nil
}
