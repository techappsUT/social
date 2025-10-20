// path: frontend/src/lib/api/social.ts
/**
 * Social Media API Client
 * Handles OAuth flows and social account management
 */

import { apiClient } from '@/lib/api-client';
import type {
  SocialPlatform,
  SocialAccountDTO,
  ConnectAccountRequest,
  PublishToSocialRequest,
  PublishResult,
  SocialAnalytics,
  OAuthURLResponse,
  SocialApiResponse,
  AccountListResponse,
} from '@/types/social';

// ============================================================================
// OAUTH FLOW
// ============================================================================

/**
 * Get OAuth authorization URL for a platform
 * GET /api/v2/social/auth/:platform
 */
export async function getOAuthURL(
  platform: SocialPlatform
): Promise<OAuthURLResponse> {
  const response = await apiClient.get<SocialApiResponse<OAuthURLResponse>>(
    `/social/auth/${platform}`
  );
  return response.data;
}

/**
 * Initiate OAuth flow by redirecting to platform
 * This opens the OAuth popup/redirect
 */
export async function initiateOAuth(platform: SocialPlatform): Promise<void> {
  try {
    const { authUrl } = await getOAuthURL(platform);
    console.log({authUrl})
    
    // Open OAuth in popup window (better UX than redirect)
    const width = 600;
    const height = 700;
    const left = window.screenX + (window.outerWidth - width) / 2;
    const top = window.screenY + (window.outerHeight - height) / 2;
    
    const popup = window.open(
      authUrl,
      `oauth_${platform}`,
      `width=${width},height=${height},left=${left},top=${top}`
    );
    
    if (!popup) {
      // Popup blocked, fallback to redirect
      window.location.href = authUrl;
    }
  } catch (error) {
    console.error('Failed to initiate OAuth:', error);
    throw error;
  }
}

/**
 * Handle OAuth callback (called from popup/redirect)
 * Backend handles GET /api/v2/social/auth/:platform/callback
 * Then we call connectAccount with the code
 */
export async function handleOAuthCallback(
  urlSearchParams: URLSearchParams
): Promise<{ code: string; state: string; platform: string }> {
  const code = urlSearchParams.get('code');
  const state = urlSearchParams.get('state');
  const platform = urlSearchParams.get('platform');
  
  if (!code || !state || !platform) {
    throw new Error('Invalid OAuth callback parameters');
  }
  
  return { code, state, platform };
}

// ============================================================================
// ACCOUNT MANAGEMENT
// ============================================================================

/**
 * Connect a social media account after OAuth
 * POST /api/v2/social/accounts
 */
export async function connectAccount(
  data: ConnectAccountRequest
): Promise<SocialAccountDTO> {
  const response = await apiClient.post<SocialApiResponse<SocialAccountDTO>>(
    '/social/accounts',
    data
  );
  return response.data;
}

/**
 * Get all connected accounts for a team
 * GET /api/v2/teams/:teamId/social/accounts
 */
export async function listAccounts(
  teamId: string
): Promise<SocialAccountDTO[]> {
  const response = await apiClient.get<SocialApiResponse<AccountListResponse>>(
    `/teams/${teamId}/social/accounts`
  );
  return response.data.accounts || [];
}

/**
 * Get a single social account by ID
 * GET /api/v2/social/accounts/:id
 */
export async function getAccount(accountId: string): Promise<SocialAccountDTO> {
  const response = await apiClient.get<SocialApiResponse<SocialAccountDTO>>(
    `/social/accounts/${accountId}`
  );
  return response.data;
}

/**
 * Disconnect a social media account
 * DELETE /api/v2/social/accounts/:id
 */
export async function disconnectAccount(accountId: string): Promise<void> {
  await apiClient.delete(`/social/accounts/${accountId}`);
}

/**
 * Refresh OAuth tokens for an account
 * POST /api/v2/social/accounts/:id/refresh
 */
