import {
  SetupStatus,
  AdminCreateRequest,
  AdminCreateResponse,
  SetupCompleteRequest,
  SetupError,
} from '../types/setup';

const API_BASE_URL = '/api/v1';

class SetupService {
  /**
   * Check the current setup status
   */
  async getStatus(): Promise<SetupStatus> {
    const response = await fetch(`${API_BASE_URL}/setup/status`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const error: SetupError = await response.json();
      throw new Error(error.message || 'Failed to get setup status');
    }

    return response.json();
  }

  /**
   * Initialize setup process (if needed)
   */
  async initialize(): Promise<{ message: string }> {
    const response = await fetch(`${API_BASE_URL}/setup/init`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const error: SetupError = await response.json();
      throw new Error(error.message || 'Failed to initialize setup');
    }

    return response.json();
  }

  /**
   * Create initial admin user
   */
  async createAdmin(request: AdminCreateRequest): Promise<AdminCreateResponse> {
    const response = await fetch(`${API_BASE_URL}/setup/admin`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error: SetupError = await response.json();
      throw new Error(error.message || 'Failed to create admin user');
    }

    return response.json();
  }

  /**
   * Complete setup with system configuration
   */
  async complete(request: SetupCompleteRequest): Promise<{ message: string }> {
    const response = await fetch(`${API_BASE_URL}/setup/complete`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error: SetupError = await response.json();
      throw new Error(error.message || 'Failed to complete setup');
    }

    return response.json();
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
