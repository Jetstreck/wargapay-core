package entity

import "time"

type Dues struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	User        User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Period      string     `gorm:"type:varchar(7);not null;uniqueIndex:idx_user_period" json:"period"`
	WasteFee    int64      `gorm:"not null" json:"waste_fee"`
	SecurityFee int64      `gorm:"not null" json:"security_fee"`
	RTFund      int64      `gorm:"not null" json:"rt_fund"`
	Total       int64      `gorm:"not null" json:"total"`
	Status      string     `gorm:"type:varchar(20);default:'unpaid';not null" json:"status"`
	PaidAt      *time.Time `json:"paid_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
