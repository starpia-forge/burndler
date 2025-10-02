import React, { useState, useEffect } from 'react';
import { UISchema, DependencyRule } from '../../types/configuration';

interface ConfigurationSchemaEditorProps {
  initialUISchema?: UISchema;
  initialDependencyRules?: DependencyRule[];
  onChange: (uiSchema: UISchema | null, dependencyRules: DependencyRule[] | null) => void;
}

const DEFAULT_UI_SCHEMA: UISchema = {
  sections: [
    {
      id: 'example',
      title: '예제 섹션',
      description: '설정 섹션 설명',
      fields: [
        {
          key: 'example.field',
          type: 'string',
          label: '예제 필드',
          description: '이것은 예제 필드입니다',
          required: false,
        },
      ],
    },
  ],
};

const DEFAULT_DEPENDENCY_RULES: DependencyRule[] = [];

export const ConfigurationSchemaEditor: React.FC<ConfigurationSchemaEditorProps> = ({
  initialUISchema,
  initialDependencyRules,
  onChange,
}) => {
  const [uiSchemaText, setUISchemaText] = useState('');
  const [dependencyRulesText, setDependencyRulesText] = useState('');
  const [uiSchemaError, setUISchemaError] = useState<string | null>(null);
  const [dependencyRulesError, setDependencyRulesError] = useState<string | null>(null);

  // Initialize with provided values or defaults
  useEffect(() => {
    const initialSchema = initialUISchema || DEFAULT_UI_SCHEMA;
    const initialRules = initialDependencyRules || DEFAULT_DEPENDENCY_RULES;

    setUISchemaText(JSON.stringify(initialSchema, null, 2));
    setDependencyRulesText(JSON.stringify(initialRules, null, 2));
  }, [initialUISchema, initialDependencyRules]);

  const validateAndNotify = (schemaText: string, rulesText: string) => {
    let schema: UISchema | null = null;
    let rules: DependencyRule[] | null = null;

    // Validate UI Schema
    try {
      schema = JSON.parse(schemaText);
      setUISchemaError(null);

      // Basic structure validation
      if (!schema || !schema.sections || !Array.isArray(schema.sections)) {
        throw new Error('UI Schema must have a "sections" array');
      }
    } catch (err: any) {
      setUISchemaError(err.message);
      schema = null;
    }

    // Validate Dependency Rules
    try {
      rules = JSON.parse(rulesText);
      setDependencyRulesError(null);

      // Basic structure validation
      if (!Array.isArray(rules)) {
        throw new Error('Dependency Rules must be an array');
      }
    } catch (err: any) {
      setDependencyRulesError(err.message);
      rules = null;
    }

    // Notify parent
    onChange(schema, rules);
  };

  const handleUISchemaChange = (value: string) => {
    setUISchemaText(value);
    validateAndNotify(value, dependencyRulesText);
  };

  const handleDependencyRulesChange = (value: string) => {
    setDependencyRulesText(value);
    validateAndNotify(uiSchemaText, value);
  };

  const loadExample = () => {
    const exampleSchema: UISchema = {
      sections: [
        {
          id: 'database',
          title: '데이터베이스 설정',
          description: '데이터베이스 연결 정보를 설정합니다',
          fields: [
            {
              key: 'Database.Host',
              type: 'string',
              label: '호스트',
              description: '데이터베이스 호스트 주소',
              required: true,
              defaultValue: 'localhost',
            },
            {
              key: 'Database.Port',
              type: 'number',
              label: '포트',
              defaultValue: 5432,
              validation: {
                min: 1,
                max: 65535,
              },
            },
            {
              key: 'Database.SSL',
              type: 'boolean',
              label: 'SSL 사용',
              defaultValue: false,
            },
          ],
        },
      ],
    };

    const exampleRules: DependencyRule[] = [
      {
        type: 'requires',
        field: 'Database.SSL',
        condition: 'Database.SSL === true',
        target: 'Database.CertPath',
        message: 'SSL이 활성화되면 인증서 경로가 필요합니다',
      },
    ];

    setUISchemaText(JSON.stringify(exampleSchema, null, 2));
    setDependencyRulesText(JSON.stringify(exampleRules, null, 2));
    validateAndNotify(
      JSON.stringify(exampleSchema, null, 2),
      JSON.stringify(exampleRules, null, 2)
    );
  };

  return (
    <div className="configuration-schema-editor space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-foreground">Configuration 템플릿</h3>
          <p className="text-sm text-muted-foreground mt-1">
            UI 스키마와 의존성 규칙을 JSON 형식으로 정의합니다
          </p>
        </div>
        <button
          type="button"
          onClick={loadExample}
          className="px-4 py-2 text-sm border border-border rounded-md hover:bg-muted transition-colors"
        >
          예제 불러오기
        </button>
      </div>

      {/* UI Schema Editor */}
      <div className="schema-section">
        <label className="block text-sm font-medium text-foreground mb-2">
          UI Schema <span className="text-red-500">*</span>
        </label>
        <textarea
          value={uiSchemaText}
          onChange={(e) => handleUISchemaChange(e.target.value)}
          className="w-full h-64 px-3 py-2 font-mono text-sm border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
          placeholder="UI Schema JSON을 입력하세요..."
        />
        {uiSchemaError && (
          <p className="mt-2 text-sm text-red-600 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded p-2">
            ❌ {uiSchemaError}
          </p>
        )}
        {!uiSchemaError && uiSchemaText && (
          <p className="mt-2 text-sm text-green-600 dark:text-green-400">✅ 유효한 UI Schema</p>
        )}
      </div>

      {/* Dependency Rules Editor */}
      <div className="rules-section">
        <label className="block text-sm font-medium text-foreground mb-2">
          Dependency Rules (선택사항)
        </label>
        <textarea
          value={dependencyRulesText}
          onChange={(e) => handleDependencyRulesChange(e.target.value)}
          className="w-full h-48 px-3 py-2 font-mono text-sm border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
          placeholder="Dependency Rules JSON을 입력하세요..."
        />
        {dependencyRulesError && (
          <p className="mt-2 text-sm text-red-600 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded p-2">
            ❌ {dependencyRulesError}
          </p>
        )}
        {!dependencyRulesError && dependencyRulesText && (
          <p className="mt-2 text-sm text-green-600 dark:text-green-400">
            ✅ 유효한 Dependency Rules
          </p>
        )}
      </div>

      {/* Help Text */}
      <div className="help-section bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
        <h4 className="text-sm font-semibold text-blue-900 dark:text-blue-200 mb-2">💡 도움말</h4>
        <ul className="text-sm text-blue-800 dark:text-blue-300 space-y-1 list-disc list-inside">
          <li>UI Schema는 사용자가 입력할 필드들을 정의합니다</li>
          <li>Dependency Rules는 필드 간 의존성과 검증 규칙을 정의합니다</li>
          <li>"예제 불러오기"를 클릭하여 샘플 템플릿을 확인하세요</li>
          <li>JSON 형식이 올바른지 자동으로 검증됩니다</li>
        </ul>
      </div>
    </div>
  );
};
