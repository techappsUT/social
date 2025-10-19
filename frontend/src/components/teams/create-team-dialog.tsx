
// ===========================================================================
// FILE: frontend/src/components/teams/create-team-dialog.tsx
// Dialog to create a new team
// ===========================================================================

'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useRouter } from 'next/navigation';
import { useCreateTeam } from '@/hooks/useTeams';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, Users, AlertCircle } from 'lucide-react';

const createTeamSchema = z.object({
  name: z
    .string()
    .min(3, 'Team name must be at least 3 characters')
    .max(100, 'Team name must be less than 100 characters'),
});

type CreateTeamFormData = z.infer<typeof createTeamSchema>;

interface CreateTeamDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateTeamDialog({ open, onOpenChange }: CreateTeamDialogProps) {
  const router = useRouter();
  const createTeam = useCreateTeam();
  const [serverError, setServerError] = useState<string | null>(null);

  const form = useForm<CreateTeamFormData>({
    resolver: zodResolver(createTeamSchema),
    defaultValues: {
      name: '',
    },
  });

  const onSubmit = async (data: CreateTeamFormData) => {
    setServerError(null);

    try {
      const newTeam = await createTeam.mutateAsync(data.name);
      
      // Success - redirect to new team
      onOpenChange(false);
      form.reset();
      router.push(`/teams/${newTeam.id}`);
    } catch (error: any) {
      const apiError = error.response?.data;
      
      if (apiError?.error === 'validation_error' && apiError.details) {
        Object.entries(apiError.details).forEach(([field, message]) => {
          form.setError(field as keyof CreateTeamFormData, {
            message: message as string,
          });
        });
      } else if (apiError?.message) {
        setServerError(apiError.message);
      }
    }
  };

  const handleClose = () => {
    if (!createTeam.isPending) {
      onOpenChange(false);
      form.reset();
      setServerError(null);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create New Team</DialogTitle>
          <DialogDescription>
            Create a new team to collaborate with others on social media management
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            {serverError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{serverError}</AlertDescription>
              </Alert>
            )}

            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Team Name</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="Marketing Team"
                      {...field}
                      disabled={createTeam.isPending}
                    />
                  </FormControl>
                  <FormDescription>
                    Choose a name that describes your team's purpose
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter className="gap-2 sm:gap-0">
              <Button
                type="button"
                variant="outline"
                onClick={handleClose}
                disabled={createTeam.isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={createTeam.isPending}>
                {createTeam.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Creating...
                  </>
                ) : (
                  <>
                    <Users className="mr-2 h-4 w-4" />
                    Create Team
                  </>
                )}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}