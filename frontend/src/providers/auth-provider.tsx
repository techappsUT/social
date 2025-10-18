// frontend/src/providers/auth-provider.tsx
// FIXED: Proper auth provider with dashboard protection
'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useRouter, usePathname } from 'next/navigation';
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

// List of public routes that don't require authentication
const publicRoutes = ['/login', '/signup', '/forgot-password', '/reset-password', '/verify-email'];

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();
  const pathname = usePathname();

  // Check authentication status
  const checkAuth = useCallback(async () => {
    try {
      const token = localStorage.getItem('accessToken');
      if (!token) {
        setUser(null);
        setIsLoading(false);
        return;
      }

      // ✅ FIX: Use correct endpoint /me
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

  // Route protection
  useEffect(() => {
    if (isLoading) return;

    const isPublicRoute = publicRoutes.some(route => pathname.startsWith(route));
    const isDashboardRoute = pathname.startsWith('/dashboard');

    if (!user && isDashboardRoute) {
      // Not authenticated, trying to access dashboard - redirect to login
      router.push('/login');
    } else if (user && !user.emailVerified && isDashboardRoute) {
      // Authenticated but email not verified - redirect to verification
      router.push('/verify-email');
    } else if (user && isPublicRoute && pathname !== '/verify-email') {
      // Authenticated user on public route (except verify-email) - redirect to dashboard
      router.push('/dashboard');
    }
  }, [user, isLoading, pathname, router]);

  // Refresh user data
  const refreshUser = useCallback(async () => {
    try {
      const userData = await apiClient.get<UserInfo>('/me');
      setUser(userData);
    } catch (error) {
      console.error('Failed to refresh user:', error);
      setUser(null);
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

// Hook to use auth context
export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}

// Protected route wrapper component
export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading, isEmailVerified } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading) {
      if (!isAuthenticated) {
        router.push('/login');
      } else if (!isEmailVerified && pathname !== '/verify-email') {
        router.push('/verify-email');
      }
    }
  }, [isAuthenticated, isEmailVerified, isLoading, pathname, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}