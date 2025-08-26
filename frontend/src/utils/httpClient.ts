// Secure HTTP client with enhanced security features

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { rateLimiter, sanitizeObject } from './validation';

// Request/Response interfaces
interface RequestConfig extends AxiosRequestConfig {
  skipAuth?: boolean;
  skipSanitization?: boolean;
  rateLimit?: {
    type: 'auth' | 'default' | 'upload';
    endpoint: string;
  };
}

interface ApiError {
  message: string;
  status?: number;
  code?: string;
  details?: any;
}

class SecureHttpClient {
  private axiosInstance: AxiosInstance;
  private csrfToken: string | null = null;
  private refreshingToken = false;
  private failedQueue: Array<{
    resolve: (value?: any) => void;
    reject: (error?: any) => void;
  }> = [];

  constructor(baseURL?: string) {
    this.axiosInstance = axios.create({
      baseURL: baseURL || '/api/v1',
      timeout: 30000, // 30 seconds
      withCredentials: true,
      headers: {
        'Content-Type': 'application/json',
        'X-Requested-With': 'XMLHttpRequest', // CSRF protection
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors(): void {
    // Request interceptor
    this.axiosInstance.interceptors.request.use(
      (config: RequestConfig) => {
        // Rate limiting check
        if (config.rateLimit) {
          const { type, endpoint } = config.rateLimit;
          if (!rateLimiter.isAllowed(endpoint, type)) {
            throw new Error(`Rate limit exceeded for ${endpoint}. Please try again later.`);
          }
        }

        // Add authentication token
        if (!config.skipAuth) {
          const token = this.getStoredToken();
          if (token) {
            config.headers = {
              ...config.headers,
              Authorization: `Bearer ${token}`,
            };
          }
        }

        // Add CSRF token
        if (this.csrfToken && ['post', 'put', 'patch', 'delete'].includes(config.method?.toLowerCase() || '')) {
          config.headers = {
            ...config.headers,
            'X-CSRF-Token': this.csrfToken,
          };
        }

        // Sanitize request data
        if (!config.skipSanitization && config.data) {
          config.data = sanitizeObject(config.data);
        }

        // Add request timestamp for debugging
        config.metadata = {
          ...config.metadata,
          startTime: Date.now(),
        };

        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor
    this.axiosInstance.interceptors.response.use(
      (response: AxiosResponse) => {
        // Log response time for monitoring
        const startTime = response.config.metadata?.startTime;
        if (startTime) {
          const duration = Date.now() - startTime;
          if (duration > 5000) {
            console.warn(`Slow API response: ${response.config.url} took ${duration}ms`);
          }
        }

        // Extract CSRF token from response headers
        const csrfToken = response.headers['x-csrf-token'];
        if (csrfToken) {
          this.csrfToken = csrfToken;
        }

        return response;
      },
      async (error: AxiosError) => {
        const originalRequest = error.config as RequestConfig & { _retry?: boolean };

        // Handle authentication errors
        if (error.response?.status === 401 && !originalRequest._retry) {
          if (this.refreshingToken) {
            return new Promise((resolve, reject) => {
              this.failedQueue.push({ resolve, reject });
            }).then(() => {
              return this.axiosInstance(originalRequest);
            }).catch((err) => {
              throw err;
            });
          }

          originalRequest._retry = true;
          this.refreshingToken = true;

          try {
            await this.refreshAccessToken();
            this.processQueue(null);
            return this.axiosInstance(originalRequest);
          } catch (refreshError) {
            this.processQueue(refreshError);
            this.handleAuthFailure();
            throw refreshError;
          } finally {
            this.refreshingToken = false;
          }
        }

        // Handle rate limiting from server
        if (error.response?.status === 429) {
          const retryAfter = error.response.headers['retry-after'];
          const message = `Rate limit exceeded. ${retryAfter ? `Please wait ${retryAfter} seconds before trying again.` : 'Please try again later.'}`;
          throw new Error(message);
        }

        // Handle other errors
        throw this.formatError(error);
      }
    );
  }

  private processQueue(error: any): void {
    this.failedQueue.forEach(({ resolve, reject }) => {
      if (error) {
        reject(error);
      } else {
        resolve();
      }
    });

    this.failedQueue = [];
  }

  private async refreshAccessToken(): Promise<void> {
    try {
      const response = await this.axiosInstance.post('/auth/refresh', {}, {
        skipAuth: true,
      } as RequestConfig);
      
      const { token } = response.data;
      if (token) {
        this.setStoredToken(token);
      }
    } catch (error) {
      throw new Error('Failed to refresh authentication token');
    }
  }

  private handleAuthFailure(): void {
    // Clear stored tokens
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
    
    // Redirect to login page
    if (typeof window !== 'undefined') {
      window.location.href = '/login';
    }
  }

  private formatError(error: AxiosError): ApiError {
    if (!error.response) {
      return {
        message: 'Network error. Please check your connection.',
        code: 'NETWORK_ERROR'
      };
    }

    const { status, data } = error.response;
    
    // Handle validation errors
    if (status === 400 && data && typeof data === 'object') {
      if ('details' in data && Array.isArray(data.details)) {
        return {
          message: 'Validation failed',
          status,
          code: 'VALIDATION_ERROR',
          details: data.details
        };
      }
    }

    // Handle server errors
    if (status >= 500) {
      return {
        message: 'Server error. Please try again later.',
        status,
        code: 'SERVER_ERROR'
      };
    }

    // Default error handling
    const message = (data as any)?.message || 
                   (data as any)?.error || 
                   `Request failed with status ${status}`;

    return {
      message,
      status,
      code: (data as any)?.code || 'API_ERROR'
    };
  }

  private getStoredToken(): string | null {
    try {
      return localStorage.getItem('token');
    } catch {
      return null;
    }
  }

  private setStoredToken(token: string): void {
    try {
      localStorage.setItem('token', token);
    } catch (error) {
      console.error('Failed to store authentication token:', error);
    }
  }

  // Public API methods
  public async get<T = any>(url: string, config?: RequestConfig): Promise<T> {
    const response = await this.axiosInstance.get(url, config);
    return response.data;
  }

  public async post<T = any>(url: string, data?: any, config?: RequestConfig): Promise<T> {
    const response = await this.axiosInstance.post(url, data, config);
    return response.data;
  }

  public async put<T = any>(url: string, data?: any, config?: RequestConfig): Promise<T> {
    const response = await this.axiosInstance.put(url, data, config);
    return response.data;
  }

  public async patch<T = any>(url: string, data?: any, config?: RequestConfig): Promise<T> {
    const response = await this.axiosInstance.patch(url, data, config);
    return response.data;
  }

  public async delete<T = any>(url: string, config?: RequestConfig): Promise<T> {
    const response = await this.axiosInstance.delete(url, config);
    return response.data;
  }

  // File upload with progress and validation
  public async uploadFile<T = any>(
    url: string,
    file: File,
    data?: any,
    onProgress?: (progress: number) => void
  ): Promise<T> {
    const formData = new FormData();
    formData.append('file', file);
    
    if (data) {
      Object.keys(data).forEach(key => {
        formData.append(key, data[key]);
      });
    }

    const config: RequestConfig = {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
      rateLimit: {
        type: 'upload',
        endpoint: url
      }
    };

    const response = await this.axiosInstance.post(url, formData, config);
    return response.data;
  }

  // Authentication methods
  public async login(email: string, password: string): Promise<any> {
    const config: RequestConfig = {
      skipAuth: true,
      rateLimit: {
        type: 'auth',
        endpoint: '/auth/login'
      }
    };

    const response = await this.post('/auth/login', { email, password }, config);
    
    if (response.token) {
      this.setStoredToken(response.token);
    }
    
    return response;
  }

  public async register(userData: any): Promise<any> {
    const config: RequestConfig = {
      skipAuth: true,
      rateLimit: {
        type: 'auth',
        endpoint: '/auth/register'
      }
    };

    return this.post('/auth/register', userData, config);
  }

  public async logout(): Promise<void> {
    try {
      await this.post('/auth/logout');
    } catch (error) {
      // Log error but don't throw, as we want to clear local state anyway
      console.error('Logout request failed:', error);
    } finally {
      localStorage.removeItem('token');
      localStorage.removeItem('refreshToken');
      this.csrfToken = null;
    }
  }

  // Utility methods
  public setAuthToken(token: string): void {
    this.setStoredToken(token);
  }

  public clearAuthToken(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
    this.csrfToken = null;
  }

  public isAuthenticated(): boolean {
    return !!this.getStoredToken();
  }

  // Health check
  public async healthCheck(): Promise<any> {
    return this.get('/health', { skipAuth: true });
  }

  // Get rate limit status
  public getRateLimitStatus(endpoint: string, type: string = 'default'): number {
    return rateLimiter.getRemainingRequests(endpoint, type);
  }
}

// Create and export singleton instance
export const httpClient = new SecureHttpClient();

// Export class for testing or custom instances
export { SecureHttpClient };

// Export common error types
export type { ApiError, RequestConfig };

// Error handling utilities
export const isApiError = (error: any): error is ApiError => {
  return error && typeof error === 'object' && 'message' in error;
};

export const getErrorMessage = (error: any): string => {
  if (isApiError(error)) {
    return error.message;
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  return 'An unexpected error occurred';
};

// Request retry utility
export const withRetry = async <T>(
  fn: () => Promise<T>,
  maxAttempts: number = 3,
  delay: number = 1000
): Promise<T> => {
  let lastError: any;
  
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;
      
      if (attempt === maxAttempts) {
        throw error;
      }
      
      // Don't retry on authentication or validation errors
      if (isApiError(error) && (error.status === 401 || error.status === 400)) {
        throw error;
      }
      
      // Exponential backoff
      await new Promise(resolve => setTimeout(resolve, delay * attempt));
    }
  }
  
  throw lastError;
};
