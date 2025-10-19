
// ===========================================================================
// FILE: frontend/src/components/teams/settings-tab.tsx
// Team settings management (timezone, preferences, etc.)
// ===========================================================================

'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Team } from '@/lib/api/teams';
import { useUpdateTeam } from '@/hooks/useTeams';
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
import { Switch } from '@/components/ui/switch';
import { Separator } from '@/components/ui/separator';
import { Loader2, Save } from 'lucide-react';

const settingsSchema = z.object({
  name: z
    .string()
    .min(3, 'Team name must be at least 3 characters')
    .max(100, 'Team name must be less than 100 characters'),
});

type SettingsFormData = z.infer<typeof settingsSchema>;

interface SettingsTabProps {
  team: Team;
  canManageSettings: boolean;
}

export function SettingsTab({ team, canManageSettings }: SettingsTabProps) {
  const updateTeam = useUpdateTeam(team.id);

  const form = useForm<SettingsFormData>({
    resolver: zodResolver(settingsSchema),
    defaultValues: {
      name: team.name,
    },
  });

  const onSubmit = async (data: SettingsFormData) => {
    try {
      await updateTeam.mutateAsync(data);
    } catch (error) {
      // Error handled by hook
    }
  };

  const hasChanges = form.formState.isDirty;

  if (!canManageSettings) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        You don't have permission to manage team settings.
        <br />
        Contact a team owner or admin for access.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          {/* Basic Settings */}
          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-medium">Basic Settings</h3>
              <p className="text-sm text-muted-foreground">
                Update your team's basic information
              </p>
            </div>
            <Separator />

            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Team Name</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder="My Awesome Team" />
                  </FormControl>
                  <FormDescription>
                    This is your team's public display name
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="space-y-4">
              <div className="space-y-2">
                <FormLabel>Team Slug</FormLabel>
                <Input value={team.slug} disabled />
                <FormDescription>
                  Your unique team identifier (cannot be changed)
                </FormDescription>
              </div>
            </div>
          </div>

          {/* Preferences */}
          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-medium">Preferences</h3>
              <p className="text-sm text-muted-foreground">
                Configure team-wide preferences
              </p>
            </div>
            <Separator />

            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <FormLabel>Timezone</FormLabel>
                  <FormDescription>
                    Default timezone for scheduling posts
                  </FormDescription>
                </div>
                <Select defaultValue={team.settings.timezone} disabled>
                  <SelectTrigger className="w-[200px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="UTC">UTC</SelectItem>
                    <SelectItem value="America/New_York">Eastern Time</SelectItem>
                    <SelectItem value="America/Los_Angeles">Pacific Time</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <FormLabel>Email Notifications</FormLabel>
                  <FormDescription>
                    Receive email updates about team activity
                  </FormDescription>
                </div>
                <Switch
                  defaultChecked={team.settings.enableNotifications}
                  disabled
                />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <FormLabel>Require Approval</FormLabel>
                  <FormDescription>
                    Posts must be approved before publishing
                  </FormDescription>
                </div>
                <Switch
                  defaultChecked={team.settings.requireApproval}
                  disabled
                />
              </div>

              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <FormLabel>Analytics</FormLabel>
                  <FormDescription>
                    Track team performance and insights
                  </FormDescription>
                </div>
                <Switch
                  defaultChecked={team.settings.enableAnalytics}
                  disabled
                />
              </div>
            </div>
          </div>

          {/* Save Button */}
          <div className="flex justify-end">
            <Button
              type="submit"
              disabled={!hasChanges || updateTeam.isPending}
            >
              {updateTeam.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Save Changes
                </>
              )}
            </Button>
          </div>
        </form>
      </Form>

      {/* Danger Zone */}
      <div className="space-y-4 pt-6">
        <div>
          <h3 className="text-lg font-medium text-destructive">Danger Zone</h3>
          <p className="text-sm text-muted-foreground">
            Irreversible actions that require caution
          </p>
        </div>
        <Separator />
        
        <div className="rounded-lg border border-destructive/50 p-4 space-y-3">
          <div>
            <h4 className="font-medium">Delete Team</h4>
            <p className="text-sm text-muted-foreground">
              Permanently delete this team and all associated data. This action cannot be undone.
            </p>
          </div>
          <Button variant="destructive" size="sm" disabled>
            Delete Team
          </Button>
        </div>
      </div>
    </div>
  );
}