// path: frontend/src/hooks/usePosts.ts
/**
 * Post Management Hooks using React Query
 * Provides data fetching, mutations, and caching for posts
 */

'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import * as postsApi from '@/lib/api/posts';
import type {
  PostDTO,
  CreatePostRequest,
  UpdatePostRequest,
  SchedulePostRequest,
  PostQueryParams,
} from '@/types/posts';

// ============================================================================
// QUERY KEYS (for cache management)
// ============================================================================

export const postKeys = {
  all: ['posts'] as const,
  lists: () => [...postKeys.all, 'list'] as const,
  list: (teamId: string, filters?: PostQueryParams) => 
    [...postKeys.lists(), { teamId, filters }] as const,
  details: () => [...postKeys.all, 'detail'] as const,
  detail: (id: string) => [...postKeys.details(), id] as const,
  scheduled: (teamId: string) => [...postKeys.all, 'scheduled', teamId] as const,
  drafts: (teamId: string) => [...postKeys.all, 'drafts', teamId] as const,
  published: (teamId: string) => [...postKeys.all, 'published', teamId] as const,
};

// ============================================================================
// QUERY HOOKS (Read Operations)
// ============================================================================

/**
 * Get a single post by ID
 */
export function usePost(postId: string | null) {
  return useQuery({
    queryKey: postKeys.detail(postId || ''),
    queryFn: () => postsApi.getPost(postId!),
    enabled: !!postId, // Only run if postId exists
    staleTime: 30000, // 30 seconds
  });
}

/**
 * Get all posts for a team with optional filters
 */
export function usePosts(teamId: string, params?: PostQueryParams) {
  return useQuery({
    queryKey: postKeys.list(teamId, params),
    queryFn: () => postsApi.listPosts(teamId, params),
    enabled: !!teamId,
    staleTime: 60000, // 1 minute
  });
}

/**
 * Get scheduled posts for calendar view
 */
export function useScheduledPosts(
  teamId: string,
  dateFrom?: string,
  dateTo?: string
) {
  return useQuery({
    queryKey: postKeys.scheduled(teamId),
    queryFn: () => postsApi.getScheduledPosts(teamId, dateFrom, dateTo),
    enabled: !!teamId,
    staleTime: 30000,
    refetchInterval: 60000, // Refetch every minute for real-time updates
  });
}

/**
 * Get draft posts
 */
export function useDraftPosts(teamId: string) {
  return useQuery({
    queryKey: postKeys.drafts(teamId),
    queryFn: () => postsApi.getDraftPosts(teamId),
    enabled: !!teamId,
    staleTime: 30000,
  });
}

/**
 * Get published posts
 */
export function usePublishedPosts(
  teamId: string,
  page: number = 1,
  pageSize: number = 20
) {
  return useQuery({
    queryKey: postKeys.published(teamId),
    queryFn: () => postsApi.getPublishedPosts(teamId, page, pageSize),
    enabled: !!teamId,
    staleTime: 60000,
  });
}

// ============================================================================
// MUTATION HOOKS (Write Operations)
// ============================================================================

/**
 * Create a new draft post
 */
export function useCreatePost() {
  const queryClient = useQueryClient();
  const router = useRouter();

  return useMutation({
    mutationFn: (data: CreatePostRequest) => postsApi.createPost(data),
    onSuccess: (post) => {
      // Invalidate and refetch post lists
      queryClient.invalidateQueries({ queryKey: postKeys.lists() });
      queryClient.invalidateQueries({ queryKey: postKeys.drafts(post.teamId) });
      
      toast.success('Post created successfully!');
      
      // Optionally redirect to post detail
      // router.push(`/dashboard/posts/${post.id}`);
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to create post';
      toast.error(message);
    },
  });
}

/**
 * Update an existing post
 */
export function useUpdatePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ postId, data }: { postId: string; data: UpdatePostRequest }) =>
      postsApi.updatePost(postId, data),
    onSuccess: (post) => {
      // Update cache for this specific post
      queryClient.setQueryData(postKeys.detail(post.id), post);
      
      // Invalidate lists to refresh
      queryClient.invalidateQueries({ queryKey: postKeys.lists() });
      
      toast.success('Post updated successfully!');
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to update post';
      toast.error(message);
    },
  });
}

