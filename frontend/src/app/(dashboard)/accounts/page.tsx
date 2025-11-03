// frontend/src/app/(dashboard)/accounts/page.tsx
'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { socialOAuthService, type ConnectedAccount } from '@/services/social-oauth.service';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { toast } from 'sonner';
import { 
  Plus, 
  Trash2, 
  RefreshCw, 
  AlertCircle, 
  CheckCircle2, 
  Loader2,
  Link2,
  ExternalLink,
  Shield,
  Clock,
  MoreVertical,
  Grid3x3,
  List,
  Activity
} from 'lucide-react';
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
// import { Toggle } from '@/components/ui/toggle';
import { cn } from '@/lib/utils';

// Platform Icons as Components
const FacebookIcon = () => (
  <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
    <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
  </svg>
);

const TwitterIcon = () => (
  <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
  </svg>
);

const LinkedInIcon = () => (
  <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
    <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
  </svg>
);

const InstagramIcon = () => (
  <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
    <path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zm0-2.163c-3.259 0-3.667.014-4.947.072-4.358.2-6.78 2.618-6.98 6.98-.059 1.281-.073 1.689-.073 4.948 0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98 1.281.058 1.689.072 4.948.072 3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98-1.281-.059-1.69-.073-4.949-.073zM5.838 12a6.162 6.162 0 1 1 12.324 0 6.162 6.162 0 0 1-12.324 0zM12 16a4 4 0 1 1 0-8 4 4 0 0 1 0 8zm4.965-10.405a1.44 1.44 0 1 1 2.881.001 1.44 1.44 0 0 1-2.881-.001z"/>
  </svg>
);

// Platform configurations
const platformConfig = {
  facebook: {
    name: 'Facebook',
    icon: FacebookIcon,
    color: 'bg-blue-600',
    textColor: 'text-blue-400',
    bgLight: 'bg-blue-500/10',
    borderColor: 'border-blue-500/20',
    description: 'Connect Facebook Pages and Groups'
  },
  twitter: {
    name: 'X (Twitter)',
    icon: TwitterIcon,
    color: 'bg-white',
    textColor: 'text-white',
    bgLight: 'bg-white/10',
    borderColor: 'border-white/20',
    description: 'Post to your X timeline'
  },
  linkedin: {
    name: 'LinkedIn',
    icon: LinkedInIcon,
    color: 'bg-blue-600',
    textColor: 'text-blue-400',
    bgLight: 'bg-blue-500/10',
    borderColor: 'border-blue-500/20',
    description: 'Share with your professional network'
  },
  instagram: {
    name: 'Instagram',
    icon: InstagramIcon,
    color: 'bg-gradient-to-br from-purple-600 to-pink-600',
    textColor: 'text-purple-400',
    bgLight: 'bg-purple-500/10',
    borderColor: 'border-purple-500/20',
    description: 'Post photos and stories'
  },
};

