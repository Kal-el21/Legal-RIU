import { useState, useRef, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Settings, LogOut, ChevronUp } from 'lucide-react'
import { useAuthStore } from '@/store/auth.store'
import { useLogout } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'

interface SidebarUserButtonProps {
  settingsPath: string
  collapsed?: boolean
}

export default function SidebarUserButton({ settingsPath, collapsed = false }: SidebarUserButtonProps) {
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
    <div ref={ref} className="relative p-3 border-t border-white/10">
      <button
        onClick={() => setOpen(!open)}
        className={cn(
          'w-full flex items-center gap-3 px-3 py-2.5 rounded-xl transition-colors hover:bg-white/5',
          collapsed && 'justify-center px-2'
        )}
      >
        <div className="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold flex-shrink-0"
          style={{ background: '#C8102E' }}>
          {user?.full_name?.charAt(0).toUpperCase()}
        </div>
        {!collapsed && (
          <div className="min-w-0 flex-1 text-left">
            <p className="text-sm font-medium truncate text-white">{user?.full_name}</p>
            <p className="text-xs truncate text-white/60">{user?.division}</p>
          </div>
        )}
        <ChevronUp className={cn(
          'w-4 h-4 flex-shrink-0 transition-transform duration-200 text-white/40',
          !open && 'rotate-180',
          collapsed ? 'hidden' : ''
        )} />
      </button>

      {/* Popover */}
      {open && (
        <div className={cn(
          'absolute bottom-full mb-2 rounded-xl border shadow-lg overflow-hidden z-50 bg-[#0f2d4d] border-white/10',
          collapsed ? 'left-16 -translate-x-0 right-auto w-48' : 'left-3 right-3'
        )}>
          {/* User info */}
          <div className="px-4 py-3 border-b border-white/10">
            <p className="text-sm font-medium text-white">{user?.full_name}</p>
            {!collapsed && <p className="text-xs mt-0.5 text-white/50">{user?.email}</p>}
          </div>

          {/* Actions */}
          <div className="py-1">
            <Link
              to={settingsPath}
              onClick={() => setOpen(false)}
              className="flex items-center gap-2.5 px-4 py-2.5 text-sm transition-colors text-white/70 hover:bg-white/5 hover:text-white"
            >
              <Settings className="w-4 h-4" />
              Pengaturan
            </Link>
            <button
              onClick={() => { setOpen(false); logout() }}
              className="w-full flex items-center gap-2.5 px-4 py-2.5 text-sm text-red-400 hover:bg-red-500/10 transition-colors"
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
