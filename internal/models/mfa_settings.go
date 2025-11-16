package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BackupCodesList represents a list of backup codes
type BackupCodesList []string

// Value implements the driver.Valuer interface for storing BackupCodesList as JSON
func (b BackupCodesList) Value() (driver.Value, error) {
	return json.Marshal(b)
}

// Scan implements the sql.Scanner interface for scanning JSON into BackupCodesList
func (b *BackupCodesList) Scan(value interface{}) error {
	if value == nil {
		*b = BackupCodesList{}
		return nil
	}
	return json.Unmarshal(value.([]byte), &b)
}

type MfaSettings struct {
	ID          uint           `gorm:"column:id;primaryKey" json:"id"`
	UserID      uint           `gorm:"column:user_id;uniqueIndex;not null" json:"user_id"`
	MfaEnabled  bool           `gorm:"column:mfa_enabled;default:false" json:"mfa_enabled"`
	TotpSecret  *string        `gorm:"column:totp_secret;type:varchar(255)" json:"totp_secret,omitempty"`
	BackupCodes datatypes.JSON `gorm:"column:backup_codes;type:json" json:"backup_codes,omitempty"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for MfaSettings model
func (MfaSettings) TableName() string {
	return "mfa_settings"
}
