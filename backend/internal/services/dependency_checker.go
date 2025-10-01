package services

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// DependencyChecker validates configuration dependencies
type DependencyChecker struct{}

// NewDependencyChecker creates a new dependency checker instance
func NewDependencyChecker() *DependencyChecker {
	return &DependencyChecker{}
}

// DependencyRule defines a dependency relationship between configuration fields
type DependencyRule struct {
	Type      string `json:"type"`      // "requires", "conflicts", "cascades"
	Field     string `json:"field"`     // Source field that triggers the rule
	Condition string `json:"condition"` // Condition expression to evaluate
	Target    string `json:"target"`    // Target field affected by the rule
	Message   string `json:"message"`   // Custom error message
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Rule    string `json:"rule"`
}

// ValidateConfiguration validates configuration values against dependency rules
func (dc *DependencyChecker) ValidateConfiguration(
	rules []DependencyRule,
	values map[string]interface{},
) []ValidationError {
	errors := []ValidationError{}

	for _, rule := range rules {
		if err := dc.validateRule(rule, values); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// validateRule validates a single dependency rule
func (dc *DependencyChecker) validateRule(
	rule DependencyRule,
	values map[string]interface{},
) *ValidationError {
	// 1. Evaluate condition
	conditionMet, err := dc.EvaluateCondition(rule.Condition, values)
	if err != nil {
		return &ValidationError{
			Field:   rule.Field,
			Message: fmt.Sprintf("Failed to evaluate condition: %v", err),
			Rule:    rule.Type,
		}
	}

	if !conditionMet {
		return nil // Condition not met, rule doesn't apply
	}

	// 2. Validate based on rule type
	switch rule.Type {
	case "requires":
		return dc.validateRequires(rule, values)
	case "conflicts":
		return dc.validateConflicts(rule, values)
	case "cascades":
		// Cascade rules are handled during value setting, not validation
		return nil
	default:
		return &ValidationError{
			Field:   rule.Field,
			Message: fmt.Sprintf("Unknown rule type: %s", rule.Type),
			Rule:    rule.Type,
		}
	}
}

// validateRequires validates that required fields are set
func (dc *DependencyChecker) validateRequires(
	rule DependencyRule,
	values map[string]interface{},
) *ValidationError {
	targetValue := dc.getNestedValue(values, rule.Target)

	if dc.isEmpty(targetValue) {
		message := rule.Message
		if message == "" {
			message = fmt.Sprintf("%s requires %s to be set", rule.Field, rule.Target)
		}

		return &ValidationError{
			Field:   rule.Target,
			Message: message,
			Rule:    "requires",
		}
	}

	return nil
}

// validateConflicts validates that conflicting fields are not both set
func (dc *DependencyChecker) validateConflicts(
	rule DependencyRule,
	values map[string]interface{},
) *ValidationError {
	targetValue := dc.getNestedValue(values, rule.Target)

	if !dc.isEmpty(targetValue) {
		message := rule.Message
		if message == "" {
			message = fmt.Sprintf("%s conflicts with %s", rule.Field, rule.Target)
		}

		return &ValidationError{
			Field:   rule.Target,
			Message: message,
			Rule:    "conflicts",
		}
	}

	return nil
}

// evaluateCondition evaluates a simple boolean condition expression
// Supports: ==, !=, >, <, >=, <=
// Format: "{{.FieldName}} operator value" or "{{.Nested.Field}} operator value"
// EvaluateCondition evaluates a template condition expression
func (dc *DependencyChecker) EvaluateCondition(
	condition string,
	values map[string]interface{},
) (bool, error) {
	if condition == "" {
		return true, nil
	}

	// Parse condition: extract field reference, operator, and expected value
	// Example: "{{.SSL.Enabled}} == true"
	condition = strings.TrimSpace(condition)

	// Extract field reference (between {{ and }})
	fieldStart := strings.Index(condition, "{{")
	fieldEnd := strings.Index(condition, "}}")
	if fieldStart == -1 || fieldEnd == -1 || fieldEnd <= fieldStart {
		return false, fmt.Errorf("invalid condition format: %s", condition)
	}

	fieldRef := strings.TrimSpace(condition[fieldStart+2 : fieldEnd])
	fieldRef = strings.TrimPrefix(fieldRef, ".")

	rest := strings.TrimSpace(condition[fieldEnd+2:])

	// Parse operator and expected value
	var operator, expectedStr string
	for _, op := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		if strings.HasPrefix(rest, op) {
			operator = op
			expectedStr = strings.TrimSpace(rest[len(op):])
			break
		}
	}

	if operator == "" {
		return false, fmt.Errorf("no valid operator found in condition: %s", condition)
	}

	// Get actual value from values map
	actualValue := dc.getNestedValue(values, fieldRef)

	// Convert expected value string to appropriate type
	expectedValue, err := dc.parseValue(expectedStr)
	if err != nil {
		return false, fmt.Errorf("failed to parse expected value: %w", err)
	}

	// Perform comparison
	return dc.compare(actualValue, operator, expectedValue)
}

// getNestedValue retrieves a nested value from a map using dot notation
// Example: "Database.Host" returns values["Database"]["Host"]
func (dc *DependencyChecker) getNestedValue(
	values map[string]interface{},
	key string,
) interface{} {
	if key == "" {
		return nil
	}

	parts := strings.Split(key, ".")
	current := values

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - return the value
			return current[part]
		}

		// Navigate deeper
		next, ok := current[part]
		if !ok {
			return nil
		}

		// Check if next is a map
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			return nil
		}
		current = nextMap
	}

	return nil
}

