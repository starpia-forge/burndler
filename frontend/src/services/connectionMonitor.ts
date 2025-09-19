export interface ConnectionStatus {
  isConnected: boolean;
  lastChecked: Date;
  error?: string;
  retryCount: number;
}

export interface ConnectionMonitorConfig {
  healthCheckUrl: string;
  checkInterval: number;
  maxRetries: number;
  retryDelays: number[];
}

class ConnectionMonitor {
  private config: ConnectionMonitorConfig;
  private status: ConnectionStatus;
  private intervalId: NodeJS.Timeout | null = null;
  private timeoutId: NodeJS.Timeout | null = null;
  private listeners: Set<(status: ConnectionStatus) => void> = new Set();

  constructor(config: Partial<ConnectionMonitorConfig> = {}) {
    this.config = {
      healthCheckUrl: '/api/v1/health',
      checkInterval: 10000, // 10 seconds
      maxRetries: 5,
      retryDelays: [1000, 2000, 4000, 8000, 16000], // exponential backoff
      ...config,
    };

    this.status = {
      isConnected: false,
      lastChecked: new Date(),
      retryCount: 0,
    };
  }

  /**
   * Start monitoring backend connection
   */
  start(): void {
    if (this.intervalId) {
      return; // Already monitoring
    }

    // Initial check
    this.checkConnection();

    // Set up periodic checks
    this.intervalId = setInterval(() => {
      this.checkConnection();
    }, this.config.checkInterval);
  }

  /**
   * Stop monitoring backend connection
   */
  stop(): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }

    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }
  }

  /**
   * Check backend connection once
   */
  async checkConnection(): Promise<ConnectionStatus> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout

      const response = await fetch(this.config.healthCheckUrl, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      const wasConnected = this.status.isConnected;

      if (response.ok) {
        this.status = {
          isConnected: true,
          lastChecked: new Date(),
          retryCount: 0,
        };

        // If we just reconnected, notify listeners
        if (!wasConnected) {
          this.notifyListeners();
        }
      } else {
        this.handleConnectionError(`HTTP ${response.status}: ${response.statusText}`);
      }
    } catch (error) {
      let errorMessage = 'Connection failed';

      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          errorMessage = 'Request timeout';
        } else {
          errorMessage = error.message;
        }
      }

      this.handleConnectionError(errorMessage);
    }

    return this.status;
  }

  /**
   * Handle connection errors with retry logic
   */
  private handleConnectionError(errorMessage: string): void {
    const wasConnected = this.status.isConnected;

    this.status = {
      isConnected: false,
      lastChecked: new Date(),
      error: errorMessage,
      retryCount: this.status.retryCount + 1,
    };

    // If we just lost connection or error changed, notify listeners
    if (wasConnected || this.status.error !== errorMessage) {
      this.notifyListeners();
    }

    // Schedule retry if we haven't exceeded max retries
    if (this.status.retryCount <= this.config.maxRetries && !this.timeoutId) {
      const retryDelayIndex = Math.min(
        this.status.retryCount - 1,
        this.config.retryDelays.length - 1
      );
      const delay = this.config.retryDelays[retryDelayIndex];

      this.timeoutId = setTimeout(() => {
        this.timeoutId = null;
        this.checkConnection();
      }, delay);
    }
  }

  /**
   * Get current connection status
   */
  getStatus(): ConnectionStatus {
    return { ...this.status };
  }

  /**
   * Subscribe to connection status changes
   */
  subscribe(listener: (status: ConnectionStatus) => void): () => void {
    this.listeners.add(listener);

    // Return unsubscribe function
    return () => {
      this.listeners.delete(listener);
    };
  }

  /**
   * Notify all listeners of status changes
   */
  private notifyListeners(): void {
    this.listeners.forEach((listener) => {
      try {
        listener(this.status);
      } catch (error) {
        console.error('Error in connection status listener:', error);
      }
    });
  }

  /**
   * Reset retry count and attempt immediate connection check
   */
  retry(): void {
    this.status.retryCount = 0;

    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }

    this.checkConnection();
  }

  /**
   * Check if backend is likely down based on error patterns
   */
  isBackendDown(): boolean {
    return (
      !this.status.isConnected &&
      (this.status.error?.includes('ECONNREFUSED') ||
        this.status.error?.includes('fetch') ||
        this.status.error?.includes('Connection failed') ||
        this.status.error?.includes('Request timeout'))
    );
  }

  /**
   * Get human-readable status message
   */
  getStatusMessage(): string {
    if (this.status.isConnected) {
      return 'Backend connected';
    }

    if (this.isBackendDown()) {
      return 'Backend server is not running';
    }

    return this.status.error || 'Connection status unknown';
  }

  /**
   * Get development-friendly debugging information
   */
  getDebugInfo(): Record<string, any> {
    return {
      status: this.status,
      config: this.config,
      isMonitoring: !!this.intervalId,
      hasRetryScheduled: !!this.timeoutId,
    };
  }
}

// Singleton instance for global use
export const connectionMonitor = new ConnectionMonitor();

export default ConnectionMonitor;
