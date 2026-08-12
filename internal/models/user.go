package models

import (
	"time"

	"gorm.io/gorm"
)

type Gender int16

const (
	GenderMale   Gender = 1
	GenderFemale Gender = 2
	GenderOther  Gender = 3
)

type User struct {
	ID            uint           `gorm:"column:id;primaryKey" json:"id"`
	Email         string         `gorm:"column:email;type:varchar(254);unique;not null" json:"email"`
	Password      string         `gorm:"column:password;type:varchar(255);not null" json:"-"`
	Name          string         `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Birthday      *time.Time     `gorm:"column:birthday;type:date;default:null" json:"birthday,omitempty"`
	Address       *string        `gorm:"column:address;type:varchar(255);default:null" json:"address,omitempty"`
	Gender        Gender         `gorm:"column:gender;type:smallint;not null" json:"gender"`
	FailedAttempts int           `gorm:"column:failed_attempts;default:0" json:"-"`
	LockedUntil   *int64         `gorm:"column:locked_until;default:null" json:"-"`
	ResetToken    *string        `gorm:"column:reset_token;type:varchar(100);default:null;unique" json:"-"`
	ResetExpiredAt *int64        `gorm:"column:reset_expired_at;type:bigint;default:null" json:"-"`
	CreatedAt     time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

func (User) TableName() string {
	return "users"
}
