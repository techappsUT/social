import React from 'react'
import { useState, useEffect } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { Menu, Settings, User } from 'lucide-react';
import Link from 'next/link';
import { useCurrentUser } from '@/hooks/useAuth';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import UserMenu from './UserMenu';
import { cn } from '@/lib/utils';
import DashboardSidebar from './DesktopSidebar';

const MobileHeader = ({navigation}:{
  navigation: {
    name: string;
    href: string;
    icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  }[];
}) => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const pathname = usePathname();
  return (
    <div className="sticky top-0 z-40 flex items-center gap-x-6 bg-white dark:bg-gray-900 px-4 py-4 shadow-sm sm:px-6 lg:hidden border-b border-gray-200 dark:border-gray-800">
        <Sheet open={isMobileMenuOpen} onOpenChange={setIsMobileMenuOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="-m-2.5">
              <Menu className="h-5 w-5" aria-hidden="true" />
              <span className="sr-only">Open sidebar</span>
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-72 p-0">
            <DashboardSidebar navigation={navigation} />
          </SheetContent>
        </Sheet>

        <div className="flex-1 text-sm font-semibold leading-6 text-gray-900 dark:text-gray-100">
          <Link href="/dashboard" className="flex items-center gap-2">
            <div className="h-7 w-7 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center">
              <span className="text-white font-bold text-sm">SQ</span>
            </div>
            <span className="text-lg font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
              SocialQueue
            </span>
          </Link>
        </div>

       <UserMenu />
      </div>
  )
}

export default MobileHeader