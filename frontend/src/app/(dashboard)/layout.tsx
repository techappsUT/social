// frontend/src/app/(dashboard)/layout.tsx
// ✅ UPDATED: Dashboard layout with Teams navigation added

'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useCurrentUser } from '@/hooks/useAuth';
import { Button } from '@/components/ui/button';
import {
  LayoutDashboard,
  PenSquare,
  Calendar,
  BarChart3,
  Link2,
  Settings,
  Loader2,
  Bell,
  Search,
  Users, // ✅ NEW: Added for Teams
} from 'lucide-react';
import DashboardLayoutWrapper from '@/components/layouts/dashboard/wrapper';
import DesktopSidebar from '@/components/layouts/dashboard/DesktopSidebar';
import UserMenu from '@/components/layouts/dashboard/UserMenu';
import MobileHeader from '@/components/layouts/dashboard/MobileHeader';

// ✅ UPDATED: Navigation items configuration with Teams added
const navigation = [
  {
    name: 'Dashboard',
    href: '/dashboard',
    icon: LayoutDashboard,
    description: 'Overview and analytics',
  },
  {
    name: 'Compose',
    href: '/compose',
    icon: PenSquare,
    description: 'Create new posts',
  },
  {
    name: 'Queue',
    href: '/queue',
    icon: Calendar,
    description: 'Scheduled posts',
  },
  {
    name: 'Teams', // ✅ NEW: Teams navigation item
    href: '/teams',
    icon: Users,
    description: 'Manage teams and collaborate',
  },
  {
    name: 'Analytics',
    href: '/analytics',
    icon: BarChart3,
    description: 'Performance metrics',
  },
  {
    name: 'Accounts',
    href: '/accounts',
    icon: Link2,
    description: 'Connected social accounts',
  },
  {
    name: 'Settings',
    href: '/settings',
    icon: Settings,
    description: 'Account settings',
  },
];


export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  
  // ✅ Use React Query directly instead of auth-provider
  const { data: user, isLoading } = useCurrentUser();
  const isAuthenticated = !!user;
  const isEmailVerified = user?.emailVerified || false;
  const [mounted, setMounted] = useState(false);

  // Prevent hydration mismatch
  useEffect(() => {
    setMounted(true);
  }, []);

  // ✅ Protect dashboard routes with proper checks
  useEffect(() => {
    if (!isLoading && mounted) {
      console.log('🔍 Dashboard auth check:', {
        isAuthenticated,
        user: user?.email,
        emailVerified: user?.emailVerified,
        isEmailVerified
      });

      // Not authenticated - redirect to login
      if (!isAuthenticated) {
        console.log('❌ Not authenticated, redirecting to login');
        router.push('/login');
        return;
      }
      
      // Email not verified - redirect to verify page
      if (!isEmailVerified) {
        console.log('⚠️ Email not verified, redirecting to verify-email');
        router.push('/verify-email');
        return;
      }

      console.log('✅ Auth check passed, staying on dashboard');
    }
  }, [isLoading, isAuthenticated, isEmailVerified, mounted, router, user]);

  // Loading state
  if (isLoading || !mounted) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-950">
        <div className="text-center space-y-4">
          <Loader2 className="h-12 w-12 animate-spin text-primary mx-auto" />
          <p className="text-muted-foreground">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  // Not authenticated or email not verified - show nothing (will redirect via useEffect)
  if (!user || !isEmailVerified) {
    return null;
  }

  return (
    <DashboardLayoutWrapper>
    <div className="min-h-screen bg-gray-50 dark:bg-gray-950">
      {/* Desktop Sidebar */}
      
    <aside className="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-72 lg:flex-col">
      <DesktopSidebar navigation={navigation}/>
    </aside>
      {/* Mobile header */}
      <MobileHeader navigation={navigation} />

      {/* Main content area */}
      <main className="lg:pl-72">
        {/* Top bar for desktop */}
        <div className="sticky top-0 z-40 hidden lg:flex h-16 shrink-0 items-center gap-x-4 border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 px-4 shadow-sm sm:gap-x-6 sm:px-6 lg:px-8">
          <div className="flex flex-1 gap-x-4 self-stretch lg:gap-x-6">
            <div className="flex flex-1 items-center">
              {/* Search bar placeholder */}
              <div className="relative flex-1 max-w-md">
                <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search..."
                  className="w-full rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 pl-10 pr-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
              </div>
            </div>

            <div className="flex items-center gap-x-4 lg:gap-x-6">
              {/* Notifications */}
              <Button variant="ghost" size="icon" className="relative">
                <Bell className="h-5 w-5" />
                <span className="absolute top-1 right-1 h-2 w-2 rounded-full bg-red-600" />
              </Button>

              {/* User menu */}
              <UserMenu />
            </div>
          </div>
        </div>

        {/* Page content */}
        <div className="px-4 py-8 sm:px-6 lg:px-8">
          {/* <TeamAccessCheck> */}
          {children}
          {/* </TeamAccessCheck> */}
        </div>
      </main>
    </div>
    </DashboardLayoutWrapper>
  );
}