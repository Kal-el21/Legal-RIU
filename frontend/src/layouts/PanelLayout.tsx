import { Outlet, Link, useLocation } from 'react-router-dom'
import { Scale, Menu, Home, ChevronRight, ChevronLeft, type LucideIcon } from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import SidebarUserButton from '@/components/common/SidebarUserButton'
import NotificationDropdown from '@/components/common/NotificationDropdown'
import { useAuthStore } from '@/store/auth.store'

export type MenuItem = {
  label: string
  href: string
  icon: LucideIcon
  permissions: string[]
  exact?: boolean
  badge?: number
}

// Grup menu yang bisa dilipat, mis. Data Master di panel Admin
export type MenuGroup = {
  label: string
  icon: LucideIcon
  children: MenuItem[]
}

interface PanelLayoutProps {
  panelLabel: string
  settingsPath: string
  nav: (MenuItem | MenuGroup)[]
}

// Kerangka semua panel setelah login: sidebar, header, dan area konten
export default function PanelLayout({ panelLabel, settingsPath, nav }: PanelLayoutProps) {
  const location = useLocation()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [openGroup, setOpenGroup] = useState<string | null>(null)

  const isActive = (href: string, exact?: boolean) =>
    exact ? location.pathname === href : location.pathname.startsWith(href)

  const permissions = useAuthStore((state) => state.permissions)
  const user = useAuthStore((state) => state.user)

  const allowed = (item: MenuItem) =>
    user?.role === 'ADMIN' || item.permissions.some((permission) => permissions.includes(permission))

  const navItems = nav
    .map((item) => ('children' in item ? { ...item, children: item.children.filter(allowed) } : item))
    .filter((item) => ('children' in item ? item.children.length > 0 : allowed(item)))

  return (
    <div className="min-h-screen flex" style={{ background: '#f8fafc' }}>
      <aside className={cn(
        'fixed inset-y-0 left-0 z-50 flex flex-col transition-all duration-200',
        sidebarCollapsed ? 'w-16' : 'w-60',
        'border-r border-white/10',
        'lg:translate-x-0 lg:h-screen',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
      )} style={{ background: '#0B2545' }}>
        <div className="h-16 flex items-center px-5 border-b border-white/10">
          <Link to="/" className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: '#C8102E' }}>
              <Scale className="w-4 h-4 text-white" />
            </div>
            {!sidebarCollapsed && (
              <div>
                <p className="font-bold text-sm text-white">Legal RIU</p>
                <p className="text-xs text-white/40 leading-none">{panelLabel}</p>
              </div>
            )}
          </Link>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
          {navItems.map((item) => 'children' in item ? (
            <div key={item.label} className="py-2">
              <button
                onClick={() => setOpenGroup(openGroup === item.label ? null : item.label)}
                className={cn(
                  'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all w-full',
                  'text-white/50 hover:bg-white/5 hover:text-white'
                )}
              >
                <item.icon className="w-4 h-4 flex-shrink-0" />
                {!sidebarCollapsed && <span>{item.label}</span>}
                {!sidebarCollapsed && <ChevronRight className={cn('w-3.5 h-3.5 ml-auto opacity-60 transition-transform', openGroup === item.label && 'rotate-90')} />}
              </button>
              {openGroup === item.label && !sidebarCollapsed && (
                <div className="ml-4 mt-1 space-y-0.5">
                  {item.children.map((child) => (
                    <Link key={child.href} to={child.href} onClick={() => setSidebarOpen(false)}
                      className={cn(
                        'flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-medium transition-all',
                        isActive(child.href)
                          ? 'bg-white/10 text-white'
                          : 'text-white/50 hover:bg-white/5 hover:text-white'
                      )}
                    >
                      <child.icon className="w-3.5 h-3.5 flex-shrink-0" />
                      {child.label}
                    </Link>
                  ))}
                </div>
              )}
            </div>
          ) : (
            <Link key={item.href} to={item.href} onClick={() => setSidebarOpen(false)}
              className={cn(
                'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all',
                sidebarCollapsed && 'justify-center px-2',
                isActive(item.href, item.exact)
                  ? 'bg-white/10 text-white'
                  : 'text-white/50 hover:bg-white/5 hover:text-white'
              )}
              title={sidebarCollapsed ? item.label : undefined}
            >
              <span className="relative flex-shrink-0">
                <item.icon className="w-4 h-4" />
                {item.badge ? (
                  <span className="absolute -top-1.5 -right-2 inline-flex items-center justify-center rounded-full text-[10px] font-bold px-1 min-w-[16px] h-4 bg-red-500 text-white border-2 border-[#0B2545]">
                    {item.badge}
                  </span>
                ) : null}
              </span>
              {!sidebarCollapsed && item.label}
              {!sidebarCollapsed && isActive(item.href, item.exact) && <ChevronRight className="w-3.5 h-3.5 ml-auto opacity-60" />}
            </Link>
          ))}
        </nav>

        <SidebarUserButton settingsPath={settingsPath} collapsed={sidebarCollapsed} />
      </aside>

      {sidebarOpen && (
        <div className="fixed inset-0 z-40 bg-black/30 lg:hidden" onClick={() => setSidebarOpen(false)} />
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
          <Link to="/" title="Ke beranda portal" className="inline-flex h-9 items-center gap-2 rounded-lg px-2 text-sm text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-800">
            <Home className="h-5 w-5" />
            <span className="hidden sm:inline">Beranda</span>
          </Link>
        </header>
        <main className="flex-1"><Outlet /></main>
      </div>
    </div>
  )
}