// isEmpty checks if a value is considered empty
func (dc *DependencyChecker) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	default:
		return false
	}
}

// parseValue converts a string value to its appropriate type
func (dc *DependencyChecker) parseValue(s string) (interface{}, error) {
	s = strings.TrimSpace(s)

	// Remove quotes if present
	if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
		(strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`)) {
		return s[1 : len(s)-1], nil
	}

	// Try boolean
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}

	// Try integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i, nil
	}

	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	}

	// Default to string
	return s, nil
}

// compare compares two values using the given operator
func (dc *DependencyChecker) compare(actual interface{}, operator string, expected interface{}) (bool, error) {
	// Handle nil cases
	if actual == nil && expected == nil {
		return operator == "==" || operator == "<=" || operator == ">=", nil
	}
	if actual == nil || expected == nil {
		return operator == "!=", nil
	}

	// Convert both to the same type for comparison
	actualVal := reflect.ValueOf(actual)
	expectedVal := reflect.ValueOf(expected)

	// Try to compare based on type
	switch operator {
	case "==":
		return dc.equals(actual, expected), nil
	case "!=":
		return !dc.equals(actual, expected), nil
	case ">", ">=", "<", "<=":
		return dc.compareNumeric(actualVal, operator, expectedVal)
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// equals checks if two values are equal
func (dc *DependencyChecker) equals(a, b interface{}) bool {
	// Direct comparison
	if a == b {
		return true
	}

	// Type conversion for numeric types
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// If both are numeric, compare as float64
	if dc.isNumeric(aVal) && dc.isNumeric(bVal) {
		return dc.toFloat64(aVal) == dc.toFloat64(bVal)
	}

	// String comparison
	if aVal.Kind() == reflect.String && bVal.Kind() == reflect.String {
		return aVal.String() == bVal.String()
	}

	// Boolean comparison
	if aVal.Kind() == reflect.Bool && bVal.Kind() == reflect.Bool {
		return aVal.Bool() == bVal.Bool()
	}

	return false
}

// compareNumeric compares two numeric values
func (dc *DependencyChecker) compareNumeric(actual reflect.Value, operator string, expected reflect.Value) (bool, error) {
	if !dc.isNumeric(actual) || !dc.isNumeric(expected) {
		return false, fmt.Errorf("non-numeric values cannot be compared with %s", operator)
	}

	actualFloat := dc.toFloat64(actual)
	expectedFloat := dc.toFloat64(expected)

	switch operator {
	case ">":
		return actualFloat > expectedFloat, nil
	case ">=":
		return actualFloat >= expectedFloat, nil
	case "<":
		return actualFloat < expectedFloat, nil
	case "<=":
		return actualFloat <= expectedFloat, nil
	default:
		return false, fmt.Errorf("invalid numeric operator: %s", operator)
	}
}

// isNumeric checks if a value is a numeric type
func (dc *DependencyChecker) isNumeric(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// toFloat64 converts a numeric reflect.Value to float64
func (dc *DependencyChecker) toFloat64(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	default:
		return 0
	}
}