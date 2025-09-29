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

// Test checkDependsOn validation with array syntax
func TestLinter_Lint_DependsOnArray(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx:latest
    depends_on:
      - db
      - cache
  db:
    image: postgres:13
  cache:
    image: redis:alpine`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid depends_on array to pass validation")
	}
}

// Test checkDependsOn validation with invalid dependency
func TestLinter_Lint_DependsOnInvalid(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx:latest
    depends_on:
      - nonexistent_service
  db:
    image: postgres:13`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid depends_on to fail validation")
	}

	// Check for specific error about invalid dependency
	foundDependsError := false
	for _, err := range result.Errors {
		if strings.Contains(err.Message, "nonexistent_service") {
			foundDependsError = true
			break
		}
	}

	if !foundDependsError {
		t.Error("Expected error about nonexistent_service dependency")
	}
}

// Test checkDependsOn validation with object syntax
func TestLinter_Lint_DependsOnObject(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx:latest
    depends_on:
      db:
        condition: service_healthy
      cache:
        condition: service_started
  db:
    image: postgres:13
  cache:
    image: redis:alpine`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid depends_on object to pass validation")
	}
}

// Test checkNetworkReferences validation
func TestLinter_Lint_NetworkReferences(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx:latest
    networks:
      - frontend
      - backend
networks:
  frontend:
  backend:`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid network references to pass validation")
	}
}

// Test checkNetworkReferences with invalid network
func TestLinter_Lint_NetworkReferencesInvalid(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx:latest
    networks:
      - undefined_network
networks:
  frontend:`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid network reference to fail validation")
	}

	// Check for specific error about invalid network
	foundNetworkError := false
	for _, err := range result.Errors {
		if strings.Contains(err.Message, "undefined_network") {
			foundNetworkError = true
			break
		}
	}

	if !foundNetworkError {
		t.Error("Expected error about undefined_network")
	}
}

// Test checkVolumeReferences validation
func TestLinter_Lint_VolumeReferences(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx:latest
    volumes:
      - data_volume:/app/data
      - logs_volume:/app/logs
volumes:
  data_volume:
  logs_volume:`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid volume references to pass validation")
	}
}

// Test checkVolumeReferences with invalid volume
func TestLinter_Lint_VolumeReferencesInvalid(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx:latest
    volumes:
      - undefined_volume:/app/data
volumes:
  data_volume:`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid volume reference to fail validation")
	}

	// Check for specific error about invalid volume
	foundVolumeError := false
	for _, err := range result.Errors {
		if strings.Contains(err.Message, "undefined_volume") {
			foundVolumeError = true
			break
		}
	}

	if !foundVolumeError {
		t.Error("Expected error about undefined_volume")
	}
}

// Test checkSecuritySettings validation
func TestLinter_Lint_SecuritySettings(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx:latest
    privileged: true
    cap_add:
      - SYS_ADMIN`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	// Security issues should generate warnings
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for security settings")
	}

	// Check for specific warnings about privileged and cap_add
	foundPrivilegedWarning := false
	foundCapAddWarning := false
	for _, warn := range result.Warnings {
		if strings.Contains(warn.Message, "privileged") {
			foundPrivilegedWarning = true
		}
		if strings.Contains(warn.Message, "capabilities") {
			foundCapAddWarning = true
		}
	}

	if !foundPrivilegedWarning {
		t.Error("Expected warning about privileged mode")
	}
	if !foundCapAddWarning {
		t.Error("Expected warning about cap_add")
	}
}

// Test checkImageFormat validation
func TestLinter_Lint_ImageFormat(t *testing.T) {
	linter := NewLinter()

	req := &LintRequest{
		Compose: `version: '3'
services:
  web:
    image: nginx@sha256:abcdef123456789
  api:
    image: node:latest`,
		StrictMode: true,
	}

	result, err := linter.Lint(req)
	if err != nil {
		t.Fatalf("Lint failed: %v", err)
	}

	// Should have warnings for latest tag
	foundLatestWarning := false
	for _, warn := range result.Warnings {
		if strings.Contains(warn.Message, "latest") {
			foundLatestWarning = true
			break
		}
	}

	if !foundLatestWarning {
		t.Error("Expected warning about latest tag")
	}
}
