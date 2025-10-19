// frontend/src/hooks/useAuth.ts
// ✅ FIXED VERSION - Properly handles backend response structure with .data wrapper

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

async function loginRequest(credentials: LoginCredentials): Promise<{ data: AuthResponse }> {
  return apiClient.post<{ data: AuthResponse }>('/auth/login', credentials, {
    skipAuth: true,
  });
}

async function signupRequest(credentials: SignupCredentials): Promise<{ data: AuthResponse }> {
  return apiClient.post<{ data: AuthResponse }>('/auth/signup', credentials, {
    skipAuth: true,
  });
}

async function logoutRequest(): Promise<MessageResponse> {
  return apiClient.post<MessageResponse>('/auth/logout');
}

async function getCurrentUser(): Promise<{ data: { user: UserInfo } }> {
  return apiClient.get<{ data: { user: UserInfo } }>('/me');
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
 * ✅ FIXED: Extracts data from response.data wrapper
 */
export function useLogin() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: loginRequest,
    onSuccess: (response) => {
      console.log('✅ Login successful:', response);
      
      // ✅ Extract from .data wrapper
      const { accessToken, user } = response.data;
      
      console.log('🔑 Access token:', accessToken);
      console.log('👤 User data:', user);
      
      // Store access token
      apiClient.setAccessToken(accessToken);

      
      // ✅ FIX: Store user ID
      localStorage.setItem('userId', user.id);
      
      // Cache user data
      queryClient.setQueryData(['user'], user);

      // Check email verification status
      const isEmailVerified = user.emailVerified === true;
      
      console.log('📋 Login verification status:', {
        emailVerified: user.emailVerified,
        willRedirectTo: isEmailVerified ? '/dashboard' : '/verify-email'
      });

      // Redirect based on verification status
      if (isEmailVerified) {
        router.push('/dashboard');
      } else {
        // Store email for resend functionality
        localStorage.setItem('userEmail', user.email);
        router.push('/verify-email');
      }
    },
    onError: (error) => {
      console.error('❌ Login error:', error);
    },
  });
}

/**
 * useSignup - Signup mutation hook
 * ✅ FIXED: Extracts data from response.data wrapper and handles emailVerified
 */
export function useSignup() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: signupRequest,
    onSuccess: (response) => {
      console.log('✅ Signup successful:', response);
      
      // ✅ Extract from .data wrapper
      const { accessToken, user } = response.data;
      
      // ✅ Validate response structure
      if (!user) {
        console.error('❌ Error: User data missing in signup response');
        router.push('/verify-email'); // Fallback to verification
        return;
      }

      // ✅ Store access token
      apiClient.setAccessToken(accessToken);

      // ✅ FIX: Store user ID
      localStorage.setItem('userId', user.id);
      
      // ✅ Cache user data
      queryClient.setQueryData(['user'], user);

      // ✅ Store email for verification resend functionality
      if (user.email) {
        localStorage.setItem('userEmail', user.email);
        console.log('📧 Stored email for verification:', user.email);
      }

      // ✅ Check emailVerified status with proper null/undefined handling
      const isEmailVerified = user.emailVerified === true;
      
      console.log('📋 Email verification status:', {
        emailVerified: user.emailVerified,
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
      console.log('✅ Logout successful');
      
      // Clear access token
      apiClient.clearAuth();

      // ✅ FIX: Clear user ID
      localStorage.removeItem('userId');

      // Clear all cached data
      queryClient.clear();

      // Clear stored email
      localStorage.removeItem('userEmail');

      // Redirect to login
      router.push('/login');
    },
    onError: (error) => {
      console.error('❌ Logout error:', error);
    },
  });
}

/**
 * useCurrentUser - Get current authenticated user
 * ✅ FIXED: Extracts user from response.data.user wrapper
 */
export function useCurrentUser() {
  return useQuery({
    queryKey: ['user'],
    queryFn: async () => {
      const response = await getCurrentUser();
      // ✅ Extract user from nested .data.user
      return response.data.user;
    },
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!apiClient.getAccessToken(),
  });
}

/**
 * useVerifyEmail - Email verification hook
 * ✅ After verification, refetch user to get updated emailVerified status
 */
export function useVerifyEmail() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: verifyEmailRequest,
    onSuccess: () => {
      console.log('✅ Email verified successfully');
      
      // Invalidate user cache to refetch with updated emailVerified status
      queryClient.invalidateQueries({ queryKey: ['user'] });
      
      // Small delay to ensure backend has updated the status
      setTimeout(() => {
        router.push('/dashboard');
      }, 500);
    },
    onError: (error) => {
      console.error('❌ Email verification error:', error);
    },
  });
}

/**
 * useForgotPassword - Password reset request hook
 */
export function useForgotPassword() {
  return useMutation({
    mutationFn: forgotPasswordRequest,
    onSuccess: () => {
      console.log('✅ Password reset email sent');
    },
    onError: (error) => {
      console.error('❌ Forgot password error:', error);
    },
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
      console.log('✅ Password reset successful');
      router.push('/login?reset=true');
    },
    onError: (error) => {
      console.error('❌ Reset password error:', error);
    },
  });
}

/**
 * useResendVerification - Resend verification email hook
 */
export function useResendVerification() {
  return useMutation({
    mutationFn: resendVerificationRequest,
    onSuccess: () => {
      console.log('✅ Verification email resent');
    },
    onError: (error) => {
      console.error('❌ Resend verification error:', error);
    },
  });
}

/**
 * useAuth - Combined auth state hook
 * Provides user state, authentication status, and helper methods
 */
export function useAuth() {
  const { data: user, isLoading, error, refetch } = useCurrentUser();
  const queryClient = useQueryClient();

  return {
    user,
    isAuthenticated: !!user,
    isLoading,
    error,
    /**
     * Manually refetch user data to get latest state
     * Use after operations that change user data (e.g., email verification)
     */
    refreshUser: async () => {
      console.log('🔄 Refreshing user data...');
      await queryClient.invalidateQueries({ queryKey: ['user'] });
      const result = await refetch();
      console.log('✅ User data refreshed:', result.data);
      return result.data;
    },
  };
}