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

// Test IsAdmin method
func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"Admin role", "Admin", true},
		{"Developer role", "Developer", false},
		{"Engineer role", "Engineer", false},
		{"Empty role", "", false},
		{"Invalid role", "Manager", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			if got := user.IsAdmin(); got != tt.expected {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.expected)
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

// Test password hashing
func TestUser_SetPassword(t *testing.T) {
	user := &User{}
	password := "testPassword123!"

	err := user.SetPassword(password)
	if err != nil {
		t.Errorf("SetPassword() error = %v", err)
	}

	if user.Password == "" {
		t.Error("Password should be hashed and set")
	}

	if user.Password == password {
		t.Error("Password should not be stored in plain text")
	}
}

// Test password validation
func TestUser_CheckPassword(t *testing.T) {
	user := &User{}
	password := "testPassword123!"
	wrongPassword := "wrongPassword"

	// Set password first
	err := user.SetPassword(password)
	if err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	// Test correct password
	if !user.CheckPassword(password) {
		t.Error("CheckPassword() should return true for correct password")
	}

	// Test wrong password
	if user.CheckPassword(wrongPassword) {
		t.Error("CheckPassword() should return false for wrong password")
	}

	// Test empty password
	if user.CheckPassword("") {
		t.Error("CheckPassword() should return false for empty password")
	}
}

// Test password validation edge cases
func TestUser_CheckPassword_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		setupPassword string
		checkPassword string
		shouldPass    bool
	}{
		{"correct password", "myPassword123!", "myPassword123!", true},
		{"wrong password", "myPassword123!", "wrongPassword", false},
		{"empty check password", "myPassword123!", "", false},
		{"case sensitive", "myPassword123!", "MyPassword123!", false},
		{"unicode password", "패스워드123!", "패스워드123!", true},
		{"special characters", "!@#$%^&*()", "!@#$%^&*()", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{}
			err := user.SetPassword(tt.setupPassword)
			if err != nil {
				t.Fatalf("SetPassword() error = %v", err)
			}

			result := user.CheckPassword(tt.checkPassword)
			if result != tt.shouldPass {
				t.Errorf("CheckPassword() = %v, want %v", result, tt.shouldPass)
			}
		})
	}
}
