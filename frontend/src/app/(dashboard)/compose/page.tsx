// path: frontend/src/app/(dashboard)/compose/page.tsx
'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import {
  PenSquare,
  Calendar,
  ImagePlus,
  Send,
  Save,
  Twitter,
  Facebook,
  Linkedin,
  Instagram,
  Loader2,
  Users,
} from 'lucide-react';
import { useCreatePost, useSchedulePost } from '@/hooks/usePosts';
import { useCurrentTeamId } from '@/contexts/team-context';
import type { Platform } from '@/types/posts';

// Platform configuration
const platforms = [
  { id: 'twitter' as Platform, name: 'Twitter', icon: Twitter, limit: 280, color: 'text-blue-400' },
  { id: 'facebook' as Platform, name: 'Facebook', icon: Facebook, limit: 5000, color: 'text-blue-600' },
  { id: 'linkedin' as Platform, name: 'LinkedIn', icon: Linkedin, limit: 3000, color: 'text-blue-700' },
  { id: 'instagram' as Platform, name: 'Instagram', icon: Instagram, limit: 2200, color: 'text-pink-600' },
];

export default function ComposePage() {
  const router = useRouter();
  const teamId = useCurrentTeamId(); // ✅ GET FROM CONTEXT
  
  const [content, setContent] = useState('');
  const [selectedPlatforms, setSelectedPlatforms] = useState<Platform[]>([]);
  const [scheduleEnabled, setScheduleEnabled] = useState(false);
  const [scheduledDate, setScheduledDate] = useState('');
  const [scheduledTime, setScheduledTime] = useState('');

  const createPost = useCreatePost();
  const schedulePost = useSchedulePost();

  const handlePlatformToggle = (platformId: Platform) => {
    setSelectedPlatforms((prev) =>
      prev.includes(platformId) ? prev.filter((id) => id !== platformId) : [...prev, platformId]
    );
  };

  const getCharacterLimit = () => {
    if (selectedPlatforms.length === 0) return 0;
    return Math.min(...selectedPlatforms.map((id) => platforms.find((p) => p.id === id)?.limit || 0));
  };

  const characterLimit = getCharacterLimit();
  const characterCount = content.length;
  const isOverLimit = characterLimit > 0 && characterCount > characterLimit;
  const canSubmit = content.trim() && selectedPlatforms.length > 0 && !isOverLimit && teamId;

  const handleSaveDraft = async () => {
    if (!canSubmit || !teamId) return;
    try {
      await createPost.mutateAsync({ teamId, content: content.trim(), platforms: selectedPlatforms });
      router.push('/dashboard/queue');
    } catch (error) {
      console.error('Failed to save draft:', error);
    }
  };

  const handlePublishNow = async () => {
    if (!canSubmit || !teamId) return;
    try {
      const post = await createPost.mutateAsync({ teamId, content: content.trim(), platforms: selectedPlatforms });
      const { publishPost } = await import('@/lib/api/posts');
      await publishPost(post.id);
      router.push('/dashboard/queue');
    } catch (error) {
      console.error('Failed to publish post:', error);
    }
  };

  const handleSchedule = async () => {
    if (!canSubmit || !teamId || !scheduleEnabled || !scheduledDate || !scheduledTime) return;
    try {
      const post = await createPost.mutateAsync({ teamId, content: content.trim(), platforms: selectedPlatforms });
      const scheduledAt = new Date(`${scheduledDate}T${scheduledTime}`).toISOString();
      await schedulePost.mutateAsync({ postId: post.id, data: { scheduledAt } });
      router.push('/dashboard/queue');
    } catch (error) {
      console.error('Failed to schedule post:', error);
    }
  };

  const isSubmitting = createPost.isPending || schedulePost.isPending;
  const today = new Date().toISOString().split('T')[0];

  if (!teamId) {
    return (
      <div className="max-w-4xl mx-auto py-12">
        <Card className="border-0 shadow-lg">
          <CardContent className="pt-12 pb-12 text-center">
            <Users className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">No Team Selected</h2>
            <p className="text-gray-600 dark:text-gray-400 mb-6">Please select a team to create posts.</p>
            <Button onClick={() => router.push('/dashboard/teams')}>Go to Teams</Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">Compose Post</h1>
        <p className="text-gray-600 dark:text-gray-400">Create and schedule your social media content</p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <Card className="border-0 shadow-lg">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <PenSquare className="h-5 w-5" />
                Post Content
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="content">What's on your mind?</Label>
                <Textarea
                  id="content"
                  placeholder="Share your thoughts..."
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  className="min-h-[200px] resize-none"
                  disabled={isSubmitting}
                />
                <div className="flex items-center justify-between text-sm">
                  <span className={isOverLimit ? 'text-red-600 dark:text-red-400 font-semibold' : 'text-gray-500 dark:text-gray-400'}>
                    {characterCount}{characterLimit > 0 && ` / ${characterLimit}`}
                  </span>
                  {isOverLimit && <Badge variant="destructive">Over character limit</Badge>}
                </div>
              </div>
              <div className="border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg p-8 text-center hover:border-indigo-400 dark:hover:border-indigo-600 transition-colors cursor-pointer">
                <ImagePlus className="h-12 w-12 text-gray-400 mx-auto mb-2" />
                <p className="text-sm text-gray-600 dark:text-gray-400">Click to upload media or drag and drop</p>
                <p className="text-xs text-gray-500 mt-1">PNG, JPG, GIF up to 10MB</p>
              </div>
            </CardContent>
          </Card>

          <Card className="border-0 shadow-lg">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Calendar className="h-5 w-5" />
                Schedule
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center space-x-2">
                <Checkbox id="schedule" checked={scheduleEnabled} onCheckedChange={(checked) => setScheduleEnabled(!!checked)} disabled={isSubmitting} />
                <Label htmlFor="schedule" className="cursor-pointer">Schedule for later</Label>
              </div>
              {scheduleEnabled && (
                <div className="grid grid-cols-2 gap-4 pt-2">
                  <div>
                    <Label htmlFor="date" className="text-sm">Date</Label>
                    <input type="date" id="date" value={scheduledDate} onChange={(e) => setScheduledDate(e.target.value)} min={today} className="w-full mt-1 px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-950 text-gray-900 dark:text-white" disabled={isSubmitting} />
                  </div>
                  <div>
                    <Label htmlFor="time" className="text-sm">Time</Label>
                    <input type="time" id="time" value={scheduledTime} onChange={(e) => setScheduledTime(e.target.value)} className="w-full mt-1 px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-950 text-gray-900 dark:text-white" disabled={isSubmitting} />
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          <div className="flex gap-3">
            {scheduleEnabled ? (
              <Button size="lg" className="flex-1 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700" disabled={!canSubmit || !scheduledDate || !scheduledTime || isSubmitting} onClick={handleSchedule}>
                {isSubmitting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Calendar className="mr-2 h-4 w-4" />}
                Schedule Post
              </Button>
            ) : (
              <Button size="lg" className="flex-1 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700" disabled={!canSubmit || isSubmitting} onClick={handlePublishNow}>
                {isSubmitting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Send className="mr-2 h-4 w-4" />}
                Publish Now
              </Button>
            )}
            <Button size="lg" variant="outline" className="flex-1" disabled={!canSubmit || isSubmitting} onClick={handleSaveDraft}>
              {isSubmitting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Save className="mr-2 h-4 w-4" />}
              Save Draft
            </Button>
          </div>
        </div>

        <div className="space-y-6">
          <Card className="border-0 shadow-lg">
            <CardHeader>
              <CardTitle>Select Platforms</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {platforms.map((platform) => (
                <div key={platform.id} className="flex items-center space-x-3 p-3 rounded-lg border border-gray-200 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
                  <Checkbox id={platform.id} checked={selectedPlatforms.includes(platform.id)} onCheckedChange={() => handlePlatformToggle(platform.id)} disabled={isSubmitting} />
                  <Label htmlFor={platform.id} className="flex items-center gap-2 cursor-pointer flex-1">
                    <platform.icon className={`h-5 w-5 ${platform.color}`} />
                    <span className="font-medium">{platform.name}</span>
                  </Label>
                  <span className="text-xs text-gray-500">{platform.limit} chars</span>
                </div>
              ))}
            </CardContent>
          </Card>

          <Card className="border-0 shadow-lg bg-gradient-to-br from-indigo-50 to-purple-50 dark:from-indigo-950 dark:to-purple-950">
            <CardContent className="p-6">
              <h3 className="font-semibold text-indigo-900 dark:text-indigo-100 mb-2">💡 Pro Tips</h3>
              <ul className="space-y-1.5 text-sm text-indigo-700 dark:text-indigo-300">
                <li>• Use hashtags to increase reach</li>
                <li>• Add images for better engagement</li>
                <li>• Schedule during peak hours</li>
                <li>• Keep messages concise and clear</li>
              </ul>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}