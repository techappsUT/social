// path: frontend/src/hooks/useSocial.ts
/**
 * Social Media Hooks using React Query
 * OAuth flows and account management
 */

'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import * as socialApi from '@/lib/api/social';
import type {
  SocialPlatform,
  SocialAccountDTO,
  ConnectAccountRequest,
  PublishToSocialRequest,
} from '@/types/social';

// ============================================================================
// QUERY KEYS (for cache management)
// ============================================================================

export const socialKeys = {
  all: ['social'] as const,
  accounts: () => [...socialKeys.all, 'accounts'] as const,
  accountsList: (teamId: string) => [...socialKeys.accounts(), 'list', teamId] as const,
  accountDetail: (id: string) => [...socialKeys.accounts(), 'detail', id] as const,
  accountHealth: (id: string) => [...socialKeys.accounts(), 'health', id] as const,
  analytics: (accountId: string, postId: string) => 
    [...socialKeys.all, 'analytics', accountId, postId] as const,
};

// ============================================================================
// QUERY HOOKS (Read Operations)
// ============================================================================

/**
 * Get all connected social accounts for a team
 */
export function useSocialAccounts(teamId: string) {
  return useQuery({
    queryKey: socialKeys.accountsList(teamId),
    queryFn: () => socialApi.listAccounts(teamId),
    enabled: !!teamId,
    staleTime: 60000, // 1 minute
    refetchInterval: 300000, // Refetch every 5 minutes
  });
}

/**
 * Get a single social account by ID
 */
export function useSocialAccount(accountId: string | null) {
  return useQuery({
    queryKey: socialKeys.accountDetail(accountId || ''),
    queryFn: () => socialApi.getAccount(accountId!),
    enabled: !!accountId,
    staleTime: 60000,
  });
}

/**
 * Check account health status
 */
export function useAccountHealth(accountId: string | null) {
  return useQuery({
    queryKey: socialKeys.accountHealth(accountId || ''),
    queryFn: () => socialApi.checkAccountHealth(accountId!),
    enabled: !!accountId,
    staleTime: 30000, // 30 seconds
    refetchInterval: 60000, // Check every minute
  });
}

/**
 * Get analytics for a published post
 */
export function usePostAnalytics(accountId: string, postId: string) {
  return useQuery({
    queryKey: socialKeys.analytics(accountId, postId),
    queryFn: () => socialApi.getPostAnalytics(accountId, postId),
    enabled: !!accountId && !!postId,
    staleTime: 300000, // 5 minutes (analytics don't change frequently)
  });
}

/**
 * Get accounts by platform
 */
export function useAccountsByPlatform(teamId: string, platform: SocialPlatform) {
  return useQuery({
    queryKey: [...socialKeys.accountsList(teamId), platform],
    queryFn: () => socialApi.getAccountsByPlatform(teamId, platform),
    enabled: !!teamId && !!platform,
    staleTime: 60000,
  });
}

/**
 * Get active accounts only
 */
export function useActiveAccounts(teamId: string) {
  return useQuery({
    queryKey: [...socialKeys.accountsList(teamId), 'active'],
    queryFn: () => socialApi.getActiveAccounts(teamId),
    enabled: !!teamId,
    staleTime: 60000,
  });
}

// ============================================================================
// MUTATION HOOKS (Write Operations)
// ============================================================================

/**
 * Connect a social media account (after OAuth callback)
 */
export function useConnectAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: ConnectAccountRequest) => socialApi.connectAccount(data),
    onSuccess: (account) => {
      // Invalidate accounts list to refetch
      queryClient.invalidateQueries({ 
        queryKey: socialKeys.accountsList(account.teamId) 
      });
      
      toast.success(`${account.displayName} connected successfully!`);
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to connect account';
      toast.error(message);
    },
  });
}

/**
 * Disconnect a social media account
 */
export function useDisconnectAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (accountId: string) => socialApi.disconnectAccount(accountId),
    onMutate: async (accountId) => {
      // Get the account details before deletion
      const account = queryClient.getQueryData<SocialAccountDTO>(
        socialKeys.accountDetail(accountId)
      );
      return { account };
    },
    onSuccess: (_, accountId, context) => {
      // Remove from cache
      queryClient.removeQueries({ 
        queryKey: socialKeys.accountDetail(accountId) 
      });
      
      // Invalidate accounts lists
      if (context?.account) {
        queryClient.invalidateQueries({ 
          queryKey: socialKeys.accountsList(context.account.teamId) 
        });
      }
      
      toast.success('Account disconnected');
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to disconnect account';
      toast.error(message);
    },
  });
}

/**
 * Refresh OAuth tokens for an account
 */
