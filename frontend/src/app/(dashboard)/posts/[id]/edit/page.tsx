// path: frontend/src/app/(dashboard)/posts/[id]/edit/page.tsx
'use client';
import React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { PostComposer } from '@/components/posts/post-composer';
import { usePost } from '@/hooks/usePosts';
import { useCurrentUser } from '@/hooks/useAuth';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Loader2, ArrowLeft, AlertCircle } from 'lucide-react';
import { useCurrentTeamId } from '@/contexts/team-context';

/**
 * Post Edit Page
 * 
 * Allows users to edit existing posts
 * 
 * Route: /dashboard/posts/[id]/edit
 * 
 * Features:
 * - Load existing post data
 * - Edit content, platforms, media
 * - Update schedule time
 * - Save changes
 * - Delete post
 */

export default function PostEditPage() {
  const params = useParams();
  const router = useRouter();
  const { data: user } = useCurrentUser();
  const postId = params.id as string;

  // Fetch post data
  const { data: post, isLoading, error } = usePost(postId);
  const teamId = useCurrentTeamId();

  // Handle success
  const handleSuccess = () => {
    router.push('/dashboard/queue');
  };

  // Handle cancel
  const handleCancel = () => {
    router.back();
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="container mx-auto py-8">
        <div className="flex flex-col items-center justify-center min-h-[60vh]">
          <Loader2 className="h-12 w-12 animate-spin text-blue-500 mb-4" />
          <p className="text-gray-600">Loading post...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (error || !post) {
    return (
      <div className="container mx-auto py-8">
        <Card className="max-w-2xl mx-auto">
          <CardContent className="pt-6">
            <div className="flex flex-col items-center text-center space-y-4">
              <div className="h-12 w-12 rounded-full bg-red-100 flex items-center justify-center">
                <AlertCircle className="h-6 w-6 text-red-600" />
              </div>
              <div>
                <h2 className="text-xl font-semibold text-gray-900 mb-2">
                  Post Not Found
                </h2>
                <p className="text-gray-600">
                  {error ? 'Failed to load post. Please try again.' : 'The post you\'re looking for doesn\'t exist.'}
                </p>
              </div>
              <Button onClick={() => router.push('/dashboard/queue')}>
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Queue
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Check if post can be edited
  const canEdit = ['draft', 'scheduled', 'failed'].includes(post.status);
  
  if (!canEdit) {
    return (
      <div className="container mx-auto py-8">
        <Card className="max-w-2xl mx-auto">
          <CardContent className="pt-6">
            <div className="flex flex-col items-center text-center space-y-4">
              <div className="h-12 w-12 rounded-full bg-yellow-100 flex items-center justify-center">
                <AlertCircle className="h-6 w-6 text-yellow-600" />
              </div>
              <div>
                <h2 className="text-xl font-semibold text-gray-900 mb-2">
                  Cannot Edit This Post
                </h2>
                <p className="text-gray-600">
                  Published posts cannot be edited. You can duplicate this post to create a new version.
                </p>
              </div>
              <Button onClick={() => router.push('/dashboard/queue')}>
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Queue
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button 
            variant="ghost" 
            size="sm"
            onClick={handleCancel}
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-2xl font-bold">Edit Post</h1>
            <p className="text-sm text-gray-600 mt-1">
              Make changes to your post
            </p>
          </div>
        </div>
      </div>

      {/* Post Composer */}
      <PostComposer
        teamId={teamId || ''}
        initialPost={post}
        onSuccess={handleSuccess}
        onCancel={handleCancel}
      />
    </div>
  );
}

/**
 * USAGE:
 * 
 * This page is automatically accessible at:
 * /dashboard/posts/[postId]/edit
 * 
 * Example: /dashboard/posts/123e4567-e89b-12d3-a456-426614174000/edit
 * 
 * Navigation:
 * 
 * 1. From post list/queue:
 *    <Button onClick={() => router.push(`/dashboard/posts/${post.id}/edit`)}>
 *      Edit
 *    </Button>
 * 
 * 2. From dropdown menu (already implemented in queue page):
 *    <DropdownMenuItem onClick={() => window.location.href = `/dashboard/posts/${post.id}/edit`}>
 *      Edit
 *    </DropdownMenuItem>
 * 
 * 3. Programmatic navigation:
 *    const router = useRouter();
 *    router.push(`/dashboard/posts/${postId}/edit`);
 * 
 * Features:
 * - Loads existing post data automatically
 * - Pre-fills all form fields
 * - Validates edits cannot be made to published posts
 * - Shows error states if post not found
 * - Redirects to queue after successful save
 */