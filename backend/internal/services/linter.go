package services

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Linter implements compose file linting according to ADR-002
type Linter struct{}

// NewLinter creates a new linter service
func NewLinter() *Linter {
	return &Linter{}
}

// LintRequest represents a lint request
type LintRequest struct {
	Compose    string `json:"compose"`
	StrictMode bool   `json:"strict_mode"`
}

// LintResult contains lint errors and warnings
type LintResult struct {
	Valid    bool         `json:"valid"`
	Errors   []LintIssue  `json:"errors"`
	Warnings []LintIssue  `json:"warnings"`
}

// LintIssue represents a single lint issue
type LintIssue struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

// Lint validates a compose file against policy
func (l *Linter) Lint(req *LintRequest) (*LintResult, error) {
	result := &LintResult{
		Valid:    true,
		Errors:   []LintIssue{},
		Warnings: []LintIssue{},
	}

	// Parse compose
	var compose map[string]interface{}
	if err := yaml.Unmarshal([]byte(req.Compose), &compose); err != nil {
		return nil, fmt.Errorf("failed to parse compose: %w", err)
	}

	// Check for forbidden build directive
	l.checkBuildDirective(compose, result)

	// Check services
	if services, ok := compose["services"].(map[string]interface{}); ok {
		l.checkServices(services, compose, result)
	}

	// Check for unresolved variables
	l.checkUnresolvedVariables(req.Compose, result)

	// Check for port collisions
	if services, ok := compose["services"].(map[string]interface{}); ok {
		l.checkPortCollisions(services, result)
	}

	// Set valid flag based on errors
	result.Valid = len(result.Errors) == 0

	return result, nil
}

// checkBuildDirective checks for forbidden build: directive
func (l *Linter) checkBuildDirective(compose map[string]interface{}, result *LintResult) {
	if services, ok := compose["services"].(map[string]interface{}); ok {
		for serviceName, serviceConfig := range services {
			if config, ok := serviceConfig.(map[string]interface{}); ok {
				if _, hasBuild := config["build"]; hasBuild {
					result.Errors = append(result.Errors, LintIssue{
						Rule:    "no-build-directive",
						Message: fmt.Sprintf("Service '%s' contains forbidden 'build:' directive. Use prebuilt images only.", serviceName),
					})
				}
			}
		}
	}
}

// checkServices validates service configurations
func (l *Linter) checkServices(services map[string]interface{}, compose map[string]interface{}, result *LintResult) {
	networks := l.getDefinedNames(compose, "networks")
	volumes := l.getDefinedNames(compose, "volumes")
	serviceNames := l.getServiceNames(services)

	for serviceName, serviceConfig := range services {
		if config, ok := serviceConfig.(map[string]interface{}); ok {
			// Check depends_on references
			l.checkDependsOn(serviceName, config, serviceNames, result)

			// Check network references
			l.checkNetworkReferences(serviceName, config, networks, result)

			// Check volume references
			l.checkVolumeReferences(serviceName, config, volumes, result)

			// Check security settings
			l.checkSecuritySettings(serviceName, config, result)

			// Check image format
			l.checkImageFormat(serviceName, config, result)
		}
	}
}

// checkDependsOn validates depends_on references
func (l *Linter) checkDependsOn(serviceName string, config map[string]interface{}, validServices []string, result *LintResult) {
	if dependsOn, ok := config["depends_on"]; ok {
		switch deps := dependsOn.(type) {
		case []interface{}:
			for _, dep := range deps {
				if depName, ok := dep.(string); ok {
					if !contains(validServices, depName) {
						result.Errors = append(result.Errors, LintIssue{
							Rule:    "invalid-depends-on",
							Message: fmt.Sprintf("Service '%s' depends on non-existent service '%s'", serviceName, depName),
						})
					}
				}
			}
		case map[string]interface{}:
			for depName := range deps {
				if !contains(validServices, depName) {
					result.Errors = append(result.Errors, LintIssue{
						Rule:    "invalid-depends-on",
						Message: fmt.Sprintf("Service '%s' depends on non-existent service '%s'", serviceName, depName),
					})
				}
			}
		}
	}
}

// checkNetworkReferences validates network references
func (l *Linter) checkNetworkReferences(serviceName string, config map[string]interface{}, validNetworks []string, result *LintResult) {
	if networks, ok := config["networks"]; ok {
		switch nets := networks.(type) {
		case []interface{}:
			for _, net := range nets {
				if netName, ok := net.(string); ok {
					if netName != "default" && !contains(validNetworks, netName) {
						result.Errors = append(result.Errors, LintIssue{
							Rule:    "invalid-network",
							Message: fmt.Sprintf("Service '%s' references non-existent network '%s'", serviceName, netName),
						})
					}
				}
			}
		}
	}
}

