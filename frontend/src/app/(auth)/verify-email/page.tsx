'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { CheckCircle, XCircle, Loader2, Mail } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { useAuth } from '@/providers/auth-provider';

export default function VerifyEmailPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { refreshUser } = useAuth();
  
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle');
  const [message, setMessage] = useState('');
  const [devCode, setDevCode] = useState('');
  
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
      
      // Refresh user data
      await refreshUser();
      
      // Redirect to dashboard after 2 seconds
      setTimeout(() => {
        router.push('/dashboard');
      }, 2000);
    } catch (error) {
      setStatus('error');
      setMessage(error instanceof Error ? error.message : 'Verification failed');
    }
  };

  const handleDevVerification = () => {
    if (devCode) {
      verifyEmail(devCode);
    }
  };

  const resendEmail = async () => {
    try {
      await apiClient.post('/auth/resend-verification', 
        { email: localStorage.getItem('userEmail') || '' },
        { skipAuth: true }
      );
      setMessage('Verification email sent!');
    } catch (error) {
      setMessage('Failed to resend email');
    }
  };

  if (status === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <Loader2 className="h-12 w-12 animate-spin mx-auto text-primary mb-4" />
            <CardTitle>Verifying Your Email</CardTitle>
            <CardDescription>Please wait...</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (status === 'success') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-green-50 to-emerald-100">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-4" />
            <CardTitle>Email Verified!</CardTitle>
            <CardDescription>{message}</CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-center text-sm text-muted-foreground">
              Redirecting to dashboard...
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <Mail className="h-12 w-12 text-primary mx-auto mb-4" />
          <CardTitle>Verify Your Email</CardTitle>
          <CardDescription>
            Check your email for the verification link
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {status === 'error' && (
            <Alert variant="destructive">
              <XCircle className="h-4 w-4" />
              <AlertDescription>{message}</AlertDescription>
            </Alert>
          )}
          
          {/* Development Mode Verification */}
          {process.env.NODE_ENV === 'development' && (
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">
                Development Mode: Enter verification code
              </p>
              <div className="flex gap-2">
                <Input
                  placeholder="Enter code (e.g., 123456)"
                  value={devCode}
                  onChange={(e) => setDevCode(e.target.value)}
                />
                <Button onClick={handleDevVerification}>
                  Verify
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                Check backend console for the verification code
              </p>
            </div>
          )}
          
          <div className="pt-4 space-y-2">
            <Button 
              variant="outline" 
              className="w-full"
              onClick={resendEmail}
            >
              Resend Verification Email
            </Button>
            
            <Button 
              variant="ghost" 
              className="w-full"
              onClick={() => router.push('/login')}
            >
              Back to Login
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}