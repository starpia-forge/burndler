package models

import (
	"testing"
	"time"
)

func TestSetup_TableName(t *testing.T) {
	setup := Setup{}
	expected := "setups"
	if got := setup.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

func TestSetup_MarkCompleted(t *testing.T) {
	setup := &Setup{
		IsCompleted: false,
	}

	// Before marking completed
	if setup.IsCompleted {
		t.Error("Setup should not be completed initially")
	}
	if setup.CompletedAt != nil {
		t.Error("CompletedAt should be nil initially")
	}

	// Mark as completed
	setup.MarkCompleted()

	// After marking completed
	if !setup.IsCompleted {
		t.Error("Setup should be completed after MarkCompleted()")
	}
	if setup.CompletedAt == nil {
		t.Error("CompletedAt should not be nil after MarkCompleted()")
	}

	// Check that CompletedAt is recent (within last minute)
	if time.Since(*setup.CompletedAt) > time.Minute {
		t.Error("CompletedAt should be recent")
	}
}

func TestSetup_IsSetupCompleted(t *testing.T) {
	tests := []struct {
		name        string
		isCompleted bool
		expected    bool
	}{
		{
			name:        "completed setup",
			isCompleted: true,
			expected:    true,
		},
		{
			name:        "incomplete setup",
			isCompleted: false,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup := &Setup{
				IsCompleted: tt.isCompleted,
			}
			if got := setup.IsSetupCompleted(); got != tt.expected {
				t.Errorf("IsSetupCompleted() = %v, want %v", got, tt.expected)
			}
		})
	}
}
