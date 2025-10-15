'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Calendar,
  Clock,
  Edit,
  Trash2,
  MoreVertical,
  Twitter,
  Facebook,
  Linkedin,
  List,
  Grid,
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

export default function QueuePage() {
  const [viewMode, setViewMode] = useState<'list' | 'calendar'>('list');

  // Mock data
  const posts = [
    {
      id: 1,
      content: 'Excited to announce our new feature launch! 🚀',
      platforms: ['twitter', 'linkedin'],
      scheduledFor: 'Tomorrow, 10:00 AM',
      status: 'scheduled',
    },
    {
      id: 2,
      content: 'Check out our latest blog post on social media trends...',
      platforms: ['facebook', 'linkedin'],
      scheduledFor: 'Jan 20, 2:00 PM',
      status: 'scheduled',
    },
    {
      id: 3,
      content: 'Behind the scenes: How we built this feature',
      platforms: ['twitter'],
      scheduledFor: 'Jan 22, 9:00 AM',
      status: 'draft',
    },
  ];

  const getPlatformIcon = (platform: string) => {
    const icons = {
      twitter: Twitter,
      facebook: Facebook,
      linkedin: Linkedin,
    };
    return icons[platform as keyof typeof icons] || Twitter;
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="space-y-2">
          <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
            Post Queue
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage your scheduled and draft posts
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={viewMode === 'list' ? 'default' : 'outline'}
            size="icon"
            onClick={() => setViewMode('list')}
          >
            <List className="h-4 w-4" />
          </Button>
          <Button
            variant={viewMode === 'calendar' ? 'default' : 'outline'}
            size="icon"
            onClick={() => setViewMode('calendar')}
          >
            <Grid className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="scheduled">
        <TabsList>
          <TabsTrigger value="scheduled">Scheduled ({posts.filter(p => p.status === 'scheduled').length})</TabsTrigger>
          <TabsTrigger value="drafts">Drafts ({posts.filter(p => p.status === 'draft').length})</TabsTrigger>
          <TabsTrigger value="published">Published</TabsTrigger>
        </TabsList>

        <TabsContent value="scheduled" className="space-y-4 mt-6">
          {posts
            .filter((post) => post.status === 'scheduled')
            .map((post) => (
              <Card key={post.id} className="border-0 shadow-md hover:shadow-lg transition-shadow">
                <CardContent className="p-6">
                  <div className="flex items-start justify-between">
                    <div className="flex-1 space-y-3">
                      {/* Platforms */}
                      <div className="flex items-center gap-2">
                        {post.platforms.map((platform) => {
                          const Icon = getPlatformIcon(platform);
                          return (
                            <div
                              key={platform}
                              className="h-8 w-8 rounded-lg bg-gradient-to-br from-gray-100 to-gray-200 dark:from-gray-800 dark:to-gray-900 flex items-center justify-center"
                            >
                              <Icon className="h-4 w-4" />
                            </div>
                          );
                        })}
                      </div>

                      {/* Content */}
                      <p className="text-gray-900 dark:text-white font-medium">
                        {post.content}
                      </p>

                      {/* Schedule Info */}
                      <div className="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400">
                        <div className="flex items-center gap-1.5">
                          <Clock className="h-4 w-4" />
                          <span>{post.scheduledFor}</span>
                        </div>
                        <Badge variant="secondary" className="capitalize">
                          {post.status}
                        </Badge>
                      </div>
                    </div>

                    {/* Actions */}
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem>
                          <Edit className="mr-2 h-4 w-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuItem>
                          <Calendar className="mr-2 h-4 w-4" />
                          Reschedule
                        </DropdownMenuItem>
                        <DropdownMenuItem className="text-red-600">
                          <Trash2 className="mr-2 h-4 w-4" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </CardContent>
              </Card>
            ))}
        </TabsContent>

        <TabsContent value="drafts" className="mt-6">
          <p className="text-center text-gray-500 py-12">
            Draft posts will appear here
          </p>
        </TabsContent>

        <TabsContent value="published" className="mt-6">
          <p className="text-center text-gray-500 py-12">
            Published posts will appear here
          </p>
        </TabsContent>
      </Tabs>
    </div>
  );
}