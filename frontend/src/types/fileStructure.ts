// File structure visualization type definitions

export interface FileStructureNode {
  name: string;
  path: string;
  type: 'file' | 'directory';
  fileType?: 'template' | 'asset' | 'static';
  condition?: string;
  visible: boolean;
  description?: string;
  size?: number;
  children?: FileStructureNode[];
  isGenerated?: boolean;
}

export interface FileStructureState {
  rootNodes: FileStructureNode[];
  totalFiles: number;
  totalSize: number;
  visibleFiles: number;
  hiddenFiles: number;
}
