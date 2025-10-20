// path: frontend/src/app/(dashboard)/social/callback/page.tsx
'use client';

import React, { useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { Card, CardContent } from '@/components/ui/card';
import { Loader2, CheckCircle, XCircle } from 'lucide-react';
import { useOAuthCallback } from '@/hooks/useSocial';
import { useCurrentUser } from '@/hooks/useAuth';

/**
 * OAuth Callback Handler Page
 * 
 * This page receives the OAuth redirect from social platforms
 * after user authorizes the connection.
 * 
 * Flow:
 * 1. User clicks "Connect Twitter" on accounts page
 * 2. Frontend opens OAuth popup with Twitter's auth page
 * 3. User authorizes on Twitter
 * 4. Twitter redirects to: /dashboard/social/callback?code=xxx&state=yyy&platform=twitter
 * 5. This page handles the callback
 * 6. Calls backend to exchange code for tokens
 * 7. Redirects user back to accounts page
 */

type CallbackState = 'loading' | 'success' | 'error';

export default function OAuthCallbackPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const { data: user } = useCurrentUser();
  
  const [state, setState] = useState<CallbackState>('loading');
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [accountName, setAccountName] = useState<string>('');

  const { handleCallback } = useOAuthCallback(
    user?.teamId || '',
    (account) => {
      // Success callback
      setState('success');
      setAccountName(account.displayName);
      
      // Wait 2 seconds to show success message, then redirect
      setTimeout(() => {
        // If this was a popup, close it
        if (window.opener) {
          window.opener.postMessage(
            { type: 'oauth_success', account },
            window.location.origin
          );
          window.close();
        } else {
          // Otherwise redirect back to accounts page
          router.push('/dashboard/accounts?success=true');
        }
      }, 2000);
    }
  );

  useEffect(() => {
    const processCallback = async () => {
      // Wait for user data
      if (!user?.teamId) {
        return;
      }

      // Check if we have the required params
      const code = searchParams.get('code');
      const platform = searchParams.get('platform');
      const error = searchParams.get('error');

      // Handle OAuth error (user denied)
      if (error) {
        setState('error');
        setErrorMessage(
          error === 'access_denied'
            ? 'You denied access. Please try again if you want to connect this account.'
            : `OAuth error: ${error}`
        );
        
        setTimeout(() => {
          if (window.opener) {
            window.opener.postMessage(
              { type: 'oauth_error', error },
              window.location.origin
            );
            window.close();
          } else {
            router.push('/dashboard/accounts');
          }
        }, 3000);
        return;
      }

      // Validate required params
      if (!code || !platform) {
        setState('error');
        setErrorMessage('Invalid callback parameters. Please try connecting again.');
        
        setTimeout(() => {
          router.push('/dashboard/accounts');
        }, 3000);
        return;
      }

      // Process the OAuth callback
      try {
        await handleCallback(searchParams);
      } catch (error: any) {
        console.error('OAuth callback error:', error);
        setState('error');
        setErrorMessage(
          error?.message || 'Failed to connect account. Please try again.'
        );
        
        setTimeout(() => {
          if (window.opener) {
            window.opener.postMessage(
              { type: 'oauth_error', error: error?.message },
              window.location.origin
            );
            window.close();
          } else {
            router.push('/dashboard/accounts');
          }
        }, 3000);
      }
    };

    processCallback();
  }, [user, searchParams, handleCallback, router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <Card className="w-full max-w-md">
        <CardContent className="pt-6">
          {state === 'loading' && (
            <div className="text-center space-y-4">
              <Loader2 className="h-16 w-16 animate-spin text-blue-500 mx-auto" />
              <div>
                <h2 className="text-xl font-semibold mb-2">
                  Connecting your account...
                </h2>
                <p className="text-gray-600 text-sm">
                  Please wait while we complete the setup
                </p>
              </div>
            </div>
          )}

          {state === 'success' && (
            <div className="text-center space-y-4">
              <div className="mx-auto h-16 w-16 rounded-full bg-green-100 flex items-center justify-center">
                <CheckCircle className="h-10 w-10 text-green-600" />
              </div>
              <div>
                <h2 className="text-xl font-semibold text-green-900 mb-2">
                  Successfully Connected!
                </h2>
                <p className="text-gray-600 text-sm">
                  {accountName ? `${accountName} is now connected to your account.` : 'Your account has been connected.'}
                </p>
                <p className="text-gray-500 text-xs mt-2">
                  Redirecting...
                </p>
              </div>
            </div>
          )}

          {state === 'error' && (
            <div className="text-center space-y-4">
              <div className="mx-auto h-16 w-16 rounded-full bg-red-100 flex items-center justify-center">
                <XCircle className="h-10 w-10 text-red-600" />
              </div>
              <div>
                <h2 className="text-xl font-semibold text-red-900 mb-2">
                  Connection Failed
                </h2>
                <p className="text-gray-600 text-sm">
                  {errorMessage}
                </p>
                <p className="text-gray-500 text-xs mt-2">
                  Redirecting back...
                </p>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

/**
 * USAGE NOTES:
 * 
 * 1. This page must be accessible at: /dashboard/social/callback
 * 
 * 2. Update your OAuth redirect URIs in each platform:
 *    - Twitter: https://yourdomain.com/dashboard/social/callback
 *    - LinkedIn: https://yourdomain.com/dashboard/social/callback
 *    - Facebook: https://yourdomain.com/dashboard/social/callback
 * 
 * 3. For development, also add:
 *    http://localhost:3000/dashboard/social/callback
 * 
 * 4. The backend should redirect to this URL after handling the OAuth callback
 * 
 * 5. If using popup flow (recommended), this page will automatically close
 *    and notify the parent window via postMessage
 * 
 * 6. If using redirect flow, it will redirect back to /dashboard/accounts
 */