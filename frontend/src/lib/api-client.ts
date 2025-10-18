// frontend/src/lib/api-client.ts
// WORKING VERSION - No CORS issues

import type { AuthResponse } from '@/types/auth';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000/api/v2';

/**
 * Enhanced API Client with automatic token refresh
 * Aligned with backend authentication flow
 */

interface RequestOptions extends RequestInit {
  skipAuth?: boolean;
  skipRefresh?: boolean;
}

// Backend response format wrapper
interface BackendResponse<T = any> {
  success?: boolean;
  data?: T;
  error?: string;
  message?: string;
}

class ApiClient {
  private isRefreshing = false;
  private refreshQueue: Array<() => void> = [];

  /**
   * Main request method with automatic token refresh
   */
  async request<T>(
    endpoint: string,
    options: RequestOptions = {}
  ): Promise<T> {
    const { skipAuth = false, skipRefresh = false, ...fetchOptions } = options;

    // Prepare headers
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(fetchOptions.headers as Record<string, string>),
    };

    // Add auth token if not skipped
    if (!skipAuth) {
      const token = this.getAccessToken();
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
    }

    // Make request
    const response = await fetch(`${API_URL}${endpoint}`, {
      ...fetchOptions,
      headers,
      credentials: 'include', // Include cookies (for refresh token)
    });

    // Handle 401 Unauthorized - attempt token refresh
    if (response.status === 401 && !skipRefresh && !skipAuth) {
      return this.handleUnauthorized(endpoint, options);
    }

    // Handle other errors
    if (!response.ok) {
      const error = await response.json().catch(() => ({
        error: 'Request failed',
        message: response.statusText,
      }));
      throw new Error(error.error || error.message || 'Request failed');
    }

    // Parse response
    const responseData: BackendResponse<T> = await response.json();
    
    // ✅ FIX: Unwrap backend response format
    // If the response has 'data' field and 'success' field, unwrap it
    if ('success' in responseData && 'data' in responseData) {
      return responseData.data as T;
    }
    
    // Otherwise return as is (for backward compatibility)
    return responseData as T;
  }

  /**
   * Handle 401 by refreshing token and retrying request
   */
  private async handleUnauthorized<T>(
    endpoint: string,
    options: RequestOptions
  ): Promise<T> {
    // If already refreshing, queue this request
    if (this.isRefreshing) {
      return new Promise((resolve, reject) => {
        this.refreshQueue.push(async () => {
          try {
            const result = await this.request<T>(endpoint, {
              ...options,
              skipRefresh: true,
            });
            resolve(result);
          } catch (error) {
            reject(error);
          }
        });
      });
    }

    // Start refresh process
    this.isRefreshing = true;

    try {
      // Attempt to refresh token
      const refreshResponse = await fetch(`${API_URL}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include', // Refresh token is in cookie
      });

      if (!refreshResponse.ok) {
        throw new Error('Token refresh failed');
      }

      const responseData: BackendResponse<AuthResponse> = await refreshResponse.json();
      
      // Unwrap the response
      const authData = responseData.data || responseData as AuthResponse;

      // Store new access token
      this.setAccessToken(authData.accessToken);

      // Process queued requests
      this.refreshQueue.forEach((callback) => callback());
      this.refreshQueue = [];

      // Retry original request
      return await this.request<T>(endpoint, {
        ...options,
        skipRefresh: true,
      });
    } catch (error) {
      // Refresh failed - clear auth and redirect to login
      this.clearAuth();
      window.location.href = '/login';
      throw error;
    } finally {
      this.isRefreshing = false;
    }
  }

  /**
   * Token management
   */
  getAccessToken(): string | null {
    return localStorage.getItem('accessToken');
  }

  setAccessToken(token: string): void {
    localStorage.setItem('accessToken', token);
  }

  clearAuth(): void {
    localStorage.removeItem('accessToken');
    // Refresh token is cleared via cookie on logout
  }

  /**
   * Convenience methods
   */
  get<T>(endpoint: string, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'GET' });
  }

  post<T>(
    endpoint: string,
    data?: unknown,
    options?: RequestOptions
  ): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  put<T>(
    endpoint: string,
    data?: unknown,
    options?: RequestOptions
  ): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  patch<T>(
    endpoint: string,
    data?: unknown,
    options?: RequestOptions
  ): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  delete<T>(endpoint: string, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' });
  }
}

// Export singleton instance
export const apiClient = new ApiClient();

// Export default for convenience
export default apiClient;