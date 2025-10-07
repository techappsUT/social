// path: frontend/src/types/auth.ts

/**
 * Auth Types - Aligned with Backend DTOs
 * All types use camelCase to match backend JSON responses
 */

// ============================================================================
// REQUEST TYPES
// ============================================================================

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface SignupCredentials {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

export interface VerifyEmailRequest {
  token: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  newPassword: string;
}

export interface RefreshTokenRequest {
  refreshToken?: string; // Optional, can come from cookie
}

// ============================================================================
// RESPONSE TYPES
// ============================================================================

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  user: UserInfo;
}

export interface UserInfo {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  fullName: string;
  role: 'user' | 'admin' | 'super_admin';
  teamId: string | null;
  emailVerified: boolean;
  avatarUrl?: string | null;
}

export interface MessageResponse {
  message: string;
  success: boolean;
}

export interface ErrorResponse {
  error: string;
  message: string;
  details?: Record<string, string>;
}

// ============================================================================
// AUTH STATE
// ============================================================================

export interface AuthState {
  user: UserInfo | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}