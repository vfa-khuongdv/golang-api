package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"column:id;primaryKey" json:"id"`
	Email     string         `gorm:"column:email;type:varchar(45);unique;not null" json:"email"`
	Password  string         `gorm:"column:password;type:varchar(255);not null" json:"-"`
	Name      string         `gorm:"column:name;type:varchar(45);not null" json:"name"`
	Birthday  *string        `gorm:"column:birthday;type:date;default:null" json:"birthday,omitempty"`
	Address   *string        `gorm:"column:address;type:varchar(255);default:null" json:"address,omitempty"`
	Gender    int16          `gorm:"column:gender;type:smallint;not null" json:"gender"` // 1. Male, 2. Felmale, 3. Other
	Token     *string        `gorm:"column:token;type:varchar(100);default:null;unique" json:"-"`
	ExpiredAt *int64         `gorm:"column:expired_at;type:bigint;default:null" json:"expiredAt,omitempty"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deletedAt,omitempty"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
