// path: frontend/src/components/auth/AuthGuard.tsx

'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { Loader2 } from 'lucide-react';

interface AuthGuardProps {
  children: React.ReactNode;
  requiredRole?: 'user' | 'admin' | 'super_admin';
  fallback?: React.ReactNode;
}

/**
 * AuthGuard - Protects routes requiring authentication
 * Redirects to login if not authenticated
 * Optionally checks for required role
 */
export function AuthGuard({ children, requiredRole, fallback }: AuthGuardProps) {
  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (!isLoading) {
      // Not authenticated - redirect to login
      if (!isAuthenticated) {
        const currentPath = window.location.pathname + window.location.search;
        router.push(`/login?redirect=${encodeURIComponent(currentPath)}`);
        return;
      }

      // Check role if required
      if (requiredRole && user?.role !== requiredRole) {
        // Insufficient permissions
        router.push('/unauthorized');
      }
    }
  }, [isAuthenticated, isLoading, user, requiredRole, router]);

  // Show loading state
  if (isLoading) {
    return (
      fallback || (
        <div className="min-h-screen flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )
    );
  }

  // Not authenticated
  if (!isAuthenticated) {
    return null; // Will redirect via useEffect
  }

  // Insufficient permissions
  if (requiredRole && user?.role !== requiredRole) {
    return null; // Will redirect via useEffect
  }

  // Authenticated and authorized
  return <>{children}</>;
}