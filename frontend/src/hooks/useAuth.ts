// frontend/src/hooks/useAuth.ts
// ✅ FIXED VERSION - Handles emailVerified and redirects properly

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
// API FUNCTIONS
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
  return apiClient.get<UserInfo>('/auth/me');
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

      // Redirect to dashboard
      router.push('/dashboard');
    },
  });
}

/**
 * useSignup - Signup mutation hook
 * ✅ FIXED: Properly handles emailVerified field and redirects
 */
export function useSignup() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: signupRequest,
    onSuccess: (data) => {
      console.log('✅ Signup successful:', data);
      
      // ✅ Store access token
      apiClient.setAccessToken(data.accessToken);
      
      // ✅ Validate response structure
      if (!data.user) {
        console.error('❌ Error: User data missing in signup response');
        router.push('/verify-email'); // Fallback to verification
        return;
      }

      // ✅ Cache user data
      queryClient.setQueryData(['user'], data.user);

      // ✅ Store email for verification resend functionality
      if (data.user.email) {
        localStorage.setItem('userEmail', data.user.email);
        console.log('📧 Stored email for verification:', data.user.email);
      }

      // ✅ Check emailVerified status with proper null/undefined handling
      const isEmailVerified = data.user.emailVerified === true;
      
      console.log('📋 Email verification status:', {
        emailVerified: data.user.emailVerified,
        isVerified: isEmailVerified,
        willRedirectTo: isEmailVerified ? '/dashboard' : '/verify-email'
      });

      // ✅ Redirect based on email verification status
      if (isEmailVerified) {
        console.log('✅ Email already verified, redirecting to dashboard');
        router.push('/dashboard');
      } else {
        console.log('📨 Email not verified, redirecting to verification page');
        router.push('/verify-email');
      }
    },
    onError: (error) => {
      console.error('❌ Signup error:', error);
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
    enabled: !!apiClient.getAccessToken(),
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
      // Clear user cache to refetch updated data
      queryClient.invalidateQueries({ queryKey: ['user'] });
      
      // Redirect to login with success message
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
    refreshUser: async () => {
      // Manual refetch to get latest user data
      const queryClient = useQueryClient();
      await queryClient.invalidateQueries({ queryKey: ['user'] });
    },
  };
}