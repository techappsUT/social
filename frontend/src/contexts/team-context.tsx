// path: frontend/src/contexts/team-context.tsx
'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getTeams } from '@/lib/api/teams';
import type { Team } from '@/lib/api/teams';

interface TeamContextType {
  currentTeam: Team | null;
  teams: Team[];
  isLoading: boolean;
  switchTeam: (teamId: string) => void;
  refreshTeams: () => void;
}

const TeamContext = createContext<TeamContextType | undefined>(undefined);

const SELECTED_TEAM_KEY = 'socialqueue_selected_team_id';

export function TeamProvider({ children }: { children: React.ReactNode }) {
  const [currentTeamId, setCurrentTeamId] = useState<string | null>(null);

  // Fetch all teams
  const { data: teams = [], isLoading, refetch } = useQuery({
    queryKey: ['teams'],
    queryFn: getTeams,
    staleTime: 300000, // 5 minutes
  });

  // Initialize selected team from localStorage or default to first team
  useEffect(() => {
    if (teams.length > 0 && !currentTeamId) {
      // Try to get from localStorage
      const savedTeamId = localStorage.getItem(SELECTED_TEAM_KEY);
      
      if (savedTeamId && teams.some(t => t.id === savedTeamId)) {
        // Use saved team if it still exists
        setCurrentTeamId(savedTeamId);
      } else {
        // Default to first team
        setCurrentTeamId(teams[0].id);
        localStorage.setItem(SELECTED_TEAM_KEY, teams[0].id);
      }
    }
  }, [teams, currentTeamId]);

  // Get current team object
  const currentTeam = teams.find(t => t.id === currentTeamId) || null;

  // Switch team function
  const switchTeam = (teamId: string) => {
    setCurrentTeamId(teamId);
    localStorage.setItem(SELECTED_TEAM_KEY, teamId);
  };

  const value: TeamContextType = {
    currentTeam,
    teams,
    isLoading,
    switchTeam,
    refreshTeams: () => refetch(),
  };

  return <TeamContext.Provider value={value}>{children}</TeamContext.Provider>;
}

// Hook to use team context
export function useTeam() {
  const context = useContext(TeamContext);
  if (context === undefined) {
    throw new Error('useTeam must be used within a TeamProvider');
  }
  return context;
}

// Convenience hook to get just the current team ID
export function useCurrentTeamId(): string | null {
  const { currentTeam } = useTeam();
  return currentTeam?.id || null;
}