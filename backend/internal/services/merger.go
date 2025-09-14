package services

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Merger implements compose file merging with namespace prefixing
type Merger struct{}

// NewMerger creates a new merger service
func NewMerger() *Merger {
	return &Merger{}
}

// MergeRequest represents a merge request
type MergeRequest struct {
	Modules         []Module          `json:"modules"`
	ProjectVariables map[string]string `json:"project_variables"`
}

// Module represents a compose module to merge
type Module struct {
	Name      string            `json:"name"`
	Compose   string            `json:"compose"`
	Variables map[string]string `json:"variables"`
}

// MergeResult contains the merged compose and mappings
type MergeResult struct {
	MergedCompose string            `json:"merged_compose"`
	Mappings      map[string]string `json:"mappings"`
	Warnings      []string          `json:"warnings"`
}

// Merge combines multiple compose files with namespace prefixing
func (m *Merger) Merge(req *MergeRequest) (*MergeResult, error) {
	result := &MergeResult{
		Mappings: make(map[string]string),
		Warnings: []string{},
	}

	mergedServices := make(map[string]interface{})
	mergedNetworks := make(map[string]interface{})
	mergedVolumes := make(map[string]interface{})

	for _, module := range req.Modules {
		// Parse module compose
		var compose map[string]interface{}
		if err := yaml.Unmarshal([]byte(module.Compose), &compose); err != nil {
			return nil, fmt.Errorf("failed to parse compose for module %s: %w", module.Name, err)
		}

		// Process services
		if services, ok := compose["services"].(map[string]interface{}); ok {
			for serviceName, serviceConfig := range services {
				// Prefix service name with namespace
				newName := fmt.Sprintf("%s__%s", module.Name, serviceName)
				result.Mappings[serviceName] = newName

				// Update depends_on references
				if config, ok := serviceConfig.(map[string]interface{}); ok {
					m.updateDependsOn(config, module.Name, result.Mappings)
					m.substituteVariables(config, module.Variables, req.ProjectVariables)
				}

				mergedServices[newName] = serviceConfig
			}
		}

		// Process networks
		if networks, ok := compose["networks"].(map[string]interface{}); ok {
			for networkName, networkConfig := range networks {
				newName := fmt.Sprintf("%s__%s", module.Name, networkName)
				result.Mappings[networkName] = newName
				mergedNetworks[newName] = networkConfig
			}
		}

		// Process volumes
		if volumes, ok := compose["volumes"].(map[string]interface{}); ok {
			for volumeName, volumeConfig := range volumes {
				newName := fmt.Sprintf("%s__%s", module.Name, volumeName)
				result.Mappings[volumeName] = newName
				mergedVolumes[newName] = volumeConfig
			}
		}
	}

	// Check for port collisions
	m.checkPortCollisions(mergedServices, result)

	// Build final compose
	finalCompose := map[string]interface{}{
		"version": "3.9",
	}

	if len(mergedServices) > 0 {
		finalCompose["services"] = mergedServices
	}
	if len(mergedNetworks) > 0 {
		finalCompose["networks"] = mergedNetworks
	}
	if len(mergedVolumes) > 0 {
		finalCompose["volumes"] = mergedVolumes
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(finalCompose)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merged compose: %w", err)
	}

	result.MergedCompose = string(yamlBytes)
	return result, nil
}

// updateDependsOn updates depends_on references with namespace prefix
func (m *Merger) updateDependsOn(service map[string]interface{}, namespace string, mappings map[string]string) {
	if dependsOn, ok := service["depends_on"]; ok {
		switch deps := dependsOn.(type) {
		case []interface{}:
			// Simple array format
			newDeps := []interface{}{}
			for _, dep := range deps {
				if depName, ok := dep.(string); ok {
					newName := fmt.Sprintf("%s__%s", namespace, depName)
					newDeps = append(newDeps, newName)
				}
			}
			service["depends_on"] = newDeps

		case map[string]interface{}:
			// Extended format with conditions
			newDeps := make(map[string]interface{})
			for depName, depConfig := range deps {
				newName := fmt.Sprintf("%s__%s", namespace, depName)
				newDeps[newName] = depConfig
			}
			service["depends_on"] = newDeps
		}
	}
}

// substituteVariables replaces variables with project overrides > module defaults
func (m *Merger) substituteVariables(config map[string]interface{}, moduleVars, projectVars map[string]string) {
	for key, value := range config {
		switch v := value.(type) {
		case string:
			// Check for variable substitution
			if strings.Contains(v, "${") {
				config[key] = m.replaceVariables(v, moduleVars, projectVars)
			}
		case map[string]interface{}:
			// Recurse into nested maps
			m.substituteVariables(v, moduleVars, projectVars)
		case []interface{}:
			// Process arrays
			for i, item := range v {
				if str, ok := item.(string); ok && strings.Contains(str, "${") {
					v[i] = m.replaceVariables(str, moduleVars, projectVars)
				}
			}
		}
	}
}

// replaceVariables replaces ${VAR} with actual values
func (m *Merger) replaceVariables(str string, moduleVars, projectVars map[string]string) string {
	result := str

	// Find all variables
	for strings.Contains(result, "${") {
		start := strings.Index(result, "${")
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		varName := result[start+2 : end]

		// Project variables override module variables
		if val, ok := projectVars[varName]; ok {
			result = strings.Replace(result, result[start:end+1], val, 1)
		} else if val, ok := moduleVars[varName]; ok {
			result = strings.Replace(result, result[start:end+1], val, 1)
		}
		// Leave unresolved variables as-is for env substitution
	}

	return result
}

// checkPortCollisions detects host port conflicts
func (m *Merger) checkPortCollisions(services map[string]interface{}, result *MergeResult) {
	usedPorts := make(map[string]string) // port -> service name

	for serviceName, serviceConfig := range services {
		if config, ok := serviceConfig.(map[string]interface{}); ok {
			if ports, ok := config["ports"].([]interface{}); ok {
				for _, port := range ports {
					if portStr, ok := port.(string); ok {
						// Extract host port
						parts := strings.Split(portStr, ":")
						if len(parts) >= 2 {
							hostPort := parts[0]
							if existingService, exists := usedPorts[hostPort]; exists {
								result.Warnings = append(result.Warnings,
									fmt.Sprintf("Port collision: %s used by both %s and %s",
										hostPort, existingService, serviceName))
							} else {
								usedPorts[hostPort] = serviceName
							}
						}
					}
				}
			}
		}
	}
}