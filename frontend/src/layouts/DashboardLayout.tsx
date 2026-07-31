import { Outlet, Link, useLocation } from 'react-router-dom'
import { Scale, LayoutDashboard, FileText, FileSearch, Menu, ChevronRight, ChevronLeft, BriefcaseBusiness, ScrollText } from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import SidebarUserButton from '@/components/common/SidebarUserButton'
import NotificationDropdown from '@/components/common/NotificationDropdown'
import { useAuthStore } from '@/store/auth.store'

const NAV = [
  { label: 'Dashboard', href: '/dashboard', icon: LayoutDashboard, exact: true, permissions: ['dashboard.user.view'] },
  { label: 'Legal Opinion', href: '/dashboard/legal-opinions', icon: FileText, permissions: ['legal_opinion.view.own', 'legal_opinion.view.all'] },
  { label: 'Review Dokumen', href: '/dashboard/review-documents', icon: FileSearch, permissions: ['document_review.view.own', 'document_review.view.all'] },
  { label: 'Dokumen Perjanjian', href: '/dashboard/agreement-documents', icon: FileText, permissions: ['agreement_document.view.own'] },
  { label: 'Case Management', href: '/dashboard/legal-cases', icon: BriefcaseBusiness, permissions: ['case_management.view'] },
  { label: 'Audit Log', href: '/dashboard/audit-logs', icon: ScrollText, permissions: ['audit_log.view'] },
]

export default function DashboardLayout() {
  const location = useLocation()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const permissions = useAuthStore((state) => state.permissions)
  const user = useAuthStore((state) => state.user)

  const isActive = (href: string, exact?: boolean) =>
    exact ? location.pathname === href : location.pathname.startsWith(href)
  const navItems = NAV.filter((item) =>
    user?.role === 'ADMIN' || item.permissions.some((permission) => permissions.includes(permission))
  )

  return (
    <div className="min-h-screen flex" style={{ background: '#f8fafc' }}>
      <aside className={cn(
        'fixed inset-y-0 left-0 z-50 flex flex-col transition-all duration-200',
        sidebarCollapsed ? 'w-16' : 'w-60',
        'bg-white border-r border-gray-100',
        'lg:translate-x-0 lg:h-screen',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
      )}>
        <div className="h-16 flex items-center px-5 border-b border-gray-100">
          <Link to="/" className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: '#C8102E' }}>
              <Scale className="w-4 h-4 text-white" />
            </div>
            {!sidebarCollapsed && (
              <div>
                <p className="font-bold text-sm" style={{ color: '#0B2545' }}>Legal RIU</p>
                <p className="text-xs text-gray-400 leading-none">Indonesia Re</p>
              </div>
            )}
          </Link>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
          {navItems.map((item) => (
            <Link key={item.href} to={item.href} onClick={() => setSidebarOpen(false)}
              className={cn(
                'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all',
                sidebarCollapsed && 'justify-center px-2',
                isActive(item.href, item.exact)
                  ? 'text-white shadow-sm'
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
              )}
              style={isActive(item.href, item.exact) ? { background: '#C8102E' } : {}}
              title={sidebarCollapsed ? item.label : undefined}
            >
              <item.icon className="w-4 h-4 flex-shrink-0" />
              {!sidebarCollapsed && item.label}
              {!sidebarCollapsed && isActive(item.href, item.exact) && <ChevronRight className="w-3.5 h-3.5 ml-auto opacity-70" />}
            </Link>
          ))}
        </nav>

        <SidebarUserButton settingsPath="/dashboard/settings" dark={false} collapsed={sidebarCollapsed} />
      </aside>

      {sidebarOpen && (
        <div className="fixed inset-0 z-40 bg-black/20 lg:hidden" onClick={() => setSidebarOpen(false)} />
      )}

      <div className={cn(
        'flex-1 flex flex-col min-w-0 transition-all duration-200',
        sidebarCollapsed ? 'lg:ml-16' : 'lg:ml-60'
      )}>
        <header className="h-16 bg-white border-b border-gray-100 flex items-center px-4 gap-3 sticky top-0 z-30">
          <button onClick={() => setSidebarOpen(true)} className="lg:hidden p-2 rounded-lg hover:bg-gray-100">
            <Menu className="w-5 h-5 text-gray-600" />
          </button>
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="hidden lg:flex p-2 rounded-lg hover:bg-gray-100"
            title={sidebarCollapsed ? 'Buka Sidebar' : 'Tutup Sidebar'}
          >
            {sidebarCollapsed ? <ChevronRight className="w-5 h-5 text-gray-600" /> : <ChevronLeft className="w-5 h-5 text-gray-600" />}
          </button>
          <div className="flex-1" />
          <NotificationDropdown />
          <Link to="/" className="text-sm text-gray-500 hover:text-gray-700 transition-colors">
            ← Kembali ke Beranda
          </Link>
        </header>
        <main className="flex-1"><Outlet /></main>
      </div>
    </div>
  )
}
