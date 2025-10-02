import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { useAuth } from '../hooks/useAuth';
import ContainerForm from '../components/containers/ContainerForm';
import { ConfigurationSchemaEditor } from '../components/configuration/ConfigurationSchemaEditor';
import {
  TemplateFilesManager,
  TemplateFileData,
} from '../components/configuration/TemplateFilesManager';
import { AssetsManager, AssetData } from '../components/configuration/AssetsManager';
import { useCreateContainerWithConfig } from '../hooks/useCreateContainerWithConfig';
import { CreateContainerRequest } from '../types/container';
import { UISchema, DependencyRule } from '../types/configuration';

type TabType = 'basic' | 'configuration' | 'files' | 'assets';

const CreateContainerPage: React.FC = () => {
  const navigate = useNavigate();
  const { isDeveloper } = useAuth();
  const { t } = useTranslation(['containers', 'common']);
  const { createContainerWithConfig, loading, error, progress } = useCreateContainerWithConfig();

  // Tab state
  const [activeTab, setActiveTab] = useState<TabType>('basic');

  // Form data states
  const [basicData, setBasicData] = useState<CreateContainerRequest | null>(null);
  const [uiSchema, setUISchema] = useState<UISchema | null>(null);
  const [dependencyRules, setDependencyRules] = useState<DependencyRule[] | null>(null);
  const [templateFiles, setTemplateFiles] = useState<TemplateFileData[]>([]);
  const [assets, setAssets] = useState<AssetData[]>([]);

  // Validation states
  const [basicFormValid, setBasicFormValid] = useState(false);

  const handleBasicSubmit = async (data: CreateContainerRequest) => {
    setBasicData(data);
    setBasicFormValid(true);
    // Move to next tab
    setActiveTab('configuration');
  };

  const handleFinalSubmit = async () => {
    if (!basicData) {
      alert('기본 정보를 먼저 입력해주세요');
      setActiveTab('basic');
      return;
    }

    // Create container with all data
    const container = await createContainerWithConfig({
      containerData: basicData,
      uiSchema,
      dependencyRules,
      templateFiles: templateFiles.length > 0 ? templateFiles : undefined,
      assets: assets.length > 0 ? assets : undefined,
    });

    if (container) {
      navigate(`/containers/${container.id}`);
    }
  };

  const handleCancel = () => {
    navigate('/containers');
  };

  const tabs = [
    { id: 'basic' as TabType, label: '기본 정보', badge: basicFormValid ? '✓' : null },
    {
      id: 'configuration' as TabType,
      label: 'Configuration 템플릿',
      badge: uiSchema && dependencyRules ? '✓' : null,
      optional: true,
    },
    {
      id: 'files' as TabType,
      label: '템플릿 파일',
      badge: templateFiles.length > 0 ? templateFiles.length.toString() : null,
      optional: true,
    },
    {
      id: 'assets' as TabType,
      label: '에셋',
      badge: assets.length > 0 ? assets.length.toString() : null,
      optional: true,
    },
  ];

  if (!isDeveloper) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <h3 className="text-lg font-medium text-red-800 dark:text-red-300 mb-2">
              {t('common:accessDenied')}
            </h3>
            <p className="text-red-700 dark:text-red-400">{t('containers:developerRequired')}</p>
            <div className="mt-4">
              <Link
                to="/containers"
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700"
              >
                <ArrowLeftIcon className="h-4 w-4 mr-2" />
                {t('containers:backToContainers')}
              </Link>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <div className="mb-8">
          <nav className="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400">
            <Link to="/containers" className="hover:text-gray-700 dark:hover:text-gray-300">
              {t('containers:title')}
            </Link>
            <span className="mx-2">/</span>
            <span className="text-gray-900 dark:text-white">{t('containers:createContainer')}</span>
          </nav>
        </div>

        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center space-x-4">
            <Link
              to="/containers"
              className="inline-flex items-center text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
            >
              <ArrowLeftIcon className="h-4 w-4 mr-1" />
              {t('containers:backToContainers')}
            </Link>
          </div>
          <h1 className="mt-4 text-2xl font-bold text-gray-900 dark:text-white">
            {t('containers:createNewContainer')}
          </h1>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
            Container를 생성하고 Configuration 템플릿, 파일, 에셋을 설정합니다
          </p>
        </div>

        {/* Progress Indicator */}
        {loading && progress && (
          <div className="mb-6 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-4">
            <div className="flex items-center space-x-3">
              <div className="animate-spin h-5 w-5 border-2 border-blue-600 border-t-transparent rounded-full" />
              <p className="text-blue-700 dark:text-blue-300">{progress}</p>
            </div>
          </div>
        )}

        {/* Error display */}
        {error && (
          <div className="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <p className="text-red-700 dark:text-red-400">{error}</p>
          </div>
        )}

        {/* Tab Navigation */}
        <div className="mb-6 border-b border-gray-200 dark:border-gray-700">
          <nav className="flex space-x-8">
            {tabs.map((tab) => {
              const isActive = activeTab === tab.id;
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  disabled={loading}
                  className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                    isActive
                      ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                      : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300'
                  } ${loading ? 'opacity-50 cursor-not-allowed' : ''}`}
                >
                  <span className="flex items-center space-x-2">
                    <span>{tab.label}</span>
                    {tab.optional && <span className="text-xs text-gray-400">(선택사항)</span>}
                    {tab.badge && (
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                          tab.badge === '✓'
                            ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300'
                            : 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300'
                        }`}
                      >
                        {tab.badge}
                      </span>
                    )}
                  </span>
                </button>
              );
            })}
          </nav>
        </div>

        {/* Tab Content */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
          {/* Tab 1: Basic Info */}
          {activeTab === 'basic' && (
            <div>
              <ContainerForm
                onSubmit={handleBasicSubmit}
                onCancel={handleCancel}
                loading={loading}
                title="Container 기본 정보"
                submitLabel="다음 단계로"
                initialData={
                  basicData
                    ? {
                        id: 0,
                        name: basicData.name,
                        description: basicData.description,
                        author: basicData.author,
                        repository: basicData.repository,
                        active: true,
                        created_at: '',
                        updated_at: '',
                      }
                    : undefined
                }
              />
            </div>
          )}

          {/* Tab 2: Configuration Template */}
          {activeTab === 'configuration' && (
            <div>
              <ConfigurationSchemaEditor
                initialUISchema={uiSchema || undefined}
                initialDependencyRules={dependencyRules || undefined}
                onChange={(schema, rules) => {
                  setUISchema(schema);
                  setDependencyRules(rules);
                }}
              />

              <div className="mt-6 flex items-center justify-end space-x-3 pt-6 border-t border-gray-200 dark:border-gray-700">
                <button
                  type="button"
                  onClick={() => setActiveTab('basic')}
                  disabled={loading}
                  className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50"
                >
                  이전
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab('files')}
                  disabled={loading}
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  다음 단계로
                </button>
              </div>
            </div>
          )}

          {/* Tab 3: Template Files */}
          {activeTab === 'files' && (
            <div>
              <TemplateFilesManager files={templateFiles} onChange={setTemplateFiles} />

              <div className="mt-6 flex items-center justify-end space-x-3 pt-6 border-t border-gray-200 dark:border-gray-700">
                <button
                  type="button"
                  onClick={() => setActiveTab('configuration')}
                  disabled={loading}
                  className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50"
                >
                  이전
                </button>
                <button
                  type="button"
                  onClick={() => setActiveTab('assets')}
                  disabled={loading}
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  다음 단계로
                </button>
              </div>
            </div>
          )}

          {/* Tab 4: Assets */}
          {activeTab === 'assets' && (
            <div>
              <AssetsManager assets={assets} onChange={setAssets} />

              <div className="mt-6 flex items-center justify-end space-x-3 pt-6 border-t border-gray-200 dark:border-gray-700">
                <button
                  type="button"
                  onClick={() => setActiveTab('files')}
                  disabled={loading}
                  className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50"
                >
                  이전
                </button>
                <button
                  type="button"
                  onClick={handleCancel}
                  disabled={loading}
                  className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50"
                >
                  취소
                </button>
                <button
                  type="button"
                  onClick={handleFinalSubmit}
                  disabled={loading || !basicFormValid}
                  className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? 'Container 생성 중...' : 'Container 생성'}
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Help Text */}
        <div className="mt-6 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
          <h4 className="text-sm font-semibold text-blue-900 dark:text-blue-200 mb-2">
            💡 안내사항
          </h4>
          <ul className="text-sm text-blue-800 dark:text-blue-300 space-y-1">
            <li>
              • <strong>기본 정보</strong>는 필수이며, 나머지 탭은 선택사항입니다
            </li>
            <li>
              • <strong>Configuration 템플릿</strong>을 설정하면 Service에서 GUI로 설정을 변경할 수
              있습니다
            </li>
            <li>
              • <strong>템플릿 파일</strong>은 변수를 포함한 설정 파일들입니다 (YAML, JSON, ENV 등)
            </li>
            <li>
              • <strong>에셋</strong>은 바이너리 파일, 대용량 데이터 등입니다
            </li>
            <li>• 모든 설정은 Container의 첫 번째 버전(v0.1.0)에 저장됩니다</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default CreateContainerPage;
