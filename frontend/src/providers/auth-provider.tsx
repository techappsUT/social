'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import type { UserInfo } from '@/types/auth';

interface AuthContextType {
  user: UserInfo | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isEmailVerified: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
  checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  // Check authentication status
  const checkAuth = useCallback(async () => {
    try {
      const token = localStorage.getItem('accessToken');
      if (!token) {
        setUser(null);
        setIsLoading(false);
        return;
      }

      const userData = await apiClient.get<UserInfo>('/me');
      setUser(userData);
    } catch (error) {
      console.error('Auth check failed:', error);
      setUser(null);
      apiClient.clearAuth();
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Initial auth check
  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  // Refresh user data
  const refreshUser = useCallback(async () => {
    try {
      const userData = await apiClient.get<UserInfo>('/me');
      setUser(userData);
    } catch (error) {
      console.error('Failed to refresh user:', error);
    }
  }, []);

  // Login
  const login = useCallback(async (email: string, password: string) => {
    const response = await apiClient.post<{
      accessToken: string;
      user: UserInfo;
    }>('/auth/login', { identifier: email, password }, { skipAuth: true });
    
    apiClient.setAccessToken(response.accessToken);
    setUser(response.user);
    
    // Redirect based on email verification status
    if (!response.user.emailVerified) {
      router.push('/verify-email');
    } else {
      router.push('/dashboard');
    }
  }, [router]);

  // Logout
  const logout = useCallback(async () => {
    try {
      await apiClient.post('/auth/logout');
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      apiClient.clearAuth();
      setUser(null);
      router.push('/login');
    }
  }, [router]);

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    isEmailVerified: user?.emailVerified || false,
    login,
    logout,
    refreshUser,
    checkAuth,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}

// Protected route component
export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading, isEmailVerified } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        router.push('/login');
      } else if (!isEmailVerified) {
        router.push('/verify-email');
      }
    }
  }, [isAuthenticated, isEmailVerified, isLoading, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (!isAuthenticated || !isEmailVerified) {
    return null;
  }

  return <>{children}</>;
}