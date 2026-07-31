import { Outlet, Link, useLocation } from 'react-router-dom'
import { Scale, LayoutDashboard, FileText, FileSearch, Users, Menu, ChevronRight, ChevronLeft, ScrollText, Building2, Target, Briefcase, MapPin, UserCog, FolderOpen, BarChart3 } from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import SidebarUserButton from '@/components/common/SidebarUserButton'
import NotificationDropdown from '@/components/common/NotificationDropdown'
import { useAuthStore } from '@/store/auth.store'

const DATA_MASTER_ITEMS = [
  { label: 'Perusahaan', href: '/admin/companies', icon: Building2, permissions: ['master_data.manage'] },
  { label: 'Tujuan Pembuatan', href: '/admin/purpose-types', icon: Target, permissions: ['master_data.manage'] },
  { label: 'Jenis Dokumen', href: '/admin/document-types', icon: FileText, permissions: ['master_data.manage'] },
  { label: 'Jenis Kasus', href: '/admin/case-types', icon: Briefcase, permissions: ['master_data.manage'] },
  { label: 'Kategori', href: '/admin/case-categories', icon: FolderOpen, permissions: ['master_data.manage'] },
  { label: 'Kabupaten/Kota', href: '/admin/regencies', icon: MapPin, permissions: ['master_data.manage'] },
  { label: 'Cedant', href: '/admin/cedants', icon: UserCog, permissions: ['master_data.manage'] },
  { label: 'Divisi', href: '/admin/divisions', icon: Users, permissions: ['master_data.manage'] },
  { label: 'Pihak Pertama', href: '/admin/agreement-company-master', icon: Building2, permissions: ['agreement_company_master.view.all'] },
]

const NAV = [
  { label: 'Dashboard', href: '/admin', icon: LayoutDashboard, exact: true, permissions: ['dashboard.admin.view'] },
  { label: 'Laporan', href: '/admin/reports', icon: BarChart3, permissions: ['report.legal_case.view', 'report.legal_opinion.view', 'report.document_review.view'] },
  { label: 'Legal Opinion', href: '/admin/legal-opinions', icon: FileText, permissions: ['legal_opinion.view.all'] },
  { label: 'Review Dokumen', href: '/admin/review-documents', icon: FileSearch, permissions: ['document_review.view.all'] },
  { label: 'Dokumen Perjanjian', href: '/admin/agreement-documents', icon: FileText, permissions: ['agreement_document.view.all'] },
  { label: 'Manajemen Kasus', href: '/admin/legal-cases', icon: Scale, permissions: ['case_management.view'] },
  { label: 'User Management', href: '/admin/users', icon: Users, permissions: ['user_management.view'] },
  { label: 'Audit Log', href: '/admin/audit-logs', icon: ScrollText, permissions: ['audit_log.view'] },
]

export default function AdminLayout() {
  const location = useLocation()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [dataMasterOpen, setDataMasterOpen] = useState(false)

  const isActive = (href: string, exact?: boolean) =>
    exact ? location.pathname === href : location.pathname.startsWith(href)

  const permissions = useAuthStore((state) => state.permissions)
  const user = useAuthStore((state) => state.user)

  const navItems = NAV.filter((item) =>
    user?.role === 'ADMIN' || item.permissions.some((permission) => permissions.includes(permission))
  )

  const dataMasterItems = DATA_MASTER_ITEMS.filter((item) =>
    user?.role === 'ADMIN' || item.permissions.some((permission) => permissions.includes(permission))
  )

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
                <p className="text-xs text-white/40 leading-none">Admin Panel</p>
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
                  ? 'bg-white/10 text-white'
                  : 'text-white/50 hover:bg-white/5 hover:text-white'
              )}
              title={sidebarCollapsed ? item.label : undefined}
            >
              <item.icon className="w-4 h-4 flex-shrink-0" />
              {!sidebarCollapsed && item.label}
              {!sidebarCollapsed && isActive(item.href, item.exact) && <ChevronRight className="w-3.5 h-3.5 ml-auto opacity-60" />}
            </Link>
          ))}

          {dataMasterItems.length > 0 && (
            <div className="pt-2">
              <button
                onClick={() => setDataMasterOpen(!dataMasterOpen)}
                className={cn(
                  'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all w-full',
                  'text-white/50 hover:bg-white/5 hover:text-white'
                )}
              >
                <FolderOpen className="w-4 h-4 flex-shrink-0" />
                {!sidebarCollapsed && <span>Data Master</span>}
                {!sidebarCollapsed && <ChevronRight className={cn('w-3.5 h-3.5 ml-auto opacity-60 transition-transform', dataMasterOpen && 'rotate-90')} />}
              </button>
              {dataMasterOpen && !sidebarCollapsed && (
                <div className="ml-4 mt-1 space-y-0.5">
                  {dataMasterItems.map((item) => (
                    <Link key={item.href} to={item.href} onClick={() => setSidebarOpen(false)}
                      className={cn(
                        'flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-medium transition-all',
                        isActive(item.href)
                          ? 'bg-white/10 text-white'
                          : 'text-white/50 hover:bg-white/5 hover:text-white'
                      )}
                    >
                      <item.icon className="w-3.5 h-3.5 flex-shrink-0" />
                      {item.label}
                    </Link>
                  ))}
                </div>
              )}
            </div>
          )}

          <Link to="/admin/materials" onClick={() => setSidebarOpen(false)}
            className={cn(
              'flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all mt-2',
              isActive('/admin/materials')
                ? 'bg-white/10 text-white'
                : 'text-white/50 hover:bg-white/5 hover:text-white'
            )}
            title={sidebarCollapsed ? 'Materi Legal' : undefined}
          >
            <FileText className="w-4 h-4 flex-shrink-0" />
            {!sidebarCollapsed && 'Materi Legal'}
          </Link>
        </nav>

        <SidebarUserButton settingsPath="/admin/settings" dark={true} collapsed={sidebarCollapsed} />
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
          <Link to="/" className="text-sm text-gray-500 hover:text-gray-700 transition-colors">
            ← Kembali ke Beranda
          </Link>
        </header>
        <main className="flex-1"><Outlet /></main>
      </div>
    </div>
  )
}
