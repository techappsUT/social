// ===========================================================================
// FILE: frontend/src/lib/api/teams.ts
// Complete team API client with proper TypeScript types
// ===========================================================================

import { apiClient } from '@/lib/api-client';

// ===========================================================================
// TYPE DEFINITIONS
// ===========================================================================

export interface TeamMember {
  id: string;
  userId: string;
  email: string;
  username: string;
  firstName: string;
  lastName: string;
  avatarUrl?: string;
  role: 'owner' | 'admin' | 'editor' | 'viewer';
  joinedAt: string;  // ISO date string
}

export interface Team {
  id: string;
  name: string;
  slug: string;
  settings: TeamSettings;
  members: TeamMember[];
  memberCount: number;
  createdAt: string;
  updatedAt: string;
  memberStatus: 'active' | 'pending' | 'removed';
}

export interface TeamSettings {
  timezone: string;
  defaultPostTime: string;
  enableNotifications: boolean;
  enableAnalytics: boolean;
  requireApproval: boolean;
  autoSchedule: boolean;
  language: string;
  dateFormat: string;
}

export interface InviteMemberRequest {
  email: string;
  role: 'owner' | 'admin' | 'editor' | 'viewer';
}

export interface UpdateMemberRoleRequest {
  role: 'owner' | 'admin' | 'editor' | 'viewer';
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}

export interface ApiError {
  success: false;
  error: string;
  message: string;
  details?: Record<string, string>;
}

// ===========================================================================
// API FUNCTIONS
// ===========================================================================

/**
 * Get all teams for current user
 */
export async function getTeams(): Promise<Team[]> {
  const response = await apiClient.get<ApiResponse<{ teams: Team[] }>>('/teams');
  return response.data.teams || [];
}

/**
 * Get single team by ID
 */
export async function getTeam(teamId: string): Promise<Team> {
  const response = await apiClient.get<ApiResponse<{ team: Team }>>(`/teams/${teamId}`);
  return response.data.team;
}

/**
 * Create new team
 */
export async function createTeam(name: string): Promise<Team> {
  const response = await apiClient.post<ApiResponse<{ team: Team }>>('/teams', { name });
  return response.data.team;
}

/**
 * Update team
 */
export async function updateTeam(teamId: string, data: { name?: string }): Promise<Team> {
  const response = await apiClient.put<ApiResponse<{ team: Team }>>(`/teams/${teamId}`, data);
  return response.data.team;
}

/**
 * Delete team
 */
export async function deleteTeam(teamId: string): Promise<void> {
  await apiClient.delete(`/teams/${teamId}`);
}

// ===========================================================================
// MEMBER MANAGEMENT FUNCTIONS
// ===========================================================================

/**
 * Remove member from team
 */
export async function removeTeamMember(teamId: string, userId: string): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/members/${userId}`);
}

/**
 * Update member role
 */
export async function updateMemberRole(
  teamId: string,
  userId: string,
  data: UpdateMemberRoleRequest
): Promise<TeamMember> {
  const response = await apiClient.patch<ApiResponse<{ member: TeamMember }>>(
    `/teams/${teamId}/members/${userId}/role`,
    data
  );
  return response.data.member;
}

export async function inviteTeamMember(
    teamId: string,
    data: InviteMemberRequest  // ✅ Type-safe
): Promise<TeamMember> {  // ✅ Known return type
    const response = await apiClient.post<ApiResponse<{ member: TeamMember }>>(
        `/teams/${teamId}/members`,
        data
    );
    return response.data.member;
}

// ===========================================================================
// FILE: frontend/src/lib/api/teams.ts
// UPDATE - Add invitation functions
// ===========================================================================

export interface PendingInvitation {
  teamId: string;
  teamName: string;
  teamSlug: string;
  role: 'owner' | 'admin' | 'editor' | 'viewer';
  invitedBy: string;
  invitedAt: string;
  inviterName?: string;
}

/**
 * Get pending team invitations for current user
 */
export async function getPendingInvitations(): Promise<PendingInvitation[]> {
  const response = await apiClient.get<ApiResponse<{ invitations: PendingInvitation[] }>>(
    '/invitations/pending'
  );
  return response.data.invitations || [];
}

/**
 * Accept a team invitation
 */
export async function acceptInvitation(teamId: string): Promise<{ team: Team; member: TeamMember }> {
  const response = await apiClient.post<ApiResponse<{ team: Team; member: TeamMember }>>(
    `/teams/${teamId}/accept`
  );
  return response.data;
}