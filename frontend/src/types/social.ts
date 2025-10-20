// path: frontend/src/types/social.ts
/**
 * Social Media Types - OAuth and Account Management
 * Based on backend/internal/domain/social/account.go
 */

// ============================================================================
// ENUMS
// ============================================================================

export type SocialPlatform = 
  | 'twitter' 
  | 'facebook' 
  | 'linkedin' 
  | 'instagram' 
  | 'tiktok' 
  | 'pinterest'
  | 'youtube';

export type AccountType = 
  | 'personal' 
  | 'business' 
  | 'page' 
  | 'group' 
  | 'channel';

export type ConnectionStatus = 
  | 'active' 
  | 'inactive' 
  | 'expired' 
  | 'revoked' 
  | 'rate_limited' 
  | 'suspended' 
  | 'reconnect_required';

// ============================================================================
// REQUEST TYPES
// ============================================================================

export interface ConnectAccountRequest {
  teamId: string;
  platform: SocialPlatform;
  code: string; // OAuth authorization code from callback
  state?: string; // CSRF protection token
}

export interface PublishToSocialRequest {
  accountId: string;
  content: string;
  mediaUrls?: string[];
  platformOptions?: Record<string, any>;
}

export interface RefreshTokensRequest {
  accountId: string;
}

// ============================================================================
// RESPONSE TYPES (DTOs)
// ============================================================================

export interface SocialAccountDTO {
  id: string;
  teamId: string;
  userId: string; // User who connected the account
  platform: SocialPlatform;
  accountType: AccountType;
  username: string;
  displayName: string;
  profileUrl: string;
  avatarUrl: string;
  platformUserId: string;
  platformAccountId: string;
  status: ConnectionStatus;
  scopes: string[];
  expiresAt: string | null;
  lastSyncAt: string | null;
  connectedAt: string;
  createdAt: string;
  updatedAt: string;
  
  // Rate limit info
  rateLimits?: {
    postsPerHour: number;
    postsPerDay: number;
    remaining: number;
    resetAt: string;
  };
  
  // Health check
  healthStatus?: {
    isHealthy: boolean;
    lastChecked: string;
    issues: string[];
  };
}

export interface OAuthURLResponse {
  authUrl: string;
  state: string;
}

export interface OAuthCallbackResponse {
  code: string;
  state: string;
  platform: SocialPlatform;
}

export interface PublishResult {
  success: boolean;
  platformPostId: string;
  url: string;
  publishedAt: string;
  error?: string;
}

export interface SocialAnalytics {
  postId: string;
  platform: SocialPlatform;
  metrics: {
    views: number;
    likes: number;
    comments: number;
    shares: number;
    clicks: number;
    impressions: number;
    engagement: number;
    engagementRate: number;
  };
  lastUpdated: string;
}

// ============================================================================
// API RESPONSE WRAPPERS
// ============================================================================

export interface SocialApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}

export interface AccountListResponse {
  accounts: SocialAccountDTO[];
  total: number;
}

// ============================================================================
// OAUTH FLOW TYPES
// ============================================================================

export interface OAuthConfig {
  platform: SocialPlatform;
  clientId: string;
  redirectUri: string;
  scope: string[];
  responseType: 'code' | 'token';
  state?: string;
}

export interface OAuthState {
  platform: SocialPlatform;
  userId: string;
  teamId: string;
  timestamp: number;
  nonce: string;
}

// ============================================================================
// PLATFORM CAPABILITIES
// ============================================================================

export interface PlatformCapabilities {
  supportsImages: boolean;
  supportsVideos: boolean;
  supportsMultipleImages: boolean;
  supportsThreads: boolean;
  supportsScheduling: boolean;
  supportsAnalytics: boolean;
  maxImageCount: number;
  maxVideoCount: number;
  maxContentLength: number;
  requiresMedia?: boolean; // Instagram requires at least one media
}

