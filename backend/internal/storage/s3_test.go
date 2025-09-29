package storage

import (
	"testing"

	"github.com/burndler/burndler/internal/config"
)

// Test NewS3Storage constructor (basic validation)
func TestNewS3Storage(t *testing.T) {
	cfg := &config.Config{
		S3Endpoint:        "https://s3.amazonaws.com",
		S3Region:          "us-east-1",
		S3Bucket:          "test-bucket",
		S3AccessKeyID:     "test-key",
		S3SecretAccessKey: "test-secret",
		S3UseSSL:          true,
		S3PathPrefix:      "packages/",
	}

	s3, err := NewS3Storage(cfg)
	if err != nil {
		t.Fatalf("NewS3Storage failed: %v", err)
	}

	if s3 == nil {
		t.Error("Expected non-nil S3Storage")
	}
}

// Test NewS3Storage with missing bucket
func TestNewS3Storage_MissingBucket(t *testing.T) {
	cfg := &config.Config{
		S3Endpoint:        "https://s3.amazonaws.com",
		S3Region:          "us-east-1",
		S3Bucket:          "", // Missing bucket
		S3AccessKeyID:     "test-key",
		S3SecretAccessKey: "test-secret",
		S3UseSSL:          true,
		S3PathPrefix:      "packages/",
	}

	_, err := NewS3Storage(cfg)
	if err == nil {
		t.Error("Expected error when bucket is missing")
	}
}

// Test NewS3Storage with missing credentials
func TestNewS3Storage_MissingCredentials(t *testing.T) {
	cfg := &config.Config{
		S3Endpoint:        "https://s3.amazonaws.com",
		S3Region:          "us-east-1",
		S3Bucket:          "test-bucket",
		S3AccessKeyID:     "", // Missing credentials
		S3SecretAccessKey: "",
		S3UseSSL:          true,
		S3PathPrefix:      "packages/",
	}

	_, err := NewS3Storage(cfg)
	if err == nil {
		t.Error("Expected error when credentials are missing")
	}
}
