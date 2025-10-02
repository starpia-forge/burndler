import React, { useState } from 'react';
import { PlusIcon, TrashIcon, PencilIcon } from '@heroicons/react/24/outline';

export interface TemplateFileData {
  id: string; // Temporary ID for UI
  file_path: string;
  file_type: 'template' | 'static';
  template_format?: 'yaml' | 'json' | 'env' | 'text';
  template_content?: string;
  display_condition?: string;
  description?: string;
}

interface TemplateFilesManagerProps {
  files: TemplateFileData[];
  onChange: (files: TemplateFileData[]) => void;
}

export const TemplateFilesManager: React.FC<TemplateFilesManagerProps> = ({ files, onChange }) => {
  const [editingFile, setEditingFile] = useState<TemplateFileData | null>(null);
  const [isAdding, setIsAdding] = useState(false);

  const handleAddFile = () => {
    const newFile: TemplateFileData = {
      id: `temp-${Date.now()}`,
      file_path: '',
      file_type: 'template',
      template_format: 'yaml',
      template_content: '',
      display_condition: '',
      description: '',
    };
    setEditingFile(newFile);
    setIsAdding(true);
  };

  const handleEditFile = (file: TemplateFileData) => {
    setEditingFile({ ...file });
    setIsAdding(false);
  };

  const handleSaveFile = () => {
    if (!editingFile) return;

    // Validate required fields
    if (!editingFile.file_path.trim()) {
      alert('파일 경로를 입력해주세요');
      return;
    }

    if (editingFile.file_type === 'template' && !editingFile.template_content?.trim()) {
      alert('템플릿 내용을 입력해주세요');
      return;
    }

    if (isAdding) {
      // Add new file
      onChange([...files, editingFile]);
    } else {
      // Update existing file
      onChange(files.map((f) => (f.id === editingFile.id ? editingFile : f)));
    }

    setEditingFile(null);
    setIsAdding(false);
  };

  const handleCancelEdit = () => {
    setEditingFile(null);
    setIsAdding(false);
  };

  const handleDeleteFile = (fileId: string) => {
    if (confirm('이 파일을 삭제하시겠습니까?')) {
      onChange(files.filter((f) => f.id !== fileId));
    }
  };

  const handleFieldChange = (field: keyof TemplateFileData, value: any) => {
    if (!editingFile) return;
    setEditingFile({ ...editingFile, [field]: value });
  };

  return (
    <div className="template-files-manager space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-foreground">템플릿 파일</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Container에 포함될 템플릿 및 정적 파일을 관리합니다
          </p>
        </div>
        <button
          type="button"
          onClick={handleAddFile}
          disabled={!!editingFile}
          className="inline-flex items-center px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
        >
          <PlusIcon className="h-4 w-4 mr-2" />
          파일 추가
        </button>
      </div>

      {/* File List */}
      {files.length === 0 && !editingFile && (
        <div className="text-center py-12 bg-muted border border-border rounded-lg">
          <p className="text-muted-foreground">추가된 파일이 없습니다</p>
          <p className="text-sm text-muted-foreground mt-2">
            "파일 추가" 버튼을 클릭하여 템플릿 파일을 추가하세요
          </p>
        </div>
      )}

      {files.length > 0 && !editingFile && (
        <div className="space-y-3">
          {files.map((file) => (
            <div
              key={file.id}
              className="file-item bg-card border border-border rounded-lg p-4 hover:border-blue-400 transition-colors"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-2">
                    <span className="text-sm font-mono text-foreground">{file.file_path}</span>
                    <span
                      className={`inline-flex items-center px-2 py-1 text-xs font-medium rounded ${
                        file.file_type === 'template'
                          ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300'
                          : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
                      }`}
                    >
                      {file.file_type === 'template'
                        ? `Template (${file.template_format})`
                        : 'Static'}
                    </span>
                  </div>
                  {file.description && (
                    <p className="text-sm text-muted-foreground mt-1">{file.description}</p>
                  )}
                  {file.display_condition && (
                    <p className="text-xs text-muted-foreground mt-1">
                      조건: <code className="bg-muted px-1 rounded">{file.display_condition}</code>
                    </p>
                  )}
                </div>
                <div className="flex items-center space-x-2 ml-4">
                  <button
                    type="button"
                    onClick={() => handleEditFile(file)}
                    className="p-2 text-gray-600 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors"
                    title="편집"
                  >
                    <PencilIcon className="h-4 w-4" />
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDeleteFile(file.id)}
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
      {editingFile && (
        <div className="edit-form bg-card border-2 border-blue-400 rounded-lg p-6 space-y-4">
          <h4 className="text-lg font-semibold text-foreground">
            {isAdding ? '새 파일 추가' : '파일 편집'}
          </h4>

          {/* File Path */}
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              파일 경로 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={editingFile.file_path}
              onChange={(e) => handleFieldChange('file_path', e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
              placeholder="예: config/app.yaml"
            />
            <p className="text-xs text-muted-foreground mt-1">인스톨러 내 상대 경로를 입력하세요</p>
          </div>

          {/* File Type */}
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              파일 타입 <span className="text-red-500">*</span>
            </label>
            <select
              value={editingFile.file_type}
              onChange={(e) =>
                handleFieldChange('file_type', e.target.value as 'template' | 'static')
              }
              className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
            >
              <option value="template">템플릿 (변수 치환)</option>
              <option value="static">정적 파일 (변경 없음)</option>
            </select>
          </div>

          {/* Template Format (only for template files) */}
          {editingFile.file_type === 'template' && (
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                템플릿 포맷 <span className="text-red-500">*</span>
              </label>
              <select
                value={editingFile.template_format}
                onChange={(e) =>
                  handleFieldChange(
                    'template_format',
                    e.target.value as 'yaml' | 'json' | 'env' | 'text'
                  )
                }
                className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
              >
                <option value="yaml">YAML</option>
                <option value="json">JSON</option>
                <option value="env">ENV</option>
                <option value="text">Text</option>
              </select>
            </div>
          )}

          {/* Template Content (only for template files) */}
          {editingFile.file_type === 'template' && (
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                템플릿 내용 <span className="text-red-500">*</span>
              </label>
              <textarea
                value={editingFile.template_content || ''}
                onChange={(e) => handleFieldChange('template_content', e.target.value)}
                rows={10}
                className="w-full px-3 py-2 font-mono text-sm border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
                placeholder={`예: database:\n  host: {{ .Database.Host }}\n  port: {{ .Database.Port | default 5432 }}`}
              />
              <p className="text-xs text-muted-foreground mt-1">
                Go 템플릿 문법을 사용하세요 (예: {'{{ .변수명 }}'})
              </p>
            </div>
          )}

          {/* Display Condition */}
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              표시 조건 (선택사항)
            </label>
            <input
              type="text"
              value={editingFile.display_condition || ''}
              onChange={(e) => handleFieldChange('display_condition', e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
              placeholder="예: .Cache.Enabled == true"
            />
            <p className="text-xs text-muted-foreground mt-1">조건이 참일 때만 파일이 포함됩니다</p>
          </div>

          {/* Description */}
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              설명 (선택사항)
            </label>
            <input
              type="text"
              value={editingFile.description || ''}
              onChange={(e) => handleFieldChange('description', e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-background text-foreground"
              placeholder="파일에 대한 설명을 입력하세요"
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
              onClick={handleSaveFile}
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
