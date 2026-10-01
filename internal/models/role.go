package models

import "time"

// RoleAdmin is the built-in role that holds every permission. It cannot be
// renamed or deleted, otherwise the system could lock itself out.
const RoleAdmin = "admin"

type Role struct {
	ID          uint         `gorm:"column:id;primaryKey" json:"id"`
	Name        string       `gorm:"column:name;type:varchar(100);not null;unique" json:"name"`
	Description string       `gorm:"column:description;type:varchar(255);not null;default:''" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	CreatedAt   time.Time    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"column:updated_at" json:"updated_at"`
}

func (Role) TableName() string {
	return "roles"
}

type Permission struct {
	ID          uint      `gorm:"column:id;primaryKey" json:"id"`
	Name        string    `gorm:"column:name;type:varchar(100);not null;unique" json:"name"`
	Description string    `gorm:"column:description;type:varchar(255);not null;default:''" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

// UserRole is the join row between users and roles.
type UserRole struct {
	UserID uint `gorm:"column:user_id;primaryKey"`
	RoleID uint `gorm:"column:role_id;primaryKey"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
