import React, { useState, useCallback } from 'react';
import {
  UISchema,
  UIField,
  ConfigurationValues,
  ValidationErrors,
} from '../../types/configuration';
import { ConfigurationSection } from './ConfigurationSection';
import { useDependencyValidation } from '../../hooks/useDependencyValidation';

interface ConfigurationFormProps {
  schema: UISchema;
  initialValues?: ConfigurationValues;
  onChange: (values: ConfigurationValues) => void;
  onValidate?: (errors: ValidationErrors) => void;
  serviceId?: string;
  containerId?: string;
}

// Field validation function
const validateField = (key: string, value: any, schema: UISchema): string | undefined => {
  // Find the field definition
  let field: UIField | undefined;
  for (const section of schema.sections) {
    field = section.fields.find((f) => f.key === key);
    if (field) break;
  }

  if (!field) return undefined;

  // Required validation
  if (field.required && (value === undefined || value === null || value === '')) {
    return `${field.label}은(는) 필수 입력 항목입니다`;
  }

  // Type-specific validation
  if (value !== undefined && value !== null && value !== '') {
    const validation = field.validation;

    if (validation) {
      // String validations
      if (field.type === 'string' || field.type === 'textarea') {
        const strValue = String(value);
        if (validation.minLength && strValue.length < validation.minLength) {
          return `최소 ${validation.minLength}자 이상 입력해야 합니다`;
        }
        if (validation.maxLength && strValue.length > validation.maxLength) {
          return `최대 ${validation.maxLength}자까지 입력 가능합니다`;
        }
        if (validation.pattern) {
          const regex = new RegExp(validation.pattern);
          if (!regex.test(strValue)) {
            return `올바른 형식이 아닙니다`;
          }
        }
      }

      // Number validations
      if (field.type === 'number') {
        const numValue = Number(value);
        if (validation.min !== undefined && numValue < validation.min) {
          return `${validation.min} 이상이어야 합니다`;
        }
        if (validation.max !== undefined && numValue > validation.max) {
          return `${validation.max} 이하여야 합니다`;
        }
      }

      // Enum validation
      if (validation.enum && !validation.enum.includes(String(value))) {
        return `허용된 값이 아닙니다`;
      }

      // Custom validation
      if (validation.custom) {
        try {
          const isValid = new Function('value', `return ${validation.custom}`)(value);
          if (!isValid) {
            return `유효하지 않은 값입니다`;
          }
        } catch {
          // Ignore custom validation errors
        }
      }
    }
  }

  return undefined;
};

export const ConfigurationForm: React.FC<ConfigurationFormProps> = ({
  schema,
  initialValues = {},
  onChange,
  onValidate,
  serviceId,
  containerId,
}) => {
  const [values, setValues] = useState<ConfigurationValues>(initialValues);
  const [errors, setErrors] = useState<ValidationErrors>({});

  // Dependency validation (real-time API-based validation)
  const { validationResult, isValidating } = useDependencyValidation(
    serviceId,
    containerId,
    values
  );

  const handleFieldChange = useCallback(
    (key: string, value: any) => {
      const newValues = {
        ...values,
        [key]: value,
      };

      setValues(newValues);
      onChange(newValues);

      // Validate field
      const fieldError = validateField(key, value, schema);
      const newErrors = {
        ...errors,
        [key]: fieldError || '',
      };

      // Remove empty error strings
      if (!fieldError) {
        delete newErrors[key];
      }

      setErrors(newErrors);

      if (onValidate) {
        onValidate(newErrors);
      }
    },
    [values, schema, onChange, onValidate, errors]
  );

  const evaluateCondition = useCallback(
    (condition: string | undefined): boolean => {
      if (!condition) return true;

      try {
        // Simple condition evaluation
        // 예: "Database.Enabled === true"
        return new Function('values', `with(values) { return ${condition} }`)(values);
      } catch {
        return false;
      }
    },
    [values]
  );

  return (
    <div className="configuration-form space-y-6">
      {/* Validation status indicator */}
      {isValidating && (
        <div className="validation-indicator text-sm text-muted-foreground flex items-center space-x-2">
          <div className="animate-spin h-4 w-4 border-2 border-primary border-t-transparent rounded-full" />
          <span>검증 중...</span>
        </div>
      )}

      {/* Dependency validation errors */}
      {!validationResult.valid && validationResult.errors.length > 0 && (
        <div className="validation-errors bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <h4 className="font-semibold text-red-800 dark:text-red-200 mb-2">설정 오류</h4>
          <ul className="list-disc list-inside space-y-1">
            {validationResult.errors.map((error, idx) => (
              <li key={idx} className="text-sm text-red-700 dark:text-red-300">
                <span className="font-medium">{error.field}:</span> {error.message}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Configuration sections */}
      {schema.sections.map((section) => {
        const shouldShow = evaluateCondition(section.condition);

        if (!shouldShow) return null;

        return (
          <ConfigurationSection
            key={section.id}
            section={section}
            values={values}
            errors={errors}
            onChange={handleFieldChange}
            evaluateCondition={evaluateCondition}
          />
        );
      })}
    </div>
  );
};