/**
 * Delete a post
 */
export function useDeletePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (postId: string) => postsApi.deletePost(postId),
    onSuccess: (_, postId) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: postKeys.detail(postId) });
      
      // Invalidate all lists
      queryClient.invalidateQueries({ queryKey: postKeys.lists() });
      
      toast.success('Post deleted successfully');
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to delete post';
      toast.error(message);
    },
  });
}

/**
 * Schedule a post
 */
export function useSchedulePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ postId, data }: { postId: string; data: SchedulePostRequest }) =>
      postsApi.schedulePost(postId, data),
    onSuccess: (post) => {
      // Update cache
      queryClient.setQueryData(postKeys.detail(post.id), post);
      
      // Invalidate lists
      queryClient.invalidateQueries({ queryKey: postKeys.lists() });
      queryClient.invalidateQueries({ queryKey: postKeys.scheduled(post.teamId) });
      queryClient.invalidateQueries({ queryKey: postKeys.drafts(post.teamId) });
      
      toast.success('Post scheduled successfully!');
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to schedule post';
      toast.error(message);
    },
  });
}

/**
 * Publish a post immediately
 */
export function usePublishPost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (postId: string) => postsApi.publishPost(postId),
    onSuccess: (result, postId) => {
      // Invalidate all related queries
      queryClient.invalidateQueries({ queryKey: postKeys.detail(postId) });
      queryClient.invalidateQueries({ queryKey: postKeys.lists() });
      
      if (result.success) {
        toast.success('Post published successfully!');
      } else {
        toast.warning(`Published to some platforms. ${result.failedPlatforms.length} failed.`);
      }
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to publish post';
      toast.error(message);
    },
  });
}

/**
 * Cancel a scheduled post
 */
export function useCancelPost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (postId: string) => postsApi.cancelScheduledPost(postId),
    onSuccess: (post) => {
      queryClient.setQueryData(postKeys.detail(post.id), post);
      queryClient.invalidateQueries({ queryKey: postKeys.lists() });
      queryClient.invalidateQueries({ queryKey: postKeys.scheduled(post.teamId) });
      
      toast.success('Post canceled');
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to cancel post';
      toast.error(message);
    },
  });
}

/**
 * Duplicate a post
 */
export function useDuplicatePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (postId: string) => postsApi.duplicatePost(postId),
    onSuccess: (post) => {
      queryClient.invalidateQueries({ queryKey: postKeys.lists() });
      queryClient.invalidateQueries({ queryKey: postKeys.drafts(post.teamId) });
      
      toast.success('Post duplicated successfully!');
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to duplicate post';
      toast.error(message);
    },
  });
}

// ============================================================================
// BULK OPERATIONS
// ============================================================================

/**
 * Bulk delete posts
 */
export function useBulkDeletePosts() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (postIds: string[]) => postsApi.bulkDeletePosts(postIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: postKeys.lists() });
      toast.success('Posts deleted successfully');
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to delete posts';
      toast.error(message);
    },
  });
}

// ============================================================================
// OPTIMISTIC UPDATES (Advanced)
// ============================================================================

/**
 * Update post with optimistic UI update
 * UI updates immediately, then syncs with server
 */
export function useOptimisticUpdatePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ postId, data }: { postId: string; data: UpdatePostRequest }) =>
      postsApi.updatePost(postId, data),
    
    // Optimistically update cache before API call completes
    onMutate: async ({ postId, data }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: postKeys.detail(postId) });

      // Snapshot the previous value
      const previousPost = queryClient.getQueryData<PostDTO>(postKeys.detail(postId));

      // Optimistically update the cache
      if (previousPost) {
        queryClient.setQueryData<PostDTO>(postKeys.detail(postId), {
          ...previousPost,
          ...data,
          updatedAt: new Date().toISOString(),
        });
      }

      // Return context with snapshot
      return { previousPost };
    },
    
    // If mutation fails, rollback using context
    onError: (error, { postId }, context) => {
      if (context?.previousPost) {
        queryClient.setQueryData(postKeys.detail(postId), context.previousPost);
      }
      toast.error('Failed to update post');
    },
    
    // Always refetch after error or success
    onSettled: (_, __, { postId }) => {
      queryClient.invalidateQueries({ queryKey: postKeys.detail(postId) });
    },
  });
}