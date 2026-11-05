package entity

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	Type      string         `gorm:"type:varchar(50);default:'info'" json:"type"` // e.g. "welcome", "dues_reminder", "payment_success", "info"
	Title     string         `gorm:"type:varchar(255);not null" json:"title"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	Date      string         `gorm:"type:varchar(50)" json:"date"`
	Priority  string         `gorm:"type:varchar(20);default:'normal'" json:"priority"` // "normal", "high"
	IsRead    bool           `gorm:"default:false" json:"is_read"`
	DuesID    uint           `json:"dues_id,omitempty"`
	Amount    int64          `json:"amount,omitempty"`
	Period    string         `gorm:"type:varchar(20)" json:"period,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

