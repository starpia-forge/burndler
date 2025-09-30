package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependencyChecker(t *testing.T) {
	checker := NewDependencyChecker()
	assert.NotNil(t, checker)
}

func TestDependencyChecker_ValidateConfiguration(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("no rules - no errors", func(t *testing.T) {
		rules := []DependencyRule{}
		values := map[string]interface{}{
			"Field1": "value1",
		}

		errors := checker.ValidateConfiguration(rules, values)
		assert.Empty(t, errors)
	})

	t.Run("multiple rules with some failing", func(t *testing.T) {
		rules := []DependencyRule{
			{
				Type:      "requires",
				Field:     "SSL.Enabled",
				Condition: "{{.SSL.Enabled}} == true",
				Target:    "SSL.Certificate",
				Message:   "SSL requires certificate",
			},
			{
				Type:      "requires",
				Field:     "Database.Enabled",
				Condition: "{{.Database.Enabled}} == true",
				Target:    "Database.Host",
				Message:   "Database requires host",
			},
		}

		values := map[string]interface{}{
			"SSL": map[string]interface{}{
				"Enabled": true,
				// Certificate missing
			},
			"Database": map[string]interface{}{
				"Enabled": true,
				"Host":    "localhost", // Provided
			},
		}

		errors := checker.ValidateConfiguration(rules, values)
		assert.Len(t, errors, 1)
		assert.Equal(t, "SSL.Certificate", errors[0].Field)
		assert.Equal(t, "SSL requires certificate", errors[0].Message)
		assert.Equal(t, "requires", errors[0].Rule)
	})

	t.Run("all rules pass", func(t *testing.T) {
		rules := []DependencyRule{
			{
				Type:      "requires",
				Field:     "SSL.Enabled",
				Condition: "{{.SSL.Enabled}} == true",
				Target:    "SSL.Certificate",
			},
		}

		values := map[string]interface{}{
			"SSL": map[string]interface{}{
				"Enabled":     true,
				"Certificate": "/path/to/cert",
			},
		}

		errors := checker.ValidateConfiguration(rules, values)
		assert.Empty(t, errors)
	})
}

func TestDependencyChecker_ValidateRequires(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("target field is set - no error", func(t *testing.T) {
		rule := DependencyRule{
			Type:      "requires",
			Field:     "SSL.Enabled",
			Condition: "{{.SSL.Enabled}} == true",
			Target:    "SSL.Certificate",
			Message:   "SSL requires certificate path",
		}

		values := map[string]interface{}{
			"SSL": map[string]interface{}{
				"Enabled":     true,
				"Certificate": "/path/to/cert",
			},
		}

		err := checker.validateRequires(rule, values)
		assert.Nil(t, err)
	})

	t.Run("target field is missing - error", func(t *testing.T) {
		rule := DependencyRule{
			Type:      "requires",
			Field:     "SSL.Enabled",
			Condition: "{{.SSL.Enabled}} == true",
			Target:    "SSL.Certificate",
			Message:   "SSL requires certificate path",
		}

		values := map[string]interface{}{
			"SSL": map[string]interface{}{
				"Enabled": true,
			},
		}

		err := checker.validateRequires(rule, values)
		require.NotNil(t, err)
		assert.Equal(t, "SSL.Certificate", err.Field)
		assert.Equal(t, "SSL requires certificate path", err.Message)
		assert.Equal(t, "requires", err.Rule)
	})

	t.Run("target field is empty string - error", func(t *testing.T) {
		rule := DependencyRule{
			Type:   "requires",
			Field:  "SSL.Enabled",
			Target: "SSL.Certificate",
		}

		values := map[string]interface{}{
			"SSL": map[string]interface{}{
				"Certificate": "",
			},
		}

		err := checker.validateRequires(rule, values)
		require.NotNil(t, err)
		assert.Equal(t, "SSL.Certificate", err.Field)
	})

	t.Run("target field is false boolean - error", func(t *testing.T) {
		rule := DependencyRule{
			Type:   "requires",
			Field:  "Feature.Main",
			Target: "Feature.Sub",
		}

		values := map[string]interface{}{
			"Feature": map[string]interface{}{
				"Sub": false,
			},
		}

		err := checker.validateRequires(rule, values)
		require.NotNil(t, err)
		assert.Equal(t, "Feature.Sub", err.Field)
	})

	t.Run("default error message", func(t *testing.T) {
		rule := DependencyRule{
			Type:   "requires",
			Field:  "SSL.Enabled",
			Target: "SSL.Certificate",
			// No custom message
		}

		values := map[string]interface{}{
			"SSL": map[string]interface{}{
				"Enabled": true,
			},
		}

		err := checker.validateRequires(rule, values)
		require.NotNil(t, err)
		assert.Contains(t, err.Message, "SSL.Enabled")
		assert.Contains(t, err.Message, "SSL.Certificate")
	})
}