export default function AccountsPage() {
  const queryClient = useQueryClient();
  const [selectedPlatform, setSelectedPlatform] = useState<string | null>(null);
  const [accountToDisconnect, setAccountToDisconnect] = useState<ConnectedAccount | null>(null);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  // Get teamId from localStorage
  const teamId = localStorage.getItem('socialqueue_selected_team_id') || '';

  // Fetch connected accounts
  const { data: accounts = [], isLoading, refetch } = useQuery({
    queryKey: ['connected-accounts', teamId],
    queryFn: () => {
      if (!teamId) {
        console.warn('No team ID available');
        return Promise.resolve([]);
      }
      return socialOAuthService.getConnectedAccounts(teamId);
    },
    enabled: !!teamId,
  });

  // Connect account mutation
  const connectMutation = useMutation({
    mutationFn: (platform: string) => {
      if (!teamId) {
        throw new Error('Please select a team first');
      }
      return socialOAuthService.initiateOAuth(platform, teamId);
    },
    onError: (error: any) => {
      toast.error('Failed to connect account', {
        description: error.message,
      });
    },
  });

  // Disconnect account mutation
  const disconnectMutation = useMutation({
    mutationFn: async (accountId: string) => {
      await socialOAuthService.disconnectAccount(accountId);
      return accountId;
    },
    onSuccess: () => {
      toast.success('Account disconnected successfully');
      queryClient.invalidateQueries({ queryKey: ['connected-accounts'] });
    },
    onError: (error: any) => {
      toast.error('Failed to disconnect account', {
        description: error.message || 'Please try again',
      });
    },
  });

  // Refresh tokens mutation
  const refreshMutation = useMutation({
    mutationFn: (accountId: string) => socialOAuthService.refreshTokens(accountId),
    onSuccess: () => {
      toast.success('Tokens refreshed successfully');
      refetch();
    },
    onError: () => {
      toast.error('Failed to refresh tokens');
    },
  });

  const handleConnect = (platform: string) => {
    if (!teamId) {
      toast.error('Please select a team first');
      return;
    }
    setSelectedPlatform(platform);
    connectMutation.mutate(platform);
  };

  const handleDisconnect = (account: ConnectedAccount) => {
    setAccountToDisconnect(account);
  };

  const confirmDisconnect = () => {
    if (accountToDisconnect) {
      disconnectMutation.mutate(accountToDisconnect.id);
      setAccountToDisconnect(null);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return (
          <Badge variant="outline" className="bg-green-500/10 text-green-400 border-green-500/20">
            <CheckCircle2 className="h-3 w-3 mr-1" />
            Active
          </Badge>
        );
      case 'expired':
        return (
          <Badge variant="outline" className="bg-yellow-500/10 text-yellow-400 border-yellow-500/20">
            <Clock className="h-3 w-3 mr-1" />
            Expired
          </Badge>
        );
      case 'revoked':
        return (
          <Badge variant="outline" className="bg-red-500/10 text-red-400 border-red-500/20">
            <AlertCircle className="h-3 w-3 mr-1" />
            Revoked
          </Badge>
        );
      default:
        return null;
    }
  };

  const connectedPlatforms = accounts.map(acc => acc.platform);

  return (
    <div className="space-y-6">
      {/* Header Section */}
      <div className="flex flex-col gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Connected Accounts</h1>
          <p className="text-muted-foreground mt-2">
            Manage your social media connections and publish content across platforms
          </p>
        </div>

        {/* Stats Cards */}
        <div className="grid gap-4 md:grid-cols-3">
          <Card className="bg-card/50 border-border/50">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Total Accounts</p>
                  <p className="text-2xl font-bold">{accounts.length}</p>
                </div>
                <div className="p-3 bg-primary/10 rounded-lg">
                  <Link2 className="h-5 w-5 text-primary" />
                </div>
              </div>
            </CardContent>
          </Card>
          
          <Card className="bg-card/50 border-border/50">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Active</p>
                  <p className="text-2xl font-bold text-green-400">
                    {accounts.filter(a => a.status === 'active').length}
                  </p>
                </div>
                <div className="p-3 bg-green-500/10 rounded-lg">
                  <CheckCircle2 className="h-5 w-5 text-green-400" />
                </div>
              </div>
            </CardContent>
          </Card>
          
          <Card className="bg-card/50 border-border/50">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Needs Attention</p>
                  <p className="text-2xl font-bold text-yellow-400">
                    {accounts.filter(a => a.status === 'expired').length}
                  </p>
                </div>
                <div className="p-3 bg-yellow-500/10 rounded-lg">
                  <AlertCircle className="h-5 w-5 text-yellow-400" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Platform Selection */}
      <Card className="bg-card/50 border-border/50">
        <CardHeader>
          <CardTitle>Add New Account</CardTitle>
          <CardDescription>
            Connect your social media accounts to start publishing content
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Object.entries(platformConfig).map(([key, platform]) => {
              const isConnected = connectedPlatforms.includes(key);
              const isAvailable = key !== 'instagram';
              const Icon = platform.icon;
              
              return (
                <button
                  key={key}
                  onClick={() => isAvailable && !isConnected && handleConnect(key)}
                  disabled={!isAvailable || isConnected || connectMutation.isPending}
                  className={cn(
                    "relative p-6 rounded-xl border-2 transition-all duration-200",
                    "hover:scale-105 disabled:hover:scale-100",
                    "disabled:opacity-50 disabled:cursor-not-allowed",
                    "bg-card/50",
                    isConnected ? platform.borderColor : "border-border/50",
                    isConnected && platform.bgLight,
                    !isConnected && isAvailable && "hover:border-primary/50"
                  )}
                >
                  {isConnected && (
                    <div className="absolute top-3 right-3">
                      <Badge variant="secondary" className="text-xs bg-primary/20 border-0">
                        Connected
                      </Badge>
                    </div>
                  )}
                  
                  {!isAvailable && (
                    <div className="absolute top-3 right-3">
                      <Badge variant="outline" className="text-xs">
                        Coming Soon
                      </Badge>
                    </div>
                  )}
                  
                  <div className="flex flex-col items-center space-y-3">
                    <div className={cn(
                      "p-3 rounded-lg",
                      isConnected ? platform.color : "bg-muted",
                      isConnected ? "text-white" : "text-muted-foreground"
                    )}>
                      <Icon />
                    </div>
                    
                    <div className="text-center">
                      <p className="font-semibold">{platform.name}</p>
                      <p className="text-xs text-muted-foreground mt-1">
                        {platform.description}
                      </p>
                    </div>
                  </div>
                  
                  {connectMutation.isPending && selectedPlatform === key && (
                    <div className="absolute inset-0 bg-background/80 rounded-xl flex items-center justify-center">
                      <Loader2 className="h-6 w-6 animate-spin" />
                    </div>
                  )}
                </button>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* Connected Accounts List */}
      <Card className="bg-card/50 border-border/50">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Your Connected Accounts</CardTitle>
              <CardDescription>
                Manage permissions and settings for connected accounts
              </CardDescription>
            </div>
            <div className="flex items-center gap-2">
              {/* View Mode Toggle */}
              <div className="flex items-center rounded-lg border border-border/50 p-1">
                <Button
                  variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
                  size="sm"
                  onClick={() => setViewMode('grid')}
                  className="h-8 px-3"
                >
                  <Grid3x3 className="h-4 w-4" />
                </Button>
                <Button
                  variant={viewMode === 'list' ? 'secondary' : 'ghost'}
                  size="sm"
                  onClick={() => setViewMode('list')}
                  className="h-8 px-3"
                >
                  <List className="h-4 w-4" />
                </Button>
              </div>
              
              <Button
                variant="outline"
                size="sm"
                onClick={() => refetch()}
                disabled={isLoading}
                className="border-border/50"
              >
                {isLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <RefreshCw className="h-4 w-4" />
                )}
                <span className="ml-2">Refresh</span>
              </Button>
            </div>
          </div>
        </CardHeader>
        
        <CardContent>
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : accounts.length > 0 ? (
            <div className={cn(
              viewMode === 'grid' 
                ? "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
                : "space-y-4"
            )}>
              {accounts.map((account) => {
                const platform = platformConfig[account.platform as keyof typeof platformConfig];
                const Icon = platform?.icon || Link2;
                
                return viewMode === 'grid' ? (
                  // Grid View
                  <Card key={account.id} className="bg-card border-border/50">
                    <CardContent className="p-6">
                      <div className="flex items-start justify-between mb-4">
                        <div className={cn(
                          "p-2 rounded-lg",
                          platform?.bgLight,
                          platform?.textColor
                        )}>
                          <Icon />
                        </div>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-8 w-8">
                              <MoreVertical className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="w-48">
                            <DropdownMenuLabel>Actions</DropdownMenuLabel>
                            <DropdownMenuSeparator />
                            
                            {account.profileUrl && (
                              <DropdownMenuItem asChild>
                                <a
                                  href={account.profileUrl}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  className="cursor-pointer"
                                >
                                  <ExternalLink className="h-4 w-4 mr-2" />
                                  View Profile
                                </a>
                              </DropdownMenuItem>
                            )}
                            
                            {account.status === 'expired' && (
                              <DropdownMenuItem
                                onClick={() => refreshMutation.mutate(account.id)}
                                disabled={refreshMutation.isPending}
                              >
                                <RefreshCw className="h-4 w-4 mr-2" />
                                Refresh Token
                              </DropdownMenuItem>
                            )}
                            
                            <DropdownMenuItem
                              onClick={() => handleDisconnect(account)}
                              className="text-red-400 focus:text-red-400"
                            >
                              <Trash2 className="h-4 w-4 mr-2" />
                              Disconnect
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                      
                      <div className="space-y-2">
                        <p className="font-medium">
                          {account.displayName || `${platform?.name} User`}
                        </p>
                        <p className="text-sm text-muted-foreground">
                          @{account.username || `user_${account.platform}`}
                        </p>
                        {getStatusBadge(account.status)}
                        <p className="text-xs text-muted-foreground mt-2">
                          Connected {new Date(account.connectedAt).toLocaleDateString()}
                        </p>
                      </div>
                    </CardContent>
                  </Card>
                ) : (
                  // List View
                  <div
                    key={account.id}
                    className="flex items-center justify-between p-4 rounded-lg border border-border/50 bg-card/50 hover:bg-accent/10 transition-colors"
                  >
                    <div className="flex items-center gap-4">
                      <div className={cn(
                        "p-2 rounded-lg",
                        platform?.bgLight,
                        platform?.textColor
                      )}>
                        <Icon />
                      </div>
                      
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <p className="font-medium">
                            {account.displayName || `${platform?.name} User`}
                          </p>
                          {getStatusBadge(account.status)}
                        </div>
                        <p className="text-sm text-muted-foreground">
                          @{account.username || `user_${account.platform}`}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          Connected {new Date(account.connectedAt).toLocaleDateString()}
                        </p>
                      </div>
                    </div>

                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" className="w-48">
                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        
                        {account.profileUrl && (
                          <DropdownMenuItem asChild>
                            <a
                              href={account.profileUrl}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="cursor-pointer"
                            >
                              <ExternalLink className="h-4 w-4 mr-2" />
                              View Profile
                            </a>
                          </DropdownMenuItem>
                        )}
                        
                        {account.status === 'expired' && (
                          <DropdownMenuItem
                            onClick={() => refreshMutation.mutate(account.id)}
                            disabled={refreshMutation.isPending}
                          >
                            <RefreshCw className="h-4 w-4 mr-2" />
                            Refresh Token
                          </DropdownMenuItem>
                        )}
                        
                        <DropdownMenuItem
                          onClick={() => handleDisconnect(account)}
                          className="text-red-400 focus:text-red-400"
                        >
                          <Trash2 className="h-4 w-4 mr-2" />
                          Disconnect
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="text-center py-12">
              <div className="mx-auto w-16 h-16 rounded-full bg-muted flex items-center justify-center mb-4">
                <Link2 className="h-8 w-8 text-muted-foreground" />
              </div>
              <p className="text-lg font-medium">No accounts connected</p>
              <p className="text-sm text-muted-foreground mt-2">
                Connect your first social media account to get started
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Disconnect Confirmation Dialog */}
      <AlertDialog open={!!accountToDisconnect} onOpenChange={() => setAccountToDisconnect(null)}>
        <AlertDialogContent className="bg-card border-border/50">
          <AlertDialogHeader>
            <AlertDialogTitle>Disconnect Account</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to disconnect {accountToDisconnect?.displayName}? 
              You'll need to reconnect to publish to this account.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel className="bg-secondary">Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDisconnect}
              className="bg-red-600 hover:bg-red-700"
            >
              Disconnect
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}