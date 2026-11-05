package dto

type ExpenseRequest struct {
	Title       string `json:"title" validate:"required"`
	Category    string `json:"category" validate:"required"`
	Amount      int    `json:"amount" validate:"required"`
	Vendor      string `json:"vendor" validate:"required"`
	Date        string `json:"date" validate:"required"`
	Description string `json:"description"`
	ReceiptRef  string `json:"receipt_ref"`
	ApprovedBy  string `json:"approved_by"`
}