func TestDependencyChecker_ValidateConflicts(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("target field is not set - no error", func(t *testing.T) {
		rule := DependencyRule{
			Type:      "conflicts",
			Field:     "Auth.OAuth",
			Condition: "{{.Auth.OAuth}} == true",
			Target:    "Auth.LDAP",
			Message:   "OAuth conflicts with LDAP",
		}

		values := map[string]interface{}{
			"Auth": map[string]interface{}{
				"OAuth": true,
				// LDAP not set
			},
		}

		err := checker.validateConflicts(rule, values)
		assert.Nil(t, err)
	})

	t.Run("target field is set - error", func(t *testing.T) {
		rule := DependencyRule{
			Type:      "conflicts",
			Field:     "Auth.OAuth",
			Condition: "{{.Auth.OAuth}} == true",
			Target:    "Auth.LDAP",
			Message:   "OAuth conflicts with LDAP",
		}

		values := map[string]interface{}{
			"Auth": map[string]interface{}{
				"OAuth": true,
				"LDAP":  true,
			},
		}

		err := checker.validateConflicts(rule, values)
		require.NotNil(t, err)
		assert.Equal(t, "Auth.LDAP", err.Field)
		assert.Equal(t, "OAuth conflicts with LDAP", err.Message)
		assert.Equal(t, "conflicts", err.Rule)
	})

	t.Run("target field is false - no error", func(t *testing.T) {
		rule := DependencyRule{
			Type:   "conflicts",
			Field:  "Auth.OAuth",
			Target: "Auth.LDAP",
		}

		values := map[string]interface{}{
			"Auth": map[string]interface{}{
				"OAuth": true,
				"LDAP":  false,
			},
		}

		err := checker.validateConflicts(rule, values)
		assert.Nil(t, err)
	})

	t.Run("default error message", func(t *testing.T) {
		rule := DependencyRule{
			Type:   "conflicts",
			Field:  "Auth.OAuth",
			Target: "Auth.LDAP",
			// No custom message
		}

		values := map[string]interface{}{
			"Auth": map[string]interface{}{
				"OAuth": true,
				"LDAP":  true,
			},
		}

		err := checker.validateConflicts(rule, values)
		require.NotNil(t, err)
		assert.Contains(t, err.Message, "Auth.OAuth")
		assert.Contains(t, err.Message, "Auth.LDAP")
	})
}

