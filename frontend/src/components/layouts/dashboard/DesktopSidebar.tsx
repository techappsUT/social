import { TeamSwitcher } from '@/components/team-switcher';
import { cn } from '@/lib/utils';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import React from 'react'

const DashboardSidebar = ({navigation}:{
  navigation: {
    name: string;
    href: string;
    icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  }[];
}) => {
  const pathname = usePathname();
  return (
    <div className="flex grow flex-col overflow-y-auto border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 px-4 pb-4">
      {/* Logo */}
      <div className="flex h-16 shrink-0 items-center">
        <Link href="/dashboard" className="flex items-center gap-2">
          <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center">
            <span className="text-white font-bold text-lg">SQ</span>
          </div>
          <span className="text-xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
            SocialQueue
          </span>
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex flex-1 flex-col">
        <ul role="list" className="space-y-1">
        <TeamSwitcher />           
          {navigation.map((item) => {
            // ✅ UPDATED: Check if current path starts with item href for nested routes
            const isActive = 
              pathname === item.href || 
              (item.href !== '/dashboard' && pathname.startsWith(item.href));
            
            return (
              <li key={item.name}>
                <Link
                  href={item.href}
                  className={cn(
                    'group flex gap-x-3 rounded-lg p-3 text-sm leading-6 font-semibold transition-all',
                    isActive
                      ? 'bg-gradient-to-r from-indigo-500 to-purple-600 text-white shadow-lg shadow-indigo-500/50'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
                  )}
                >
                  <item.icon
                    className={cn(
                      'h-6 w-6 shrink-0',
                      isActive
                        ? 'text-white'
                        : 'text-gray-400 group-hover:text-indigo-600 dark:group-hover:text-indigo-400'
                    )}
                    aria-hidden="true"
                  />
                  {item.name}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>
    </div>
  )
}

export default DashboardSidebar