// path: frontend/src/components/team-switcher.tsx
import React from 'react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Check, ChevronsUpDown, Users, Plus } from 'lucide-react';
import { useTeam } from '@/contexts/team-context';
import { useRouter } from 'next/navigation';

export function TeamSwitcher() {
  const router = useRouter();
  const { currentTeam, teams, switchTeam } = useTeam();

  if (!currentTeam) {
    return null;
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          className="w-full justify-between rounded-full mb-5 py-1.5"
        >
          <div className="flex items-center gap-2">
            <Users className="h-4 w-4" />
            <span className="truncate">{currentTeam.name}</span>
          </div>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56" align='start'> 
        <DropdownMenuLabel>Your Teams</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {teams.map((team) => (
          <DropdownMenuItem
            key={team.id}
            onClick={() => switchTeam(team.id)}
            className="cursor-pointer"
          >
            <div className="flex items-center justify-between w-full">
              <span className="truncate">{team.name}</span>
              {team.id === currentTeam.id && (
                <Check className="h-4 w-4 text-blue-600" />
              )}
            </div>
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => router.push('/dashboard/teams/create')}
          className="cursor-pointer"
        >
          <Plus className="mr-2 h-4 w-4" />
          Create Team
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}