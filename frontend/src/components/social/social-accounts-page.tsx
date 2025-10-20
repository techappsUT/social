// path: frontend/src/components/social/social-accounts-page.tsx
import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
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
  Twitter, 
  Facebook, 
  Linkedin, 
  Instagram, 
  Link2Off, 
  RefreshCw,
  AlertCircle,
  CheckCircle,
  Loader2,
  Plus
} from 'lucide-react';
import {
  useSocialAccounts,
  useDisconnectAccount,
  useRefreshTokens,
  useInitiateOAuth,
  useAccountsNeedingAttention,
} from '@/hooks/useSocial';
import type { SocialPlatform, SocialAccountDTO } from '@/types/social';
import { 
  getPlatformName, 
  getPlatformColor, 
  getStatusLabel,
  getStatusColor,
  needsReconnection,
  canPublish
} from '@/types/social';

// ============================================================================
// PLATFORM CONFIGS
// ============================================================================

const PLATFORMS: Array<{
  value: SocialPlatform;
  icon: React.ComponentType<any>;
  description: string;
}> = [
  { value: 'twitter', icon: Twitter, description: 'Connect your Twitter/X account' },
  { value: 'facebook', icon: Facebook, description: 'Connect your Facebook pages' },
  { value: 'linkedin', icon: Linkedin, description: 'Connect your LinkedIn profile' },
  { value: 'instagram', icon: Instagram, description: 'Connect your Instagram account' },
];

// ============================================================================
// COMPONENT
// ============================================================================

interface SocialAccountsPageProps {
  teamId: string;
}

