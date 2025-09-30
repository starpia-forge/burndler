// Container configuration UI type definitions

export interface UISchema {
  sections: UISection[];
}

export interface UISection {
  id: string;
  title: string;
  description?: string;
  fields: UIField[];
  condition?: string; // 섹션 표시 조건
}

export interface UIField {
  key: string; // 변수 키 (예: "Database.Host")
  type: UIFieldType;
  label: string;
  description?: string;
  defaultValue?: any;
  required?: boolean;
  validation?: FieldValidation;
  affects?: string[]; // 영향받는 파일 경로들
  dependencies?: string[]; // 의존하는 다른 필드들
  ui?: FieldUIOptions;
}

export type UIFieldType = 'boolean' | 'string' | 'number' | 'select' | 'multiselect' | 'textarea';

export interface FieldValidation {
  min?: number;
  max?: number;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  enum?: string[];
  custom?: string; // 커스텀 검증 표현식
}

export interface FieldUIOptions {
  placeholder?: string;
  helpText?: string;
  options?: SelectOption[]; // select/multiselect용
  rows?: number; // textarea용
  unit?: string; // 단위 표시 (예: "MB", "초")
}

export interface SelectOption {
  label: string;
  value: string | number;
}

export interface ConfigurationValues {
  [key: string]: any;
}

export interface DependencyRule {
  type: 'requires' | 'conflicts' | 'cascades';
  field: string;
  condition: string;
  target: string;
  targetValue?: any;
  message?: string;
}

export interface ValidationErrors {
  [key: string]: string;
}

// Dependency validation types (from backend API)
export interface DependencyValidationError {
  field: string;
  message: string;
  rule: string;
}

export interface ValidationResult {
  valid: boolean;
  errors: DependencyValidationError[];
}