// checkVolumeReferences validates volume references
func (l *Linter) checkVolumeReferences(serviceName string, config map[string]interface{}, validVolumes []string, result *LintResult) {
	if volumes, ok := config["volumes"]; ok {
		if vols, ok := volumes.([]interface{}); ok {
			for _, vol := range vols {
				if volStr, ok := vol.(string); ok {
					// Check if it's a named volume (not a bind mount)
					if !strings.Contains(volStr, "/") && !strings.Contains(volStr, ":") {
						if !contains(validVolumes, volStr) {
							result.Errors = append(result.Errors, LintIssue{
								Rule:    "invalid-volume",
								Message: fmt.Sprintf("Service '%s' references non-existent volume '%s'", serviceName, volStr),
							})
						}
					}
				}
			}
		}
	}
}

// checkSecuritySettings checks for security concerns
func (l *Linter) checkSecuritySettings(serviceName string, config map[string]interface{}, result *LintResult) {
	// Check for privileged mode
	if privileged, ok := config["privileged"].(bool); ok && privileged {
		result.Warnings = append(result.Warnings, LintIssue{
			Rule:    "privileged-container",
			Message: fmt.Sprintf("Service '%s' runs in privileged mode", serviceName),
		})
	}

	// Check for cap_add
	if _, hasCapAdd := config["cap_add"]; hasCapAdd {
		result.Warnings = append(result.Warnings, LintIssue{
			Rule:    "capability-add",
			Message: fmt.Sprintf("Service '%s' adds Linux capabilities", serviceName),
		})
	}
}

// checkImageFormat checks if images use SHA256 digests
func (l *Linter) checkImageFormat(serviceName string, config map[string]interface{}, result *LintResult) {
	if image, ok := config["image"].(string); ok {
		if !strings.Contains(image, "@sha256:") {
			result.Warnings = append(result.Warnings, LintIssue{
				Rule:    "image-digest",
				Message: fmt.Sprintf("Service '%s' image '%s' doesn't use SHA256 digest", serviceName, image),
			})
		}
	} else {
		result.Errors = append(result.Errors, LintIssue{
			Rule:    "missing-image",
			Message: fmt.Sprintf("Service '%s' missing image specification", serviceName),
		})
	}
}

// checkUnresolvedVariables checks for unresolved environment variables
func (l *Linter) checkUnresolvedVariables(compose string, result *LintResult) {
	// Simple check for ${VAR} patterns that might be unresolved
	// In production, this would be more sophisticated
	lines := strings.Split(compose, "\n")
	for i, line := range lines {
		if strings.Contains(line, "${") && strings.Contains(line, "}") {
			// Extract variable name
			start := strings.Index(line, "${")
			end := strings.Index(line[start:], "}")
			if end > 0 {
				varName := line[start+2 : start+end]
				// Check if it looks like an unresolved variable (simple heuristic)
				if !strings.Contains(varName, ":-") && !strings.Contains(varName, "-") {
					result.Warnings = append(result.Warnings, LintIssue{
						Rule:    "unresolved-variable",
						Message: fmt.Sprintf("Possible unresolved variable: ${%s}", varName),
						Line:    i + 1,
					})
				}
			}
		}
	}
}

// checkPortCollisions checks for host port duplications
func (l *Linter) checkPortCollisions(services map[string]interface{}, result *LintResult) {
	usedPorts := make(map[string][]string) // port -> service names

	for serviceName, serviceConfig := range services {
		if config, ok := serviceConfig.(map[string]interface{}); ok {
			if ports, ok := config["ports"].([]interface{}); ok {
				for _, port := range ports {
					if portStr, ok := port.(string); ok {
						// Extract host port
						parts := strings.Split(portStr, ":")
						if len(parts) >= 2 {
							hostPort := parts[0]
							usedPorts[hostPort] = append(usedPorts[hostPort], serviceName)
						}
					}
				}
			}
		}
	}

	// Report collisions
	for port, services := range usedPorts {
		if len(services) > 1 {
			result.Errors = append(result.Errors, LintIssue{
				Rule:    "port-collision",
				Message: fmt.Sprintf("Port %s used by multiple services: %s", port, strings.Join(services, ", ")),
			})
		}
	}
}

// Helper functions

func (l *Linter) getDefinedNames(compose map[string]interface{}, key string) []string {
	var names []string
	if items, ok := compose[key].(map[string]interface{}); ok {
		for name := range items {
			names = append(names, name)
		}
	}
	return names
}

func (l *Linter) getServiceNames(services map[string]interface{}) []string {
	var names []string
	for name := range services {
		names = append(names, name)
	}
	return names
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}