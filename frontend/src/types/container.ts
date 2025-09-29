// Container management type definitions

export interface Container {
  id: number;
  name: string;
  description: string;
  author: string;
  repository: string;
  active: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
  versions?: ContainerVersion[];
}

export interface ContainerVersion {
  id: number;
  container_id: number;
  version: string;
  compose_content: string;
  variables: Record<string, any>;
  resource_paths: string[];
  dependencies: Record<string, string>;
  published: boolean;
  published_at?: string;
  created_at: string;
  updated_at: string;
  container?: Container;
}

// API Request Types
export interface CreateContainerRequest {
  name: string;
  description?: string;
  author?: string;
  repository?: string;
}

export interface UpdateContainerRequest {
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
export interface ContainerListResponse {
  data: Container[];
  pagination: PaginationInfo;
}

export interface ContainerVersionListResponse {
  data: ContainerVersion[];
}

export interface PaginationInfo {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

// Query Parameters
export interface ContainerFilters {
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
export interface ContainerListState {
  containers: Container[];
  loading: boolean;
  initialLoading: boolean;
  isRefreshing: boolean;
  error: string | null;
  pagination: PaginationInfo | null;
  filters: ContainerFilters;
}

export interface ContainerVersionState {
  versions: ContainerVersion[];
  loading: boolean;
  error: string | null;
}

// Form State Types
export interface ContainerFormState {
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
export enum ContainerStatus {
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

export const CONTAINER_SORT_OPTIONS: SortOption[] = [
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
