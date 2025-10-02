import { useState } from 'react';
import containerService from '../services/containerService';
import { createContainerConfiguration } from '../services/configurationService';
import { CreateContainerRequest, Container } from '../types/container';
import { UISchema, DependencyRule } from '../types/configuration';
import { TemplateFileData } from '../components/configuration/TemplateFilesManager';
import { AssetData } from '../components/configuration/AssetsManager';

export interface CreateContainerWithConfigData {
  // Basic container info
  containerData: CreateContainerRequest;

  // Configuration template (optional)
  uiSchema?: UISchema | null;
  dependencyRules?: DependencyRule[] | null;

  // Template files (optional)
  templateFiles?: TemplateFileData[];

  // Assets (optional)
  assets?: AssetData[];
}

export const useCreateContainerWithConfig = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [progress, setProgress] = useState<string>('');

  const createContainerWithConfig = async (
    data: CreateContainerWithConfigData
  ): Promise<Container | null> => {
    try {
      setLoading(true);
      setError(null);

      // Step 1: Create Container
      setProgress('Container 생성 중...');
      const container = await containerService.createContainer(data.containerData);

      // Step 2: Create first version (v0.1.0)
      setProgress('첫 번째 버전 생성 중...');
      const version = await containerService.createVersion(container.id, {
        version: 'v0.1.0',
        changelog: 'Initial version',
      });

      // Step 3: Create Configuration (if provided)
      // Using Container-level configuration API (Phase 6)
      if (data.uiSchema && data.dependencyRules) {
        setProgress('Configuration 템플릿 생성 중...');
        await createContainerConfiguration(container.id.toString(), {
          name: 'default',
          minimum_version: version.version,
          ui_schema: data.uiSchema,
          dependency_rules: data.dependencyRules,
        });
      }

      // TODO: File and Asset upload endpoints need to be implemented in backend
      // The model migration changed FK from ContainerVersionID to ContainerConfigurationID,
      // but the actual file/asset upload handlers haven't been created yet.
      // This will be implemented in a future phase.

      // Step 4: Upload Template Files (if provided) - DISABLED until backend endpoints are ready
      // if (data.templateFiles && data.templateFiles.length > 0) {
      //   setProgress(`템플릿 파일 업로드 중 (${data.templateFiles.length}개)...`);
      //   // Need: POST /containers/:id/configurations/:name/files
      // }

      // Step 5: Upload Assets (if provided) - DISABLED until backend endpoints are ready
      // if (data.assets && data.assets.length > 0) {
      //   setProgress(`에셋 업로드 중 (${data.assets.length}개)...`);
      //   // Need: POST /containers/:id/configurations/:name/assets
      // }

      setProgress('완료!');
      setLoading(false);

      return container;
    } catch (err: any) {
      console.error('Failed to create container with config:', err);
      const errorMessage =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        'Container 생성에 실패했습니다';
      setError(errorMessage);
      setLoading(false);
      return null;
    }
  };

  return {
    createContainerWithConfig,
    loading,
    error,
    progress,
  };
};
