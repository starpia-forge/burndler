import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { ConfigurationForm } from '../components/configuration/ConfigurationForm';
import { FileStructureViewer } from '../components/configuration/FileStructureViewer';
import { useFileStructureResolver } from '../hooks/useFileStructureResolver';
import { UISchema, ConfigurationValues, ValidationErrors } from '../types/configuration';
// import api from '../services/api'; // TODO: Uncomment when backend API is ready

export const ServiceConfigurationPage: React.FC = () => {
  const { serviceId, containerId } = useParams<{ serviceId: string; containerId: string }>();
  const navigate = useNavigate();

  const [schema, setSchema] = useState<UISchema | null>(null);
  const [values, setValues] = useState<ConfigurationValues>({});
  const [errors, setErrors] = useState<ValidationErrors>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Resolve file structure based on current values
  const fileStructure = useFileStructureResolver(schema, values);

  // Load configuration schema and current values
  useEffect(() => {
    const loadConfiguration = async () => {
      try {
        setLoading(true);
        setError(null);

        // TODO: Replace with actual API endpoint when backend is ready
        // const response = await api.get(
        //   `/api/v1/services/${serviceId}/containers/${containerId}/configuration`
        // );
        // setSchema(response.data.ui_schema);
        // setValues(response.data.current_values || {});

        // Placeholder: Empty schema for now
        setSchema({ sections: [] });
        setValues({});
      } catch (err: any) {
        console.error('Failed to load configuration:', err);
        setError(err.message || 'Failed to load configuration');
      } finally {
        setLoading(false);
      }
    };

    if (serviceId && containerId) {
      loadConfiguration();
    }
  }, [serviceId, containerId]);

  const handleValuesChange = (newValues: ConfigurationValues) => {
    setValues(newValues);
  };

  const handleValidate = (newErrors: ValidationErrors) => {
    setErrors(newErrors);
  };

  const handleSave = async () => {
    // Check for validation errors
    if (Object.keys(errors).length > 0) {
      alert('설정에 오류가 있습니다. 모든 필드를 올바르게 입력해주세요.');
      return;
    }

    try {
      setSaving(true);

      // TODO: Replace with actual API endpoint when backend is ready
      // await api.put(
      //   `/api/v1/services/${serviceId}/containers/${containerId}/configuration`,
      //   {
      //     configuration_values: values,
      //   }
      // );

      alert('설정이 저장되었습니다');
      // Optionally navigate back
      // navigate(-1);
    } catch (err: any) {
      console.error('Failed to save configuration:', err);
      alert(err.message || '설정 저장에 실패했습니다');
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    navigate(-1);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
          <h3 className="font-semibold mb-2">Error</h3>
          <p>{error}</p>
          <button
            onClick={() => navigate(-1)}
            className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
          >
            Go Back
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="service-configuration-page p-6">
      {/* Page Header */}
      <div className="page-header mb-6">
        <button
          onClick={() => navigate(-1)}
          className="inline-flex items-center text-muted-foreground hover:text-foreground mb-4"
        >
          <ArrowLeftIcon className="h-4 w-4 mr-2" />
          Back
        </button>
        <h1 className="text-2xl font-bold text-foreground">컨테이너 설정</h1>
        <p className="text-muted-foreground mt-2">서비스에 포함될 컨테이너의 설정을 변경합니다</p>
      </div>

      {/* Two-column layout */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Left: Configuration Panel */}
        <div className="config-panel">
          <h2 className="text-xl font-semibold text-foreground mb-4">설정</h2>

          {schema && (
            <div className="bg-card border border-border rounded-lg p-6">
              <ConfigurationForm
                schema={schema}
                initialValues={values}
                onChange={handleValuesChange}
                onValidate={handleValidate}
              />

              {/* Action Buttons */}
              <div className="actions mt-6 flex space-x-3">
                <button
                  onClick={handleSave}
                  disabled={saving}
                  className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed transition-colors"
                >
                  {saving ? '저장 중...' : '저장'}
                </button>
                <button
                  onClick={handleCancel}
                  disabled={saving}
                  className="px-6 py-2 bg-muted text-foreground rounded-md hover:bg-muted/80 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  취소
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Right: Preview Panel */}
        <div className="preview-panel">
          <h2 className="text-xl font-semibold text-foreground mb-4">파일 구조 미리보기</h2>
          <FileStructureViewer structure={fileStructure} />
        </div>
      </div>
    </div>
  );
};
