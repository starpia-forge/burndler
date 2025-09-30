import React from 'react';
import { UISection, ConfigurationValues, ValidationErrors } from '../../types/configuration';
import { ConfigurationField } from './ConfigurationField';

interface ConfigurationSectionProps {
  section: UISection;
  values: ConfigurationValues;
  errors: ValidationErrors;
  onChange: (key: string, value: any) => void;
  evaluateCondition: (condition: string | undefined) => boolean;
}

export const ConfigurationSection: React.FC<ConfigurationSectionProps> = ({
  section,
  values,
  errors,
  onChange,
  evaluateCondition,
}) => {
  return (
    <div className="section border border-border rounded-lg p-6 bg-card">
      <h3 className="text-lg font-semibold text-foreground mb-2">{section.title}</h3>
      {section.description && (
        <p className="text-sm text-muted-foreground mb-4">{section.description}</p>
      )}

      <div className="fields space-y-4">
        {section.fields.map((field) => {
          const shouldShow = evaluateCondition(field.condition);

          if (!shouldShow) return null;

          return (
            <ConfigurationField
              key={field.key}
              field={field}
              value={values[field.key]}
              error={errors[field.key]}
              disabled={false}
              onChange={(value) => onChange(field.key, value)}
            />
          );
        })}
      </div>
    </div>
  );
};
