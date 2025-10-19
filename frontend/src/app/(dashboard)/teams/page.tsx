// ===========================================================================
// FILE: frontend/src/app/(dashboard)/teams/page.tsx
// Teams list page with create team functionality
// ===========================================================================

'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useTeams, useCreateTeam } from '@/hooks/useTeams';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { 
  Plus, 
  Users, 
  ArrowRight, 
  AlertCircle,
  Calendar,
  Crown,
  Shield,
  Edit,
  Eye
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

import { CreateTeamDialog } from '@/components/teams/create-team-dialog';

export default function TeamsPage() {
  const router = useRouter();
  const { data: teams, isLoading, error } = useTeams();
  const [showCreateDialog, setShowCreateDialog] = useState(false);

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'owner':
        return <Crown className="h-3 w-3" />;
      case 'admin':
        return <Shield className="h-3 w-3" />;
      case 'editor':
        return <Edit className="h-3 w-3" />;
      case 'viewer':
        return <Eye className="h-3 w-3" />;
      default:
        return null;
    }
  };

  const getRoleBadgeVariant = (role: string) => {
    switch (role) {
      case 'owner':
        return 'default';
      case 'admin':
        return 'secondary';
      default:
        return 'outline';
    }
  };

  const getTeamInitials = (name: string) => {
    return name
      .split(' ')
      .map(word => word.charAt(0))
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="container mx-auto py-8 space-y-6">
        <div className="flex justify-between items-center">
          <Skeleton className="h-10 w-48" />
          <Skeleton className="h-10 w-32" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-48" />
          ))}
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="container mx-auto py-8">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Failed to load teams. Please try again.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  const currentUserId = localStorage.getItem('userId');

  return (
    <div className="container mx-auto py-8 space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Teams</h1>
          <p className="text-muted-foreground">
            Manage your teams and collaborate with others
          </p>
        </div>
        <Button onClick={() => setShowCreateDialog(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Team
        </Button>
      </div>

      {/* Teams Grid */}
      {teams && teams.length > 0 ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {teams.map((team) => {
            const currentUserMember = team.members.find(
              m => m.userId === currentUserId
            );
            const userRole = currentUserMember?.role || 'viewer';

            return (
              <Card
                key={team.id}
                className="hover:border-primary/50 transition-colors cursor-pointer group"
                onClick={() => router.push(`/teams/${team.id}`)}
              >
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <Avatar className="h-12 w-12">
                      <AvatarFallback className="bg-primary text-primary-foreground">
                        {getTeamInitials(team.name)}
                      </AvatarFallback>
                    </Avatar>
                    <Badge 
                      variant={getRoleBadgeVariant(userRole)}
                      className="flex items-center gap-1"
                    >
                      {getRoleIcon(userRole)}
                      {userRole.charAt(0).toUpperCase() + userRole.slice(1)}
                    </Badge>
                  </div>
                  <CardTitle className="mt-4">{team.name}</CardTitle>
                  <CardDescription>@{team.slug}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Team Stats */}
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <div className="flex items-center gap-1">
                      <Users className="h-4 w-4" />
                      <span>{team.memberCount}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Calendar className="h-4 w-4" />
                      <span>
                        {formatDistanceToNow(new Date(team.createdAt), { 
                          addSuffix: true 
                        })}
                      </span>
                    </div>
                  </div>

                  {/* View Team Button */}
                  <Button
                    variant="outline"
                    className="w-full group-hover:border-primary group-hover:text-primary"
                    onClick={(e) => {
                      e.stopPropagation();
                      router.push(`/teams/${team.id}`);
                    }}
                  >
                    View Team
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </Button>
                </CardContent>
              </Card>
            );
          })}
        </div>
      ) : (
        // Empty state
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="rounded-full bg-muted p-6">
              <Users className="h-12 w-12 text-muted-foreground" />
            </div>
            <div className="text-center space-y-2">
              <h3 className="text-xl font-semibold">No teams yet</h3>
              <p className="text-muted-foreground max-w-md">
                Create your first team to start collaborating with others on social media management
              </p>
            </div>
            <Button onClick={() => setShowCreateDialog(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Your First Team
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Create Team Dialog */}
      <CreateTeamDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
      />
    </div>
  );
}
