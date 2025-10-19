// ===========================================================================
// FILE: frontend/src/app/(dashboard)/invitations/page.tsx
// NEW - Display and accept pending invitations
// ===========================================================================

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { usePendingInvitations, useAcceptInvitation } from '@/hooks/useInvitations';
import { Mail, Users, CheckCircle, Loader2, AlertCircle } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { useRouter } from 'next/navigation';

const roleColors = {
  owner: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
  admin: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
  editor: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
  viewer: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200',
};

export default function InvitationsPage() {
  const router = useRouter();
  const { data: invitations, isLoading, error } = usePendingInvitations();
  const acceptMutation = useAcceptInvitation();

  const handleAccept = async (teamId: string, teamName: string) => {
    await acceptMutation.mutateAsync(teamId);
    // Redirect to the team dashboard after accepting
    router.push(`/dashboard?team=${teamId}`);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="container max-w-4xl mx-auto py-8">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Failed to load invitations. Please refresh the page.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  if (!invitations || invitations.length === 0) {
    return (
      <div className="container max-w-4xl mx-auto py-16">
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <Mail className="h-16 w-16 text-muted-foreground mb-4" />
            <h2 className="text-2xl font-semibold mb-2">No Pending Invitations</h2>
            <p className="text-muted-foreground text-center">
              You don't have any pending team invitations at the moment.
            </p>
            <Button className="mt-6" onClick={() => router.push('/dashboard')}>
              Go to Dashboard
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container max-w-4xl mx-auto py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Team Invitations</h1>
        <p className="text-muted-foreground">
          You have {invitations.length} pending {invitations.length === 1 ? 'invitation' : 'invitations'}
        </p>
      </div>

      <div className="space-y-4">
        {invitations.map((invitation) => (
          <Card key={invitation.teamId}>
            <CardHeader>
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <CardTitle className="flex items-center gap-2">
                    <Users className="h-5 w-5" />
                    {invitation.teamName}
                  </CardTitle>
                  <CardDescription className="mt-2">
                    {invitation.inviterName ? (
                      <>Invited by {invitation.inviterName}</>
                    ) : (
                      <>You've been invited to join this team</>
                    )}
                    {' · '}
                    {formatDistanceToNow(new Date(invitation.invitedAt), { addSuffix: true })}
                  </CardDescription>
                </div>
                <Badge className={roleColors[invitation.role as keyof typeof roleColors]}>
                  {invitation.role}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="flex gap-2">
                <Button
                  onClick={() => handleAccept(invitation.teamId, invitation.teamName)}
                  disabled={acceptMutation.isPending}
                  className="flex-1"
                >
                  {acceptMutation.isPending ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Accepting...
                    </>
                  ) : (
                    <>
                      <CheckCircle className="mr-2 h-4 w-4" />
                      Accept Invitation
                    </>
                  )}
                </Button>
                <Button variant="outline" className="flex-1">
                  Decline
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}