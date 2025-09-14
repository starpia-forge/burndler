package services

import (
	"strings"
	"testing"
)

// Test NewLinter constructor
func TestNewLinter(t *testing.T) {
	linter := NewLinter()
	if linter == nil {
		t.Error("Expected NewLinter to return non-nil linter")
	}
}

// Test valid compose validation
func TestLinter_Lint_ValidCompose(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
    environment:
      - ENV_VAR=value`,
		StrictMode: false,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid compose to pass validation")
	}

	if len(result.Errors) > 0 {
		t.Errorf("Expected no errors, got %d", len(result.Errors))
	}
}

// Test compose with build directive (should fail)
func TestLinter_Lint_BuildDirective(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    build: .
    image: myapp:latest`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected compose with build directive to fail validation")
	}

	// Check for specific error about build directive
	foundBuildError := false
	for _, err := range result.Errors {
		if strings.Contains(err.Message, "build") {
			foundBuildError = true
			break
		}
	}

	if !foundBuildError {
		t.Error("Expected error about build directive")
	}
}

// Test compose with unresolved variables
func TestLinter_Lint_UnresolvedVariables(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx:latest
    environment:
      - UNRESOLVED=${UNDEFINED_VAR}`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	// Unresolved variables should generate warnings
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for unresolved variables")
	}

	// Check for specific warning about unresolved variable
	foundVariableWarning := false
	for _, warn := range result.Warnings {
		if strings.Contains(warn.Message, "UNDEFINED_VAR") {
			foundVariableWarning = true
			break
		}
	}

	if !foundVariableWarning {
		t.Error("Expected warning about UNDEFINED_VAR")
	}
}