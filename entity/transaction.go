package entity

import "time"

type Transaction struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	UserID          uint          `gorm:"not null;index" json:"user_id"`
	User            User          `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"user,omitempty"`
	OrderID         string        `gorm:"type:varchar(100);uniqueIndex;not null" json:"order_id"`
	Type            string        `gorm:"type:varchar(50);not null" json:"type"` // "dues" atau "waste_service"
	DuesID          *uint         `gorm:"index" json:"dues_id,omitempty"`
	Dues            *Dues         `gorm:"foreignKey:DuesID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"dues,omitempty"`
	WasteServiceID  *uint         `gorm:"index" json:"waste_service_id,omitempty"`
	WasteService    *WasteService `gorm:"foreignKey:WasteServiceID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"waste_service,omitempty"`
	Amount          int64         `gorm:"not null" json:"amount"`
	Status          string        `gorm:"type:varchar(30);default:'pending';not null" json:"status"` // pending, paid, failed, expired, cancelled
	PaymentType     string        `gorm:"type:varchar(50)" json:"payment_type"`
	SnapToken       string        `gorm:"type:varchar(255)" json:"snap_token,omitempty"`
	SnapRedirectURL string        `gorm:"type:text" json:"snap_redirect_url,omitempty"`
	Notes           string        `gorm:"type:text" json:"notes,omitempty"`
	PickupAddress   string        `gorm:"type:text" json:"pickup_address,omitempty"`
	PickupDate      *time.Time    `json:"pickup_date,omitempty"`
	PaidAt          *time.Time    `json:"paid_at,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}