export async function refreshAccountTokens(
  accountId: string
): Promise<SocialAccountDTO> {
  const response = await apiClient.post<SocialApiResponse<SocialAccountDTO>>(
    `/social/accounts/${accountId}/refresh`,
    {}
  );
  return response.data;
}

// ============================================================================
// PUBLISHING
// ============================================================================

/**
 * Publish content directly to a social platform
 * POST /api/v2/social/accounts/:id/publish
 */
export async function publishToSocial(
  accountId: string,
  data: PublishToSocialRequest
): Promise<PublishResult> {
  const response = await apiClient.post<SocialApiResponse<PublishResult>>(
    `/social/accounts/${accountId}/publish`,
    data
  );
  return response.data;
}

// ============================================================================
// ANALYTICS
// ============================================================================

/**
 * Get analytics for a published post
 * GET /api/v2/social/accounts/:accountId/posts/:postId/analytics
 */
export async function getPostAnalytics(
  accountId: string,
  postId: string
): Promise<SocialAnalytics> {
  const response = await apiClient.get<SocialApiResponse<SocialAnalytics>>(
    `/social/accounts/${accountId}/posts/${postId}/analytics`
  );
  return response.data;
}

// ============================================================================
// ACCOUNT HEALTH CHECKS
// ============================================================================

/**
 * Check account health and connection status
 * GET /api/v2/social/accounts/:id/health
 */
export async function checkAccountHealth(accountId: string): Promise<{
  isHealthy: boolean;
  status: string;
  issues: string[];
  lastChecked: string;
}> {
  const response = await apiClient.get<SocialApiResponse<any>>(
    `/social/accounts/${accountId}/health`
  );
  return response.data;
}

/**
 * Test if account can post (check rate limits, permissions)
 * GET /api/v2/social/accounts/:id/can-post
 */
export async function canAccountPost(accountId: string): Promise<{
  canPost: boolean;
  reason?: string;
  rateLimits?: {
    remaining: number;
    resetAt: string;
  };
}> {
  const response = await apiClient.get<SocialApiResponse<any>>(
    `/social/accounts/${accountId}/can-post`
  );
  return response.data;
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Complete OAuth flow: initiate → callback → connect
 * High-level helper that combines the entire flow
 */
export async function completeOAuthFlow(
  platform: SocialPlatform,
  teamId: string
): Promise<SocialAccountDTO> {
  // 1. Initiate OAuth
  await initiateOAuth(platform);
  
  // 2. Wait for callback (in real app, this is handled by callback page)
  // The callback page will call connectAccount with the code
  
  // 3. This function would be called from the callback page:
  // const { code, state } = await handleOAuthCallback(searchParams);
  // const account = await connectAccount({ teamId, platform, code, state });
  // return account;
  
  throw new Error('OAuth flow must be completed via callback page');
}

/**
 * Get accounts by platform
 */
export async function getAccountsByPlatform(
  teamId: string,
  platform: SocialPlatform
): Promise<SocialAccountDTO[]> {
  const accounts = await listAccounts(teamId);
  return accounts.filter(acc => acc.platform === platform);
}

/**
 * Get active accounts only
 */
export async function getActiveAccounts(
  teamId: string
): Promise<SocialAccountDTO[]> {
  const accounts = await listAccounts(teamId);
  return accounts.filter(acc => acc.status === 'active');
}

/**
 * Check if team has any connected accounts
 */
export async function hasConnectedAccounts(teamId: string): Promise<boolean> {
  const accounts = await listAccounts(teamId);
  return accounts.length > 0;
}

/**
 * Get account count by platform
 */
export async function getAccountCountByPlatform(
  teamId: string
): Promise<Record<SocialPlatform, number>> {
  const accounts = await listAccounts(teamId);
  
  const counts: Record<string, number> = {};
  accounts.forEach(acc => {
    counts[acc.platform] = (counts[acc.platform] || 0) + 1;
  });
  
  return counts as Record<SocialPlatform, number>;
}