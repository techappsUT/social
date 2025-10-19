// ===========================================================================
// FILE: frontend/src/components/teams/members-tab.tsx
// Members list with role management and remove actions
// ===========================================================================

'use client';

import { useState } from 'react';
import { Team, TeamMember } from '@/lib/api/teams';
import { useRemoveTeamMember, useUpdateMemberRole } from '@/hooks/useTeams';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
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
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { MoreVertical, Crown, Shield, Edit, Eye, Trash2, Mail } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { toast } from 'sonner';

interface MembersTabProps {
  team: Team;
  canManageMembers: boolean;
}

export function MembersTab({ team, canManageMembers }: MembersTabProps) {
  const [memberToRemove, setMemberToRemove] = useState<TeamMember | null>(null);
  const removeMember = useRemoveTeamMember(team.id);
  const updateMemberRole = useUpdateMemberRole(team.id);

  const currentUserId = localStorage.getItem('userId');
  const currentUserRole = team.members.find(m => m.userId === currentUserId)?.role;
  const isOwner = currentUserRole === 'owner';

  const handleRemoveMember = async () => {
    if (!memberToRemove) return;

    try {
      await removeMember.mutateAsync(memberToRemove.userId);
      setMemberToRemove(null);
    } catch (error) {
      // Error handled by hook
    }
  };

  const handleChangeRole = async (member: TeamMember, newRole: string) => {
    if (member.role === newRole) return;

    try {
      await updateMemberRole.mutateAsync({
        userId: member.userId,
        role: newRole,
      });
    } catch (error) {
      // Error handled by hook
    }
  };

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
      case 'editor':
        return 'outline';
      case 'viewer':
        return 'outline';
      default:
        return 'outline';
    }
  };

  const canModifyMember = (member: TeamMember) => {
    if (!canManageMembers) return false;
    if (member.userId === currentUserId) return false; // Can't modify self
    if (member.role === 'owner' && !isOwner) return false; // Only owner can modify owner
    return true;
  };

  const getInitials = (firstName: string, lastName: string) => {
    return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase();
  };

  // Sort members: Owner first, then by role, then by name
  const sortedMembers = [...team.members].sort((a, b) => {
    const roleOrder = { owner: 0, admin: 1, editor: 2, viewer: 3 };
    const roleComparison = (roleOrder[a.role] || 4) - (roleOrder[b.role] || 4);
    
    if (roleComparison !== 0) return roleComparison;
    
    return a.firstName.localeCompare(b.firstName);
  });

  return (
    <div className="space-y-4">
      {/* Members count */}
      <div className="flex justify-between items-center">
        <p className="text-sm text-muted-foreground">
          {team.memberCount} {team.memberCount === 1 ? 'member' : 'members'}
        </p>
      </div>

      {/* Members table */}
      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Member</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Joined</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {sortedMembers.map((member) => {
              const canModify = canModifyMember(member);
              const isCurrentUser = member.userId === currentUserId;

              return (
                <TableRow key={member.id}>
                  {/* Member Info */}
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <Avatar className="h-9 w-9">
                        <AvatarImage src={member.avatarUrl} alt={member.firstName} />
                        <AvatarFallback>
                          {getInitials(member.firstName, member.lastName)}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium flex items-center gap-2">
                          {member.firstName} {member.lastName}
                          {isCurrentUser && (
                            <Badge variant="outline" className="text-xs">
                              You
                            </Badge>
                          )}
                        </div>
                        <div className="text-sm text-muted-foreground flex items-center gap-1">
                          <Mail className="h-3 w-3" />
                          {member.email}
                        </div>
                      </div>
                    </div>
                  </TableCell>

                  {/* Role Badge */}
                  <TableCell>
                    <Badge 
                      variant={getRoleBadgeVariant(member.role)}
                      className="flex items-center gap-1 w-fit"
                    >
                      {getRoleIcon(member.role)}
                      {member.role.charAt(0).toUpperCase() + member.role.slice(1)}
                    </Badge>
                  </TableCell>

                  {/* Joined Date */}
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDistanceToNow(new Date(member.joinedAt), { addSuffix: true })}
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="text-right">
                    {canModify ? (
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          {/* Change Role Options (only for owner) */}
                          {isOwner && member.role !== 'owner' && (
                            <>
                              <DropdownMenuItem
                                onClick={() => handleChangeRole(member, 'admin')}
                                disabled={member.role === 'admin'}
                              >
                                <Shield className="mr-2 h-4 w-4" />
                                Make Admin
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleChangeRole(member, 'editor')}
                                disabled={member.role === 'editor'}
                              >
                                <Edit className="mr-2 h-4 w-4" />
                                Make Editor
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => handleChangeRole(member, 'viewer')}
                                disabled={member.role === 'viewer'}
                              >
                                <Eye className="mr-2 h-4 w-4" />
                                Make Viewer
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                            </>
                          )}

                          {/* Remove Member */}
                          <DropdownMenuItem
                            onClick={() => setMemberToRemove(member)}
                            className="text-destructive focus:text-destructive"
                          >
                            <Trash2 className="mr-2 h-4 w-4" />
                            Remove from Team
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    ) : (
                      <span className="text-sm text-muted-foreground">—</span>
                    )}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>

      {/* Remove Member Confirmation Dialog */}
      <AlertDialog
        open={!!memberToRemove}
        onOpenChange={(open) => !open && setMemberToRemove(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Team Member</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to remove{' '}
              <span className="font-semibold">
                {memberToRemove?.firstName} {memberToRemove?.lastName}
              </span>{' '}
              from the team? They will lose access to all team resources.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRemoveMember}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {removeMember.isPending ? 'Removing...' : 'Remove Member'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}