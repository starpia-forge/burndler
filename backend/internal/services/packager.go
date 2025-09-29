package services

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/burndler/burndler/internal/storage"
	"github.com/google/uuid"
)

// Packager creates offline installer packages
type Packager struct {
	storage storage.Storage
}

// NewPackager creates a new packager service
func NewPackager(storage storage.Storage) *Packager {
	return &Packager{
		storage: storage,
	}
}

// PackageRequest represents a package creation request
type PackageRequest struct {
	Name      string     `json:"name"`
	Compose   string     `json:"compose"`
	Resources []Resource `json:"resources"`
}

// Resource represents a static resource to include
type Resource struct {
	Module  string   `json:"module"`
	Version string   `json:"version"`
	Files   []string `json:"files"`
}

// PackageManifest represents the manifest.json content
type PackageManifest struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	CreatedAt time.Time         `json:"created_at"`
	Images    []ImageInfo       `json:"images"`
	Resources []ResourceInfo    `json:"resources"`
	Checksums map[string]string `json:"checksums"`
}

// ImageInfo represents Docker image metadata
type ImageInfo struct {
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
	File   string `json:"file"`
}

// ResourceInfo represents static resource metadata
type ResourceInfo struct {
	Module  string   `json:"module"`
	Version string   `json:"version"`
	Files   []string `json:"files"`
}

// CreatePackage builds an offline installer package
func (p *Packager) CreatePackage(ctx context.Context, req *PackageRequest) (string, error) {
	buildID := uuid.New().String()
	packageName := fmt.Sprintf("%s-%s.tar.gz", req.Name, buildID)

	// Create manifest
	manifest := PackageManifest{
		Name:      req.Name,
		Version:   "1.0.0",
		CreatedAt: time.Now(),
		Images:    []ImageInfo{},
		Resources: []ResourceInfo{},
		Checksums: make(map[string]string),
	}

	// Create tar.gz buffer
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	// Add compose file
	if err := p.addFileToTar(tarWriter, "compose/docker-compose.yaml", []byte(req.Compose)); err != nil {
		return "", fmt.Errorf("failed to add compose file: %w", err)
	}

	// Add .env.example
	envExample := p.generateEnvExample()
	if err := p.addFileToTar(tarWriter, "env/.env.example", []byte(envExample)); err != nil {
		return "", fmt.Errorf("failed to add .env.example: %w", err)
	}

	// Add install.sh
	installScript := p.generateInstallScript()
	if err := p.addFileToTar(tarWriter, "bin/install.sh", []byte(installScript)); err != nil {
		return "", fmt.Errorf("failed to add install.sh: %w", err)
	}

	// Add verify.sh
	verifyScript := p.generateVerifyScript()
	if err := p.addFileToTar(tarWriter, "bin/verify.sh", []byte(verifyScript)); err != nil {
		return "", fmt.Errorf("failed to add verify.sh: %w", err)
	}

	// Add resources
	for _, resource := range req.Resources {
		manifest.Resources = append(manifest.Resources, ResourceInfo(resource))
	}

	// Add manifest.json
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal manifest: %w", err)
	}
	if err := p.addFileToTar(tarWriter, "manifest.json", manifestJSON); err != nil {
		return "", fmt.Errorf("failed to add manifest: %w", err)
	}

	// Close tar and gzip writers
	if err := tarWriter.Close(); err != nil {
		return "", fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return "", fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// Upload to storage
	reader := bytes.NewReader(buf.Bytes())
	url, err := p.storage.Upload(ctx, packageName, reader, int64(buf.Len()))
	if err != nil {
		return "", fmt.Errorf("failed to upload package: %w", err)
	}

	return url, nil
}

// addFileToTar adds a file to the tar archive
func (p *Packager) addFileToTar(tw *tar.Writer, name string, content []byte) error {
	header := &tar.Header{
		Name:    name,
		Mode:    0644,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}

	// Make scripts executable
	if name == "bin/install.sh" || name == "bin/verify.sh" {
		header.Mode = 0755
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := tw.Write(content); err != nil {
		return err
	}

	return nil
}

// generateEnvExample creates a template .env file
func (p *Packager) generateEnvExample() string {
	return `# Burndler Environment Configuration
# Copy to .env and update values

DB_HOST=localhost
DB_PORT=5432
DB_NAME=burndler
DB_USER=burndler
DB_PASSWORD=changeme

STORAGE_MODE=local
LOCAL_STORAGE_PATH=/var/lib/burndler/storage

JWT_SECRET=changeme-generate-secure-secret
SERVER_PORT=8080
`
}

// generateInstallScript creates the installation script
func (p *Packager) generateInstallScript() string {
	return `#!/bin/bash
set -e

echo "Burndler Offline Installer"
echo "=========================="

# Check prerequisites
echo "Checking prerequisites..."

if ! command -v docker &> /dev/null; then
    echo "ERROR: Docker is not installed"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "ERROR: Docker Compose is not installed"
    exit 1
fi

# Load images
echo "Loading Docker images..."
for image in images/*.tar; do
    if [ -f "$image" ]; then
        echo "Loading $image..."
        docker load < "$image"
    fi
done

# Copy resources
if [ -d "resources" ]; then
    echo "Copying resources..."
    mkdir -p /var/lib/burndler/resources
    cp -r resources/* /var/lib/burndler/resources/
fi

# Setup environment
if [ ! -f ".env" ]; then
    echo "Creating .env from template..."
    cp env/.env.example .env
    echo "Please edit .env with your configuration"
fi

# Start services
echo "Starting services..."
cd compose
docker-compose up -d

# Wait for health checks
echo "Waiting for services to be healthy..."
sleep 10

echo "Installation complete!"
echo "Access the application at http://localhost:8080"
`
}

// generateVerifyScript creates the verification script
func (p *Packager) generateVerifyScript() string {
	return `#!/bin/bash
set -e

echo "Burndler Installation Verification"
echo "=================================="

# Check Docker version
echo "Docker version:"
docker --version

# Check Docker Compose version
echo "Docker Compose version:"
docker-compose --version

# Check disk space
echo "Available disk space:"
df -h /var/lib/docker

# Verify manifest
if [ -f "manifest.json" ]; then
    echo "Package manifest:"
    cat manifest.json | head -20
fi

# Check for required files
echo "Checking required files..."
required_files=(
    "compose/docker-compose.yaml"
    "env/.env.example"
    "bin/install.sh"
    "manifest.json"
)

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "✓ $file exists"
    else
        echo "✗ $file missing"
    fi
done

echo "Verification complete!"
`
}
