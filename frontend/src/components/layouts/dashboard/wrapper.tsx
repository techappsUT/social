import { TeamProvider } from '@/contexts/team-context'
import React from 'react'

const DashboardLayoutWrapper = ({children}: {children: React.ReactNode}) => {
  return (
    <TeamProvider>
      {children}
    </TeamProvider>
  )
}

export default DashboardLayoutWrapper