func TestDependencyChecker_EvaluateCondition(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("empty condition - always true", func(t *testing.T) {
		result, err := checker.evaluateCondition("", map[string]interface{}{})
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("simple boolean equality - true", func(t *testing.T) {
		condition := "{{.SSL.Enabled}} == true"
		values := map[string]interface{}{
			"SSL": map[string]interface{}{
				"Enabled": true,
			},
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("simple boolean equality - false", func(t *testing.T) {
		condition := "{{.SSL.Enabled}} == true"
		values := map[string]interface{}{
			"SSL": map[string]interface{}{
				"Enabled": false,
			},
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("boolean inequality", func(t *testing.T) {
		condition := "{{.Debug}} != false"
		values := map[string]interface{}{
			"Debug": true,
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("numeric equality", func(t *testing.T) {
		condition := "{{.Port}} == 8080"
		values := map[string]interface{}{
			"Port": 8080,
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("numeric greater than", func(t *testing.T) {
		condition := "{{.Count}} > 5"
		values := map[string]interface{}{
			"Count": 10,
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("numeric greater than or equal", func(t *testing.T) {
		condition := "{{.Count}} >= 10"
		values := map[string]interface{}{
			"Count": 10,
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("numeric less than", func(t *testing.T) {
		condition := "{{.Count}} < 5"
		values := map[string]interface{}{
			"Count": 3,
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("numeric less than or equal", func(t *testing.T) {
		condition := "{{.Count}} <= 5"
		values := map[string]interface{}{
			"Count": 5,
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("string equality with quotes", func(t *testing.T) {
		condition := `{{.Environment}} == "production"`
		values := map[string]interface{}{
			"Environment": "production",
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("string inequality", func(t *testing.T) {
		condition := `{{.Environment}} != "development"`
		values := map[string]interface{}{
			"Environment": "production",
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("nested field access", func(t *testing.T) {
		condition := "{{.Database.Port}} == 5432"
		values := map[string]interface{}{
			"Database": map[string]interface{}{
				"Port": 5432,
			},
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("deeply nested field access", func(t *testing.T) {
		condition := "{{.Server.Database.Primary.Port}} == 5432"
		values := map[string]interface{}{
			"Server": map[string]interface{}{
				"Database": map[string]interface{}{
					"Primary": map[string]interface{}{
						"Port": 5432,
					},
				},
			},
		}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("missing field - nil comparison", func(t *testing.T) {
		condition := "{{.MissingField}} == true"
		values := map[string]interface{}{}

		result, err := checker.evaluateCondition(condition, values)
		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("invalid condition format - no braces", func(t *testing.T) {
		condition := ".Field == true"
		values := map[string]interface{}{}

		_, err := checker.evaluateCondition(condition, values)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid condition format")
	})

	t.Run("invalid condition format - no operator", func(t *testing.T) {
		condition := "{{.Field}} true"
		values := map[string]interface{}{}

		_, err := checker.evaluateCondition(condition, values)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid operator")
	})
}

func TestDependencyChecker_GetNestedValue(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("simple field", func(t *testing.T) {
		values := map[string]interface{}{
			"Field": "value",
		}

		result := checker.getNestedValue(values, "Field")
		assert.Equal(t, "value", result)
	})

	t.Run("nested field - one level", func(t *testing.T) {
		values := map[string]interface{}{
			"Database": map[string]interface{}{
				"Host": "localhost",
			},
		}

		result := checker.getNestedValue(values, "Database.Host")
		assert.Equal(t, "localhost", result)
	})

	t.Run("nested field - multiple levels", func(t *testing.T) {
		values := map[string]interface{}{
			"Server": map[string]interface{}{
				"Database": map[string]interface{}{
					"Primary": map[string]interface{}{
						"Host": "primary.db.local",
					},
				},
			},
		}

		result := checker.getNestedValue(values, "Server.Database.Primary.Host")
		assert.Equal(t, "primary.db.local", result)
	})

	t.Run("missing field - returns nil", func(t *testing.T) {
		values := map[string]interface{}{
			"Field": "value",
		}

		result := checker.getNestedValue(values, "MissingField")
		assert.Nil(t, result)
	})

	t.Run("missing nested field - returns nil", func(t *testing.T) {
		values := map[string]interface{}{
			"Database": map[string]interface{}{
				"Host": "localhost",
			},
		}

		result := checker.getNestedValue(values, "Database.Port")
		assert.Nil(t, result)
	})

	t.Run("invalid nested path - returns nil", func(t *testing.T) {
		values := map[string]interface{}{
			"Field": "not a map",
		}

		result := checker.getNestedValue(values, "Field.SubField")
		assert.Nil(t, result)
	})

	t.Run("empty key - returns nil", func(t *testing.T) {
		values := map[string]interface{}{
			"Field": "value",
		}

		result := checker.getNestedValue(values, "")
		assert.Nil(t, result)
	})
}

func TestDependencyChecker_IsEmpty(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("nil is empty", func(t *testing.T) {
		assert.True(t, checker.isEmpty(nil))
	})

	t.Run("empty string is empty", func(t *testing.T) {
		assert.True(t, checker.isEmpty(""))
	})

	t.Run("non-empty string is not empty", func(t *testing.T) {
		assert.False(t, checker.isEmpty("value"))
	})

	t.Run("false boolean is empty", func(t *testing.T) {
		assert.True(t, checker.isEmpty(false))
	})

	t.Run("true boolean is not empty", func(t *testing.T) {
		assert.False(t, checker.isEmpty(true))
	})

	t.Run("zero integer is empty", func(t *testing.T) {
		assert.True(t, checker.isEmpty(0))
	})

	t.Run("non-zero integer is not empty", func(t *testing.T) {
		assert.False(t, checker.isEmpty(42))
	})

	t.Run("empty slice is empty", func(t *testing.T) {
		assert.True(t, checker.isEmpty([]string{}))
	})

	t.Run("non-empty slice is not empty", func(t *testing.T) {
		assert.False(t, checker.isEmpty([]string{"item"}))
	})

	t.Run("empty map is empty", func(t *testing.T) {
		assert.True(t, checker.isEmpty(map[string]string{}))
	})

	t.Run("non-empty map is not empty", func(t *testing.T) {
		assert.False(t, checker.isEmpty(map[string]string{"key": "value"}))
	})
}

func TestDependencyChecker_ParseValue(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("boolean true", func(t *testing.T) {
		result, err := checker.parseValue("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("boolean false", func(t *testing.T) {
		result, err := checker.parseValue("false")
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("integer", func(t *testing.T) {
		result, err := checker.parseValue("42")
		require.NoError(t, err)
		assert.Equal(t, int64(42), result)
	})

	t.Run("negative integer", func(t *testing.T) {
		result, err := checker.parseValue("-100")
		require.NoError(t, err)
		assert.Equal(t, int64(-100), result)
	})

	t.Run("float", func(t *testing.T) {
		result, err := checker.parseValue("3.14")
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
	})

	t.Run("quoted string - double quotes", func(t *testing.T) {
		result, err := checker.parseValue(`"hello"`)
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("quoted string - single quotes", func(t *testing.T) {
		result, err := checker.parseValue(`'hello'`)
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("unquoted string", func(t *testing.T) {
		result, err := checker.parseValue("production")
		require.NoError(t, err)
		assert.Equal(t, "production", result)
	})

	t.Run("string with spaces", func(t *testing.T) {
		result, err := checker.parseValue("  value  ")
		require.NoError(t, err)
		assert.Equal(t, "value", result)
	})
}

func TestDependencyChecker_ValidateRule_CascadeType(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("cascade rule - no validation error", func(t *testing.T) {
		rule := DependencyRule{
			Type:      "cascades",
			Field:     "Environment",
			Condition: "{{.Environment}} == \"production\"",
			Target:    "Debug",
		}

		values := map[string]interface{}{
			"Environment": "production",
		}

		err := checker.validateRule(rule, values)
		assert.Nil(t, err) // Cascade rules don't produce validation errors
	})
}

func TestDependencyChecker_ValidateRule_UnknownType(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("unknown rule type - error", func(t *testing.T) {
		rule := DependencyRule{
			Type:      "unknown",
			Field:     "Field1",
			Condition: "{{.Field1}} == true",
			Target:    "Field2",
		}

		values := map[string]interface{}{
			"Field1": true,
		}

		err := checker.validateRule(rule, values)
		require.NotNil(t, err)
		assert.Contains(t, err.Message, "Unknown rule type")
	})
}

func TestDependencyChecker_ValidateRule_ConditionNotMet(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("condition not met - rule skipped", func(t *testing.T) {
		rule := DependencyRule{
			Type:      "requires",
			Field:     "SSL.Enabled",
			Condition: "{{.SSL.Enabled}} == true",
			Target:    "SSL.Certificate",
		}

		values := map[string]interface{}{
			"SSL": map[string]interface{}{
				"Enabled": false,
				// Certificate not provided, but condition is false so rule doesn't apply
			},
		}

		err := checker.validateRule(rule, values)
		assert.Nil(t, err)
	})
}

func TestDependencyChecker_IntegrationScenarios(t *testing.T) {
	checker := NewDependencyChecker()

	t.Run("SSL configuration scenario", func(t *testing.T) {
		rules := []DependencyRule{
			{
				Type:      "requires",
				Field:     "SSL.Enabled",
				Condition: "{{.SSL.Enabled}} == true",
				Target:    "SSL.Certificate",
				Message:   "SSL enabled requires certificate path",
			},
			{
				Type:      "requires",
				Field:     "SSL.Enabled",
				Condition: "{{.SSL.Enabled}} == true",
				Target:    "SSL.PrivateKey",
				Message:   "SSL enabled requires private key path",
			},
		}

		t.Run("SSL disabled - no errors", func(t *testing.T) {
			values := map[string]interface{}{
				"SSL": map[string]interface{}{
					"Enabled": false,
				},
			}

			errors := checker.ValidateConfiguration(rules, values)
			assert.Empty(t, errors)
		})

		t.Run("SSL enabled without certificate - error", func(t *testing.T) {
			values := map[string]interface{}{
				"SSL": map[string]interface{}{
					"Enabled":    true,
					"PrivateKey": "/path/to/key",
					// Certificate missing
				},
			}

			errors := checker.ValidateConfiguration(rules, values)
			assert.Len(t, errors, 1)
			assert.Equal(t, "SSL.Certificate", errors[0].Field)
		})

		t.Run("SSL enabled with all required fields - no errors", func(t *testing.T) {
			values := map[string]interface{}{
				"SSL": map[string]interface{}{
					"Enabled":     true,
					"Certificate": "/path/to/cert",
					"PrivateKey":  "/path/to/key",
				},
			}

			errors := checker.ValidateConfiguration(rules, values)
			assert.Empty(t, errors)
		})
	})

	t.Run("authentication conflict scenario", func(t *testing.T) {
		rules := []DependencyRule{
			{
				Type:      "conflicts",
				Field:     "Auth.OAuth",
				Condition: "{{.Auth.OAuth}} == true",
				Target:    "Auth.LDAP",
				Message:   "Cannot use OAuth and LDAP simultaneously",
			},
		}

		t.Run("only OAuth enabled - no error", func(t *testing.T) {
			values := map[string]interface{}{
				"Auth": map[string]interface{}{
					"OAuth": true,
					"LDAP":  false,
				},
			}

			errors := checker.ValidateConfiguration(rules, values)
			assert.Empty(t, errors)
		})

		t.Run("both OAuth and LDAP enabled - error", func(t *testing.T) {
			values := map[string]interface{}{
				"Auth": map[string]interface{}{
					"OAuth": true,
					"LDAP":  true,
				},
			}

			errors := checker.ValidateConfiguration(rules, values)
			assert.Len(t, errors, 1)
			assert.Equal(t, "Auth.LDAP", errors[0].Field)
			assert.Contains(t, errors[0].Message, "OAuth")
			assert.Contains(t, errors[0].Message, "LDAP")
		})
	})
}