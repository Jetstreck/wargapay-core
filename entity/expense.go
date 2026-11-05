package entity

import "time"

type Expense struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"type:varchar(150);not null" json:"title"`
	Category    string    `gorm:"type:varchar(50);not null" json:"category"` // Keamanan, Kebersihan, Infrastruktur, Operasional
	Amount      int       `gorm:"not null" json:"amount"`
	Vendor      string    `gorm:"type:varchar(100);not null" json:"vendor"`
	Date        string    `gorm:"type:varchar(30);not null" json:"date"`
	Description string    `gorm:"type:text" json:"description"`
	ReceiptRef  string    `gorm:"type:varchar(50)" json:"receipt_ref"`
	ApprovedBy  string    `gorm:"type:varchar(100)" json:"approved_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

