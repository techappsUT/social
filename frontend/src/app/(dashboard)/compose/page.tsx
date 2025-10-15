'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import {
  PenSquare,
  Calendar,
  Image as ImageIcon,
  Send,
  Save,
  Twitter,
  Facebook,
  Linkedin,
} from 'lucide-react';

export default function ComposePage() {
  const [content, setContent] = useState('');
  const [selectedPlatforms, setSelectedPlatforms] = useState<string[]>([]);

  const platforms = [
    { id: 'twitter', name: 'Twitter', icon: Twitter, limit: 280, color: 'text-blue-400' },
    { id: 'facebook', name: 'Facebook', icon: Facebook, limit: 5000, color: 'text-blue-600' },
    { id: 'linkedin', name: 'LinkedIn', icon: Linkedin, limit: 3000, color: 'text-blue-700' },
  ];

  const handlePlatformToggle = (platformId: string) => {
    setSelectedPlatforms((prev) =>
      prev.includes(platformId)
        ? prev.filter((id) => id !== platformId)
        : [...prev, platformId]
    );
  };

  const getCharacterLimit = () => {
    if (selectedPlatforms.length === 0) return 0;
    return Math.min(
      ...selectedPlatforms.map(
        (id) => platforms.find((p) => p.id === id)?.limit || 0
      )
    );
  };

  const characterLimit = getCharacterLimit();
  const characterCount = content.length;
  const isOverLimit = characterLimit > 0 && characterCount > characterLimit;

  return (
    <div className="max-w-4xl mx-auto space-y-8">
      {/* Header */}
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
          Compose Post
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Create and schedule your social media content
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Main Content Area */}
        <div className="lg:col-span-2 space-y-6">
          {/* Post Content */}
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
                />
                <div className="flex items-center justify-between text-sm">
                  <span
                    className={`${
                      isOverLimit
                        ? 'text-red-600 dark:text-red-400 font-semibold'
                        : 'text-gray-500 dark:text-gray-400'
                    }`}
                  >
                    {characterCount}
                    {characterLimit > 0 && ` / ${characterLimit}`}
                  </span>
                  {isOverLimit && (
                    <Badge variant="destructive">Over character limit</Badge>
                  )}
                </div>
              </div>

              {/* Media Upload */}
              <div className="border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg p-8 text-center hover:border-indigo-400 dark:hover:border-indigo-600 transition-colors cursor-pointer">
                <ImageIcon className="h-12 w-12 text-gray-400 mx-auto mb-2" />
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Click to upload media or drag and drop
                </p>
                <p className="text-xs text-gray-500 mt-1">
                  PNG, JPG, GIF up to 10MB
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Schedule Section */}
          <Card className="border-0 shadow-lg">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Calendar className="h-5 w-5" />
                Schedule
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center space-x-2">
                <Checkbox id="schedule" />
                <Label htmlFor="schedule" className="cursor-pointer">
                  Schedule for later
                </Label>
              </div>
              {/* Date/Time picker would go here */}
              <p className="text-sm text-gray-500 dark:text-gray-400">
                Date and time picker coming soon...
              </p>
            </CardContent>
          </Card>

          {/* Action Buttons */}
          <div className="flex gap-3">
            <Button
              size="lg"
              className="flex-1 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700"
              disabled={!content.trim() || selectedPlatforms.length === 0 || isOverLimit}
            >
              <Send className="mr-2 h-4 w-4" />
              Publish Now
            </Button>
            <Button size="lg" variant="outline" className="flex-1">
              <Save className="mr-2 h-4 w-4" />
              Save Draft
            </Button>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Platform Selection */}
          <Card className="border-0 shadow-lg">
            <CardHeader>
              <CardTitle>Select Platforms</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {platforms.map((platform) => (
                <div
                  key={platform.id}
                  className="flex items-center space-x-3 p-3 rounded-lg border border-gray-200 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
                >
                  <Checkbox
                    id={platform.id}
                    checked={selectedPlatforms.includes(platform.id)}
                    onCheckedChange={() => handlePlatformToggle(platform.id)}
                  />
                  <Label
                    htmlFor={platform.id}
                    className="flex items-center gap-2 cursor-pointer flex-1"
                  >
                    <platform.icon className={`h-5 w-5 ${platform.color}`} />
                    <span className="font-medium">{platform.name}</span>
                  </Label>
                  <span className="text-xs text-gray-500">
                    {platform.limit} chars
                  </span>
                </div>
              ))}
            </CardContent>
          </Card>

          {/* Tips */}
          <Card className="border-0 shadow-lg bg-gradient-to-br from-indigo-50 to-purple-50 dark:from-indigo-950 dark:to-purple-950">
            <CardContent className="p-6">
              <h3 className="font-semibold text-indigo-900 dark:text-indigo-100 mb-2">
                💡 Pro Tips
              </h3>
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