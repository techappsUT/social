// path: frontend/src/components/posts/post-composer.tsx
'use client';
import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Calendar, Save, Send, Loader2 } from 'lucide-react';
import { useCreatePost, useUpdatePost, useSchedulePost } from '@/hooks/usePosts';
import type { Platform, CreatePostRequest, PostDTO } from '@/types/posts';

// ============================================================================
// VALIDATION SCHEMA
// ============================================================================

const postSchema = z.object({
  content: z.string()
    .min(1, 'Content is required')
    .max(3000, 'Content too long'),
  platforms: z.array(z.enum(['twitter', 'facebook', 'linkedin', 'instagram', 'tiktok', 'pinterest']))
    .min(1, 'Select at least one platform'),
  mediaUrls: z.array(z.string().url()).optional(),
  scheduledAt: z.string().optional(),
});

type PostFormData = z.infer<typeof postSchema>;

// ============================================================================
// PLATFORM CONFIG
// ============================================================================

const PLATFORMS: Array<{ value: Platform; label: string; color: string; maxChars: number }> = [
  { value: 'twitter', label: 'Twitter/X', color: '#1DA1F2', maxChars: 280 },
  { value: 'facebook', label: 'Facebook', color: '#1877F2', maxChars: 63206 },
  { value: 'linkedin', label: 'LinkedIn', color: '#0A66C2', maxChars: 3000 },
  { value: 'instagram', label: 'Instagram', color: '#E4405F', maxChars: 2200 },
  { value: 'tiktok', label: 'TikTok', color: '#000000', maxChars: 150 },
  { value: 'pinterest', label: 'Pinterest', color: '#BD081C', maxChars: 500 },
];

// ============================================================================
// COMPONENT
// ============================================================================

interface PostComposerProps {
  teamId: string;
  initialPost?: PostDTO;
  onSuccess?: (post: PostDTO) => void;
  onCancel?: () => void;
}

