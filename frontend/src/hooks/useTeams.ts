
// ===========================================================================
// REACT QUERY HOOKS
// ===========================================================================
// FILE: frontend/src/hooks/useTeams.ts

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import * as teamsApi from '@/lib/api/teams';

/**
 * Get all teams
 */
export function useTeams() {
  return useQuery({
    queryKey: ['teams'],
    queryFn: teamsApi.getTeams,
  });
}

/**
 * Get single team
 */
export function useTeam(teamId: string) {
  return useQuery({
    queryKey: ['teams', teamId],
    queryFn: () => teamsApi.getTeam(teamId),
    enabled: !!teamId,
  });
}

/**
 * Create team mutation
 */
export function useCreateTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (name: string) => teamsApi.createTeam(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      toast.success('Team created successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to create team', {
        description: error.response?.data?.message || 'Please try again',
      });
    },
  });
}

/**
 * Update team mutation
 */
export function useUpdateTeam(teamId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name?: string }) => teamsApi.updateTeam(teamId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      queryClient.invalidateQueries({ queryKey: ['teams', teamId] });
      toast.success('Team updated successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to update team', {
        description: error.response?.data?.message || 'Please try again',
      });
    },
  });
}

/**
 * Delete team mutation
 */
export function useDeleteTeam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (teamId: string) => teamsApi.deleteTeam(teamId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams'] });
      toast.success('Team deleted successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to delete team', {
        description: error.response?.data?.message || 'Please try again',
      });
    },
  });
}

// ===========================================================================
// MEMBER MANAGEMENT HOOKS
// ===========================================================================


/**
 * Remove team member mutation
 */
export function useRemoveTeamMember(teamId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: string) => teamsApi.removeTeamMember(teamId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams', teamId] });
      toast.success('Member removed successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to remove member', {
        description: error.response?.data?.message || 'Please try again',
      });
    },
  });
}

/**
 * Update member role mutation
 */
export function useUpdateMemberRole(teamId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      teamsApi.updateMemberRole(teamId, userId, { role: role as any }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['teams', teamId] });
      toast.success('Member role updated successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to update role', {
        description: error.response?.data?.message || 'Please try again',
      });
    },
  });
}

// ===========================================================================
// USAGE EXAMPLE IN COMPONENT
// ===========================================================================
/*
import { useTeam, useInviteTeamMember, useRemoveTeamMember } from '@/hooks/useTeams';

export function TeamMembersPage({ teamId }: { teamId: string }) {
  const { data: team, isLoading } = useTeam(teamId);
  const inviteMember = useInviteTeamMember(teamId);
  const removeMember = useRemoveTeamMember(teamId);
  
  const handleInvite = (email: string, role: string) => {
    inviteMember.mutate({ email, role: role as any });
  };
  
  const handleRemove = (userId: string) => {
    removeMember.mutate(userId);
  };
  
  if (isLoading) return <div>Loading...</div>;
  
  return (
    <div>
      <h2>{team?.name} Members</h2>
      <MemberList 
        members={team?.members || []} 
        onRemove={handleRemove}
      />
      <InviteMemberForm onSubmit={handleInvite} />
    </div>
  );
}
*/


// File: frontend/src/hooks/useTeams.ts
export function useInviteTeamMember(teamId: string) {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (data: teamsApi.InviteMemberRequest) => 
            teamsApi.inviteTeamMember(teamId, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['teams', teamId] });
            toast.success('Member invited successfully');
        },
        onError: (error: any) => {
            // ✅ Consistent error handling
            const apiError = error.response?.data as teamsApi.ApiError;
            
            if (apiError?.error === 'validation_error' && apiError.details) {
                const fieldErrors = Object.values(apiError.details).join(', ');
                toast.error('Validation error', { description: fieldErrors });
            } else {
                toast.error('Failed to invite member', {
                    description: apiError?.message || 'Please try again',
                });
            }
        },
    });
}

// Usage in component:
// function TeamMembersPage({ teamId }: { teamId: string }) {
//     const inviteMember = useInviteTeamMember(teamId);
    
//     const handleInvite = (email: string, role: string) => {
//         // ✅ Type-safe, handles loading, errors, success automatically
//         inviteMember.mutate({ email, role: role as any });
//     };
    
//     return (
//         <InviteMemberForm 
//             onSubmit={handleInvite}
//             isLoading={inviteMember.isPending}  // ✅ Built-in loading state
//         />
//     );
// }
