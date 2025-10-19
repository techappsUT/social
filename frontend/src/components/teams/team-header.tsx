// ===========================================================================
// FILE: frontend/src/components/teams/team-header.tsx
// Team header with name, stats, and quick actions
// ===========================================================================

'use client';

import { Team } from '@/lib/api/teams';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Users, Calendar } from 'lucide-react';
import { format } from 'date-fns';

interface TeamHeaderProps {
  team: Team;
}

export function TeamHeader({ team }: TeamHeaderProps) {
  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight">{team.name}</h1>
            <p className="text-muted-foreground">@{team.slug}</p>
          </div>
          <Badge variant="outline" className="text-sm">
            Active
          </Badge>
        </div>

        <div className="mt-6 flex items-center gap-6 text-sm text-muted-foreground">
          <div className="flex items-center gap-2">
            <Users className="h-4 w-4" />
            <span>
              {team.memberCount} {team.memberCount === 1 ? 'member' : 'members'}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Calendar className="h-4 w-4" />
            <span>Created {format(new Date(team.createdAt), 'MMM d, yyyy')}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
