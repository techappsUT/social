// path: frontend/src/app/(dashboard)/queue/page.tsx
'use client';

import { useState } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Calendar,
  Clock,
  MoreVertical,
  Edit,
  Trash2,
  Copy,
  Send,
  XCircle,
  Loader2,
} from 'lucide-react';
import { usePosts, useDeletePost, usePublishPost, useCancelPost, useDuplicatePost } from '@/hooks/usePosts';
import { useCurrentTeamId } from '@/contexts/team-context'; // ✅ CHANGED
import type { PostDTO, PostStatus } from '@/types/posts';
import { getStatusLabel, getStatusColor, getPlatformLabel } from '@/types/posts';

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
    hour12: true
  });
}

export default function QueuePage() {
  const teamId = useCurrentTeamId(); // ✅ USE TEAM CONTEXT
  const [activeTab, setActiveTab] = useState<PostStatus>('scheduled');
  const [deletingPostId, setDeletingPostId] = useState<string | null>(null);

  const { data: postsData, isLoading } = usePosts(teamId || '', {
    status: activeTab,
    sortBy: activeTab === 'published' ? 'publishedAt' : 'scheduledAt',
    sortOrder: 'asc',
  });

  const deletePost = useDeletePost();
  const publishPost = usePublishPost();
  const cancelPost = useCancelPost();
  const duplicatePost = useDuplicatePost();

  const handleDelete = async () => {
    if (!deletingPostId) return;
    try {
      await deletePost.mutateAsync(deletingPostId);
      setDeletingPostId(null);
    } catch (error) {
      console.error('Failed to delete post:', error);
    }
  };

  const handlePublishNow = async (postId: string) => {
    try {
      await publishPost.mutateAsync(postId);
    } catch (error) {
      console.error('Failed to publish post:', error);
    }
  };

  const handleCancel = async (postId: string) => {
    try {
      await cancelPost.mutateAsync(postId);
    } catch (error) {
      console.error('Failed to cancel post:', error);
    }
  };

  const handleDuplicate = async (postId: string) => {
    try {
      await duplicatePost.mutateAsync(postId);
    } catch (error) {
      console.error('Failed to duplicate post:', error);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
      </div>
    );
  }

  const posts = postsData?.posts || [];

  return (
    <div className="container mx-auto py-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Content Queue</h1>
          <p className="text-gray-600 mt-1">Manage your scheduled and published posts</p>
        </div>
        <Button onClick={() => window.location.href = '/dashboard/compose'}>
          <Calendar className="mr-2 h-4 w-4" />
          Create Post
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <StatsCard title="Drafts" count={posts.filter(p => p.status === 'draft').length} icon={<Edit className="h-5 w-5" />} color="gray" />
        <StatsCard title="Scheduled" count={posts.filter(p => p.status === 'scheduled').length} icon={<Clock className="h-5 w-5" />} color="blue" />
        <StatsCard title="Published" count={posts.filter(p => p.status === 'published').length} icon={<Send className="h-5 w-5" />} color="green" />
        <StatsCard title="Failed" count={posts.filter(p => p.status === 'failed').length} icon={<XCircle className="h-5 w-5" />} color="red" />
      </div>

      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as PostStatus)}>
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="draft">Drafts</TabsTrigger>
          <TabsTrigger value="scheduled">Scheduled</TabsTrigger>
          <TabsTrigger value="published">Published</TabsTrigger>
          <TabsTrigger value="failed">Failed</TabsTrigger>
        </TabsList>

        <TabsContent value={activeTab} className="mt-6">
          {posts.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Calendar className="h-12 w-12 text-gray-400 mb-4" />
                <h3 className="text-lg font-semibold text-gray-700 mb-2">No {activeTab} posts</h3>
                <p className="text-gray-500 text-center max-w-md">
                  {activeTab === 'draft' && "Create your first draft to get started"}
                  {activeTab === 'scheduled' && "Schedule posts to see them here"}
                  {activeTab === 'published' && "Published posts will appear here"}
                  {activeTab === 'failed' && "Failed posts will appear here"}
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {posts.map((post) => (
                <PostCard
                  key={post.id}
                  post={post}
                  onDelete={() => setDeletingPostId(post.id)}
                  onPublishNow={() => handlePublishNow(post.id)}
                  onCancel={() => handleCancel(post.id)}
                  onDuplicate={() => handleDuplicate(post.id)}
                />
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>

      <AlertDialog open={!!deletingPostId} onOpenChange={() => setDeletingPostId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Post?</AlertDialogTitle>
            <AlertDialogDescription>This action cannot be undone. The post will be permanently deleted.</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-red-600 hover:bg-red-700">
              {deletePost.isPending ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

interface StatsCardProps {
  title: string;
  count: number;
  icon: React.ReactNode;
  color: 'gray' | 'blue' | 'green' | 'red';
}

function StatsCard({ title, count, icon, color }: StatsCardProps) {
  const colorClasses = {
    gray: 'bg-gray-100 text-gray-600',
    blue: 'bg-blue-100 text-blue-600',
    green: 'bg-green-100 text-green-600',
    red: 'bg-red-100 text-red-600',
  };

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-gray-600">{title}</p>
            <p className="text-3xl font-bold mt-1">{count}</p>
          </div>
          <div className={`p-3 rounded-lg ${colorClasses[color]}`}>{icon}</div>
        </div>
      </CardContent>
    </Card>
  );
}

interface PostCardProps {
  post: PostDTO;
  onDelete: () => void;
  onPublishNow: () => void;
  onCancel: () => void;
  onDuplicate: () => void;
}

function PostCard({ post, onDelete, onPublishNow, onCancel, onDuplicate }: PostCardProps) {
  const statusColor = getStatusColor(post.status);

  return (
    <Card className="hover:shadow-md transition-shadow">
      <CardContent className="pt-6">
        <div className="flex gap-4">
          <div className="flex-1 space-y-3">
            <div className="flex items-center gap-2 flex-wrap">
              <Badge className={getBadgeColor(statusColor)}>{getStatusLabel(post.status)}</Badge>
              {post.scheduledAt && post.status === 'scheduled' && (
                <span className="text-sm text-gray-600 flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  {formatDate(post.scheduledAt)}
                </span>
              )}
              {post.publishedAt && post.status === 'published' && (
                <span className="text-sm text-gray-600 flex items-center gap-1">
                  <Send className="h-3 w-3" />
                  {formatDate(post.publishedAt)}
                </span>
              )}
            </div>
            <p className="text-gray-900 line-clamp-3">{post.content}</p>
            <div className="flex items-center gap-2 flex-wrap">
              {post.platforms.map((platform) => (
                <Badge key={platform} variant="outline" className="text-xs">{getPlatformLabel(platform)}</Badge>
              ))}
            </div>
            {post.mediaUrls && post.mediaUrls.length > 0 && (
              <div className="flex gap-2">
                {post.mediaUrls.slice(0, 4).map((url, index) => (
                  <div key={index} className="h-16 w-16 rounded border overflow-hidden">
                    <img src={url} alt={`Media ${index + 1}`} className="h-full w-full object-cover" />
                  </div>
                ))}
                {post.mediaUrls.length > 4 && (
                  <div className="h-16 w-16 rounded border bg-gray-100 flex items-center justify-center text-sm text-gray-600">
                    +{post.mediaUrls.length - 4}
                  </div>
                )}
              </div>
            )}
          </div>
          <div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm"><MoreVertical className="h-4 w-4" /></Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => window.location.href = `/dashboard/posts/${post.id}/edit`}>
                  <Edit className="mr-2 h-4 w-4" />Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={onDuplicate}><Copy className="mr-2 h-4 w-4" />Duplicate</DropdownMenuItem>
                {post.status === 'draft' && (
                  <DropdownMenuItem onClick={onPublishNow}><Send className="mr-2 h-4 w-4" />Publish Now</DropdownMenuItem>
                )}
                {post.status === 'scheduled' && (
                  <>
                    <DropdownMenuItem onClick={onPublishNow}><Send className="mr-2 h-4 w-4" />Publish Now</DropdownMenuItem>
                    <DropdownMenuItem onClick={onCancel}><XCircle className="mr-2 h-4 w-4" />Cancel Schedule</DropdownMenuItem>
                  </>
                )}
                <DropdownMenuItem onClick={onDelete} className="text-red-600">
                  <Trash2 className="mr-2 h-4 w-4" />Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function getBadgeColor(color: string): string {
  const colors: Record<string, string> = {
    gray: 'bg-gray-100 text-gray-700 hover:bg-gray-200',
    blue: 'bg-blue-100 text-blue-700 hover:bg-blue-200',
    yellow: 'bg-yellow-100 text-yellow-700 hover:bg-yellow-200',
    purple: 'bg-purple-100 text-purple-700 hover:bg-purple-200',
    green: 'bg-green-100 text-green-700 hover:bg-green-200',
    red: 'bg-red-100 text-red-700 hover:bg-red-200',
  };
  return colors[color] || colors.gray;
}