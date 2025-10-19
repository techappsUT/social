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

      // ✅ Extract from .data.user wrapper
      const response = await apiClient.get<{ data: { user: UserInfo } }>('/me');
      console.log('🔍 Auth check response:', response);

      // ✅ FIX: Store user ID when checking auth
      const userData = response.data.user;
      localStorage.setItem('userId', userData.id);

      setUser(response.data.user);
    } catch (error) {
      console.error('Auth check failed:', error);
      setUser(null);
      
      // ✅ FIX: Clear user ID on error
      localStorage.removeItem('userId');

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
      const response = await apiClient.get<{ data: { user: UserInfo } }>('/me');
      console.log('🔄 User refreshed:', response.data.user);
      setUser(response.data.user);
    } catch (error) {
      console.error('Failed to refresh user:', error);
    }
  }, []);

  // Logout
  const logout = useCallback(async () => {
    try {
      await apiClient.post('/auth/logout');
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      apiClient.clearAuth();
      setUser(null);
      
      // ✅ FIX: Clear user ID
      localStorage.removeItem('userId');
      localStorage.removeItem('userEmail');
      router.push('/login');
    }
  }, [router]);

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    isEmailVerified: user?.emailVerified || false,
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

// ❌ REMOVE ProtectedRoute component - it causes redirect loops
// The dashboard layout will handle protection instead