package models

import (
	"testing"
)

// Test IsDeveloper method
func TestUser_IsDeveloper(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"Developer role", "Developer", true},
		{"Engineer role", "Engineer", false},
		{"Admin role", "Admin", false},
		{"Empty role", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			if got := user.IsDeveloper(); got != tt.expected {
				t.Errorf("IsDeveloper() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test IsEngineer method
func TestUser_IsEngineer(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"Engineer role", "Engineer", true},
		{"Developer role", "Developer", false},
		{"Admin role", "Admin", false},
		{"Empty role", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			if got := user.IsEngineer(); got != tt.expected {
				t.Errorf("IsEngineer() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test TableName method
func TestUser_TableName(t *testing.T) {
	user := User{}
	expected := "users"
	if got := user.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}