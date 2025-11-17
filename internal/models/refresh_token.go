package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	ID           uint           `gorm:"column:id;primaryKey" json:"id"`
	RefreshToken string         `gorm:"column:refresh_token;type:varchar(60);not null;unique" json:"refreshToken"`
	IpAddress    string         `gorm:"column:ip_address;type:varchar(45);not null" json:"ipAddress"`
	UsedCount    int64          `gorm:"column:used_count;default:0" json:"usedCount"`
	ExpiredAt    int64          `gorm:"column:expired_at;not null" json:"expiredAt"`
	UserID       uint           `gorm:"column:user_id;not null" json:"user_id"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`

	// Relations
	User User `gorm:"constraint:OnDelete:CASCADE;foreignKey:UserID" json:"user"`
}

// TableName specifies the table name for RefreshToken model
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
