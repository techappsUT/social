// ===========================================================================
// FILE: frontend/src/app/(dashboard)/teams/[teamId]/page.tsx
// Main team management page with members list and settings
// ===========================================================================

'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useTeam } from '@/hooks/useTeams';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  Users, 
  Settings, 
  Plus, 
  ArrowLeft,
  AlertCircle 
} from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import Link from 'next/link';

import { MembersTab } from '@/components/teams/members-tab';
import { SettingsTab } from '@/components/teams/settings-tab';
import { InviteMemberDialog } from '@/components/teams/invite-member-dialog';
import { TeamHeader } from '@/components/teams/team-header';

export default function TeamManagementPage() {
  const params = useParams();
  const router = useRouter();
  const teamId = params.teamId as string;
  
  const { data: team, isLoading, error } = useTeam(teamId);
  const [showInviteDialog, setShowInviteDialog] = useState(false);
  const [activeTab, setActiveTab] = useState('members');

  // Loading state
  if (isLoading) {
    return (
      <div className="container mx-auto py-8 space-y-6">
        <Skeleton className="h-12 w-full" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  // Error state
  if (error || !team) {
    return (
      <div className="container mx-auto py-8">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            {error?.message || 'Failed to load team. Please try again.'}
          </AlertDescription>
        </Alert>
        <Button
          variant="outline"
          className="mt-4"
          onClick={() => router.push('/teams')}
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Teams
        </Button>
      </div>
    );
  }

  // Get current user's role
  const currentUserRole = team.members.find(
    m => m.userId === localStorage.getItem('userId')
  )?.role;
  
  const canManageMembers = currentUserRole === 'owner' || currentUserRole === 'admin';
  const canManageSettings = currentUserRole === 'owner' || currentUserRole === 'admin';

  return (
    <div className="container mx-auto py-8 space-y-6">
      {/* Team Header */}
      <TeamHeader team={team} />

      {/* Action Buttons */}
      <div className="flex justify-between items-center">
        <Button
          variant="outline"
          onClick={() => router.push('/teams')}
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Teams
        </Button>

        {canManageMembers && (
          <Button onClick={() => setShowInviteDialog(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Invite Member
          </Button>
        )}
      </div>

      {/* Main Content Tabs */}
      <Card>
        <CardHeader>
          <CardTitle>Team Management</CardTitle>
          <CardDescription>
            Manage your team members and settings
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="members" className="flex items-center gap-2">
                <Users className="h-4 w-4" />
                Members ({team.memberCount})
              </TabsTrigger>
              <TabsTrigger 
                value="settings" 
                className="flex items-center gap-2"
                disabled={!canManageSettings}
              >
                <Settings className="h-4 w-4" />
                Settings
              </TabsTrigger>
            </TabsList>

            <TabsContent value="members" className="mt-6">
              <MembersTab 
                team={team} 
                canManageMembers={canManageMembers}
              />
            </TabsContent>

            <TabsContent value="settings" className="mt-6">
              <SettingsTab 
                team={team} 
                canManageSettings={canManageSettings}
              />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Invite Member Dialog */}
      {canManageMembers && (
        <InviteMemberDialog
          teamId={teamId}
          open={showInviteDialog}
          onOpenChange={setShowInviteDialog}
        />
      )}
    </div>
  );
}