export function PostComposer({ teamId, initialPost, onSuccess, onCancel }: PostComposerProps) {
  const [selectedPlatforms, setSelectedPlatforms] = useState<Platform[]>(
    initialPost?.platforms || []
  );
  const [showScheduler, setShowScheduler] = useState(false);

  const createPost = useCreatePost();
  const updatePost = useUpdatePost();
  const schedulePost = useSchedulePost();

  const {
    register,
    watch,
    setValue,
    formState: { errors, isSubmitting },
    getValues,
  } = useForm<PostFormData>({
    resolver: zodResolver(postSchema),
    defaultValues: {
      content: initialPost?.content || '',
      platforms: initialPost?.platforms || [],
      mediaUrls: initialPost?.mediaUrls || [],
    },
  });

  const content = watch('content');
  const contentLength = content?.length || 0;

  // Calculate max allowed characters based on selected platforms
  const maxChars = selectedPlatforms.length > 0
    ? Math.min(...selectedPlatforms.map(p => 
        PLATFORMS.find(pl => pl.value === p)?.maxChars || 280
      ))
    : 280;

  const isOverLimit = contentLength > maxChars;

  // Handle platform toggle
  const togglePlatform = (platform: Platform) => {
    const updated = selectedPlatforms.includes(platform)
      ? selectedPlatforms.filter(p => p !== platform)
      : [...selectedPlatforms, platform];
    
    setSelectedPlatforms(updated);
    setValue('platforms', updated);
  };

  // Submit as draft
  const handleSaveDraft = async () => {
    const data = getValues();
    
    try {
      const payload: CreatePostRequest = {
        teamId,
        content: data.content,
        platforms: data.platforms,
        mediaUrls: data.mediaUrls,
      };

      if (initialPost) {
        const updated = await updatePost.mutateAsync({
          postId: initialPost.id,
          data: { content: data.content, platforms: data.platforms },
        });
        onSuccess?.(updated);
      } else {
        const created = await createPost.mutateAsync(payload);
        onSuccess?.(created);
      }
    } catch (error) {
      console.error('Failed to save draft:', error);
    }
  };

  // Submit and schedule
  const handleSchedule = async () => {
    const data = getValues();
    
    if (!data.scheduledAt) {
      alert('Please select a date and time');
      return;
    }

    try {
      let postId = initialPost?.id;

      // Create draft first if new post
      if (!postId) {
        const created = await createPost.mutateAsync({
          teamId,
          content: data.content,
          platforms: data.platforms,
        });
        postId = created.id;
      }

      // Then schedule it
      const scheduled = await schedulePost.mutateAsync({
        postId: postId!,
        data: { scheduledAt: data.scheduledAt },
      });

      onSuccess?.(scheduled);
    } catch (error) {
      console.error('Failed to schedule post:', error);
    }
  };

  // Submit and publish immediately
  const handlePublishNow = async () => {
    const data = getValues();
    
    try {
      let postId = initialPost?.id;

      // Create draft first if new post
      if (!postId) {
        const created = await createPost.mutateAsync({
          teamId,
          content: data.content,
          platforms: data.platforms,
        });
        postId = created.id;
      }

      // Then publish immediately (will be handled by backend worker)
      const { publishPost } = await import('@/lib/api/posts');
      await publishPost(postId!);
      
      onSuccess?.({ ...initialPost!, status: 'publishing' } as PostDTO);
    } catch (error) {
      console.error('Failed to publish post:', error);
    }
  };

  return (
    <Card className="w-full max-w-2xl mx-auto">
      <CardHeader>
        <CardTitle>{initialPost ? 'Edit Post' : 'Create New Post'}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {/* Platform Selection */}
          <div>
            <Label className="text-base mb-3 block">Select Platforms *</Label>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
              {PLATFORMS.map((platform) => (
                <div
                  key={platform.value}
                  onClick={() => togglePlatform(platform.value)}
                  className={`
                    flex items-center p-3 rounded-lg border-2 cursor-pointer transition-all
                    ${selectedPlatforms.includes(platform.value)
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 hover:border-gray-300'
                    }
                  `}
                >
                  <Checkbox
                    checked={selectedPlatforms.includes(platform.value)}
                    className="mr-2"
                  />
                  <span className="text-sm font-medium">{platform.label}</span>
                </div>
              ))}
            </div>
            {errors.platforms && (
              <p className="text-red-500 text-sm mt-1">{errors.platforms.message}</p>
            )}
          </div>

          {/* Content Textarea */}
          <div>
            <div className="flex justify-between items-center mb-2">
              <Label htmlFor="content">Post Content *</Label>
              <span
                className={`text-sm ${isOverLimit ? 'text-red-500 font-bold' : 'text-gray-500'}`}
              >
                {contentLength} / {maxChars}
              </span>
            </div>
            <Textarea
              id="content"
              {...register('content')}
              placeholder="What's on your mind?"
              rows={6}
              className={`resize-none ${isOverLimit ? 'border-red-500' : ''}`}
            />
            {errors.content && (
              <p className="text-red-500 text-sm mt-1">{errors.content.message}</p>
            )}
            {isOverLimit && (
              <p className="text-red-500 text-sm mt-1">
                Content exceeds maximum length for selected platforms
              </p>
            )}
          </div>

          {/* Schedule Date/Time (conditionally shown) */}
          {showScheduler && (
            <div>
              <Label htmlFor="scheduledAt">Schedule for</Label>
              <input
                type="datetime-local"
                {...register('scheduledAt')}
                className="w-full p-2 border rounded-md mt-1"
                min={new Date().toISOString().slice(0, 16)}
              />
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex flex-wrap gap-3">
            <Button
              type="button"
              variant="outline"
              onClick={handleSaveDraft}
              disabled={isSubmitting || isOverLimit}
            >
              {isSubmitting ? <Loader2 className="animate-spin mr-2 h-4 w-4" /> : <Save className="mr-2 h-4 w-4" />}
              Save Draft
            </Button>

            <Button
              type="button"
              variant="default"
              onClick={() => setShowScheduler(!showScheduler)}
            >
              <Calendar className="mr-2 h-4 w-4" />
              {showScheduler ? 'Hide Scheduler' : 'Schedule'}
            </Button>

            {showScheduler && (
              <Button
                type="button"
                variant="default"
                onClick={handleSchedule}
                disabled={isSubmitting || isOverLimit}
              >
                {isSubmitting ? <Loader2 className="animate-spin mr-2 h-4 w-4" /> : <Calendar className="mr-2 h-4 w-4" />}
                Schedule Post
              </Button>
            )}

            <Button
              type="button"
              onClick={handlePublishNow}
              disabled={isSubmitting || isOverLimit}
            >
              {isSubmitting ? <Loader2 className="animate-spin mr-2 h-4 w-4" /> : <Send className="mr-2 h-4 w-4" />}
              Publish Now
            </Button>

            {onCancel && (
              <Button type="button" variant="ghost" onClick={onCancel}>
                Cancel
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}