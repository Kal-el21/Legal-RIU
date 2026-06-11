import { useState, useRef, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Settings, LogOut, ChevronUp } from 'lucide-react'
import { useAuthStore } from '@/store/auth.store'
import { useLogout } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'

interface SidebarUserButtonProps {
  settingsPath: string
  dark?: boolean // true = AdminLayout (dark sidebar)
}

export default function SidebarUserButton({ settingsPath, dark = false }: SidebarUserButtonProps) {
  const { user } = useAuthStore()
  const logout = useLogout()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  // Close on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  return (
    <div ref={ref} className="relative p-3 border-t" style={{ borderColor: dark ? 'rgba(255,255,255,0.1)' : '#F1F5F9' }}>
      <button
        onClick={() => setOpen(!open)}
        className={cn(
          'w-full flex items-center gap-3 px-3 py-2.5 rounded-xl transition-colors',
          dark ? 'hover:bg-white/5' : 'hover:bg-gray-50'
        )}
      >
        <div className="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold flex-shrink-0"
          style={{ background: dark ? '#C8102E' : '#0B2545' }}>
          {user?.full_name?.charAt(0).toUpperCase()}
        </div>
        <div className="min-w-0 flex-1 text-left">
          <p className={cn('text-sm font-medium truncate', dark ? 'text-white' : 'text-gray-900')}>{user?.full_name}</p>
          <p className={cn('text-xs truncate', dark ? 'text-white/40' : 'text-gray-500')}>{user?.division}</p>
        </div>
        <ChevronUp className={cn(
          'w-4 h-4 flex-shrink-0 transition-transform duration-200',
          dark ? 'text-white/40' : 'text-gray-400',
          !open && 'rotate-180'
        )} />
      </button>

      {/* Popover */}
      {open && (
        <div className={cn(
          'absolute bottom-full left-3 right-3 mb-2 rounded-xl border shadow-lg overflow-hidden z-50',
          dark ? 'bg-[#0f2d4d] border-white/10' : 'bg-white border-gray-100'
        )}>
          {/* User info */}
          <div className={cn('px-4 py-3 border-b', dark ? 'border-white/10' : 'border-gray-100')}>
            <p className={cn('text-sm font-medium', dark ? 'text-white' : 'text-gray-900')}>{user?.full_name}</p>
            <p className={cn('text-xs mt-0.5', dark ? 'text-white/50' : 'text-gray-500')}>{user?.email}</p>
          </div>

          {/* Actions */}
          <div className="py-1">
            <Link
              to={settingsPath}
              onClick={() => setOpen(false)}
              className={cn(
                'flex items-center gap-2.5 px-4 py-2.5 text-sm transition-colors',
                dark ? 'text-white/70 hover:bg-white/5 hover:text-white' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
              )}
            >
              <Settings className="w-4 h-4" />
              Pengaturan
            </Link>
            <button
              onClick={() => { setOpen(false); logout() }}
              className="w-full flex items-center gap-2.5 px-4 py-2.5 text-sm text-red-500 hover:bg-red-50 transition-colors"
            >
              <LogOut className="w-4 h-4" />
              Keluar
            </button>
          </div>
        </div>
      )}
    </div>
  )
}