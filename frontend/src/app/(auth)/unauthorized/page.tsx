// frontend/src/app/unauthorized/page.tsx
// Unauthorized Access Page - Beautiful Design

'use client';

import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
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
  ShieldAlert, 
  Home, 
  ArrowLeft, 
  Lock,
  Info
} from 'lucide-react';

export default function UnauthorizedPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const returnUrl = searchParams.get('returnUrl');
  const reason = searchParams.get('reason');

  // Get descriptive message based on reason
  const getMessage = () => {
    switch (reason) {
      case 'expired':
        return 'Your session has expired. Please log in again to continue.';
      case 'invalid':
        return 'Your authentication token is invalid. Please log in again.';
      case 'insufficient':
        return 'You don\'t have sufficient permissions to access this resource.';
      case 'team':
        return 'You don\'t have access to this team or resource.';
      default:
        return 'You don\'t have permission to access this page.';
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 px-4 py-8">
      <Card className="w-full max-w-md shadow-2xl border-0 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm">
        <CardHeader className="text-center pb-8">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-amber-500 to-orange-600 shadow-lg">
            <ShieldAlert className="h-8 w-8 text-white" />
          </div>
          <CardTitle className="text-3xl font-bold bg-gradient-to-r from-amber-600 to-orange-600 bg-clip-text text-transparent">
            Access Denied
          </CardTitle>
          <CardDescription className="text-base mt-2">
            {getMessage()}
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          {/* Main Alert */}
          <Alert className="border-amber-500 bg-amber-50 dark:bg-amber-900/20">
            <Lock className="h-4 w-4 text-amber-600 dark:text-amber-400" />
            <AlertDescription className="text-amber-800 dark:text-amber-300">
              <strong>Authorization Required</strong>
              <p className="mt-1 text-sm">
                {reason === 'expired' || reason === 'invalid' 
                  ? 'Your session has ended. Log in again to continue.'
                  : 'Contact your team administrator if you believe this is an error.'
                }
              </p>
            </AlertDescription>
          </Alert>

          {/* Helpful Info */}
          <div className="rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 p-4">
            <div className="flex items-start gap-3">
              <Info className="h-5 w-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
              <div className="space-y-2">
                <p className="text-sm font-medium text-blue-900 dark:text-blue-100">
                  Common reasons for access denial:
                </p>
                <ul className="text-xs text-blue-700 dark:text-blue-300 space-y-1 list-disc list-inside">
                  <li>Your login session has expired</li>
                  <li>You're not a member of the requested team</li>
                  <li>Your account role doesn't have the required permissions</li>
                  <li>The resource has been deleted or moved</li>
                </ul>
              </div>
            </div>
          </div>
        </CardContent>

        <CardFooter className="flex flex-col space-y-3 pt-6">
          {/* Primary Action - Login or Go Back */}
          {reason === 'expired' || reason === 'invalid' ? (
            <Link href={`/login${returnUrl ? `?returnUrl=${encodeURIComponent(returnUrl)}` : ''}`} className="w-full">
              <Button className="w-full h-11 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 text-white shadow-lg font-medium">
                <Lock className="mr-2 h-5 w-5" />
                Log In Again
              </Button>
            </Link>
          ) : (
            <Button
              onClick={() => router.back()}
              className="w-full h-11 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 text-white shadow-lg font-medium"
            >
              <ArrowLeft className="mr-2 h-5 w-5" />
              Go Back
            </Button>
          )}

          {/* Secondary Action - Home */}
          <Link href="/dashboard" className="w-full">
            <Button variant="outline" className="w-full h-11 font-medium">
              <Home className="mr-2 h-5 w-5" />
              Go to Dashboard
            </Button>
          </Link>

          {/* Contact Support */}
          <p className="text-center text-sm text-muted-foreground">
            Need help?{' '}
            <Link
              href="/support"
              className="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300"
            >
              Contact Support
            </Link>
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}