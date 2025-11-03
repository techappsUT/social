'use client';

import { useEffect, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { socialOAuthService } from '@/services/social-oauth.service';
import { toast } from 'sonner';
import { Loader2 } from 'lucide-react';

function OAuthCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [isProcessing, setIsProcessing] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const handleCallback = async () => {
      const platform = searchParams.get('platform');
      const code = searchParams.get('code');
      const state = searchParams.get('state');
      const errorParam = searchParams.get('error');
      const errorDesc = searchParams.get('description');

      // Handle OAuth errors
      if (errorParam) {
        setError(errorDesc || `OAuth failed: ${errorParam}`);
        setIsProcessing(false);
        toast.error('Connection failed', {
          description: errorDesc || errorParam,
        });
        setTimeout(() => router.push('/accounts'), 3000);
        return;
      }

      // Validate required parameters
      if (!platform || !code || !state) {
        setError('Missing required parameters');
        setIsProcessing(false);
        toast.error('Invalid callback parameters');
        setTimeout(() => router.push('/accounts'), 3000);
        return;
      }

      try {
        // Complete OAuth connection
        const account = await socialOAuthService.completeOAuth(platform, code, state);
        
        toast.success('Account connected!', {
          description: `Successfully connected ${account.displayName} on ${platform}`,
        });
        
        // Redirect to accounts page
        router.push('/accounts');
      } catch (error: any) {
        console.error('OAuth completion failed:', error);
        setError(error.message || 'Failed to complete connection');
        setIsProcessing(false);
        
        toast.error('Connection failed', {
          description: error.message || 'Please try again',
        });
        
        // Redirect after delay
        setTimeout(() => router.push('/accounts'), 3000);
      }
    };

    handleCallback();
  }, [searchParams, router]);

  // In frontend/src/app/accounts/callback/page.tsx
useEffect(() => {
  // Debug session storage
  console.log('Session storage contents:', {
    oauth_state: sessionStorage.getItem('oauth_state'),
    oauth_platform: sessionStorage.getItem('oauth_platform'),
    oauth_team_id: sessionStorage.getItem('oauth_team_id'),
    allKeys: Object.keys(sessionStorage),
  });
  
  // Debug URL params
  console.log('URL params:', {
    platform: searchParams.get('platform'),
    code: searchParams.get('code'),
    state: searchParams.get('state'),
  });
  
  // Continue with existing handleCallback logic...
}, [searchParams]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="max-w-md w-full space-y-8 p-8">
        <div className="text-center">
          {isProcessing ? (
            <>
              <Loader2 className="h-12 w-12 animate-spin mx-auto text-primary" />
              <h2 className="mt-6 text-2xl font-bold">Completing connection...</h2>
              <p className="mt-2 text-sm text-muted-foreground">
                Please wait while we finalize your social account connection.
              </p>
            </>
          ) : error ? (
            <>
              <div className="rounded-full bg-destructive/10 p-3 mx-auto w-fit">
                <span className="text-4xl">❌</span>
              </div>
              <h2 className="mt-6 text-2xl font-bold">Connection Failed</h2>
              <p className="mt-2 text-sm text-destructive">{error}</p>
              <p className="mt-4 text-sm text-muted-foreground">
                Redirecting you back to accounts page...
              </p>
            </>
          ) : null}
        </div>
      </div>
    </div>
  );
}

// Main component with Suspense boundary
export default function OAuthCallbackPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="h-12 w-12 animate-spin text-primary" />
      </div>
    }>
      <OAuthCallbackContent />
    </Suspense>
  );
}