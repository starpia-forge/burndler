import axios, { AxiosInstance } from 'axios'
import { MergeRequest, MergeResult, LintRequest, LintResult, PackageRequest, Build } from '../types'

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_API_URL || '/api/v1',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Add auth token to requests
    this.client.interceptors.request.use((config) => {
      const token = localStorage.getItem('burndler_token')
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      return config
    })

    // Handle auth errors
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          localStorage.removeItem('burndler_token')
          window.location.href = '/login'
        }
        return Promise.reject(error)
      }
    )
  }

  // Health check
  async health() {
    const response = await this.client.get('/health')
    return response.data
  }

  // Compose operations
  async mergeCompose(request: MergeRequest): Promise<MergeResult> {
    const response = await this.client.post('/compose/merge', request)
    return response.data
  }

  async lintCompose(request: LintRequest): Promise<LintResult> {
    const response = await this.client.post('/compose/lint', request)
    return response.data
  }

  // Package operations
  async createPackage(request: PackageRequest): Promise<{ build_id: string; status: string }> {
    const response = await this.client.post('/build/package', request)
    return response.data
  }

  async getBuildStatus(buildId: string): Promise<Build> {
    const response = await this.client.get(`/build/status/${buildId}`)
    return response.data
  }
}

export default new ApiClient()