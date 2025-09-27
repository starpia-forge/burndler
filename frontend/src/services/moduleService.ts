import apiClient from './api';
import {
  Module,
  ModuleVersion,
  ModuleListResponse,
  ModuleVersionListResponse,
  CreateModuleRequest,
  UpdateModuleRequest,
  CreateVersionRequest,
  UpdateVersionRequest,
  ModuleFilters,
  VersionFilters,
  ApiError,
} from '../types/module';

class ModuleService {
  private client = apiClient;

  // Module CRUD Operations
  async listModules(filters: ModuleFilters = {}): Promise<ModuleListResponse> {
    try {
      const params = new URLSearchParams();

      if (filters.page) params.append('page', filters.page.toString());
      if (filters.page_size) params.append('page_size', filters.page_size.toString());
      if (filters.active !== undefined) params.append('active', filters.active.toString());
      if (filters.author) params.append('author', filters.author);
      if (filters.show_deleted) params.append('show_deleted', filters.show_deleted.toString());
      if (filters.published_only)
        params.append('published_only', filters.published_only.toString());

      const queryString = params.toString();
      const url = queryString ? `/modules?${queryString}` : '/modules';

      const response = await this.client.client.get(url);
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async getModule(id: number, includeVersions = false): Promise<Module> {
    try {
      const params = includeVersions ? '?include_versions=true' : '';
      const response = await this.client.client.get(`/modules/${id}${params}`);
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async createModule(data: CreateModuleRequest): Promise<Module> {
    try {
      const response = await this.client.client.post('/modules', data);
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async updateModule(id: number, data: UpdateModuleRequest): Promise<Module> {
    try {
      const response = await this.client.client.put(`/modules/${id}`, data);
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async deleteModule(id: number): Promise<void> {
    try {
      await this.client.client.delete(`/modules/${id}`);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  // Module Version Operations
  async listVersions(
    moduleId: number,
    filters: VersionFilters = {}
  ): Promise<ModuleVersionListResponse> {
    try {
      const params = new URLSearchParams();

      if (filters.published_only)
        params.append('published_only', filters.published_only.toString());

      const queryString = params.toString();
      const url = queryString
        ? `/modules/${moduleId}/versions?${queryString}`
        : `/modules/${moduleId}/versions`;

      const response = await this.client.client.get(url);
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async createVersion(moduleId: number, data: CreateVersionRequest): Promise<ModuleVersion> {
    try {
      const response = await this.client.client.post(`/modules/${moduleId}/versions`, data);
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async getVersion(moduleId: number, version: string): Promise<ModuleVersion> {
    try {
      const response = await this.client.client.get(`/modules/${moduleId}/versions/${version}`);
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async updateVersion(
    moduleId: number,
    version: string,
    data: UpdateVersionRequest
  ): Promise<ModuleVersion> {
    try {
      const response = await this.client.client.put(
        `/modules/${moduleId}/versions/${version}`,
        data
      );
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async publishVersion(moduleId: number, version: string): Promise<ModuleVersion> {
    try {
      const response = await this.client.client.post(
        `/modules/${moduleId}/versions/${version}/publish`
      );
      return response.data;
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  // Utility Methods
  validateSemVer(version: string): boolean {
    // Simple semantic version validation
    const semverRegex =
      /^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$/;
    return semverRegex.test(version);
  }

  formatVersion(version: string): string {
    // Ensure version starts with 'v'
    return version.startsWith('v') ? version : `v${version}`;
  }

  private handleError(error: any): ApiError {
    if (error.response?.data) {
      return {
        error: error.response.data.error || 'UNKNOWN_ERROR',
        message: error.response.data.message || 'An unknown error occurred',
        status: error.response.status,
      };
    }

    if (error.message) {
      return {
        error: 'NETWORK_ERROR',
        message: error.message,
      };
    }

    return {
      error: 'UNKNOWN_ERROR',
      message: 'An unknown error occurred',
    };
  }
}

export default new ModuleService();
