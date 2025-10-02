import { useState } from 'react';
import containerService from '../services/containerService';
import api from '../services/api';
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
      if (data.uiSchema && data.dependencyRules) {
        setProgress('Configuration 템플릿 생성 중...');
        await api.post(`/containers/${container.id}/versions/${version.id}/configuration`, {
          ui_schema: data.uiSchema,
          dependency_rules: data.dependencyRules,
        });
      }

      // Step 4: Upload Template Files (if provided)
      if (data.templateFiles && data.templateFiles.length > 0) {
        setProgress(`템플릿 파일 업로드 중 (${data.templateFiles.length}개)...`);

        for (const file of data.templateFiles) {
          await api.post(`/containers/${container.id}/versions/${version.id}/files`, {
            file_path: file.file_path,
            file_type: file.file_type,
            template_format: file.template_format,
            template_content: file.template_content,
            display_condition: file.display_condition,
            description: file.description,
          });
        }
      }

      // Step 5: Upload Assets (if provided)
      if (data.assets && data.assets.length > 0) {
        setProgress(`에셋 업로드 중 (${data.assets.length}개)...`);

        for (const asset of data.assets) {
          if (asset.storage_type === 'embedded' && asset.file_content) {
            // Upload file as multipart/form-data
            const formData = new FormData();
            formData.append('file', asset.file_content);
            formData.append('file_path', asset.file_path);
            formData.append('asset_type', asset.asset_type);
            formData.append('storage_type', asset.storage_type);
            if (asset.include_condition)
              formData.append('include_condition', asset.include_condition);
            if (asset.description) formData.append('description', asset.description);

            await api.post(`/containers/${container.id}/versions/${version.id}/assets`, formData, {
              headers: {
                'Content-Type': 'multipart/form-data',
              },
            });
          } else if (asset.storage_type === 'download') {
            // For download type, just send metadata
            await api.post(`/containers/${container.id}/versions/${version.id}/assets`, {
              original_file_name: asset.original_file_name,
              file_path: asset.file_path,
              asset_type: asset.asset_type,
              storage_type: asset.storage_type,
              source_url: asset.source_url,
              include_condition: asset.include_condition,
              description: asset.description,
            });
          }
        }
      }

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
