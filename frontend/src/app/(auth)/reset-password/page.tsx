// frontend/src/app/(auth)/reset-password/page.tsx
// Updated Version - Beautiful Modern Design with Dark/Light Theme

'use client';

import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useSearchParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { useResetPassword } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
  Lock, 
  CheckCircle, 
  AlertCircle,
  Eye,
  EyeOff,
  Shield,
  Check
} from 'lucide-react';
import { toast } from 'sonner';

// Validation schema
const resetPasswordSchema = z
  .object({
    newPassword: z
      .string()
      .min(8, 'Password must be at least 8 characters')
      .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
      .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
      .regex(/[0-9]/, 'Password must contain at least one number'),
    confirmPassword: z.string(),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
  });

type ResetPasswordFormData = z.infer<typeof resetPasswordSchema>;

// Password strength calculator
function calculatePasswordStrength(password: string) {
  let level = 0;
  if (password.length >= 8) level++;
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) level++;
  if (/[0-9]/.test(password)) level++;
  if (/[^a-zA-Z0-9]/.test(password)) level++;

  const labels = ['weak', 'fair', 'good', 'strong'];
  const colors = {
    weak: 'bg-red-500',
    fair: 'bg-yellow-500',
    good: 'bg-blue-500',
    strong: 'bg-green-500',
  };

  return { 
    level, 
    label: labels[level - 1] || 'weak',
    color: colors[labels[level - 1] as keyof typeof colors] || 'bg-gray-300'
  };
}

// Password strength indicator
function PasswordStrength({ password }: { password: string }) {
  if (!password) return null;

  const strength = calculatePasswordStrength(password);

  return (
    <div className="space-y-2 mt-3">
      <div className="flex gap-1">
        {[...Array(4)].map((_, i) => (
          <div
            key={i}
            className={`h-1.5 w-full rounded-full transition-all ${
              i < strength.level
                ? strength.color
                : 'bg-gray-200 dark:bg-gray-700'
            }`}
          />
        ))}
      </div>
      <p className="text-xs text-muted-foreground">
        Password strength:{' '}
        <span className={`font-medium ${
          strength.level >= 3 ? 'text-green-600 dark:text-green-400' :
          strength.level >= 2 ? 'text-blue-600 dark:text-blue-400' :
          'text-yellow-600 dark:text-yellow-400'
        }`}>
          {strength.label}
        </span>
      </p>
    </div>
  );
}

