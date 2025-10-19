// ===========================================================================
// FILE: frontend/src/components/teams/delete-team-dialog.tsx
// Dialog to confirm team deletion with name verification
// ===========================================================================

'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useDeleteTeam } from '@/hooks/useTeams';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Label } from '@/components/ui/label';
import { Loader2, AlertTriangle } from 'lucide-react';

interface DeleteTeamDialogProps {
  teamId: string;
  teamName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function DeleteTeamDialog({
  teamId,
  teamName,
  open,
  onOpenChange,
}: DeleteTeamDialogProps) {
  const router = useRouter();
  const deleteTeam = useDeleteTeam();
  const [confirmationText, setConfirmationText] = useState('');
  const [error, setError] = useState<string | null>(null);

  const isConfirmationValid = confirmationText === teamName;

  const handleDelete = async () => {
    if (!isConfirmationValid) {
      setError('Team name does not match');
      return;
    }

    setError(null);

    try {
      await deleteTeam.mutateAsync(teamId);
      
      // Success - redirect to teams list
      onOpenChange(false);
      router.push('/dashboard/teams');
    } catch (error: any) {
      const apiError = error.response?.data;
      setError(apiError?.message || 'Failed to delete team. Please try again.');
    }
  };

  const handleClose = () => {
    if (!deleteTeam.isPending) {
      setConfirmationText('');
      setError(null);
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-destructive/10">
              <AlertTriangle className="h-5 w-5 text-destructive" />
            </div>
            <div>
              <DialogTitle>Delete Team</DialogTitle>
              <DialogDescription>
                This action cannot be undone
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              <strong>Warning:</strong> Deleting this team will permanently remove:
              <ul className="mt-2 list-disc list-inside space-y-1 text-sm">
                <li>All team members and invitations</li>
                <li>All scheduled posts</li>
                <li>All connected social accounts</li>
                <li>All analytics and historical data</li>
              </ul>
            </AlertDescription>
          </Alert>

          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label htmlFor="confirm-name">
              Type <span className="font-bold">{teamName}</span> to confirm
            </Label>
            <Input
              id="confirm-name"
              placeholder={`Enter "${teamName}"`}
              value={confirmationText}
              onChange={(e) => setConfirmationText(e.target.value)}
              disabled={deleteTeam.isPending}
              autoComplete="off"
              className={
                confirmationText && !isConfirmationValid
                  ? 'border-destructive focus-visible:ring-destructive'
                  : ''
              }
            />
            {confirmationText && !isConfirmationValid && (
              <p className="text-sm text-destructive">
                Team name does not match
              </p>
            )}
          </div>

          <div className="rounded-lg border border-muted bg-muted/50 p-3">
            <p className="text-sm text-muted-foreground">
              <strong>Note:</strong> Only the team owner can delete a team. This
              action is permanent and cannot be reversed.
            </p>
          </div>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            type="button"
            variant="outline"
            onClick={handleClose}
            disabled={deleteTeam.isPending}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={handleDelete}
            disabled={!isConfirmationValid || deleteTeam.isPending}
          >
            {deleteTeam.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Deleting Team...
              </>
            ) : (
              'Delete Team Permanently'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}