// ===========================================================================
// FILE: frontend/src/hooks/useInvitations.ts
// NEW - React hooks for invitation management
// ===========================================================================

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getPendingInvitations, acceptInvitation } from '@/lib/api/teams';
import { toast } from 'sonner';

/**
 * Hook to get pending invitations
 */
export function usePendingInvitations() {
  return useQuery({
    queryKey: ['pending-invitations'],
    queryFn: getPendingInvitations,
    staleTime: 30 * 1000, // 30 seconds
  });
}

/**
 * Hook to accept invitation
 */
export function useAcceptInvitation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: acceptInvitation,
    onSuccess: (data) => {
      // Invalidate queries
      queryClient.invalidateQueries({ queryKey: ['pending-invitations'] });
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      queryClient.invalidateQueries({ queryKey: ['team', data.team.id] });

      toast.success('Invitation accepted!', {
        description: `You've joined ${data.team.name}`,
      });
    },
    onError: (error: any) => {
      toast.error('Failed to accept invitation', {
        description: error.message || 'Please try again',
      });
    },
  });
}