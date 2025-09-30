import React from 'react';
import { UIField } from '../../types/configuration';

interface ConfigurationFieldProps {
  field: UIField;
  value: any;
  error?: string;
  disabled?: boolean;
  onChange: (value: any) => void;
}

export const ConfigurationField: React.FC<ConfigurationFieldProps> = ({
  field,
  value,
  error,
  disabled,
  onChange,
}) => {
  const renderField = () => {
    switch (field.type) {
      case 'boolean':
        return (
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={value || false}
              disabled={disabled}
              onChange={(e) => onChange(e.target.checked)}
              className="form-checkbox h-4 w-4 text-blue-600 border-border rounded focus:ring-2 focus:ring-blue-500"
            />
            <span className="text-sm font-medium text-foreground">{field.label}</span>
            {field.required && <span className="text-red-500 ml-1">*</span>}
          </label>
        );

      case 'string':
        return (
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              {field.label}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <input
              type="text"
              value={value || ''}
              disabled={disabled}
              placeholder={field.ui?.placeholder}
              onChange={(e) => onChange(e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground disabled:bg-muted disabled:text-muted-foreground"
            />
          </div>
        );

      case 'number':
        return (
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              {field.label}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <div className="flex items-center">
              <input
                type="number"
                value={value || ''}
                disabled={disabled}
                min={field.validation?.min}
                max={field.validation?.max}
                onChange={(e) => onChange(Number(e.target.value))}
                className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground disabled:bg-muted disabled:text-muted-foreground"
              />
              {field.ui?.unit && (
                <span className="ml-2 text-sm text-muted-foreground">{field.ui.unit}</span>
              )}
            </div>
          </div>
        );

      case 'select':
        return (
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              {field.label}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <select
              value={value || ''}
              disabled={disabled}
              onChange={(e) => onChange(e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground disabled:bg-muted disabled:text-muted-foreground"
            >
              <option value="">선택하세요</option>
              {field.ui?.options?.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        );

      case 'multiselect':
        return (
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              {field.label}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <select
              multiple
              value={Array.isArray(value) ? value : []}
              disabled={disabled}
              onChange={(e) => {
                const selected = Array.from(e.target.selectedOptions, (option) => option.value);
                onChange(selected);
              }}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground disabled:bg-muted disabled:text-muted-foreground"
              size={Math.min(field.ui?.options?.length || 5, 5)}
            >
              {field.ui?.options?.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        );

      case 'textarea':
        return (
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              {field.label}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <textarea
              value={value || ''}
              disabled={disabled}
              rows={field.ui?.rows || 3}
              placeholder={field.ui?.placeholder}
              onChange={(e) => onChange(e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground resize-vertical disabled:bg-muted disabled:text-muted-foreground"
            />
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className={`field ${disabled ? 'opacity-50' : ''}`}>
      {renderField()}
      {field.description && !error && (
        <p className="text-xs text-muted-foreground mt-1">{field.description}</p>
      )}
      {error && <p className="text-xs text-red-500 mt-1">{error}</p>}
      {field.ui?.helpText && <p className="text-xs text-blue-500 mt-1">{field.ui.helpText}</p>}
    </div>
  );
};
