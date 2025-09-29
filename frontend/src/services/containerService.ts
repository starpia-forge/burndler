import apiClient from './api';
import {
  Container,
  ContainerVersion,
  ContainerListResponse,
  ContainerVersionListResponse,
  CreateContainerRequest,
  UpdateContainerRequest,
  CreateVersionRequest,
  UpdateVersionRequest,
  ContainerFilters,
  VersionFilters,
  ApiError,
} from '../types/container';

class ContainerService {
  private client = apiClient;

  // Container CRUD Operations
  async listContainers(filters: ContainerFilters = {}): Promise<ContainerListResponse> {
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
      const url = queryString ? `/containers?${queryString}` : '/containers';

      return await this.client.get(url);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async getContainer(id: number, includeVersions = false): Promise<Container> {
    try {
      const params = includeVersions ? '?include_versions=true' : '';
      return await this.client.get(`/containers/${id}${params}`);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async createContainer(data: CreateContainerRequest): Promise<Container> {
    try {
      return await this.client.post('/containers', data);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async updateContainer(id: number, data: UpdateContainerRequest): Promise<Container> {
    try {
      return await this.client.put(`/containers/${id}`, data);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async deleteContainer(id: number): Promise<void> {
    try {
      await this.client.delete(`/containers/${id}`);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  // Container Version Operations
  async listVersions(
    containerId: number,
    filters: VersionFilters = {}
  ): Promise<ContainerVersionListResponse> {
    try {
      const params = new URLSearchParams();

      if (filters.published_only)
        params.append('published_only', filters.published_only.toString());

      const queryString = params.toString();
      const url = queryString
        ? `/containers/${containerId}/versions?${queryString}`
        : `/containers/${containerId}/versions`;

      return await this.client.get(url);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async createVersion(containerId: number, data: CreateVersionRequest): Promise<ContainerVersion> {
    try {
      return await this.client.post(`/containers/${containerId}/versions`, data);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async getVersion(containerId: number, version: string): Promise<ContainerVersion> {
    try {
      return await this.client.get(`/containers/${containerId}/versions/${version}`);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async updateVersion(
    containerId: number,
    version: string,
    data: UpdateVersionRequest
  ): Promise<ContainerVersion> {
    try {
      return await this.client.put(`/containers/${containerId}/versions/${version}`, data);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async publishVersion(containerId: number, version: string): Promise<ContainerVersion> {
    try {
      return await this.client.post(`/containers/${containerId}/versions/${version}/publish`);
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

export default new ContainerService();
