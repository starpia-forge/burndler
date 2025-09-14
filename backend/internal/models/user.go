package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a system user with RBAC role
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Name      string         `json:"name"`
	Role      string         `gorm:"not null;default:'Engineer'" json:"role"` // Developer or Engineer
	Active    bool           `gorm:"default:true" json:"active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// IsDeveloper checks if user has Developer role
func (u *User) IsDeveloper() bool {
	return u.Role == "Developer"
}

// IsEngineer checks if user has Engineer role
func (u *User) IsEngineer() bool {
	return u.Role == "Engineer"
}