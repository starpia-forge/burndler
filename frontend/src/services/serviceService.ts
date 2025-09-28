import apiClient from './api';
import {
  Service,
  ServiceContainer,
  ServiceListResponse,
  ServiceContainerListResponse,
  CreateServiceRequest,
  UpdateServiceRequest,
  AddContainerToServiceRequest,
  UpdateServiceContainerRequest,
  ServiceFilters,
  ApiError,
} from '../types/service';

class ServiceService {
  private client = apiClient;

  // Service CRUD Operations
  async listServices(filters: ServiceFilters = {}): Promise<ServiceListResponse> {
    try {
      const params = new URLSearchParams();

      if (filters.page) params.append('page', filters.page.toString());
      if (filters.page_size) params.append('page_size', filters.page_size.toString());
      if (filters.active !== undefined) params.append('active', filters.active.toString());
      if (filters.search) params.append('search', filters.search);

      const queryString = params.toString();
      const url = queryString ? `/services?${queryString}` : '/services';

      return await this.client.get(url);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async getService(id: number, includeContainers = false): Promise<Service> {
    try {
      const params = includeContainers ? '?include_containers=true' : '';
      return await this.client.get(`/services/${id}${params}`);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async createService(data: CreateServiceRequest): Promise<Service> {
    try {
      return await this.client.post('/services', data);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async updateService(id: number, data: UpdateServiceRequest): Promise<Service> {
    try {
      return await this.client.put(`/services/${id}`, data);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async deleteService(id: number): Promise<void> {
    try {
      await this.client.delete(`/services/${id}`);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  // Service Container Operations
  async getServiceContainers(serviceId: number): Promise<ServiceContainerListResponse> {
    try {
      return await this.client.get(`/services/${serviceId}/containers`);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async addContainerToService(
    serviceId: number,
    data: AddContainerToServiceRequest
  ): Promise<ServiceContainer> {
    try {
      return await this.client.post(`/services/${serviceId}/containers`, data);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async updateServiceContainer(
    serviceId: number,
    containerId: number,
    data: UpdateServiceContainerRequest
  ): Promise<ServiceContainer> {
    try {
      return await this.client.put(`/services/${serviceId}/containers/${containerId}`, data);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async removeContainerFromService(serviceId: number, containerId: number): Promise<void> {
    try {
      await this.client.delete(`/services/${serviceId}/containers/${containerId}`);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  // Service Operations
  async validateService(serviceId: number): Promise<any> {
    try {
      return await this.client.post(`/services/${serviceId}/validate`);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  async buildService(serviceId: number): Promise<any> {
    try {
      return await this.client.post(`/services/${serviceId}/build`);
    } catch (error: any) {
      throw this.handleError(error);
    }
  }

  // Error handling
  private handleError(error: any): ApiError {
    if (error.response) {
      // Server responded with error status
      return {
        error: error.response.data?.error || 'API_ERROR',
        message: error.response.data?.message || 'An error occurred',
        status: error.response.status,
      };
    } else if (error.request) {
      // Network error
      return {
        error: 'NETWORK_ERROR',
        message: 'Network error - please check your connection',
      };
    } else {
      // Other error
      return {
        error: 'UNKNOWN_ERROR',
        message: error.message || 'An unknown error occurred',
      };
    }
  }
}

// Export singleton instance
const serviceService = new ServiceService();
export default serviceService;