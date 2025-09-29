import {
  SetupStatus,
  AdminCreateRequest,
  AdminCreateResponse,
  SetupCompleteRequest,
  SetupError,
} from '../types/setup';

const API_BASE_URL = '/api/v1';

// Error types for better error handling
export enum ErrorType {
  NETWORK_ERROR = 'NETWORK_ERROR',
  API_ERROR = 'API_ERROR',
  PARSE_ERROR = 'PARSE_ERROR',
  BACKEND_DOWN = 'BACKEND_DOWN',
}

export interface ApiError extends Error {
  type: ErrorType;
  details?: any;
}

/**
 * Safely parse JSON response, handling empty/invalid responses
 */
async function safeJsonParse(response: Response): Promise<any> {
  const contentType = response.headers.get('content-type');

  // Check if response has content
  if (!contentType || !contentType.includes('application/json')) {
    // If no content-type or not JSON, check if there's any content
    const text = await response.text();
    if (!text || text.trim() === '') {
      throw createApiError(
        ErrorType.BACKEND_DOWN,
        'Backend server returned empty response (server may be down)',
        { status: response.status, statusText: response.statusText }
      );
    }

    // Try to parse as JSON anyway
    try {
      return JSON.parse(text);
    } catch {
      throw createApiError(ErrorType.PARSE_ERROR, 'Invalid JSON response from backend', {
        responseText: text,
        contentType,
      });
    }
  }

  // Content-Type indicates JSON, try to parse
  try {
    const text = await response.text();
    if (!text || text.trim() === '') {
      throw createApiError(ErrorType.BACKEND_DOWN, 'Backend server returned empty JSON response', {
        status: response.status,
      });
    }
    return JSON.parse(text);
  } catch (error) {
    if (error instanceof SyntaxError) {
      throw createApiError(
        ErrorType.PARSE_ERROR,
        'Unexpected end of JSON input - backend may have returned invalid response',
        { originalError: error.message }
      );
    }
    throw error;
  }
}

/**
 * Create a typed API error with additional context
 */
function createApiError(type: ErrorType, message: string, details?: any): ApiError {
  const error = new Error(message) as ApiError;
  error.type = type;
  error.details = details;
  return error;
}

/**
 * Enhanced fetch wrapper with better error handling
 */
async function apiRequest(url: string, options: RequestInit = {}): Promise<any> {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000); // 10 second timeout

    const response = await fetch(url, {
      ...options,
      signal: controller.signal,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      let errorData: SetupError;

      try {
        errorData = await safeJsonParse(response);
      } catch (parseError) {
        // If we can't parse the error response, create a generic one
        throw createApiError(
          ErrorType.API_ERROR,
          `HTTP ${response.status}: ${response.statusText}`,
          { status: response.status, statusText: response.statusText }
        );
      }

      throw createApiError(
        ErrorType.API_ERROR,
        errorData.message || `HTTP ${response.status}: ${response.statusText}`,
        { status: response.status, errorData }
      );
    }

    return safeJsonParse(response);
  } catch (error) {
    // Handle network errors and timeouts
    if (error instanceof Error) {
      if (error.name === 'AbortError') {
        throw createApiError(
          ErrorType.NETWORK_ERROR,
          'Request timeout - backend server may be down',
          { timeout: true }
        );
      }

      if (error.message.includes('fetch')) {
        throw createApiError(
          ErrorType.BACKEND_DOWN,
          'Cannot connect to backend server (server may not be running)',
          { originalError: error.message }
        );
      }
    }

    // If it's already an ApiError, re-throw as is
    if (error && typeof error === 'object' && 'type' in error) {
      throw error;
    }

    // Wrap unknown errors
    throw createApiError(
      ErrorType.NETWORK_ERROR,
      error instanceof Error ? error.message : 'Unknown network error',
      { originalError: error }
    );
  }
}

class SetupService {
  /**
   * Check the current setup status
   */
  async getStatus(): Promise<SetupStatus> {
    return apiRequest(`${API_BASE_URL}/setup/status`, {
      method: 'GET',
    });
  }

  /**
   * Initialize setup process (if needed)
   */
  async initialize(): Promise<{ message: string }> {
    return apiRequest(`${API_BASE_URL}/setup/init`, {
      method: 'POST',
    });
  }

  /**
   * Create initial admin user
   */
  async createAdmin(request: AdminCreateRequest): Promise<AdminCreateResponse> {
    return apiRequest(`${API_BASE_URL}/setup/admin`, {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  /**
   * Complete setup with system configuration
   */
  async complete(request: SetupCompleteRequest): Promise<{ message: string }> {
    return apiRequest(`${API_BASE_URL}/setup/complete`, {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  /**
   * Check if setup is completed
   */
  async isSetupCompleted(): Promise<boolean> {
    try {
      const status = await this.getStatus();
      return status.is_completed;
    } catch (error) {
      // If we can't get setup status, assume setup is not completed
      console.error('Error checking setup status:', error);
      return false;
    }
  }

  /**
   * Check if setup is required
   */
  async isSetupRequired(): Promise<boolean> {
    try {
      const status = await this.getStatus();
      return status.requires_setup;
    } catch (error) {
      // If we can't get setup status, assume setup is required
      console.error('Error checking setup requirements:', error);
      return true;
    }
  }

  /**
   * Get setup token from status response
   */
  async getSetupToken(): Promise<string | null> {
    try {
      const status = await this.getStatus();
      return status.setup_token || null;
    } catch (error) {
      console.error('Error getting setup token:', error);
      return null;
    }
  }
}

export const setupService = new SetupService();
