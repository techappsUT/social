import { AuthResponse, UserInfo } from '@/types/auth';

interface RequestOptions extends RequestInit {
  skipAuth?: boolean;
  skipRefresh?: boolean;
}

class ApiClient {
  private baseURL: string;
  private accessToken: string | null = null;
  private refreshPromise: Promise<void> | null = null;
  private refreshQueue: Array<() => void> = [];
  private csrfToken: string | null = null;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000/api/v2';
    this.loadStoredToken();
  }

  private loadStoredToken(): void {
    if (typeof window !== 'undefined') {
      this.accessToken = localStorage.getItem('accessToken');
    }
  }

  private async getFingerprint(): Promise<string> {
    // Simple fingerprint - enhance with more data in production
    const userAgent = navigator.userAgent;
    const language = navigator.language;
    const platform = navigator.platform;
    const screenResolution = `${screen.width}x${screen.height}`;
    
    const fingerprintData = `${userAgent}|${language}|${platform}|${screenResolution}`;
    const encoder = new TextEncoder();
    const data = encoder.encode(fingerprintData);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
  }

  async request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    const {
      skipAuth = false,
      skipRefresh = false,
      headers = {},
      ...fetchOptions
    } = options;

    // Build headers
    const requestHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
      ...headers as Record<string, string>,
    };

    // Add auth header if token exists
    if (!skipAuth && this.accessToken) {
      requestHeaders['Authorization'] = `Bearer ${this.accessToken}`;
    }

    // Add fingerprint
    requestHeaders['X-Device-Fingerprint'] = await this.getFingerprint();

    // Add CSRF token if available
    if (this.csrfToken) {
      requestHeaders['X-CSRF-Token'] = this.csrfToken;
    }

    const url = `${this.baseURL}${endpoint}`;
    
    try {
      const response = await fetch(url, {
        ...fetchOptions,
        headers: requestHeaders,
        credentials: 'include', // Include cookies
      });

      // Extract CSRF token from response
      const csrfToken = response.headers.get('X-CSRF-Token');
      if (csrfToken) {
        this.csrfToken = csrfToken;
      }

      // Handle 401 - try refresh
      if (response.status === 401 && !skipRefresh && !skipAuth) {
        // Wait for any ongoing refresh
        if (this.refreshPromise) {
          await this.refreshPromise;
          return this.request<T>(endpoint, { ...options, skipRefresh: true });
        }

        // Start refresh
        this.refreshPromise = this.refreshAccessToken();
        
        try {
          await this.refreshPromise;
          // Retry original request
          return this.request<T>(endpoint, { ...options, skipRefresh: true });
        } catch (error) {
          // Refresh failed - logout
          this.clearAuth();
          window.location.href = '/login';
          throw error;
        } finally {
          this.refreshPromise = null;
        }
      }

      if (!response.ok) {
        const error = await response.json().catch(() => ({ error: 'Request failed' }));
        throw new Error(error.error || error.message || 'Request failed');
      }

      return response.json();
    } catch (error) {
      if (error instanceof Error) {
        throw error;
      }
      throw new Error('Network error');
    }
  }

  private async refreshAccessToken(): Promise<void> {
    const response = await fetch(`${this.baseURL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Device-Fingerprint': await this.getFingerprint(),
      },
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Token refresh failed');
    }

    const data: AuthResponse = await response.json();
    this.setAccessToken(data.accessToken);
  }

  setAccessToken(token: string): void {
    this.accessToken = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('accessToken', token);
    }
  }

  clearAuth(): void {
    this.accessToken = null;
    this.csrfToken = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('accessToken');
    }
  }

  // Convenience methods
  get<T>(endpoint: string, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'GET' });
  }

  post<T>(endpoint: string, data?: unknown, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  put<T>(endpoint: string, data?: unknown, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  delete<T>(endpoint: string, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' });
  }
}

export const apiClient = new ApiClient();
export default apiClient;