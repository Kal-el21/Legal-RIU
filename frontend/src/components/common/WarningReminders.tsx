import { Link } from 'react-router-dom'
import { AlertTriangle } from 'lucide-react'
import WarningBadge from '@/components/common/WarningBadge'
import { useNotifications } from '@/hooks/useNotifications'
import { useAuthStore } from '@/store/auth.store'
import { cn } from '@/lib/utils'
import { getReminderDetailPath, getReminderMessage } from '@/lib/reminders'
import type { ReminderItem } from '@/types'

const GROUPS = [
  { level: 'RED', title: 'Terlambat', border: 'border-red-100', divider: 'border-red-50 divide-red-50', icon: 'text-red-500', heading: 'text-red-700', count: 'text-red-500' },
  { level: 'YELLOW', title: 'Perlu Perhatian', border: 'border-amber-100', divider: 'border-amber-50 divide-amber-50', icon: 'text-amber-500', heading: 'text-amber-700', count: 'text-amber-600' },
] as const

// Pengajuan yang sudah melewati ambang peringatan, untuk dashboard pemroses.
// Tidak tampil sama sekali kalau tidak ada yang terlambat.
export default function WarningReminders() {
  const role = useAuthStore((state) => state.user?.role)
  const { data } = useNotifications({ warning_only: true, limit: 100 })
  const items = data?.items ?? []

  if (!items.length) return null

  return (
    <div className="space-y-4">
      {GROUPS.map((group) => {
        // Yang paling lama menunggu di atas
        const rows = items
          .filter((item: ReminderItem) => item.warning_level === group.level)
          .sort((a, b) => b.days_waiting - a.days_waiting)
        if (!rows.length) return null

        return (
          <div key={group.level} className={cn('bg-white rounded-2xl border overflow-hidden', group.border)}>
            <div className={cn('px-6 py-4 border-b flex items-center gap-2', group.divider)}>
              <AlertTriangle className={cn('w-4 h-4', group.icon)} />
              <h2 className={cn('text-sm font-semibold', group.heading)}>{group.title}</h2>
              <span className={cn('text-xs', group.count)}>{rows.length}</span>
            </div>
            <div className={cn('divide-y', group.divider)}>
              {rows.map((item) => (
                <Link
                  key={`${item.submission_type}-${item.id}`}
                  to={getReminderDetailPath(item, role)}
                  className="flex items-center gap-3 px-6 py-3.5 hover:bg-gray-50/50 transition-colors"
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">{item.ticket_number} — {item.title}</p>
                    <p className="text-xs text-gray-400 mt-0.5">
                      {getReminderMessage(item)} · menunggu {item.days_waiting} hari
                    </p>
                  </div>
                  <WarningBadge level={group.level} />
                </Link>
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}