export function useRefreshTokens() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (accountId: string) => socialApi.refreshAccountTokens(accountId),
    onSuccess: (account) => {
      // Update cache with refreshed account data
      queryClient.setQueryData(
        socialKeys.accountDetail(account.id),
        account
      );
      
      // Invalidate lists
      queryClient.invalidateQueries({ 
        queryKey: socialKeys.accountsList(account.teamId) 
      });
      
      toast.success('Tokens refreshed successfully');
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to refresh tokens';
      toast.error(message);
    },
  });
}

/**
 * Publish content directly to a social platform
 */
export function usePublishToSocial() {
  return useMutation({
    mutationFn: ({ 
      accountId, 
      data 
    }: { 
      accountId: string; 
      data: PublishToSocialRequest 
    }) => socialApi.publishToSocial(accountId, data),
    onSuccess: (result) => {
      if (result.success) {
        toast.success('Published successfully!');
      } else {
        toast.error(result.error || 'Publishing failed');
      }
    },
    onError: (error: any) => {
      const message = error?.message || 'Failed to publish';
      toast.error(message);
    },
  });
}

// ============================================================================
// OAUTH FLOW HELPERS
// ============================================================================

/**
 * Initiate OAuth flow
 * Opens OAuth popup/redirect
 */
export function useInitiateOAuth() {
  return useMutation({
    mutationFn: (platform: SocialPlatform) => socialApi.initiateOAuth(platform),
    onError: (error: any) => {
      const message = error?.message || 'Failed to start OAuth flow';
      toast.error(message);
    },
  });
}

/**
 * Hook to handle OAuth callback in callback page
 * Usage in /dashboard/social/callback page:
 * 
 * const searchParams = useSearchParams();
 * const connectAccount = useConnectAccount();
 * 
 * useEffect(() => {
 *   const { code, state, platform } = await handleOAuthCallback(searchParams);
 *   connectAccount.mutate({ teamId, platform, code, state });
 * }, []);
 */
export function useOAuthCallback(
  teamId: string,
  onSuccess?: (account: SocialAccountDTO) => void
) {
  const connectAccount = useConnectAccount();
  
  const handleCallback = async (searchParams: URLSearchParams) => {
    try {
      const { code, state, platform } = await socialApi.handleOAuthCallback(searchParams);
      
      const account = await connectAccount.mutateAsync({
        teamId,
        platform: platform as SocialPlatform,
        code,
        state,
      });
      
      onSuccess?.(account);
      
      return account;
    } catch (error) {
      console.error('OAuth callback error:', error);
      toast.error('Failed to complete account connection');
      throw error;
    }
  };
  
  return {
    handleCallback,
    isLoading: connectAccount.isPending,
    error: connectAccount.error,
  };
}

// ============================================================================
// BULK OPERATIONS & UTILITIES
// ============================================================================

/**
 * Refresh all expired accounts for a team
 */
export function useRefreshExpiredAccounts() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (teamId: string) => {
      const accounts = await socialApi.listAccounts(teamId);
      const expired = accounts.filter(acc => 
        acc.status === 'expired' || acc.status === 'reconnect_required'
      );
      
      const results = await Promise.allSettled(
        expired.map(acc => socialApi.refreshAccountTokens(acc.id))
      );
      
      return {
        total: expired.length,
        succeeded: results.filter(r => r.status === 'fulfilled').length,
        failed: results.filter(r => r.status === 'rejected').length,
      };
    },
    onSuccess: (results, teamId) => {
      queryClient.invalidateQueries({ 
        queryKey: socialKeys.accountsList(teamId) 
      });
      
      toast.success(
        `Refreshed ${results.succeeded}/${results.total} accounts`
      );
    },
    onError: () => {
      toast.error('Failed to refresh accounts');
    },
  });
}

/**
 * Check if any accounts need attention
 */
export function useAccountsNeedingAttention(teamId: string) {
  const { data: accounts } = useSocialAccounts(teamId);
  
  const needsAttention = accounts?.filter(acc => 
    ['expired', 'revoked', 'reconnect_required', 'suspended'].includes(acc.status)
  ) || [];
  
  return {
    needsAttention,
    count: needsAttention.length,
    hasIssues: needsAttention.length > 0,
  };
}

/**
 * Get platform connection status
 */
export function usePlatformStatus(teamId: string, platform: SocialPlatform) {
  const { data: accounts } = useSocialAccounts(teamId);
  
  const platformAccounts = accounts?.filter(acc => acc.platform === platform) || [];
  const activeCount = platformAccounts.filter(acc => acc.status === 'active').length;
  
  return {
    isConnected: platformAccounts.length > 0,
    isActive: activeCount > 0,
    totalAccounts: platformAccounts.length,
    activeAccounts: activeCount,
    accounts: platformAccounts,
  };
}