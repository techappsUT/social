// path: frontend/src/hooks/useAuth.ts

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
// API FUNCTIONS (Aligned with Backend)
// ============================================================================

async function loginRequest(credentials: LoginCredentials): Promise<AuthResponse> {
  return apiClient.post<AuthResponse>('/auth/login', credentials, {
    skipAuth: true,
  });
}

async function signupRequest(credentials: SignupCredentials): Promise<MessageResponse> {
  return apiClient.post<MessageResponse>('/auth/signup', credentials, {
    skipAuth: true,
  });
}

async function logoutRequest(): Promise<MessageResponse> {
  return apiClient.post<MessageResponse>('/auth/logout');
}

async function getCurrentUser(): Promise<UserInfo> {
  return apiClient.get<UserInfo>('/api/me');
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
      // Note: refresh token is automatically stored in HTTP-only cookie by backend

      // Cache user data
      queryClient.setQueryData(['user'], data.user);

      // Redirect to dashboard
      router.push('/dashboard');
    },
  });
}

/**
 * useSignup - Signup mutation hook
 */
export function useSignup() {
  const router = useRouter();

  return useMutation({
    mutationFn: signupRequest,
    onSuccess: () => {
      // Redirect to login with success message
      router.push('/login?registered=true');
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

  return useMutation({
    mutationFn: verifyEmailRequest,
    onSuccess: () => {
      router.push('/login?verified=true');
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