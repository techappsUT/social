// frontend/src/hooks/useAuth.ts
// FIXED: Correct endpoint and response handling
'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import type {
  LoginCredentials,
  SignupCredentials,
  AuthResponse,
  UserInfo,
  MessageResponse,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  VerifyEmailRequest,
} from '@/types/auth';

// ============================================================================
// API FUNCTIONS (Fixed endpoints)
// ============================================================================

async function loginRequest(credentials: LoginCredentials): Promise<AuthResponse> {
  return apiClient.post<AuthResponse>('/auth/login', credentials, {
    skipAuth: true,
  });
}

async function signupRequest(credentials: SignupCredentials): Promise<AuthResponse> {
  return apiClient.post<AuthResponse>('/auth/signup', credentials, {
    skipAuth: true,
  });
}

async function logoutRequest(): Promise<MessageResponse> {
  return apiClient.post<MessageResponse>('/auth/logout');
}

async function getCurrentUser(): Promise<UserInfo> {
  // ✅ FIX: Correct endpoint - /me is at root level, not under /auth
  return apiClient.get<UserInfo>('/me');
}

async function verifyEmailRequest(data: VerifyEmailRequest): Promise<MessageResponse> {
  return apiClient.post<MessageResponse>('/auth/verify-email', data, {
    skipAuth: true,
  });
}

async function forgotPasswordRequest(data: ForgotPasswordRequest): Promise<MessageResponse> {
  return apiClient.post<MessageResponse>('/auth/forgot-password', data, {
    skipAuth: true,
  });
}

async function resetPasswordRequest(data: ResetPasswordRequest): Promise<MessageResponse> {
  return apiClient.post<MessageResponse>('/auth/reset-password', data, {
    skipAuth: true,
  });
}

async function resendVerificationRequest(email: string): Promise<MessageResponse> {
  return apiClient.post<MessageResponse>('/auth/resend-verification', { email }, {
    skipAuth: true,
  });
}

// ============================================================================
// HOOKS
// ============================================================================

/**
 * useLogin - Login mutation hook
 */
export function useLogin() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: loginRequest,
    onSuccess: (data) => {
      // Store access token
      apiClient.setAccessToken(data.accessToken);
      
      // Cache user data
      queryClient.setQueryData(['user'], data.user);

      // Redirect based on email verification status
      if (!data.user.emailVerified) {
        router.push('/verify-email');
      } else {
        router.push('/dashboard');
      }
    },
    onError: (error) => {
      console.error('Login error:', error);
    },
  });
}

/**
 * useSignup - Signup mutation hook
 */
export function useSignup() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: signupRequest,
    onSuccess: (data) => {
      // Backend auto-logs in on signup, so store token
      apiClient.setAccessToken(data.accessToken);
      queryClient.setQueryData(['user'], data.user);

      // Redirect based on email verification status
      if (!data.user.emailVerified) {
        router.push('/verify-email');
      } else {
        router.push('/dashboard');
      }
    },
    onError: (error) => {
      console.error('Signup error:', error);
    },
  });
}

/**
 * useLogout - Logout mutation hook
 */
export function useLogout() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: logoutRequest,
    onSuccess: () => {
      // Clear access token
      apiClient.clearAuth();

      // Clear all cached data
      queryClient.clear();

      // Redirect to login
      router.push('/login');
    },
    onError: (error) => {
      console.error('Logout error:', error);
      // Still clear auth and redirect even on error
      apiClient.clearAuth();
      queryClient.clear();
      router.push('/login');
    },
  });
}

/**
 * useCurrentUser - Get current authenticated user
 */
export function useCurrentUser() {
  return useQuery({
    queryKey: ['user'],
    queryFn: getCurrentUser,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!apiClient.getAccessToken(), // Only run if token exists
  });
}

/**
 * useVerifyEmail - Email verification hook
 */
export function useVerifyEmail() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: verifyEmailRequest,
    onSuccess: () => {
      // Invalidate user cache to refresh verification status
      queryClient.invalidateQueries({ queryKey: ['user'] });
      router.push('/dashboard');
    },
  });
}

/**
 * useForgotPassword - Password reset request hook
 */
export function useForgotPassword() {
  return useMutation({
    mutationFn: forgotPasswordRequest,
  });
}

/**
 * useResetPassword - Password reset hook
 */
export function useResetPassword() {
  const router = useRouter();

  return useMutation({
    mutationFn: resetPasswordRequest,
    onSuccess: () => {
      router.push('/login?reset=true');
    },
  });
}

/**
 * useResendVerification - Resend verification email hook
 */
export function useResendVerification() {
  return useMutation({
    mutationFn: resendVerificationRequest,
  });
}

/**
 * useAuth - Combined auth state hook
 */
export function useAuth() {
  const { data: user, isLoading, error } = useCurrentUser();

  return {
    user,
    isAuthenticated: !!user,
    isLoading,
    error,
  };
}