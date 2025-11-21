package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	ID           uint           `gorm:"column:id;primaryKey" json:"id"`
	RefreshToken string         `gorm:"column:refresh_token;type:varchar(60);not null;unique" json:"refresh_token"`
	IpAddress    string         `gorm:"column:ip_address;type:varchar(45);not null" json:"ip_address"`
	UsedCount    int64          `gorm:"column:used_count;default:0" json:"used_count"`
	ExpiredAt    int64          `gorm:"column:expired_at;not null" json:"expired_at"`
	UserID       uint           `gorm:"column:user_id;not null" json:"user_id"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`

	// Relations
	User User `gorm:"constraint:OnDelete:CASCADE;foreignKey:UserID" json:"user"`
}

// TableName specifies the table name for RefreshToken model
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
