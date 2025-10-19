// ===========================================================================
// FILE: frontend/src/components/auth/TeamAccessCheck.tsx
// NEW - Check team access on dashboard
// ===========================================================================

'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { usePendingInvitations } from '@/hooks/useInvitations';
import { useQuery } from '@tanstack/react-query';
import { getTeams } from '@/lib/api/teams';
import { Loader2 } from 'lucide-react';

export function TeamAccessCheck({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [checking, setChecking] = useState(true);
  
  const { data: teams, isLoading: teamsLoading } = useQuery({
    queryKey: ['teams'],
    queryFn: getTeams,
  });
  
  const { data: invitations, isLoading: invitationsLoading } = usePendingInvitations();

  useEffect(() => {
    if (teamsLoading || invitationsLoading) return;

    // Check for pending invitations first
    if (invitations && invitations.length > 0) {
      router.push('/invitations');
      return;
    }

    // Check for active teams
    const activeTeams = teams?.filter(t => t.memberStatus === 'active') || [];
    
    if (activeTeams.length === 0) {
      // No teams - redirect to create team
      router.push('/onboarding/create-team');
      return;
    }

    // Has active teams - allow access
    setChecking(false);
  }, [teams, invitations, teamsLoading, invitationsLoading, router]);

  if (checking || teamsLoading || invitationsLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return <>{children}</>;
}