import React, { useState } from 'react';
import { PlusIcon, TrashIcon, PencilIcon, DocumentIcon } from '@heroicons/react/24/outline';

export interface AssetData {
  id: string; // Temporary ID for UI
  original_file_name: string;
  file_path: string;
  asset_type: 'config' | 'data' | 'script' | 'binary' | 'document';
  storage_type: 'embedded' | 'download';
  file_content?: File | null; // For file upload
  source_url?: string; // For download type
  include_condition?: string;
  description?: string;
}

interface AssetsManagerProps {
  assets: AssetData[];
  onChange: (assets: AssetData[]) => void;
}

export const AssetsManager: React.FC<AssetsManagerProps> = ({ assets, onChange }) => {
  const [editingAsset, setEditingAsset] = useState<AssetData | null>(null);
  const [isAdding, setIsAdding] = useState(false);

  const handleAddAsset = () => {
    const newAsset: AssetData = {
      id: `temp-${Date.now()}`,
      original_file_name: '',
      file_path: '',
      asset_type: 'config',
      storage_type: 'embedded',
      file_content: null,
      source_url: '',
      include_condition: '',
      description: '',
    };
    setEditingAsset(newAsset);
    setIsAdding(true);
  };

  const handleEditAsset = (asset: AssetData) => {
    setEditingAsset({ ...asset });
    setIsAdding(false);
  };

  const handleSaveAsset = () => {
    if (!editingAsset) return;

    // Validate required fields
    if (!editingAsset.file_path.trim()) {
      alert('파일 경로를 입력해주세요');
      return;
    }

    if (editingAsset.storage_type === 'embedded' && !editingAsset.file_content && isAdding) {
      alert('업로드할 파일을 선택해주세요');
      return;
    }

    if (editingAsset.storage_type === 'download' && !editingAsset.source_url?.trim()) {
      alert('다운로드 URL을 입력해주세요');
      return;
    }

    if (isAdding) {
      // Add new asset
      onChange([...assets, editingAsset]);
    } else {
      // Update existing asset
      onChange(assets.map((a) => (a.id === editingAsset.id ? editingAsset : a)));
    }

    setEditingAsset(null);
    setIsAdding(false);
  };

  const handleCancelEdit = () => {
    setEditingAsset(null);
    setIsAdding(false);
  };

  const handleDeleteAsset = (assetId: string) => {
    if (confirm('이 에셋을 삭제하시겠습니까?')) {
      onChange(assets.filter((a) => a.id !== assetId));
    }
  };

  const handleFieldChange = (field: keyof AssetData, value: any) => {
    if (!editingAsset) return;
    setEditingAsset({ ...editingAsset, [field]: value });
  };

  const handleFileUpload = (file: File | null) => {
    if (!editingAsset || !file) return;

    setEditingAsset({
      ...editingAsset,
      file_content: file,
      original_file_name: file.name,
      file_path: editingAsset.file_path || `assets/${file.name}`,
    });
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  return (
    <div className="assets-manager space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-foreground">에셋 파일</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Container에 포함될 바이너리 파일, 데이터 파일 등을 관리합니다
          </p>
        </div>
        <button
          type="button"
          onClick={handleAddAsset}
          disabled={!!editingAsset}
          className="inline-flex items-center px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
        >
          <PlusIcon className="h-4 w-4 mr-2" />
          에셋 추가
        </button>
      </div>

      {/* Asset List */}
      {assets.length === 0 && !editingAsset && (
        <div className="text-center py-12 bg-muted border border-border rounded-lg">
          <DocumentIcon className="h-12 w-12 mx-auto text-muted-foreground" />
          <p className="text-muted-foreground mt-2">추가된 에셋이 없습니다</p>
          <p className="text-sm text-muted-foreground mt-1">
            "에셋 추가" 버튼을 클릭하여 파일을 추가하세요
          </p>
        </div>
      )}

      {assets.length > 0 && !editingAsset && (
        <div className="space-y-3">
          {assets.map((asset) => (
            <div
              key={asset.id}
              className="asset-item bg-card border border-border rounded-lg p-4 hover:border-blue-400 transition-colors"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-2">
                    <span className="text-sm font-mono text-foreground">{asset.file_path}</span>
                    <span className="inline-flex items-center px-2 py-1 text-xs font-medium rounded bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300">
                      {asset.asset_type}
                    </span>
                    <span
                      className={`inline-flex items-center px-2 py-1 text-xs font-medium rounded ${
                        asset.storage_type === 'embedded'
                          ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300'
                          : 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300'
                      }`}
                    >
                      {asset.storage_type === 'embedded' ? '내장' : '다운로드'}
                    </span>
                  </div>
                  {asset.original_file_name && (
                    <p className="text-sm text-muted-foreground mt-1">
                      원본: {asset.original_file_name}
                      {asset.file_content && ` (${formatFileSize(asset.file_content.size)})`}
                    </p>
                  )}
                  {asset.description && (
                    <p className="text-sm text-muted-foreground mt-1">{asset.description}</p>
                  )}
                  {asset.include_condition && (
                    <p className="text-xs text-muted-foreground mt-1">
                      조건: <code className="bg-muted px-1 rounded">{asset.include_condition}</code>
                    </p>
                  )}
                </div>
                <div className="flex items-center space-x-2 ml-4">
                  <button
                    type="button"
                    onClick={() => handleEditAsset(asset)}
                    className="p-2 text-gray-600 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors"
                    title="편집"
                  >
                    <PencilIcon className="h-4 w-4" />
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDeleteAsset(asset.id)}
                    className="p-2 text-gray-600 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors"
                    title="삭제"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Edit Form */}
      {editingAsset && (
        <div className="edit-form bg-card border-2 border-blue-400 rounded-lg p-6 space-y-4">
          <h4 className="text-lg font-semibold text-foreground">
            {isAdding ? '새 에셋 추가' : '에셋 편집'}
          </h4>

          {/* Storage Type */}
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              저장 방식 <span className="text-red-500">*</span>
            </label>
            <select
              value={editingAsset.storage_type}
              onChange={(e) =>
                handleFieldChange('storage_type', e.target.value as 'embedded' | 'download')
              }
              className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
            >
              <option value="embedded">내장 (인스톨러에 포함)</option>
              <option value="download">다운로드 (URL로 제공)</option>
            </select>
            <p className="text-xs text-muted-foreground mt-1">
              {editingAsset.storage_type === 'embedded'
                ? '파일이 인스톨러 패키지에 직접 포함됩니다'
                : '파일이 외부 URL에서 다운로드됩니다 (대용량 파일 권장)'}
            </p>
          </div>

          {/* File Upload (for embedded) */}
          {editingAsset.storage_type === 'embedded' && (
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                파일 업로드 {isAdding && <span className="text-red-500">*</span>}
              </label>
              <input
                type="file"
                onChange={(e) => handleFileUpload(e.target.files?.[0] || null)}
                className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
              />
              {editingAsset.file_content && (
                <p className="text-sm text-muted-foreground mt-1">
                  선택된 파일: {editingAsset.file_content.name} (
                  {formatFileSize(editingAsset.file_content.size)})
                </p>
              )}
            </div>
          )}

          {/* Source URL (for download) */}
          {editingAsset.storage_type === 'download' && (
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                다운로드 URL <span className="text-red-500">*</span>
              </label>
              <input
                type="url"
                value={editingAsset.source_url || ''}
                onChange={(e) => handleFieldChange('source_url', e.target.value)}
                className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
                placeholder="https://example.com/large-file.tar.gz"
              />
            </div>
          )}

          {/* File Path */}
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              파일 경로 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={editingAsset.file_path}
              onChange={(e) => handleFieldChange('file_path', e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
              placeholder="예: assets/large-dataset.tar.gz"
            />
            <p className="text-xs text-muted-foreground mt-1">인스톨러 내 상대 경로를 입력하세요</p>
          </div>

          {/* Asset Type */}
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              에셋 타입 <span className="text-red-500">*</span>
            </label>
            <select
              value={editingAsset.asset_type}
              onChange={(e) =>
                handleFieldChange(
                  'asset_type',
                  e.target.value as 'config' | 'data' | 'script' | 'binary' | 'document'
                )
              }
              className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
            >
              <option value="config">Config</option>
              <option value="data">Data</option>
              <option value="script">Script</option>
              <option value="binary">Binary</option>
              <option value="document">Document</option>
            </select>
          </div>

          {/* Include Condition */}
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              포함 조건 (선택사항)
            </label>
            <input
              type="text"
              value={editingAsset.include_condition || ''}
              onChange={(e) => handleFieldChange('include_condition', e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
              placeholder="예: .AdvancedFeatures.Enabled == true"
            />
            <p className="text-xs text-muted-foreground mt-1">조건이 참일 때만 에셋이 포함됩니다</p>
          </div>

          {/* Description */}
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              설명 (선택사항)
            </label>
            <input
              type="text"
              value={editingAsset.description || ''}
              onChange={(e) => handleFieldChange('description', e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
              placeholder="에셋에 대한 설명을 입력하세요"
            />
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end space-x-3 pt-4 border-t border-border">
            <button
              type="button"
              onClick={handleCancelEdit}
              className="px-4 py-2 border border-border rounded-md text-foreground hover:bg-muted transition-colors"
            >
              취소
            </button>
            <button
              type="button"
              onClick={handleSaveAsset}
              className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
            >
              {isAdding ? '추가' : '저장'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
