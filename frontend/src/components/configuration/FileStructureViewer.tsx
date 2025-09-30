import React from 'react';
import { FileStructureState } from '../../types/fileStructure';
import { FileTreeNode } from './FileTreeNode';
import { formatFileSize } from '../../utils/format';

interface FileStructureViewerProps {
  structure: FileStructureState;
}

export const FileStructureViewer: React.FC<FileStructureViewerProps> = ({ structure }) => {
  return (
    <div className="file-structure-viewer border border-border rounded-lg p-4 bg-muted">
      <div className="structure-header mb-4">
        <h3 className="text-lg font-semibold text-foreground mb-2">📁 파일 구조 미리보기</h3>
        <div className="structure-stats flex space-x-4 text-sm text-foreground">
          <span className="flex items-center">
            <span className="font-medium mr-1">📄</span>
            {structure.visibleFiles}개 파일
          </span>
          <span className="flex items-center">
            <span className="font-medium mr-1">📦</span>
            {formatFileSize(structure.totalSize)}
          </span>
          {structure.hiddenFiles > 0 && (
            <span className="text-muted-foreground">({structure.hiddenFiles}개 숨김)</span>
          )}
        </div>
      </div>

      <div className="structure-tree bg-card border border-border rounded p-3 max-h-96 overflow-y-auto">
        {structure.rootNodes.length === 0 ? (
          <div className="text-center text-muted-foreground py-8">파일 구조가 없습니다</div>
        ) : (
          structure.rootNodes.map((node) => <FileTreeNode key={node.path} node={node} level={0} />)
        )}
      </div>
    </div>
  );
};
