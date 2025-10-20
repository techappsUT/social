// frontend/src/app/(dashboard)/dashboard/page.tsx
// Beautiful Modern Dashboard Home Page

'use client';

import Link from 'next/link';
import { useAuth } from '@/providers/auth-provider';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  TrendingUp,
  TrendingDown,
  Calendar,
  Eye,
  Heart,
  MessageCircle,
  Share2,
  PenSquare,
  Link2,
  BarChart3,
  ArrowRight,
  Clock,
  CheckCircle2,
} from 'lucide-react';
import { TeamSwitcher } from '@/components/team-switcher';

export default function DashboardPage() {
  const { user } = useAuth();

  // Mock data - replace with real API calls
  const stats = [
    {
      name: 'Scheduled Posts',
      value: '12',
      change: '+4.5%',
      trend: 'up' as const,
      icon: Calendar,
      color: 'from-blue-500 to-cyan-500',
    },
    {
      name: 'Total Impressions',
      value: '24.5K',
      change: '+12.3%',
      trend: 'up' as const,
      icon: Eye,
      color: 'from-purple-500 to-pink-500',
    },
    {
      name: 'Engagement Rate',
      value: '3.8%',
      change: '+0.5%',
      trend: 'up' as const,
      icon: Heart,
      color: 'from-orange-500 to-red-500',
    },
    {
      name: 'Connected Accounts',
      value: '5',
      change: '+2',
      trend: 'up' as const,
      icon: Link2,
      color: 'from-green-500 to-emerald-500',
    },
  ];

  const recentPosts = [
    {
      id: 1,
      content: 'Just launched our new feature! Check it out 🚀',
      platform: 'Twitter',
      status: 'published',
      scheduledFor: '2 hours ago',
      engagement: { likes: 245, comments: 32, shares: 18 },
    },
    {
      id: 2,
      content: 'Behind the scenes of our product development process...',
      platform: 'LinkedIn',
      status: 'scheduled',
      scheduledFor: 'Tomorrow at 10:00 AM',
      engagement: null,
    },
    {
      id: 3,
      content: 'New blog post: 10 Tips for Social Media Management',
      platform: 'Facebook',
      status: 'scheduled',
      scheduledFor: 'Jan 20 at 2:00 PM',
      engagement: null,
    },
  ];

  const quickActions = [
    {
      title: 'Create Post',
      description: 'Compose and schedule a new post',
      icon: PenSquare,
      href: '/dashboard/compose',
      color: 'from-indigo-500 to-purple-600',
    },
    {
      title: 'View Queue',
      description: 'Manage your scheduled posts',
      icon: Calendar,
      href: '/dashboard/queue',
      color: 'from-blue-500 to-cyan-500',
    },
    {
      title: 'Connect Account',
      description: 'Add a new social media account',
      icon: Link2,
      href: '/dashboard/accounts',
      color: 'from-green-500 to-emerald-500',
    },
    {
      title: 'View Analytics',
      description: 'Check your performance metrics',
      icon: BarChart3,
      href: '/dashboard/analytics',
      color: 'from-orange-500 to-red-500',
    },
  ];

  return (
    <div className="space-y-8">
      {/* Welcome Header */}
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
          Welcome back, {user?.firstName}! 👋
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Here's what's happening with your social media today.
        </p>
      </div>

      <TeamSwitcher />

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <Card
            key={stat.name}
            className="relative overflow-hidden border-0 shadow-lg hover:shadow-xl transition-shadow"
          >
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-gray-600 dark:text-gray-400">
                {stat.name}
              </CardTitle>
              <div
                className={`h-10 w-10 rounded-lg bg-gradient-to-br ${stat.color} flex items-center justify-center shadow-md`}
              >
                <stat.icon className="h-5 w-5 text-white" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {stat.value}
              </div>
              <div className="flex items-center text-xs mt-1">
                {stat.trend === 'up' ? (
                  <TrendingUp className="h-4 w-4 text-green-500 mr-1" />
                ) : (
                  <TrendingDown className="h-4 w-4 text-red-500 mr-1" />
                )}
                <span
                  className={
                    stat.trend === 'up' ? 'text-green-600' : 'text-red-600'
                  }
                >
                  {stat.change}
                </span>
                <span className="text-gray-500 dark:text-gray-400 ml-1">
                  from last month
                </span>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Quick Actions */}
      <div>
        <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-4">
          Quick Actions
        </h2>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {quickActions.map((action) => (
            <Link key={action.title} href={action.href}>
              <Card className="group relative overflow-hidden border-0 shadow-md hover:shadow-xl transition-all cursor-pointer h-full">
                <div
                  className={`absolute inset-0 bg-gradient-to-br ${action.color} opacity-0 group-hover:opacity-5 transition-opacity`}
                />
                <CardHeader className="space-y-3">
                  <div
                    className={`h-12 w-12 rounded-lg bg-gradient-to-br ${action.color} flex items-center justify-center shadow-md group-hover:scale-110 transition-transform`}
                  >
                    <action.icon className="h-6 w-6 text-white" />
                  </div>
                  <div>
                    <CardTitle className="text-base group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                      {action.title}
                    </CardTitle>
                    <CardDescription className="text-xs mt-1">
                      {action.description}
                    </CardDescription>
                  </div>
                </CardHeader>
              </Card>
            </Link>
          ))}
        </div>
      </div>

      {/* Recent Posts */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold text-gray-900 dark:text-white">
            Recent Activity
          </h2>
          <Link href="/dashboard/queue">
            <Button variant="ghost" size="sm" className="gap-2">
              View All
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>

        <div className="space-y-4">
          {recentPosts.map((post) => (
            <Card
              key={post.id}
              className="border-0 shadow-md hover:shadow-lg transition-shadow"
            >
              <CardContent className="p-6">
                <div className="flex items-start justify-between">
                  <div className="flex-1 space-y-3">
                    {/* Post Header */}
                    <div className="flex items-center gap-3">
                      <Badge
                        variant={
                          post.status === 'published' ? 'default' : 'secondary'
                        }
                        className="capitalize"
                      >
                        {post.status === 'published' ? (
                          <CheckCircle2 className="h-3 w-3 mr-1" />
                        ) : (
                          <Clock className="h-3 w-3 mr-1" />
                        )}
                        {post.status}
                      </Badge>
                      <span className="text-sm text-gray-600 dark:text-gray-400">
                        {post.platform}
                      </span>
                      <span className="text-sm text-gray-500 dark:text-gray-500">
                        •
                      </span>
                      <span className="text-sm text-gray-600 dark:text-gray-400">
                        {post.scheduledFor}
                      </span>
                    </div>

                    {/* Post Content */}
                    <p className="text-gray-900 dark:text-white font-medium">
                      {post.content}
                    </p>

                    {/* Engagement Stats */}
                    {post.engagement && (
                      <div className="flex items-center gap-6 pt-2">
                        <div className="flex items-center gap-1.5 text-sm text-gray-600 dark:text-gray-400">
                          <Heart className="h-4 w-4" />
                          <span>{post.engagement.likes}</span>
                        </div>
                        <div className="flex items-center gap-1.5 text-sm text-gray-600 dark:text-gray-400">
                          <MessageCircle className="h-4 w-4" />
                          <span>{post.engagement.comments}</span>
                        </div>
                        <div className="flex items-center gap-1.5 text-sm text-gray-600 dark:text-gray-400">
                          <Share2 className="h-4 w-4" />
                          <span>{post.engagement.shares}</span>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Actions */}
                  <div>
                    <Button variant="ghost" size="sm">
                      View
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* Empty State CTA (if no posts) */}
      {recentPosts.length === 0 && (
        <Card className="border-2 border-dashed border-gray-300 dark:border-gray-700">
          <CardContent className="flex flex-col items-center justify-center py-12 text-center">
            <div className="h-16 w-16 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center mb-4">
              <PenSquare className="h-8 w-8 text-white" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              No posts yet
            </h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4 max-w-sm">
              Get started by creating your first social media post
            </p>
            <Link href="/dashboard/compose">
              <Button className="gap-2 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700">
                <PenSquare className="h-4 w-4" />
                Create Your First Post
              </Button>
            </Link>
          </CardContent>
        </Card>
      )}
    </div>
  );
}