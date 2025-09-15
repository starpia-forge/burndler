package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a system user with RBAC role
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Name      string         `json:"name"`
	Password  string         `gorm:"not null" json:"-"`                        // Bcrypt hashed password, excluded from JSON
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

// SetPassword hashes a plain text password and stores it
func (u *User) SetPassword(password string) error {
	// Use cost factor 12 for good security-performance balance
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword compares a plain text password with the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
