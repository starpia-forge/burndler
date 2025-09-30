import { useState, useEffect, useCallback, useRef } from 'react';
import { ConfigurationValues, ValidationResult } from '../types/configuration';
import api from '../services/api';

// Custom debounce implementation (since lodash is not installed)
function debounce<T extends (...args: any[]) => void>(
  func: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout> | null = null;

  return (...args: Parameters<T>) => {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }

    timeoutId = setTimeout(() => {
      func(...args);
      timeoutId = null;
    }, delay);
  };
}

export const useDependencyValidation = (
  serviceId: string | undefined,
  containerId: string | undefined,
  values: ConfigurationValues
) => {
  const [validationResult, setValidationResult] = useState<ValidationResult>({
    valid: true,
    errors: [],
  });
  const [isValidating, setIsValidating] = useState(false);

  // Use ref to store the debounced function so it doesn't change on every render
  const debouncedValidateRef = useRef<(vals: ConfigurationValues) => void>();

  const validateValues = useCallback(
    async (vals: ConfigurationValues) => {
      // Skip validation if serviceId or containerId is not available
      if (!serviceId || !containerId) {
        setValidationResult({ valid: true, errors: [] });
        return;
      }

      setIsValidating(true);

      try {
        const response = await api.post(
          `/services/${serviceId}/containers/${containerId}/validate`,
          { values: vals }
        );

        setValidationResult(response);
      } catch (error) {
        console.error('Validation failed:', error);
        // On error, reset to valid state to not block the user
        setValidationResult({ valid: true, errors: [] });
      } finally {
        setIsValidating(false);
      }
    },
    [serviceId, containerId]
  );

  // Create debounced version only once
  useEffect(() => {
    debouncedValidateRef.current = debounce(validateValues, 500);
  }, [validateValues]);

  // Trigger validation when values change
  useEffect(() => {
    if (debouncedValidateRef.current) {
      debouncedValidateRef.current(values);
    }
  }, [values]);

  return { validationResult, isValidating };
};
