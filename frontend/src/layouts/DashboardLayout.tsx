import { Outlet, Link, useLocation } from 'react-router-dom'
import { Scale, LayoutDashboard, FileText, FileSearch, Menu, ChevronRight } from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import SidebarUserButton from '@/components/common/SidebarUserButton'

const NAV = [
  { label: 'Dashboard', href: '/dashboard', icon: LayoutDashboard, exact: true },
  { label: 'Legal Opinion', href: '/dashboard/legal-opinions', icon: FileText },
  { label: 'Review Dokumen', href: '/dashboard/review-documents', icon: FileSearch },
]

export default function DashboardLayout() {
  const location = useLocation()
  const [sidebarOpen, setSidebarOpen] = useState(false)

  const isActive = (href: string, exact?: boolean) =>
    exact ? location.pathname === href : location.pathname.startsWith(href)

  return (
    <div className="min-h-screen flex" style={{ background: '#f8fafc' }}>
      <aside className={cn(
        'fixed inset-y-0 left-0 z-50 w-60 bg-white border-r border-gray-100 flex flex-col transition-transform duration-200',
         'lg:translate-x-0 lg:static lg:h-screen lg:z-auto',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
      )}>
        <div className="h-16 flex items-center px-5 border-b border-gray-100">
          <Link to="/" className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: '#C8102E' }}>
              <Scale className="w-4 h-4 text-white" />
            </div>
            <div>
              <p className="font-bold text-sm" style={{ color: '#0B2545' }}>Legal RIU</p>
              <p className="text-xs text-gray-400 leading-none">Indonesia Re</p>
            </div>
          </Link>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
          {NAV.map((item) => (
            <Link key={item.href} to={item.href} onClick={() => setSidebarOpen(false)}
              className={cn(
                'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all',
                isActive(item.href, item.exact)
                  ? 'text-white shadow-sm'
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
              )}
              style={isActive(item.href, item.exact) ? { background: '#C8102E' } : {}}>
              <item.icon className="w-4 h-4 flex-shrink-0" />
              {item.label}
              {isActive(item.href, item.exact) && <ChevronRight className="w-3.5 h-3.5 ml-auto opacity-70" />}
            </Link>
          ))}
        </nav>

        {/* User button with popover */}
        <SidebarUserButton settingsPath="/dashboard/settings" dark={false} />
      </aside>

      {sidebarOpen && (
        <div className="fixed inset-0 z-40 bg-black/20 lg:hidden" onClick={() => setSidebarOpen(false)} />
      )}

      <div className="flex-1 flex flex-col min-w-0">
        <header className="h-16 bg-white border-b border-gray-100 flex items-center px-4 gap-3 sticky top-0 z-30">
          <button onClick={() => setSidebarOpen(true)} className="lg:hidden p-2 rounded-lg hover:bg-gray-100">
            <Menu className="w-5 h-5 text-gray-600" />
          </button>
          <div className="flex-1" />
          <Link to="/" className="text-sm text-gray-500 hover:text-gray-700 transition-colors">
            ← Kembali ke Beranda
          </Link>
        </header>
        <main className="flex-1"><Outlet /></main>
      </div>
    </div>
  )
}