// frontend/src/providers/auth-provider.tsx
// Complete Auth Provider with global state management

'use client';

import React, { createContext, useContext, useEffect } from 'react';
import { useCurrentUser } from '@/hooks/useAuth';
import type { UserInfo } from '@/types/auth';

interface AuthContextType {
  user: UserInfo | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: Error | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { data: user, isLoading, error } = useCurrentUser();

  const value: AuthContextType = {
    user: user || null,
    isAuthenticated: !!user,
    isLoading,
    error: error as Error | null,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuthContext() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuthContext must be used within AuthProvider');
  }
  return context;
}