package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// TemplateEngine handles rendering of templates in various formats
type TemplateEngine struct {
	funcMap template.FuncMap
}

// NewTemplateEngine creates a new template engine with standard functions
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		funcMap: getTemplateFuncMap(),
	}
}

// RenderYAML renders a YAML template with structure preservation
func (te *TemplateEngine) RenderYAML(templateContent string, variables map[string]interface{}) (string, error) {
	// 1. Parse and execute Go template
	tmpl, err := template.New("yaml").Funcs(te.funcMap).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	// 2. Validate YAML structure
	var yamlData interface{}
	if err := yaml.Unmarshal(buf.Bytes(), &yamlData); err != nil {
		return "", fmt.Errorf("invalid YAML after rendering: %w", err)
	}

	// 3. Format and return YAML
	formatted, err := yaml.Marshal(yamlData)
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}

// RenderJSON renders a JSON template with structure preservation
func (te *TemplateEngine) RenderJSON(templateContent string, variables map[string]interface{}) (string, error) {
	// 1. Parse and execute Go template
	tmpl, err := template.New("json").Funcs(te.funcMap).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	// 2. Validate JSON structure
	var jsonData interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonData); err != nil {
		return "", fmt.Errorf("invalid JSON after rendering: %w", err)
	}

	// 3. Format and return JSON with 2-space indentation
	formatted, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}

// RenderEnv renders an ENV file template
func (te *TemplateEngine) RenderEnv(templateContent string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("env").Funcs(te.funcMap).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// RenderText renders a plain text template
func (te *TemplateEngine) RenderText(templateContent string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("text").Funcs(te.funcMap).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// Render automatically selects the appropriate renderer based on format
func (te *TemplateEngine) Render(templateContent string, format string, variables map[string]interface{}) (string, error) {
	switch format {
	case "yaml":
		return te.RenderYAML(templateContent, variables)
	case "json":
		return te.RenderJSON(templateContent, variables)
	case "env":
		return te.RenderEnv(templateContent, variables)
	case "text":
		return te.RenderText(templateContent, variables)
	default:
		return "", fmt.Errorf("unsupported template format: %s", format)
	}
}

// getTemplateFuncMap returns template helper functions
func getTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper":   strings.ToUpper,
		"lower":   strings.ToLower,
		"trim":    strings.TrimSpace,
		"replace": strings.ReplaceAll,

		// Default value function
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},

		// Math functions
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },

		// Conditional functions
		"eq": func(a, b interface{}) bool { return a == b },
		"ne": func(a, b interface{}) bool { return a != b },
	}
}
