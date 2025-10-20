// path: frontend/src/lib/api/posts.ts
/**
 * Post API Client
 * Handles all post-related API calls
 */

import { apiClient } from '@/lib/api-client';
import type {
  PostDTO,
  CreatePostRequest,
  UpdatePostRequest,
  SchedulePostRequest,
  PublishResult,
  PostApiResponse,
  PostListResponse,
  PostQueryParams,
} from '@/types/posts';

// ============================================================================
// POST CRUD OPERATIONS
// ============================================================================

/**
 * Create a new draft post
 * POST /api/v2/posts
 */
export async function createPost(data: CreatePostRequest): Promise<PostDTO> {
  const response = await apiClient.post<PostApiResponse<PostDTO>>(
    '/posts',
    data
  );
  return response.data;
}

/**
 * Get a single post by ID
 * GET /api/v2/posts/:id
 */
export async function getPost(postId: string): Promise<PostDTO> {
  const response = await apiClient.get<PostApiResponse<PostDTO>>(
    `/posts/${postId}`
  );
  return response.data;
}

/**
 * Update an existing post
 * PUT /api/v2/posts/:id
 */
export async function updatePost(
  postId: string,
  data: UpdatePostRequest
): Promise<PostDTO> {
  const response = await apiClient.put<PostApiResponse<PostDTO>>(
    `/posts/${postId}`,
    data
  );
  return response.data;
}

/**
 * Delete a post (soft delete)
 * DELETE /api/v2/posts/:id
 */
export async function deletePost(postId: string): Promise<void> {
  await apiClient.delete(`/posts/${postId}`);
}

// ============================================================================
// POST ACTIONS
// ============================================================================

/**
 * Schedule a post for future publishing
 * POST /api/v2/posts/:id/schedule
 */
export async function schedulePost(
  postId: string,
  data: SchedulePostRequest
): Promise<PostDTO> {
  const response = await apiClient.post<PostApiResponse<PostDTO>>(
    `/posts/${postId}/schedule`,
    data
  );
  return response.data;
}

/**
 * Publish a post immediately
 * POST /api/v2/posts/:id/publish
 */
export async function publishPost(postId: string): Promise<PublishResult> {
  const response = await apiClient.post<PostApiResponse<PublishResult>>(
    `/posts/${postId}/publish`,
    {}
  );
  return response.data;
}

/**
 * Cancel a scheduled post
 * POST /api/v2/posts/:id/cancel
 */
export async function cancelScheduledPost(postId: string): Promise<PostDTO> {
  const response = await apiClient.post<PostApiResponse<PostDTO>>(
    `/posts/${postId}/cancel`,
    {}
  );
  return response.data;
}

// ============================================================================
// POST LISTING & FILTERING
// ============================================================================

/**
 * Get posts for a team with filters
 * GET /api/v2/teams/:teamId/posts
 */
export async function listPosts(
  teamId: string,
  params?: PostQueryParams
): Promise<PostListResponse> {
  const queryString = params
    ? '?' + new URLSearchParams(params as any).toString()
    : '';
  
  const response = await apiClient.get<PostApiResponse<PostListResponse>>(
    `/teams/${teamId}/posts${queryString}`
  );
  return response.data;
}

/**
 * Get posts by status
 * GET /api/v2/teams/:teamId/posts?status=:status
 */
export async function getPostsByStatus(
  teamId: string,
  status: string,
  params?: PostQueryParams
): Promise<PostListResponse> {
  const queryParams = { ...params, status };
  return listPosts(teamId, queryParams );
}

/**
 * Get scheduled posts (calendar view)
 * GET /api/v2/teams/:teamId/posts?status=scheduled
 */
export async function getScheduledPosts(
  teamId: string,
  dateFrom?: string,
  dateTo?: string
): Promise<PostListResponse> {
  return listPosts(teamId, {
    status: 'scheduled',
    dateFrom,
    dateTo,
    sortBy: 'scheduledAt',
    sortOrder: 'asc',
  });
}

/**
 * Get draft posts
 * GET /api/v2/teams/:teamId/posts?status=draft
 */
export async function getDraftPosts(teamId: string): Promise<PostListResponse> {
  return listPosts(teamId, {
    status: 'draft',
    sortBy: 'updatedAt',
    sortOrder: 'desc',
  });
}

/**
 * Get published posts
 * GET /api/v2/teams/:teamId/posts?status=published
 */
export async function getPublishedPosts(
  teamId: string,
  page: number = 1,
  pageSize: number = 20
): Promise<PostListResponse> {
  return listPosts(teamId, {
    status: 'published',
    page,
    pageSize,
    sortBy: 'publishedAt',
    sortOrder: 'desc',
  });
}

// ============================================================================
// POST SEARCH
// ============================================================================

/**
 * Search posts by content
 * GET /api/v2/teams/:teamId/posts?search=:query
 */
export async function searchPosts(
  teamId: string,
  query: string,
  params?: PostQueryParams
): Promise<PostListResponse> {
  return listPosts(teamId, { ...params, search: query });
}

// ============================================================================
// BULK OPERATIONS
// ============================================================================

/**
 * Bulk delete posts
 * DELETE /api/v2/posts/bulk
 */
export async function bulkDeletePosts(postIds: string[]): Promise<void> {
  await apiClient.post('/posts/bulk-delete', { postIds });
}

/**
 * Bulk schedule posts
 * POST /api/v2/posts/bulk-schedule
 */
export async function bulkSchedulePosts(
  posts: Array<{ postId: string; scheduledAt: string }>
): Promise<void> {
  await apiClient.post('/posts/bulk-schedule', { posts });
}

// ============================================================================
// POST ANALYTICS
// ============================================================================

/**
 * Get analytics for a specific post
 * GET /api/v2/posts/:id/analytics
 */
export async function getPostAnalytics(postId: string) {
  const response = await apiClient.get<PostApiResponse<any>>(
    `/posts/${postId}/analytics`
  );
  return response.data;
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Duplicate a post (create a copy as draft)
 */
export async function duplicatePost(postId: string): Promise<PostDTO> {
  const originalPost = await getPost(postId);
  
  // Create a new draft with same content
  const duplicateData: CreatePostRequest = {
    teamId: originalPost.teamId,
    content: originalPost.content,
    platforms: originalPost.platforms,
    mediaUrls: originalPost.mediaUrls,
    mediaTypes: originalPost.mediaTypes,
    hashtags: originalPost.hashtags,
    mentions: originalPost.mentions,
    link: originalPost.link || undefined,
    priority: originalPost.priority,
    metadata: originalPost.metadata,
  };
  
  return createPost(duplicateData);
}

/**
 * Reschedule a post to a new time
 */
export async function reschedulePost(
  postId: string,
  newScheduledAt: string
): Promise<PostDTO> {
  return schedulePost(postId, { scheduledAt: newScheduledAt });
}