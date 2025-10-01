import React, { useRef } from 'react';
import { ArrowDownTrayIcon, ArrowUpTrayIcon } from '@heroicons/react/24/outline';
import api from '../../services/api';

interface ConfigurationActionsProps {
  serviceId: string;
  onImportComplete?: () => void;
}

export const ConfigurationActions: React.FC<ConfigurationActionsProps> = ({
  serviceId,
  onImportComplete,
}) => {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleExport = async () => {
    try {
      const response = await api.get(`/services/${serviceId}/configuration/export`);

      // Create blob and download
      const blob = new Blob([JSON.stringify(response.data, null, 2)], {
        type: 'application/json',
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `service-${serviceId}-config.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (error: any) {
      console.error('Export failed:', error);
      alert(error.response?.data?.error || '설정 내보내기에 실패했습니다');
    }
  };

  const handleImportClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    try {
      const text = await file.text();
      const data = JSON.parse(text);

      const response = await api.post(`/services/${serviceId}/configuration/import`, data);

      const importedCount = response.data.imported || 0;
      const skipped = response.data.skipped || [];

      let message = `설정을 가져왔습니다 (${importedCount}개 컨테이너)`;
      if (skipped.length > 0) {
        message += `\n건너뛴 컨테이너: ${skipped.join(', ')}`;
      }

      alert(message);

      // Clear file input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }

      // Callback to reload page or refresh data
      if (onImportComplete) {
        onImportComplete();
      }
    } catch (error: any) {
      console.error('Import failed:', error);

      let errorMessage = '설정 가져오기에 실패했습니다';
      if (error instanceof SyntaxError) {
        errorMessage = '유효하지 않은 JSON 파일입니다';
      } else if (error.response?.data?.error) {
        errorMessage = error.response.data.error;
      }

      alert(errorMessage);

      // Clear file input on error
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  return (
    <div className="configuration-actions flex items-center space-x-3">
      <button
        onClick={handleExport}
        className="inline-flex items-center px-4 py-2 border border-border rounded-md text-foreground bg-background hover:bg-muted transition-colors"
        title="설정을 JSON 파일로 내보내기"
      >
        <ArrowDownTrayIcon className="h-4 w-4 mr-2" />
        내보내기
      </button>

      <button
        onClick={handleImportClick}
        className="inline-flex items-center px-4 py-2 border border-border rounded-md text-foreground bg-background hover:bg-muted transition-colors"
        title="JSON 파일에서 설정 가져오기"
      >
        <ArrowUpTrayIcon className="h-4 w-4 mr-2" />
        가져오기
      </button>

      <input
        ref={fileInputRef}
        type="file"
        accept=".json"
        onChange={handleFileChange}
        className="hidden"
        aria-label="설정 파일 선택"
      />
    </div>
  );
};
