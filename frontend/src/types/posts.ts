// path: frontend/src/types/post.ts
/**
 * Post Types - Aligned with Backend Domain Models
 * Based on backend/internal/domain/post/post.go
 */

// ============================================================================
// ENUMS
// ============================================================================

export type Platform = 
  | 'twitter' 
  | 'facebook' 
  | 'linkedin' 
  | 'instagram' 
  | 'tiktok' 
  | 'pinterest';

export type PostStatus = 
  | 'draft' 
  | 'scheduled' 
  | 'queued' 
  | 'publishing' 
  | 'published' 
  | 'failed' 
  | 'canceled';

export type PostPriority = 0 | 1 | 2 | 3; // Low, Normal, High, Urgent

export type MediaType = 'image' | 'video' | 'gif';

// ============================================================================
// REQUEST TYPES
// ============================================================================

export interface CreatePostRequest {
  teamId: string;
  content: string;
  platforms: Platform[];
  mediaUrls?: string[];
  mediaTypes?: MediaType[];
  hashtags?: string[];
  mentions?: string[];
  link?: string;
  scheduledAt?: string; // ISO date string
  priority?: PostPriority;
  metadata?: {
    campaign?: string;
    tags?: string[];
    internalNote?: string;
  };
}

export interface UpdatePostRequest {
  content?: string;
  platforms?: Platform[];
  mediaUrls?: string[];
  mediaTypes?: MediaType[];
  hashtags?: string[];
  mentions?: string[];
  link?: string;
  priority?: PostPriority;
  metadata?: {
    campaign?: string;
    tags?: string[];
    internalNote?: string;
  };
}

export interface SchedulePostRequest {
  scheduledAt: string; // ISO date string
  timezone?: string;
}

// ============================================================================
// RESPONSE TYPES (DTOs)
// ============================================================================

export interface PostDTO {
  id: string;
  teamId: string;
  createdBy: string;
  content: string;
  platforms: Platform[];
  mediaUrls: string[];
  mediaTypes: MediaType[];
  hashtags: string[];
  mentions: string[];
  link: string | null;
  linkPreview: LinkPreview | null;
  scheduledAt: string | null;
  publishedAt: string | null;
  status: PostStatus;
  priority: PostPriority;
  metadata: PostMetadata;
  analytics: PostAnalytics | null;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
}

export interface LinkPreview {
  title: string;
  description: string;
  imageUrl: string;
  siteName: string;
}

export interface PostMetadata {
  campaign: string;
  tags: string[];
  internalNote: string;
  requiresApproval: boolean;
  approvedBy: string | null;
  approvedAt: string | null;
}

export interface PostAnalytics {
  views: number;
  likes: number;
  comments: number;
  shares: number;
  clicks: number;
  impressions: number;
  engagementRate: number;
  platformAnalytics: Record<Platform, PlatformAnalytics>;
}

export interface PlatformAnalytics {
  views: number;
  likes: number;
  comments: number;
  shares: number;
  clicks: number;
}

export interface PublishResult {
  postId: string;
  platformResults: PlatformPublishResult[];
  success: boolean;
  failedPlatforms: Platform[];
}

export interface PlatformPublishResult {
  platform: Platform;
  platformPostId: string;
  url: string;
  publishedAt: string;
  success: boolean;
  error: string | null;
}

// ============================================================================
// API RESPONSE WRAPPERS
// ============================================================================

export interface PostApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}

export interface PostListResponse {
  posts: PostDTO[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}

// ============================================================================
// FILTER & QUERY TYPES
// ============================================================================

export interface PostFilters {
  teamId?: string;
  status?: PostStatus;
  platforms?: Platform[];
  dateFrom?: string;
  dateTo?: string;
  search?: string;
}

export interface PostQueryParams extends PostFilters {
  page?: number;
  pageSize?: number;
  sortBy?: 'createdAt' | 'scheduledAt' | 'updatedAt';
  sortOrder?: 'asc' | 'desc';
}

// ============================================================================
// VALIDATION CONSTANTS (from backend)
// ============================================================================

export const POST_VALIDATION = {
  MAX_CONTENT_LENGTH: {
    twitter: 280,
    facebook: 63206,
    linkedin: 3000,
    instagram: 2200,
    tiktok: 150,
    pinterest: 500,
  },
  MAX_MEDIA_FILES: {
    twitter: 4,
    facebook: 10,
    linkedin: 9,
    instagram: 10,
    tiktok: 1,
    pinterest: 5,
  },
  MAX_MEDIA_SIZE_MB: 50,
  SUPPORTED_IMAGE_TYPES: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
  SUPPORTED_VIDEO_TYPES: ['video/mp4', 'video/quicktime', 'video/x-msvideo'],
} as const;

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

export function getMaxContentLength(platforms: Platform[]): number {
  if (platforms.length === 0) return POST_VALIDATION.MAX_CONTENT_LENGTH.twitter;
  
  // Return the shortest max length among selected platforms
  return Math.min(
    ...platforms.map(p => POST_VALIDATION.MAX_CONTENT_LENGTH[p])
  );
}

export function getMaxMediaFiles(platforms: Platform[]): number {
  if (platforms.length === 0) return POST_VALIDATION.MAX_MEDIA_FILES.twitter;
  
  // Return the smallest max media count among selected platforms
  return Math.min(
    ...platforms.map(p => POST_VALIDATION.MAX_MEDIA_FILES[p])
  );
}

export function getPlatformLabel(platform: Platform): string {
  const labels: Record<Platform, string> = {
    twitter: 'Twitter/X',
    facebook: 'Facebook',
    linkedin: 'LinkedIn',
    instagram: 'Instagram',
    tiktok: 'TikTok',
    pinterest: 'Pinterest',
  };
  return labels[platform];
}

export function getStatusColor(status: PostStatus): string {
  const colors: Record<PostStatus, string> = {
    draft: 'gray',
    scheduled: 'blue',
    queued: 'yellow',
    publishing: 'purple',
    published: 'green',
    failed: 'red',
    canceled: 'gray',
  };
  return colors[status];
}

export function getStatusLabel(status: PostStatus): string {
  const labels: Record<PostStatus, string> = {
    draft: 'Draft',
    scheduled: 'Scheduled',
    queued: 'In Queue',
    publishing: 'Publishing...',
    published: 'Published',
    failed: 'Failed',
    canceled: 'Canceled',
  };
  return labels[status];
}