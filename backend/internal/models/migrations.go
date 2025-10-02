package models

import (
	"fmt"

	"gorm.io/gorm"
)

// MigrateContainerConfigurationToContainerLevel migrates ContainerConfiguration
// from version-level to container-level structure.
//
// This migration is needed when upgrading from the old structure where:
// - ContainerConfiguration belonged to ContainerVersion (1:1)
// - Files and Assets belonged to ContainerVersion
//
// To the new structure where:
// - ContainerConfiguration belongs to Container (N:1) with Name and MinimumVersion
// - ContainerVersion optionally references a Configuration
// - Files and Assets belong to ContainerConfiguration
//
// Migration steps:
// 1. Check if migration is needed (look for data in old structure)
// 2. For each old ContainerConfiguration:
//    - Get ContainerID from its ContainerVersion
//    - Set Name = "default"
//    - Set MinimumVersion = version's Version value
//    - Update ContainerVersion to reference this Configuration
// 3. Update Files and Assets to reference ContainerConfiguration instead of Version
func MigrateContainerConfigurationToContainerLevel(db *gorm.DB) error {
	// Check if old structure exists by looking for configurations
	// that don't have Name or MinimumVersion (old structure indicators)
	var needsMigration bool
	err := db.Raw(`
		SELECT EXISTS(
			SELECT 1 FROM container_configurations
			WHERE name IS NULL OR name = '' OR minimum_version IS NULL OR minimum_version = ''
		)
	`).Scan(&needsMigration).Error

	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !needsMigration {
		// No migration needed - either already migrated or clean database
		return nil
	}

	// Start transaction for data migration
	return db.Transaction(func(tx *gorm.DB) error {
		// Step 1: Get all configurations that need migration
		type OldConfig struct {
			ID                 uint
			ContainerVersionID uint
		}

		var oldConfigs []OldConfig
		if err := tx.Raw(`
			SELECT id, container_version_id
			FROM container_configurations
			WHERE (name IS NULL OR name = '') OR (minimum_version IS NULL OR minimum_version = '')
		`).Scan(&oldConfigs).Error; err != nil {
			return fmt.Errorf("failed to fetch old configurations: %w", err)
		}

		// Step 2: Migrate each configuration
		for _, oldConfig := range oldConfigs {
			// Get version information
			var version struct {
				ContainerID uint
				Version     string
			}
			if err := tx.Raw(`
				SELECT container_id, version
				FROM container_versions
				WHERE id = ?
			`, oldConfig.ContainerVersionID).Scan(&version).Error; err != nil {
				return fmt.Errorf("failed to fetch version for config %d: %w", oldConfig.ID, err)
			}

			// Update configuration with new structure
			if err := tx.Exec(`
				UPDATE container_configurations
				SET container_id = ?, name = 'default', minimum_version = ?
				WHERE id = ?
			`, version.ContainerID, version.Version, oldConfig.ID).Error; err != nil {
				return fmt.Errorf("failed to update config %d: %w", oldConfig.ID, err)
			}

			// Update version to reference this configuration
			if err := tx.Exec(`
				UPDATE container_versions
				SET configuration_id = ?
				WHERE id = ?
			`, oldConfig.ID, oldConfig.ContainerVersionID).Error; err != nil {
				return fmt.Errorf("failed to update version %d: %w", oldConfig.ContainerVersionID, err)
			}

			// Migrate files
			if err := tx.Exec(`
				UPDATE container_files
				SET container_configuration_id = ?
				WHERE container_version_id = ?
			`, oldConfig.ID, oldConfig.ContainerVersionID).Error; err != nil {
				return fmt.Errorf("failed to migrate files for config %d: %w", oldConfig.ID, err)
			}

			// Migrate assets
			if err := tx.Exec(`
				UPDATE container_assets
				SET container_configuration_id = ?
				WHERE container_version_id = ?
			`, oldConfig.ID, oldConfig.ContainerVersionID).Error; err != nil {
				return fmt.Errorf("failed to migrate assets for config %d: %w", oldConfig.ID, err)
			}
		}

		return nil
	})
}

// RollbackContainerConfigurationMigration rolls back the migration
// (mainly useful for testing)
func RollbackContainerConfigurationMigration(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// This is a simplified rollback - in production you'd need more careful handling
		// Reset Name and MinimumVersion to empty to indicate old structure
		if err := tx.Exec(`
			UPDATE container_configurations
			SET name = '', minimum_version = ''
			WHERE name = 'default'
		`).Error; err != nil {
			return fmt.Errorf("failed to rollback configurations: %w", err)
		}

		// Clear ConfigurationID from versions
		if err := tx.Exec(`
			UPDATE container_versions
			SET configuration_id = NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to rollback versions: %w", err)
		}

		return nil
	})
}
