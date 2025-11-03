import { apiClient } from '@/lib/api-client';
import { v4 as uuidv4 } from 'uuid';

export interface SocialPlatform {
  id: string;
  name: string;
  icon: string;
  color: string;
  available: boolean;
}

export interface ConnectedAccount {
  id: string;
  teamId: string;
  platform: string;
  username: string;
  displayName: string;
  profileUrl: string;
  avatarUrl: string;
  status: 'active' | 'expired' | 'revoked';
  connectedAt: string;
  expiresAt?: string;
}

class SocialOAuthService {
  private readonly API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000/api/v2';

  // Available platforms
  platforms: SocialPlatform[] = [
    { id: 'facebook', name: 'Facebook', icon: '📘', color: '#1877F2', available: true },
    { id: 'twitter', name: 'X (Twitter)', icon: '🐦', color: '#000000', available: true },
    { id: 'linkedin', name: 'LinkedIn', icon: '💼', color: '#0A66C2', available: true },
    { id: 'instagram', name: 'Instagram', icon: '📷', color: '#E4405F', available: false },
  ];

  // Step 1: Initiate OAuth flow
  async initiateOAuth(platform: string, teamId?: string): Promise<void> {
    try {
      // Generate state for CSRF protection
      const state = uuidv4();
      
      // Store state and optional teamId in session storage
      sessionStorage.setItem('oauth_state', state);
      sessionStorage.setItem('oauth_platform', platform); // Store platform too
      if (teamId) {
        sessionStorage.setItem('oauth_team_id', teamId);
      }

      // Get OAuth URL from backend, passing our state
      const response = await apiClient.get<{ data: { authUrl: string, state: string } }>(
        `/social/auth/${platform}?state=${encodeURIComponent(state)}`
      );

      // Use the authUrl directly from backend
      window.location.href = response.data.authUrl;
    } catch (error) {
      console.error('Failed to initiate OAuth:', error);
      throw new Error(`Failed to connect ${platform}`);
    }
  }

  // Step 2: Complete OAuth connection (called after redirect back)
  async completeOAuth(platform: string, code: string, state: string): Promise<ConnectedAccount> {
  try {
    // Verify state
    const savedState = sessionStorage.getItem('oauth_state');
    
    console.log('Completing OAuth:', {
      savedState,
      receivedState: state,
      platform,
      hasCode: !!code
    });
    
    if (!savedState || state !== savedState) {
      throw new Error('Invalid state parameter - possible CSRF attack');
    }

    // Get teamId from localStorage with correct key
    const teamId = localStorage.getItem('socialqueue_selected_team_id');
    
    if (!teamId) {
      throw new Error('No team selected. Please select a team first.');
    }

    // Complete the connection
    const response = await apiClient.post<{ data: { account: ConnectedAccount } }>(
      '/social/auth/complete',
      {
        platform,
        code,
        state,
        teamId, // This is now guaranteed to exist
      }
    );

    // Clean up
    sessionStorage.removeItem('oauth_state');
    sessionStorage.removeItem('oauth_platform');

    return response.data.account;
  } catch (error: any) {
    sessionStorage.removeItem('oauth_state');
    sessionStorage.removeItem('oauth_platform');
    
    console.error('OAuth completion error:', error);
    throw error;
  }
}

// Get connected accounts for a team
async getConnectedAccounts(teamId: string): Promise<ConnectedAccount[]> {
  try {
    const response = await apiClient.get<{ data: { accounts: ConnectedAccount[] } }>(
      `/teams/${teamId}/social/accounts`
    );
    return response.data.accounts;
  } catch (error: any) {
    // If 404, return empty array (no accounts yet)
    if (error.response?.status === 404) {
      console.log('No social accounts found for team');
      return [];
    }
    throw error;
  }
}

  // Disconnect account
  async disconnectAccount(accountId: string): Promise<void> {
    await apiClient.delete(`/social/accounts/${accountId}`);
  }

  // Refresh tokens for an account
  async refreshTokens(accountId: string): Promise<void> {
    await apiClient.post(`/social/accounts/${accountId}/refresh`);
  }
}

export const socialOAuthService = new SocialOAuthService();