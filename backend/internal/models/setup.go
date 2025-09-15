package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Setup represents the system setup state
type Setup struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	IsCompleted  bool           `gorm:"not null;default:false" json:"is_completed"`
	CompletedAt  *time.Time     `json:"completed_at"`
	AdminEmail   string         `gorm:"type:varchar(255)" json:"admin_email"`
	CompanyName  string         `gorm:"type:varchar(255)" json:"company_name"`
	SystemConfig datatypes.JSON `gorm:"type:jsonb" json:"system_config"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Setup model
func (Setup) TableName() string {
	return "setups"
}

// MarkCompleted marks the setup as completed
func (s *Setup) MarkCompleted() {
	s.IsCompleted = true
	now := time.Now()
	s.CompletedAt = &now
}

// IsSetupCompleted checks if the setup process is completed
func (s *Setup) IsSetupCompleted() bool {
	return s.IsCompleted
}
