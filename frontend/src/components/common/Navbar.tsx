import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { ChevronDown, Menu, X, Scale, LayoutDashboard, LogOut } from 'lucide-react'
import { useAuthStore } from '@/store/auth.store'
import { useLogout } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'

const NAV_LINKS = [
  { label: 'Home', href: '/' },
  {
    label: 'Dokumen Perusahaan',
    children: [
      { label: 'Akta Perusahaan', href: '/akta-perusahaan' },
      { label: 'Asset Perusahaan', href: '/asset-perusahaan', soon: true },
      { label: 'SK SOP Legal', href: '/sk-sop-legal', soon: true },
    ],
  },
  { label: 'Materi Legal', href: '/materi-legal', soon: true },
  { label: 'Profil Legal', href: '/profil-legal', soon: true },
]

export default function Navbar() {
  const location = useLocation()
  const { isAuthenticated, user } = useAuthStore()
  const logout = useLogout()
  const [mobileOpen, setMobileOpen] = useState(false)
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const [profileOpen, setProfileOpen] = useState(false)

  return (
    <nav className="sticky top-0 z-50 bg-white border-b border-gray-100 shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">

          {/* Logo */}
          <Link to="/" className="flex items-center gap-2.5 flex-shrink-0">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: '#C8102E' }}>
              <Scale className="w-4 h-4 text-white" />
            </div>
            <div>
              <span className="font-bold text-sm" style={{ color: '#0B2545' }}>Legal RIU</span>
              <p className="text-xs text-gray-400 leading-none">Indonesia Re</p>
            </div>
          </Link>

          {/* Desktop nav */}
          {isAuthenticated && (
            <div className="hidden lg:flex items-center gap-1">
              {NAV_LINKS.map((link) =>
                link.children ? (
                  <div key={link.label} className="relative">
                    <button
                      onMouseEnter={() => setDropdownOpen(true)}
                      onMouseLeave={() => setDropdownOpen(false)}
                      className="flex items-center gap-1 px-3 py-2 rounded-lg text-sm font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-50 transition-colors"
                    >
                      {link.label}
                      <ChevronDown className={cn('w-3.5 h-3.5 transition-transform', dropdownOpen && 'rotate-180')} />
                    </button>
                    {dropdownOpen && (
                      <div
                        onMouseEnter={() => setDropdownOpen(true)}
                        onMouseLeave={() => setDropdownOpen(false)}
                        className="absolute top-full left-0 mt-1 w-52 bg-white rounded-xl shadow-lg border border-gray-100 py-1.5 z-50"
                      >
                        {link.children.map((child) => (
                          <Link key={child.href} to={child.href}
                            className="flex items-center justify-between px-4 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-50">
                            {child.label}
                            {child.soon && <span className="text-xs px-1.5 py-0.5 rounded-full bg-orange-100 text-orange-600 font-medium">Soon</span>}
                          </Link>
                        ))}
                      </div>
                    )}
                  </div>
                ) : (
                  <Link key={link.href} to={link.href!}
                    className={cn('flex items-center gap-1 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
                      location.pathname === link.href ? 'text-gray-900 bg-gray-100' : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                    )}>
                    {link.label}
                    {link.soon && <span className="text-xs px-1.5 py-0.5 rounded-full bg-orange-100 text-orange-600 font-medium">Soon</span>}
                  </Link>
                )
              )}
            </div>
          )}

          {/* Right side */}
          <div className="hidden lg:flex items-center gap-2">
            {!isAuthenticated ? (
              <Link to="/login" className="px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors" style={{ background: '#C8102E' }}>
                Login
              </Link>
            ) : (
              <div className="flex items-center gap-2">
                <Link to={user?.role === 'ADMIN' ? '/admin' : '/dashboard'}
                  className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium text-gray-600 hover:bg-gray-50 transition-colors">
                  <LayoutDashboard className="w-4 h-4" /> Dashboard
                </Link>
                <div className="relative">
                  <button onClick={() => setProfileOpen(!profileOpen)}
                    className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-50 transition-colors">
                    <div className="w-7 h-7 rounded-full flex items-center justify-center text-white text-xs font-bold" style={{ background: '#0B2545' }}>
                      {user?.full_name?.charAt(0).toUpperCase()}
                    </div>
                    <span className="text-sm font-medium text-gray-700 max-w-[96px] truncate">{user?.full_name}</span>
                    <ChevronDown className={cn('w-3.5 h-3.5 text-gray-400 transition-transform', profileOpen && 'rotate-180')} />
                  </button>
                  {profileOpen && (
                    <div className="absolute right-0 top-full mt-1 w-48 bg-white rounded-xl shadow-lg border border-gray-100 py-1.5 z-50">
                      <div className="px-4 py-2 border-b border-gray-100 mb-1">
                        <p className="text-sm font-medium text-gray-900 truncate">{user?.full_name}</p>
                        <p className="text-xs text-gray-500 truncate">{user?.email}</p>
                      </div>
                      <button onClick={logout}
                        className="w-full flex items-center gap-2.5 px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors">
                        <LogOut className="w-4 h-4" /> Keluar
                      </button>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>

          {/* Mobile toggle */}
          {isAuthenticated ? (
            <button onClick={() => setMobileOpen(!mobileOpen)} className="lg:hidden p-2 rounded-lg text-gray-600 hover:bg-gray-50">
              {mobileOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          ) : (
            <Link to="/login" className="lg:hidden px-4 py-2 rounded-lg text-sm font-medium text-white transition-colors" style={{ background: '#C8102E' }}>
              Login
            </Link>
          )}
        </div>
      </div>

      {/* Mobile menu */}
      {mobileOpen && isAuthenticated && (
        <div className="lg:hidden border-t border-gray-100 bg-white px-4 py-3 space-y-1">
          {NAV_LINKS.map((link) =>
            link.children ? (
              <div key={link.label}>
                <p className="px-3 py-2 text-xs font-semibold text-gray-400 uppercase tracking-wide">{link.label}</p>
                {link.children.map((child) => (
                  <Link key={child.href} to={child.href} onClick={() => setMobileOpen(false)}
                    className="flex items-center justify-between px-3 py-2 rounded-lg text-sm text-gray-600 hover:bg-gray-50">
                    {child.label}
                    {child.soon && <span className="text-xs px-1.5 py-0.5 rounded-full bg-orange-100 text-orange-600">Soon</span>}
                  </Link>
                ))}
              </div>
            ) : (
              <Link key={link.href} to={link.href!} onClick={() => setMobileOpen(false)}
                className="flex items-center justify-between px-3 py-2 rounded-lg text-sm text-gray-600 hover:bg-gray-50">
                {link.label}
                {link.soon && <span className="text-xs px-1.5 py-0.5 rounded-full bg-orange-100 text-orange-600">Soon</span>}
              </Link>
            )
          )}
          <div className="pt-2 border-t border-gray-100">
            {!isAuthenticated ? (
              <Link to="/login" onClick={() => setMobileOpen(false)}
                className="block px-3 py-2 rounded-lg text-sm font-medium text-white text-center" style={{ background: '#C8102E' }}>
                Login
              </Link>
            ) : (
              <div className="space-y-1">
                <Link to={user?.role === 'ADMIN' ? '/admin' : '/dashboard'} onClick={() => setMobileOpen(false)}
                  className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-gray-600 hover:bg-gray-50">
                  <LayoutDashboard className="w-4 h-4" /> Dashboard
                </Link>
                <button onClick={logout} className="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-red-600 hover:bg-red-50">
                  <LogOut className="w-4 h-4" /> Keluar
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </nav>
  )
}
