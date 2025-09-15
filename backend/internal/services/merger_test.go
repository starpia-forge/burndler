package services

import (
	"strings"
	"testing"
)

// Test NewMerger constructor
func TestNewMerger(t *testing.T) {
	merger := NewMerger()
	if merger == nil {
		t.Error("Expected NewMerger to return non-nil merger")
	}
}

// Test basic merge with single module
func TestMerger_Merge_SingleModule(t *testing.T) {
	merger := NewMerger()

	req := &MergeRequest{
		Modules: []Module{
			{
				Name: "web",
				Compose: `version: '3'
services:
  app:
    image: nginx:latest
    ports:
      - "80:80"`,
			},
		},
	}

	result, err := merger.Merge(req)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.MergedCompose == "" {
		t.Error("Expected non-empty merged compose")
	}

	// Check that the service was prefixed
	if !strings.Contains(result.MergedCompose, "web__app") {
		t.Error("Expected service to be prefixed with namespace")
	}
}

// Test merge with multiple modules
func TestMerger_Merge_MultipleModules(t *testing.T) {
	merger := NewMerger()

	req := &MergeRequest{
		Modules: []Module{
			{
				Name: "frontend",
				Compose: `version: '3'
services:
  web:
    image: nginx:latest`,
			},
			{
				Name: "backend",
				Compose: `version: '3'
services:
  api:
    image: node:14`,
			},
		},
	}

	result, err := merger.Merge(req)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Check both services are prefixed
	if !strings.Contains(result.MergedCompose, "frontend__web") {
		t.Error("Expected frontend service to be prefixed")
	}

	if !strings.Contains(result.MergedCompose, "backend__api") {
		t.Error("Expected backend service to be prefixed")
	}
}
