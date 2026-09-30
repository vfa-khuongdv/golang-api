package models

import "time"

type Setting struct {
	ID        uint      `gorm:"column:id;primaryKey" json:"id"`
	Key       string    `gorm:"column:key;type:varchar(100);not null;unique" json:"key"`
	Value     string    `gorm:"column:value;type:text;not null" json:"value"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Setting) TableName() string {
	return "settings"
}
