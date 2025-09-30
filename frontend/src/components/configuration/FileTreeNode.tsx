import React, { useState } from 'react';
import { FileStructureNode } from '../../types/fileStructure';
import { formatFileSize } from '../../utils/format';

interface FileTreeNodeProps {
  node: FileStructureNode;
  level: number;
}

export const FileTreeNode: React.FC<FileTreeNodeProps> = ({ node, level }) => {
  const [expanded, setExpanded] = useState(true);

  const getNodeIcon = () => {
    if (node.type === 'directory') {
      return expanded ? '📁' : '📂';
    }

    switch (node.fileType) {
      case 'template':
        return '📝';
      case 'asset':
        return '🗂️';
      case 'static':
        return '📄';
      default:
        return '📄';
    }
  };

  const getStatusClass = () => {
    if (!node.condition) return 'text-foreground';
    return node.visible ? 'text-green-700' : 'text-muted-foreground line-through';
  };

  return (
    <div className="file-node" style={{ marginLeft: `${level * 20}px` }}>
      <div
        className={`node-content flex items-center py-1 px-2 hover:bg-muted rounded cursor-pointer ${getStatusClass()}`}
        onClick={() => node.type === 'directory' && setExpanded(!expanded)}
      >
        <span className="node-icon mr-2">{getNodeIcon()}</span>
        <span className="node-name flex-1 text-sm">{node.name}</span>

        {node.size && (
          <span className="node-size text-xs text-muted-foreground mr-2">
            {formatFileSize(node.size)}
          </span>
        )}

        {node.condition && (
          <span
            className={`condition-badge text-xs px-2 py-1 rounded ${
              node.visible ? 'bg-green-100 text-green-700' : 'bg-muted text-muted-foreground'
            }`}
          >
            {node.visible ? '✅' : '❌'} 조건부
          </span>
        )}

        {node.isGenerated && (
          <span className="generated-badge text-xs px-2 py-1 rounded bg-blue-100 text-blue-700 ml-2">
            🔄 생성됨
          </span>
        )}
      </div>

      {node.description && (
        <div className="node-description text-xs text-muted-foreground ml-8 mb-1">
          {node.description}
        </div>
      )}

      {node.condition && (
        <div className="node-condition text-xs text-muted-foreground ml-8 mb-1">
          조건: <code className="bg-muted px-1 rounded">{node.condition}</code>
        </div>
      )}

      {node.type === 'directory' && expanded && node.children && (
        <div className="node-children">
          {node.children.map((child) => (
            <FileTreeNode key={child.path} node={child} level={level + 1} />
          ))}
        </div>
      )}
    </div>
  );
};