export const PLATFORM_CAPABILITIES: Record<SocialPlatform, PlatformCapabilities> = {
  twitter: {
    supportsImages: true,
    supportsVideos: true,
    supportsMultipleImages: true,
    supportsThreads: true,
    supportsScheduling: true,
    supportsAnalytics: true,
    maxImageCount: 4,
    maxVideoCount: 1,
    maxContentLength: 280,
  },
  facebook: {
    supportsImages: true,
    supportsVideos: true,
    supportsMultipleImages: true,
    supportsThreads: false,
    supportsScheduling: true,
    supportsAnalytics: true,
    maxImageCount: 10,
    maxVideoCount: 1,
    maxContentLength: 63206,
  },
  linkedin: {
    supportsImages: true,
    supportsVideos: true,
    supportsMultipleImages: true,
    supportsThreads: false,
    supportsScheduling: true,
    supportsAnalytics: true,
    maxImageCount: 9,
    maxVideoCount: 1,
    maxContentLength: 3000,
  },
  instagram: {
    supportsImages: true,
    supportsVideos: true,
    supportsMultipleImages: true,
    supportsThreads: false,
    supportsScheduling: true,
    supportsAnalytics: true,
    maxImageCount: 10,
    maxVideoCount: 1,
    maxContentLength: 2200,
    requiresMedia: true,
  },
  tiktok: {
    supportsImages: false,
    supportsVideos: true,
    supportsMultipleImages: false,
    supportsThreads: false,
    supportsScheduling: true,
    supportsAnalytics: false,
    maxImageCount: 0,
    maxVideoCount: 1,
    maxContentLength: 150,
  },
  pinterest: {
    supportsImages: true,
    supportsVideos: true,
    supportsMultipleImages: false,
    supportsThreads: false,
    supportsScheduling: true,
    supportsAnalytics: false,
    maxImageCount: 1,
    maxVideoCount: 1,
    maxContentLength: 500,
    requiresMedia: true,
  },
  youtube: {
    supportsImages: false,
    supportsVideos: true,
    supportsMultipleImages: false,
    supportsThreads: false,
    supportsScheduling: true,
    supportsAnalytics: true,
    maxImageCount: 0,
    maxVideoCount: 1,
    maxContentLength: 5000,
    requiresMedia: true,
  },
};

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

export function getPlatformName(platform: SocialPlatform): string {
  const names: Record<SocialPlatform, string> = {
    twitter: 'Twitter/X',
    facebook: 'Facebook',
    linkedin: 'LinkedIn',
    instagram: 'Instagram',
    tiktok: 'TikTok',
    pinterest: 'Pinterest',
    youtube: 'YouTube',
  };
  return names[platform];
}

export function getPlatformColor(platform: SocialPlatform): string {
  const colors: Record<SocialPlatform, string> = {
    twitter: '#1DA1F2',
    facebook: '#1877F2',
    linkedin: '#0A66C2',
    instagram: '#E4405F',
    tiktok: '#000000',
    pinterest: '#BD081C',
    youtube: '#FF0000',
  };
  return colors[platform];
}

export function getStatusColor(status: ConnectionStatus): string {
  const colors: Record<ConnectionStatus, string> = {
    active: 'green',
    inactive: 'gray',
    expired: 'orange',
    revoked: 'red',
    rate_limited: 'yellow',
    suspended: 'red',
    reconnect_required: 'orange',
  };
  return colors[status];
}

export function getStatusLabel(status: ConnectionStatus): string {
  const labels: Record<ConnectionStatus, string> = {
    active: 'Connected',
    inactive: 'Inactive',
    expired: 'Token Expired',
    revoked: 'Access Revoked',
    rate_limited: 'Rate Limited',
    suspended: 'Suspended',
    reconnect_required: 'Reconnect Required',
  };
  return labels[status];
}

export function isAccountHealthy(account: SocialAccountDTO): boolean {
  return account.status === 'active' && 
         (!account.expiresAt || new Date(account.expiresAt) > new Date());
}

export function needsReconnection(account: SocialAccountDTO): boolean {
  return ['expired', 'revoked', 'reconnect_required'].includes(account.status);
}

export function canPublish(account: SocialAccountDTO): boolean {
  return account.status === 'active' && 
         (!account.rateLimits || account.rateLimits.remaining > 0);
}