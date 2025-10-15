'use client';

import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Link2,
  Twitter,
  Facebook,
  Linkedin,
  Instagram,
  CheckCircle,
  AlertCircle,
  Plus,
} from 'lucide-react';

export default function AccountsPage() {
  // Mock data
  const connectedAccounts = [
    {
      id: 1,
      platform: 'Twitter',
      handle: '@johndoe',
      icon: Twitter,
      status: 'connected',
      lastSync: '2 hours ago',
      color: 'text-blue-400',
    },
    {
      id: 2,
      platform: 'LinkedIn',
      handle: 'John Doe',
      icon: Linkedin,
      status: 'connected',
      lastSync: '1 hour ago',
      color: 'text-blue-700',
    },
  ];

  const availablePlatforms = [
    {
      name: 'Facebook',
      icon: Facebook,
      description: 'Connect your Facebook pages',
      color: 'from-blue-600 to-blue-700',
    },
    {
      name: 'Instagram',
      icon: Instagram,
      description: 'Connect your Instagram business account',
      color: 'from-purple-500 to-pink-500',
    },
  ];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
          Connected Accounts
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Manage your social media account connections
        </p>
      </div>

      {/* Connected Accounts */}
      <div>
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
          Active Connections
        </h2>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {connectedAccounts.map((account) => (
            <Card key={account.id} className="border-0 shadow-lg">
              <CardContent className="p-6">
                <div className="flex items-start justify-between mb-4">
                  <div
                    className={`h-12 w-12 rounded-lg bg-gradient-to-br from-gray-100 to-gray-200 dark:from-gray-800 dark:to-gray-900 flex items-center justify-center`}
                  >
                    <account.icon className={`h-6 w-6 ${account.color}`} />
                  </div>
                  <Badge
                    variant={account.status === 'connected' ? 'default' : 'destructive'}
                    className="flex items-center gap-1"
                  >
                    {account.status === 'connected' ? (
                      <CheckCircle className="h-3 w-3" />
                    ) : (
                      <AlertCircle className="h-3 w-3" />
                    )}
                    {account.status}
                  </Badge>
                </div>

                <div className="space-y-2">
                  <h3 className="font-semibold text-gray-900 dark:text-white">
                    {account.platform}
                  </h3>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    {account.handle}
                  </p>
                  <p className="text-xs text-gray-500">
                    Last synced {account.lastSync}
                  </p>
                </div>

                <div className="mt-4 flex gap-2">
                  <Button variant="outline" size="sm" className="flex-1">
                    Refresh
                  </Button>
                  <Button variant="outline" size="sm" className="flex-1 text-red-600 hover:text-red-700">
                    Disconnect
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* Available Platforms */}
      <div>
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
          Add New Connection
        </h2>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {availablePlatforms.map((platform) => (
            <Card key={platform.name} className="border-0 shadow-lg hover:shadow-xl transition-shadow">
              <CardContent className="p-6">
                <div
                  className={`h-12 w-12 rounded-lg bg-gradient-to-br ${platform.color} flex items-center justify-center mb-4`}
                >
                  <platform.icon className="h-6 w-6 text-white" />
                </div>

                <h3 className="font-semibold text-gray-900 dark:text-white mb-2">
                  {platform.name}
                </h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                  {platform.description}
                </p>

                <Button className="w-full gap-2">
                  <Plus className="h-4 w-4" />
                  Connect {platform.name}
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
}