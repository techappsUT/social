// path: frontend/src/lib/api-client.ts

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000';

export async function apiClient(url: string, options: RequestInit = {}) {
  const token = localStorage.getItem('accessToken');
  
  const response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': token ? `Bearer ${token}` : '',
      'Content-Type': 'application/json',
    },
    credentials: 'include',
  });

  // Handle 401 - refresh token
  if (response.status === 401) {
    const refreshToken = localStorage.getItem('refreshToken');
    
    const refreshResponse = await fetch(`${API_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    });

    if (refreshResponse.ok) {
      const { accessToken } = await refreshResponse.json();
      localStorage.setItem('accessToken', accessToken);
      
      // Retry original request
      return apiClient(url, options);
    } else {
      // Refresh failed - logout
      localStorage.clear();
      window.location.href = '/login';
    }
  }

  return response;
}