export function SocialAccountsPage({ teamId }: SocialAccountsPageProps) {
  const [disconnectingId, setDisconnectingId] = useState<string | null>(null);
  
  // Fetch accounts
  const { data: accounts = [], isLoading } = useSocialAccounts(teamId);
  const { needsAttention, count: issuesCount } = useAccountsNeedingAttention(teamId);
  
  // Mutations
  const initiateOAuth = useInitiateOAuth();
  const disconnectAccount = useDisconnectAccount();
  const refreshTokens = useRefreshTokens();

  // Handle connect
  const handleConnect = async (platform: SocialPlatform) => {
    try {
      await initiateOAuth.mutateAsync(platform);
      // OAuth popup will open, callback will handle connection
    } catch (error) {
      console.error('Failed to initiate OAuth:', error);
    }
  };

  // Handle disconnect
  const handleDisconnect = async (accountId: string) => {
    try {
      await disconnectAccount.mutateAsync(accountId);
      setDisconnectingId(null);
    } catch (error) {
      console.error('Failed to disconnect:', error);
    }
  };

  // Handle token refresh
  const handleRefresh = async (accountId: string) => {
    try {
      await refreshTokens.mutateAsync(accountId);
    } catch (error) {
      console.error('Failed to refresh tokens:', error);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold mb-2">Social Accounts</h1>
        <p className="text-gray-600">
          Connect and manage your social media accounts
        </p>
      </div>

      {/* Issues Alert */}
      {issuesCount > 0 && (
        <Card className="border-orange-200 bg-orange-50">
          <CardContent className="pt-6">
            <div className="flex items-start gap-3">
              <AlertCircle className="h-5 w-5 text-orange-600 mt-0.5" />
              <div>
                <h3 className="font-semibold text-orange-900">
                  {issuesCount} account{issuesCount > 1 ? 's' : ''} need attention
                </h3>
                <p className="text-sm text-orange-700 mt-1">
                  Some accounts have expired tokens or connection issues
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Connected Accounts */}
      {accounts.length > 0 && (
        <div>
          <h2 className="text-xl font-semibold mb-4">Connected Accounts</h2>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {accounts.map((account) => (
              <AccountCard
                key={account.id}
                account={account}
                onDisconnect={() => setDisconnectingId(account.id)}
                onRefresh={() => handleRefresh(account.id)}
                isRefreshing={refreshTokens.isPending}
              />
            ))}
          </div>
        </div>
      )}

      {/* Available Platforms */}
      <div>
        <h2 className="text-xl font-semibold mb-4">
          {accounts.length > 0 ? 'Add More Accounts' : 'Connect Your First Account'}
        </h2>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {PLATFORMS.map((platform) => {
            const Icon = platform.icon;
            const connectedCount = accounts.filter(
              acc => acc.platform === platform.value
            ).length;

            return (
              <Card 
                key={platform.value}
                className="hover:shadow-md transition-shadow cursor-pointer"
                onClick={() => handleConnect(platform.value)}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <Icon 
                      className="h-8 w-8" 
                      style={{ color: getPlatformColor(platform.value) }}
                    />
                    {connectedCount > 0 && (
                      <Badge variant="secondary">
                        {connectedCount} connected
                      </Badge>
                    )}
                  </div>
                  <CardTitle className="text-lg">
                    {getPlatformName(platform.value)}
                  </CardTitle>
                  <CardDescription className="text-sm">
                    {platform.description}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Button 
                    className="w-full" 
                    variant="outline"
                    disabled={initiateOAuth.isPending}
                  >
                    {initiateOAuth.isPending ? (
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    ) : (
                      <Plus className="h-4 w-4 mr-2" />
                    )}
                    Connect Account
                  </Button>
                </CardContent>
              </Card>
            );
          })}
        </div>
      </div>

      {/* Disconnect Confirmation Dialog */}
      <AlertDialog 
        open={!!disconnectingId} 
        onOpenChange={() => setDisconnectingId(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Disconnect Account?</AlertDialogTitle>
            <AlertDialogDescription>
              This will remove the account from your team. You'll need to reconnect
              to schedule posts to this account again.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => disconnectingId && handleDisconnect(disconnectingId)}
              className="bg-red-600 hover:bg-red-700"
            >
              {disconnectAccount.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
              ) : null}
              Disconnect
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

// ============================================================================
// ACCOUNT CARD COMPONENT
// ============================================================================

interface AccountCardProps {
  account: SocialAccountDTO;
  onDisconnect: () => void;
  onRefresh: () => void;
  isRefreshing: boolean;
}

function AccountCard({ account, onDisconnect, onRefresh, isRefreshing }: AccountCardProps) {
  const statusColor = getStatusColor(account.status);
  const needsReauth = needsReconnection(account);
  const canPost = canPublish(account);

  return (
    <Card className={needsReauth ? 'border-orange-300' : undefined}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <Avatar>
              <AvatarImage src={account.avatarUrl} />
              <AvatarFallback>
                {account.displayName.substring(0, 2).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div>
              <h3 className="font-semibold">{account.displayName}</h3>
              <p className="text-sm text-gray-500">@{account.username}</p>
            </div>
          </div>
        </div>
        
        <div className="flex items-center gap-2 mt-2">
          <Badge 
            variant={account.status === 'active' ? 'default' : 'secondary'}
            className={`
              ${account.status === 'active' && 'bg-green-500 hover:bg-green-600'}
              ${needsReauth && 'bg-orange-500 hover:bg-orange-600'}
            `}
          >
            {account.status === 'active' ? (
              <CheckCircle className="h-3 w-3 mr-1" />
            ) : (
              <AlertCircle className="h-3 w-3 mr-1" />
            )}
            {getStatusLabel(account.status)}
          </Badge>
          
          <Badge variant="outline">
            {getPlatformName(account.platform)}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-2">
        {/* Rate Limits */}
        {account.rateLimits && (
          <div className="text-xs text-gray-600">
            <p>Posts remaining today: {account.rateLimits.remaining}</p>
          </div>
        )}

        {/* Warning if needs reconnection */}
        {needsReauth && (
          <div className="bg-orange-50 border border-orange-200 rounded p-2 text-xs text-orange-800">
            <AlertCircle className="h-3 w-3 inline mr-1" />
            Account needs reconnection
          </div>
        )}

        {/* Actions */}
        <div className="flex gap-2 pt-2">
          {needsReauth && (
            <Button 
              size="sm" 
              variant="default"
              className="flex-1"
              onClick={onRefresh}
              disabled={isRefreshing}
            >
              {isRefreshing ? (
                <Loader2 className="h-3 w-3 animate-spin mr-1" />
              ) : (
                <RefreshCw className="h-3 w-3 mr-1" />
              )}
              Reconnect
            </Button>
          )}
          
          <Button 
            size="sm" 
            variant="ghost"
            className="text-red-600 hover:text-red-700 hover:bg-red-50"
            onClick={onDisconnect}
          >
            <Link2Off className="h-3 w-3 mr-1" />
            Disconnect
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}