export default function ResetPasswordPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const token = searchParams.get('token');

  const [tokenValid, setTokenValid] = useState<boolean | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const { mutate: resetPassword, isPending, error } = useResetPassword();

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<ResetPasswordFormData>({
    resolver: zodResolver(resetPasswordSchema),
  });

  const password = watch('newPassword');

  useEffect(() => {
    // Validate token on mount
    if (!token) {
      setTokenValid(false);
    } else {
      setTokenValid(true);
    }
  }, [token]);

  const onSubmit = (data: ResetPasswordFormData) => {
    if (!token) return;

    resetPassword(
      {
        token,
        newPassword: data.newPassword,
      },
      {
        onSuccess: () => {
          toast.success('Password reset successful!');
          // Redirect to login
          setTimeout(() => {
            router.push('/login?reset=true');
          }, 2000);
        },
        onError: () => {
          toast.error('Failed to reset password');
        },
      }
    );
  };

  // Invalid Token State
  if (tokenValid === false) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 px-4 py-8">
        <Card className="w-full max-w-md shadow-2xl border-0 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm">
          <CardHeader className="text-center pb-8">
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-red-500 to-rose-600 shadow-lg">
              <AlertCircle className="h-8 w-8 text-white" />
            </div>
            <CardTitle className="text-2xl font-bold">Invalid Reset Link</CardTitle>
            <CardDescription className="text-base">
              This password reset link is invalid or has expired
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-4">
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Password reset links expire after 1 hour for security reasons.
                Please request a new reset link.
              </AlertDescription>
            </Alert>
          </CardContent>

          <CardFooter className="flex flex-col space-y-3">
            <Link href="/forgot-password" className="w-full">
              <Button className="w-full h-11 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 text-white shadow-lg font-medium">
                Request New Reset Link
              </Button>
            </Link>
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

  // Form State
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 px-4 py-8">
      <Card className="w-full max-w-md shadow-2xl border-0 bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm">
        <CardHeader className="space-y-1 text-center pb-8">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 shadow-lg">
            <Lock className="h-8 w-8 text-white" />
          </div>
          <CardTitle className="text-3xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
            Reset Password
          </CardTitle>
          <CardDescription className="text-base">
            Create a strong new password for your account
          </CardDescription>
        </CardHeader>

        <form onSubmit={handleSubmit(onSubmit)}>
          <CardContent className="space-y-6">
            {/* Error Alert */}
            {error && (
              <Alert variant="destructive">
                <AlertDescription>
                  {error.message || 'Failed to reset password. Please try again.'}
                </AlertDescription>
              </Alert>
            )}

            {/* New Password Input */}
            <div className="space-y-2">
              <Label htmlFor="newPassword" className="text-sm font-medium">
                New Password
              </Label>
              <div className="relative">
                <Input
                  id="newPassword"
                  type={showPassword ? 'text' : 'password'}
                  placeholder="Enter your new password"
                  autoComplete="new-password"
                  autoFocus
                  {...register('newPassword')}
                  disabled={isPending}
                  className={`h-11 pr-10 ${
                    errors.newPassword
                      ? 'border-red-500 focus-visible:ring-red-500'
                      : ''
                  }`}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showPassword ? (
                    <EyeOff className="h-5 w-5" />
                  ) : (
                    <Eye className="h-5 w-5" />
                  )}
                </button>
              </div>
              {errors.newPassword && (
                <p className="text-sm text-red-600 dark:text-red-400">
                  {errors.newPassword.message}
                </p>
              )}
              
              {/* Password Strength Indicator */}
              <PasswordStrength password={password} />

              {/* Password Requirements */}
              <div className="rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 p-3 mt-3">
                <p className="text-xs font-medium text-blue-900 dark:text-blue-100 mb-2">
                  Password must contain:
                </p>
                <ul className="space-y-1">
                  <li className={`text-xs flex items-center gap-2 ${
                    password?.length >= 8
                      ? 'text-green-700 dark:text-green-400'
                      : 'text-blue-700 dark:text-blue-300'
                  }`}>
                    <Check className={`h-3 w-3 ${
                      password?.length >= 8 ? 'opacity-100' : 'opacity-30'
                    }`} />
                    At least 8 characters
                  </li>
                  <li className={`text-xs flex items-center gap-2 ${
                    /[A-Z]/.test(password || '')
                      ? 'text-green-700 dark:text-green-400'
                      : 'text-blue-700 dark:text-blue-300'
                  }`}>
                    <Check className={`h-3 w-3 ${
                      /[A-Z]/.test(password || '') ? 'opacity-100' : 'opacity-30'
                    }`} />
                    One uppercase letter
                  </li>
                  <li className={`text-xs flex items-center gap-2 ${
                    /[a-z]/.test(password || '')
                      ? 'text-green-700 dark:text-green-400'
                      : 'text-blue-700 dark:text-blue-300'
                  }`}>
                    <Check className={`h-3 w-3 ${
                      /[a-z]/.test(password || '') ? 'opacity-100' : 'opacity-30'
                    }`} />
                    One lowercase letter
                  </li>
                  <li className={`text-xs flex items-center gap-2 ${
                    /[0-9]/.test(password || '')
                      ? 'text-green-700 dark:text-green-400'
                      : 'text-blue-700 dark:text-blue-300'
                  }`}>
                    <Check className={`h-3 w-3 ${
                      /[0-9]/.test(password || '') ? 'opacity-100' : 'opacity-30'
                    }`} />
                    One number
                  </li>
                </ul>
              </div>
            </div>

            {/* Confirm Password Input */}
            <div className="space-y-2">
              <Label htmlFor="confirmPassword" className="text-sm font-medium">
                Confirm New Password
              </Label>
              <div className="relative">
                <Input
                  id="confirmPassword"
                  type={showConfirmPassword ? 'text' : 'password'}
                  placeholder="Confirm your new password"
                  autoComplete="new-password"
                  {...register('confirmPassword')}
                  disabled={isPending}
                  className={`h-11 pr-10 ${
                    errors.confirmPassword
                      ? 'border-red-500 focus-visible:ring-red-500'
                      : ''
                  }`}
                />
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showConfirmPassword ? (
                    <EyeOff className="h-5 w-5" />
                  ) : (
                    <Eye className="h-5 w-5" />
                  )}
                </button>
              </div>
              {errors.confirmPassword && (
                <p className="text-sm text-red-600 dark:text-red-400">
                  {errors.confirmPassword.message}
                </p>
              )}
            </div>

            {/* Security Notice */}
            <Alert className="border-amber-500 bg-amber-50 dark:bg-amber-900/20">
              <Shield className="h-4 w-4 text-amber-600 dark:text-amber-400" />
              <AlertDescription className="text-amber-800 dark:text-amber-300">
                <strong>Security Note:</strong> After resetting your password, you'll
                be logged out of all devices.
              </AlertDescription>
            </Alert>
          </CardContent>

          <CardFooter className="flex flex-col space-y-4 pt-6">
            {/* Submit Button */}
            <Button
              type="submit"
              className="w-full h-11 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 text-white shadow-lg font-medium"
              disabled={isPending}
            >
              {isPending ? (
                <>
                  <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                  Resetting Password...
                </>
              ) : (
                <>
                  <CheckCircle className="mr-2 h-5 w-5" />
                  Reset Password
                </>
              )}
            </Button>

            {/* Back to Login */}
            <Link
              href="/login"
              className="text-sm text-muted-foreground hover:text-primary text-center"
            >
              Back to Login
            </Link>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}