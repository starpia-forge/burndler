package models

import (
	"time"

	"gorm.io/datatypes"
)

// ContainerConfiguration represents template system metadata for a container version
type ContainerConfiguration struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	ContainerVersionID uint           `gorm:"not null;uniqueIndex" json:"container_version_id"`
	UISchema           datatypes.JSON `gorm:"type:jsonb" json:"ui_schema"`
	DependencyRules    datatypes.JSON `gorm:"type:jsonb" json:"dependency_rules"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`

	// Relationships
	ContainerVersion ContainerVersion `gorm:"foreignKey:ContainerVersionID" json:"container_version,omitempty"`
	Files            []ContainerFile  `gorm:"foreignKey:ContainerVersionID;references:ContainerVersionID" json:"files,omitempty"`
	Assets           []ContainerAsset `gorm:"foreignKey:ContainerVersionID;references:ContainerVersionID" json:"assets,omitempty"`
}

// TableName specifies the table name for ContainerConfiguration model
func (ContainerConfiguration) TableName() string {
	return "container_configurations"
}

// ContainerFile represents template or static files in a container version
type ContainerFile struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	ContainerVersionID uint      `gorm:"not null;index" json:"container_version_id"`
	FilePath           string    `gorm:"size:512;not null" json:"file_path"`
	FileType           string    `gorm:"size:20;not null" json:"file_type"` // 'template', 'static'
	StoragePath        string    `gorm:"size:512" json:"storage_path"`
	TemplateFormat     string    `gorm:"size:20" json:"template_format"` // 'yaml', 'json', 'env', 'text'
	DisplayCondition   string    `gorm:"type:text" json:"display_condition"`
	IsDirectory        bool      `gorm:"default:false" json:"is_directory"`
	Description        string    `gorm:"type:text" json:"description"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// Relationships
	ContainerVersion ContainerVersion `gorm:"foreignKey:ContainerVersionID" json:"container_version,omitempty"`
}

// TableName specifies the table name for ContainerFile model
func (ContainerFile) TableName() string {
	return "container_files"
}

// ContainerAsset represents asset files in a container version
type ContainerAsset struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	ContainerVersionID uint      `gorm:"not null;index" json:"container_version_id"`
	OriginalFileName   string    `gorm:"size:255;not null" json:"original_file_name"`
	FilePath           string    `gorm:"size:512;not null" json:"file_path"`
	AssetType          string    `gorm:"size:20;not null" json:"asset_type"` // 'config', 'data', 'script', 'binary', 'document'
	MimeType           string    `gorm:"size:100" json:"mime_type"`
	FileSize           int64     `gorm:"not null" json:"file_size"`
	Checksum           string    `gorm:"size:64;not null" json:"checksum"` // SHA256
	Compressed         bool      `gorm:"default:false" json:"compressed"`
	IncludeCondition   string    `gorm:"type:text" json:"include_condition"`
	StorageType        string    `gorm:"size:20;not null" json:"storage_type"` // 'embedded', 'download'
	StoragePath        string    `gorm:"size:512;not null" json:"storage_path"`
	DownloadURL        string    `gorm:"type:text" json:"download_url"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// Relationships
	ContainerVersion ContainerVersion `gorm:"foreignKey:ContainerVersionID" json:"container_version,omitempty"`
}

// TableName specifies the table name for ContainerAsset model
func (ContainerAsset) TableName() string {
	return "container_assets"
}

// ServiceConfiguration represents container configuration values for a service
type ServiceConfiguration struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	ServiceID           uint           `gorm:"not null;uniqueIndex:idx_service_container" json:"service_id"`
	ContainerID         uint           `gorm:"not null;uniqueIndex:idx_service_container" json:"container_id"`
	ConfigurationValues datatypes.JSON `gorm:"type:jsonb" json:"configuration_values"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`

	// Relationships
	Service   Service   `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Container Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
}

// TableName specifies the table name for ServiceConfiguration model
func (ServiceConfiguration) TableName() string {
	return "service_configurations"
}