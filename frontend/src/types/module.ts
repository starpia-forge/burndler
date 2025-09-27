// Module management type definitions

export interface Module {
  id: number;
  name: string;
  description: string;
  author: string;
  repository: string;
  active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
  versions?: ModuleVersion[];
}

export interface ModuleVersion {
  id: number;
  module_id: number;
  version: string;
  compose_content: string;
  variables: Record<string, any>;
  resource_paths: string[];
  dependencies: Record<string, string>;
  published: boolean;
  published_at?: string;
  created_at: string;
  updated_at: string;
  module?: Module;
}

// API Request Types
export interface CreateModuleRequest {
  name: string;
  description?: string;
  author?: string;
  repository?: string;
}

export interface UpdateModuleRequest {
  description?: string;
  author?: string;
  repository?: string;
  active?: boolean;
}

export interface CreateVersionRequest {
  version: string;
  compose: string;
  variables?: Record<string, any>;
  resource_paths?: string[];
  dependencies?: Record<string, string>;
}

export interface UpdateVersionRequest {
  compose?: string;
  variables?: Record<string, any>;
  resource_paths?: string[];
  dependencies?: Record<string, string>;
}

// API Response Types
export interface ModuleListResponse {
  data: Module[];
  pagination: PaginationInfo;
}

export interface ModuleVersionListResponse {
  data: ModuleVersion[];
}

export interface PaginationInfo {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

// Query Parameters
export interface ModuleFilters {
  page?: number;
  page_size?: number;
  active?: boolean;
  author?: string;
  show_deleted?: boolean;
  published_only?: boolean;
  search?: string;
}

export interface VersionFilters {
  published_only?: boolean;
}

// UI State Types
export interface ModuleListState {
  modules: Module[];
  loading: boolean;
  error: string | null;
  pagination: PaginationInfo | null;
  filters: ModuleFilters;
}

export interface ModuleVersionState {
  versions: ModuleVersion[];
  loading: boolean;
  error: string | null;
}

// Form State Types
export interface ModuleFormState {
  name: string;
  description: string;
  author: string;
  repository: string;
  active: boolean;
}

export interface VersionFormState {
  version: string;
  compose: string;
  variables: Record<string, any>;
  resource_paths: string[];
  dependencies: Record<string, string>;
}

// Error Types
export interface ApiError {
  error: string;
  message: string;
  status?: number;
}

// Status Enums
export enum ModuleStatus {
  Active = 'active',
  Inactive = 'inactive',
  Deleted = 'deleted',
}

export enum VersionStatus {
  Draft = 'draft',
  Published = 'published',
}

// Sort Options
export interface SortOption {
  key: string;
  label: string;
  direction: 'asc' | 'desc';
}

export const MODULE_SORT_OPTIONS: SortOption[] = [
  { key: 'name', label: 'Name', direction: 'asc' },
  { key: 'created_at', label: 'Created Date', direction: 'desc' },
  { key: 'updated_at', label: 'Updated Date', direction: 'desc' },
  { key: 'author', label: 'Author', direction: 'asc' },
];

export const VERSION_SORT_OPTIONS: SortOption[] = [
  { key: 'version', label: 'Version', direction: 'desc' },
  { key: 'created_at', label: 'Created Date', direction: 'desc' },
  { key: 'published_at', label: 'Published Date', direction: 'desc' },
];
