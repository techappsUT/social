// frontend/src/app/(auth)/verify-email/page.tsx
// Updated to match login/signup design

'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { 
  CheckCircle, 
  XCircle, 
  Loader2, 
  Mail, 
  ArrowLeft,
  Send,
  ShieldCheck,
  Clock
} from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { useAuth } from '@/providers/auth-provider';
import { toast } from 'sonner';

export default function VerifyEmailPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { refreshUser } = useAuth();
  
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle');
  const [message, setMessage] = useState('');
  const [devCode, setDevCode] = useState('');
  const [isResending, setIsResending] = useState(false);
  
  // Check if we have a token in URL
  const token = searchParams.get('token');
  
  useEffect(() => {
    if (token) {
      verifyEmail(token);
    }
  }, [token]);

  const verifyEmail = async (verificationToken: string) => {
    setStatus('loading');
    
    try {
      await apiClient.post('/auth/verify-email', 
        { token: verificationToken },
        { skipAuth: true }
      );
      
      setStatus('success');
      setMessage('Email verified successfully!');
      toast.success('Email verified!', {
        description: 'Redirecting to dashboard...'
      });
      
      // Refresh user data
      await refreshUser();
      
      // Redirect to dashboard after 2 seconds
      setTimeout(() => {
        router.push('/dashboard');
      }, 2000);
    } catch (error) {
      setStatus('error');
      setMessage(error instanceof Error ? error.message : 'Verification failed. The link may be expired or invalid.');
      toast.error('Verification failed', {
        description: 'Please try requesting a new verification link.'
      });
    }
  };

  const handleDevVerification = () => {
    if (devCode) {
      verifyEmail(devCode);
    }
  };

  const resendEmail = async () => {
    setIsResending(true);
    try {
      const email = localStorage.getItem('userEmail');
      if (!email) {
        toast.error('Email not found', {
          description: 'Please log in again to resend verification.'
        });
        return;
      }

      await apiClient.post('/auth/resend-verification', 
        { email },
        { skipAuth: true }
      );
      
      toast.success('Verification email sent!', {
        description: 'Check your inbox for the verification link.'
      });
      setMessage('Verification email sent! Check your inbox.');
    } catch (error) {
      toast.error('Failed to resend email', {
        description: 'Please try again in a few moments.'
      });
      setMessage('Failed to resend email. Please try again.');
    } finally {
      setIsResending(false);
    }
  };

  // Loading State
  if (status === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 px-4 py-8">
        <Card className="w-full max-w-md shadow-2xl border-0 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm">
          <CardHeader className="text-center pb-8">
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 shadow-lg">
              <Loader2 className="h-8 w-8 animate-spin text-white" />
            </div>
            <CardTitle className="text-3xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
              Verifying Your Email
            </CardTitle>
            <CardDescription className="text-base mt-2">
              Please wait while we verify your email address...
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center gap-3 text-sm text-muted-foreground">
                <Clock className="h-5 w-5 text-indigo-500" />
                <span>This usually takes just a few seconds</span>
              </div>
              <div className="flex items-center gap-3 text-sm text-muted-foreground">
                <ShieldCheck className="h-5 w-5 text-purple-500" />
                <span>Verifying your account security</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Success State
  if (status === 'success') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 px-4 py-8">
        <Card className="w-full max-w-md shadow-2xl border-0 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm">
          <CardHeader className="text-center pb-8">
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-green-500 to-emerald-600 shadow-lg">
              <CheckCircle className="h-8 w-8 text-white" />
            </div>
            <CardTitle className="text-3xl font-bold bg-gradient-to-r from-green-600 to-emerald-600 bg-clip-text text-transparent">
              Email Verified!
            </CardTitle>
            <CardDescription className="text-base mt-2">
              {message}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Alert className="border-green-500 bg-green-50 dark:bg-green-900/20">
              <CheckCircle className="h-4 w-4 text-green-600 dark:text-green-400" />
              <AlertDescription className="text-green-800 dark:text-green-300">
                <strong>Success!</strong> Your email has been verified. Redirecting you to the dashboard...
              </AlertDescription>
            </Alert>
          </CardContent>
          <CardFooter>
            <div className="w-full flex items-center justify-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>Redirecting to dashboard...</span>
            </div>
          </CardFooter>
        </Card>
      </div>
    );
  }

  // Idle/Error State (Main Verify Email Screen)
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 px-4 py-8">
      <Card className="w-full max-w-md shadow-2xl border-0 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm">
        <CardHeader className="text-center pb-8">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 shadow-lg">
            <Mail className="h-8 w-8 text-white" />
          </div>
          <CardTitle className="text-3xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
            Verify Your Email
          </CardTitle>
          <CardDescription className="text-base mt-2">
            Check your email for the verification link
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          {/* Error Alert */}
          {status === 'error' && (
            <Alert variant="destructive">
              <XCircle className="h-4 w-4" />
              <AlertDescription>{message}</AlertDescription>
            </Alert>
          )}

          {/* Success Message (after resend) */}
          {status === 'idle' && message && (
            <Alert className="border-green-500 bg-green-50 dark:bg-green-900/20">
              <CheckCircle className="h-4 w-4 text-green-600 dark:text-green-400" />
              <AlertDescription className="text-green-800 dark:text-green-300">
                {message}
              </AlertDescription>
            </Alert>
          )}

          {/* Instructions */}
          <div className="space-y-3 text-sm text-muted-foreground">
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 mt-0.5">
                <div className="flex h-6 w-6 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 text-xs font-bold">
                  1
                </div>
              </div>
              <p>Check your inbox for an email from SocialQueue</p>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 mt-0.5">
                <div className="flex h-6 w-6 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 text-xs font-bold">
                  2
                </div>
              </div>
              <p>Click the verification link in the email</p>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 mt-0.5">
                <div className="flex h-6 w-6 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 text-xs font-bold">
                  3
                </div>
              </div>
              <p>You'll be automatically redirected to your dashboard</p>
            </div>
          </div>

          {/* Help Text */}
          <div className="rounded-lg bg-muted/50 p-4">
            <p className="text-xs text-muted-foreground">
              <strong>Didn't receive the email?</strong> Check your spam folder or click the button below to resend.
            </p>
          </div>

          {/* Development Mode Code Input */}
          {process.env.NODE_ENV === 'development' && (
            <div className="space-y-3 p-4 rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800">
              <p className="text-sm font-medium text-amber-800 dark:text-amber-300">
                🔧 Development Mode
              </p>
              <p className="text-xs text-amber-700 dark:text-amber-400">
                Enter the verification code from the backend console
              </p>
              <div className="space-y-2">
                <Label htmlFor="devCode" className="text-sm text-amber-800 dark:text-amber-300">
                  Verification Code
                </Label>
                <div className="flex gap-2">
                  <Input
                    id="devCode"
                    placeholder="e.g., 123456"
                    value={devCode}
                    onChange={(e) => setDevCode(e.target.value)}
                    className="bg-white dark:bg-gray-800"
                  />
                  <Button 
                    onClick={handleDevVerification}
                    className="bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-700 hover:to-orange-700 text-white"
                  >
                    Verify
                  </Button>
                </div>
              </div>
            </div>
          )}
        </CardContent>

        <CardFooter className="flex flex-col space-y-4 pt-6">
          {/* Resend Email Button */}
          <Button
            onClick={resendEmail}
            disabled={isResending}
            className="w-full h-11 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 text-white shadow-lg font-medium"
          >
            {isResending ? (
              <>
                <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                Sending...
              </>
            ) : (
              <>
                <Send className="mr-2 h-5 w-5" />
                Resend Verification Email
              </>
            )}
          </Button>

          {/* Divider */}
          <div className="relative w-full">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-white dark:bg-gray-800 px-2 text-muted-foreground">
                or
              </span>
            </div>
          </div>

          {/* Back to Login */}
          <Link href="/login" className="w-full">
            <Button 
              variant="outline" 
              className="w-full h-11 border-2 hover:bg-indigo-50 dark:hover:bg-indigo-950 hover:border-indigo-600 transition-colors font-medium"
            >
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back to Login
            </Button>
          </Link>

          {/* Help Text */}
          <p className="text-center text-sm text-muted-foreground">
            Need help?{' '}
            <Link
              href="/support"
              className="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300 underline-offset-4 hover:underline"
            >
              Contact Support
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}