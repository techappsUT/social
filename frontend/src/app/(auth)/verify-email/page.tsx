// frontend/src/app/(auth)/verify-email/page.tsx
// Updated Version - Beautiful Modern Design with Dark/Light Theme

'use client';

import { useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useVerifyEmail, useResendVerification } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { 
  Loader2, 
  CheckCircle, 
  AlertCircle, 
  Mail, 
  ArrowRight,
  RefreshCcw 
} from 'lucide-react';
import { toast } from 'sonner';

export default function VerifyEmailPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const token = searchParams.get('token');
  
  const [verificationStatus, setVerificationStatus] = useState<
    'loading' | 'success' | 'error'
  >('loading');
  const [errorMessage, setErrorMessage] = useState('');

  const { mutate: verifyEmail } = useVerifyEmail();
  const { mutate: resendVerification, isPending: isResending } = useResendVerification();

  useEffect(() => {
    if (token) {
      // Automatically verify on mount if token is present
      verifyEmail(
        { token },
        {
          onSuccess: () => {
            setVerificationStatus('success');
            toast.success('Email verified successfully!');
            // Redirect to login after 3 seconds
            setTimeout(() => {
              router.push('/login?verified=true');
            }, 3000);
          },
          onError: (error) => {
            setVerificationStatus('error');
            setErrorMessage(
              error.message || 'Verification failed. The link may have expired.'
            );
            toast.error('Verification failed');
          },
        }
      );
    } else {
      // No token provided
      setVerificationStatus('error');
      setErrorMessage('Invalid verification link. No token provided.');
    }
  }, [token, verifyEmail, router]);

  const handleResend = () => {
    const email = searchParams.get('email');
    if (!email) {
      toast.error('Email address is required to resend verification');
      return;
    }

    resendVerification(email, {
      onSuccess: () => {
        toast.success('Verification email sent! Check your inbox.');
      },
      onError: () => {
        toast.error('Failed to resend verification email');
      },
    });
  };

  // Loading State
  if (verificationStatus === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 px-4 py-8">
        <Card className="w-full max-w-md shadow-2xl border-0 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm">
          <CardHeader className="text-center pb-8">
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 shadow-lg">
              <Loader2 className="h-8 w-8 text-white animate-spin" />
            </div>
            <CardTitle className="text-2xl font-bold">
              Verifying Your Email
            </CardTitle>
            <CardDescription className="text-base">
              Please wait while we verify your email address...
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-4">
            <div className="flex justify-center">
              <div className="animate-pulse text-center">
                <p className="text-sm text-muted-foreground">
                  This will only take a moment
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Success State
  if (verificationStatus === 'success') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 px-4 py-8">
        <Card className="w-full max-w-md shadow-2xl border-0 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm">
          <CardHeader className="text-center pb-8">
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-green-500 to-emerald-600 shadow-lg animate-in zoom-in duration-300">
              <CheckCircle className="h-8 w-8 text-white" />
            </div>
            <CardTitle className="text-3xl font-bold bg-gradient-to-r from-green-600 to-emerald-600 bg-clip-text text-transparent">
              Email Verified!
            </CardTitle>
            <CardDescription className="text-base mt-2">
              Your email has been successfully verified
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-6">
            <Alert className="border-green-500 bg-green-50 dark:bg-green-900/20">
              <CheckCircle className="h-4 w-4 text-green-600 dark:text-green-400" />
              <AlertDescription className="text-green-800 dark:text-green-300">
                <strong>Success!</strong> You can now access all features of your
                account.
              </AlertDescription>
            </Alert>

            <div className="text-center space-y-2">
              <p className="text-sm text-muted-foreground">
                Redirecting you to login in a few seconds...
              </p>
              <div className="flex justify-center">
                <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
              </div>
            </div>
          </CardContent>

          <CardFooter className="flex flex-col space-y-3">
            <Link href="/login?verified=true" className="w-full">
              <Button className="w-full h-11 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 text-white shadow-lg font-medium">
                Continue to Login
                <ArrowRight className="ml-2 h-5 w-5" />
              </Button>
            </Link>
          </CardFooter>
        </Card>
      </div>
    );
  }

  // Error State
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 px-4 py-8">
      <Card className="w-full max-w-md shadow-2xl border-0 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm">
        <CardHeader className="text-center pb-8">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-red-500 to-rose-600 shadow-lg">
            <AlertCircle className="h-8 w-8 text-white" />
          </div>
          <CardTitle className="text-2xl font-bold">
            Verification Failed
          </CardTitle>
          <CardDescription className="text-base">
            We couldn't verify your email address
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{errorMessage}</AlertDescription>
          </Alert>

          <div className="rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 p-4">
            <div className="flex items-start gap-3">
              <Mail className="h-5 w-5 text-blue-600 dark:text-blue-400 mt-0.5" />
              <div className="space-y-1">
                <p className="text-sm font-medium text-blue-900 dark:text-blue-100">
                  Need a new verification link?
                </p>
                <p className="text-xs text-blue-700 dark:text-blue-300">
                  Verification links expire after 24 hours for security reasons.
                </p>
              </div>
            </div>
          </div>
        </CardContent>

        <CardFooter className="flex flex-col space-y-3">
          <Button
            onClick={handleResend}
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
                <RefreshCcw className="mr-2 h-5 w-5" />
                Resend Verification Email
              </>
            )}
          </Button>

          <Link
            href="/login"
            className="text-sm text-muted-foreground hover:text-primary text-center"
          >
            Back to Login
          </Link>
        </CardFooter>
      </Card>
    </div>
  );
}