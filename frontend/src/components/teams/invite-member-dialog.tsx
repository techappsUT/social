// ===========================================================================
// FILE: frontend/src/components/teams/invite-member-dialog.tsx
// Dialog form to invite new team members with validation
// ===========================================================================

'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useInviteTeamMember } from '@/hooks/useTeams';
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { 
  Mail, 
  Crown, 
  Shield, 
  Edit, 
  Eye, 
  Loader2,
  AlertCircle 
} from 'lucide-react';

// Validation schema
const inviteMemberSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Please enter a valid email address'),
  role: z.enum(['owner', 'admin', 'editor', 'viewer'] as const, {
    error: 'Please select a role',
  }),
});

type InviteMemberFormData = z.infer<typeof inviteMemberSchema>;

interface InviteMemberDialogProps {
  teamId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function InviteMemberDialog({
  teamId,
  open,
  onOpenChange,
}: InviteMemberDialogProps) {
  const inviteMember = useInviteTeamMember(teamId);
  const [serverError, setServerError] = useState<string | null>(null);

  const form = useForm<InviteMemberFormData>({
    resolver: zodResolver(inviteMemberSchema),
    defaultValues: {
      email: '',
      role: 'editor',
    },
  });

  const onSubmit = async (data: InviteMemberFormData) => {
    setServerError(null);

    try {
      await inviteMember.mutateAsync(data);
      
      // Success - close dialog and reset form
      onOpenChange(false);
      form.reset();
    } catch (error: any) {
      // Error is handled by the hook (toast), but we can show additional context
      const apiError = error.response?.data;
      
      if (apiError?.error === 'validation_error' && apiError.details) {
        // Set field-specific errors
        Object.entries(apiError.details).forEach(([field, message]) => {
          form.setError(field as keyof InviteMemberFormData, {
            message: message as string,
          });
        });
      } else if (apiError?.message) {
        // Set general server error
        setServerError(apiError.message);
      }
    }
  };

  const handleClose = () => {
    if (!inviteMember.isPending) {
      onOpenChange(false);
      form.reset();
      setServerError(null);
    }
  };

  const roleDescriptions = {
    owner: 'Full control over team, including billing and deletion',
    admin: 'Can manage team members and all settings',
    editor: 'Can create and manage posts, but not team settings',
    viewer: 'Can view posts and analytics, but cannot make changes',
  };

  const roleIcons = {
    owner: Crown,
    admin: Shield,
    editor: Edit,
    viewer: Eye,
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Invite Team Member</DialogTitle>
          <DialogDescription>
            Send an invitation to join your team. They'll receive an email with instructions.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            {/* Server Error Alert */}
            {serverError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{serverError}</AlertDescription>
              </Alert>
            )}

            {/* Email Field */}
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email Address</FormLabel>
                  <FormControl>
                    <div className="relative">
                      <Mail className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                      <Input
                        placeholder="colleague@example.com"
                        className="pl-9"
                        {...field}
                        disabled={inviteMember.isPending}
                      />
                    </div>
                  </FormControl>
                  <FormDescription>
                    Enter the email address of the person you want to invite
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Role Field */}
            <FormField
              control={form.control}
              name="role"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Role</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                    disabled={inviteMember.isPending}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a role" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {Object.entries(roleDescriptions).map(([role, description]) => {
                        const Icon = roleIcons[role as keyof typeof roleIcons];
                        return (
                          <SelectItem key={role} value={role}>
                            <div className="flex items-start gap-2">
                              <Icon className="h-4 w-4 mt-0.5" />
                              <div className="flex-1">
                                <div className="font-medium capitalize">{role}</div>
                                <div className="text-xs text-muted-foreground">
                                  {description}
                                </div>
                              </div>
                            </div>
                          </SelectItem>
                        );
                      })}
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    {roleDescriptions[field.value as keyof typeof roleDescriptions]}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Footer Buttons */}
            <DialogFooter className="gap-2 sm:gap-0">
              <Button
                type="button"
                variant="outline"
                onClick={handleClose}
                disabled={inviteMember.isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={inviteMember.isPending}>
                {inviteMember.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Sending Invitation...
                  </>
                ) : (
                  <>
                    <Mail className="mr-2 h-4 w-4" />
                    Send Invitation
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