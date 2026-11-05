package dto

import "wargapay-core/entity"

type CreateTransactionRequest struct {
	Type           string `json:"type"`             // "dues" atau "waste_service"
	DuesID         *uint  `json:"dues_id"`          // Wajib jika type = dues
	WasteServiceID *uint  `json:"waste_service_id"` // Wajib jika type = waste_service
	PickupAddress  string `json:"pickup_address"`   // Alamat penjemputan sampah
	PickupDate     string `json:"pickup_date"`      // Format: YYYY-MM-DD
	Notes          string `json:"notes"`            // Catatan tambahan
}

type MidtransNotificationRequest struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
	TransactionTime   string `json:"transaction_time"`
}

type TransactionReportResponse struct {
	TotalRevenue        int64                `json:"total_revenue"`
	TotalTransactions   int64                `json:"total_transactions"`
	TotalPaid           int64                `json:"total_paid"`
	TotalPending        int64                `json:"total_pending"`
	TotalFailed         int64                `json:"total_failed"`
	DuesRevenue         int64                `json:"dues_revenue"`
	WasteServiceRevenue int64                `json:"waste_service_revenue"`
	Transactions        []entity.Transaction `json:"transactions,omitempty"`
}

