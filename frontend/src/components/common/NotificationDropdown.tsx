import { useEffect, useMemo, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Bell, ChevronRight, Loader2 } from 'lucide-react'
import WarningBadge from '@/components/common/WarningBadge'
import { dashboardService } from '@/services/dashboard.service'
import { useAuthStore } from '@/store/auth.store'
import { cn, formatDate } from '@/lib/utils'
import { getNotificationBasePath, getReminderDetailPath, getReminderTypeLabel } from '@/lib/reminders'
import type { ReminderItem } from '@/types'

function normalizeItems(dataItems?: ReminderItem[], red?: ReminderItem[], yellow?: ReminderItem[]) {
  if (dataItems?.length) return dataItems
  return [...(red ?? []), ...(yellow ?? [])].slice(0, 5)
}

export default function NotificationDropdown() {
  const [open, setOpen] = useState(false)
  const rootRef = useRef<HTMLDivElement>(null)
  const user = useAuthStore((state) => state.user)
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['notifications', 'dropdown'],
    queryFn: () => dashboardService.getReminders({ page: 1, limit: 5 }),
    staleTime: 30_000,
    enabled: Boolean(user),
  })

  const markRead = useMutation({
    mutationFn: dashboardService.markReminderRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['dashboard'] })
    },
  })

  const basePath = getNotificationBasePath(user?.role)
  const items = useMemo(() => normalizeItems(data?.items, data?.red, data?.yellow), [data])
  const total = data?.total ?? ((data?.red?.length ?? 0) + (data?.yellow?.length ?? 0))
  const unreadTotal = data?.unread_total ?? items.filter((item) => !item.is_read).length
  const countLabel = unreadTotal > 99 ? '99+' : String(unreadTotal)

  const handleItemClick = (item: ReminderItem) => {
    if (item.is_read) return
    markRead.mutate({
      submission_type: item.submission_type,
      submission_id: item.id,
    })
  }

  useEffect(() => {
    if (!open) return

    const handlePointerDown = (event: MouseEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false)
      }
    }

    document.addEventListener('mousedown', handlePointerDown)
    return () => document.removeEventListener('mousedown', handlePointerDown)
  }, [open])

  return (
    <div ref={rootRef} className="relative">
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        className="relative inline-flex h-9 w-9 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-800"
        aria-label="Buka notifikasi"
        aria-expanded={open}
      >
        <Bell className="h-5 w-5" />
        {unreadTotal > 0 ? (
          <span className="absolute -right-1 -top-1 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-red-600 px-1 text-[10px] font-bold text-white ring-2 ring-white">
            {countLabel}
          </span>
        ) : null}
      </button>

      {open ? (
        <div className="absolute right-0 top-full z-50 mt-2 w-80 max-w-[calc(100vw-2rem)] overflow-hidden rounded-lg border border-gray-100 bg-white shadow-xl">
          <div className="flex items-center justify-between border-b border-gray-100 px-3 py-2.5">
            <div>
              <p className="text-sm font-semibold text-gray-900">Notifikasi</p>
              <p className="text-xs text-gray-500">{unreadTotal} belum dibaca dari {total}</p>
            </div>
            {isLoading ? <Loader2 className="h-4 w-4 animate-spin text-gray-400" /> : null}
          </div>

          <div className="max-h-72 overflow-y-auto">
            {!isLoading && items.length === 0 ? (
              <div className="px-3 py-7 text-center">
                <p className="text-sm font-medium text-gray-700">Tidak ada notifikasi</p>
                <p className="mt-1 text-xs text-gray-400">Semua pengajuan masih dalam batas waktu.</p>
              </div>
            ) : null}

            {items.map((item) => (
              <Link
                key={`${item.submission_type}-${item.id}`}
                to={getReminderDetailPath(item, user?.role)}
                onClick={() => { handleItemClick(item); setOpen(false) }}
                className={cn(
                  'block border-b border-gray-50 px-3 py-2.5 transition-colors last:border-b-0 hover:bg-gray-50',
                  !item.is_read && 'bg-red-50/30',
                )}
              >
                <div className="flex items-start gap-2.5">
                  <div className={cn(
                    'mt-1 h-2 w-2 shrink-0 rounded-full',
                    item.is_read ? 'bg-gray-300' : item.warning_level === 'RED' ? 'bg-red-500' : 'bg-amber-400',
                  )} />
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-1.5">
                      <p className="truncate text-xs font-medium text-gray-500">{item.ticket_number}</p>
                      <span className="text-xs text-gray-300">-</span>
                      <p className="truncate text-xs text-gray-500">{getReminderTypeLabel(item)}</p>
                    </div>
                    <p className={cn(
                      'mt-0.5 line-clamp-2 text-sm text-gray-900',
                      item.is_read ? 'font-medium' : 'font-semibold',
                    )}>{item.title}</p>
                    <div className="mt-1.5 flex flex-wrap items-center gap-1.5">
                      {item.warning_level !== 'NONE' ? <WarningBadge level={item.warning_level} /> : null}
                      <span className="text-xs text-gray-400">{item.days_since_submission} hari sejak pengajuan</span>
                    </div>
                    <p className="mt-0.5 text-xs text-gray-400">{formatDate(item.submitted_at)}</p>
                  </div>
                </div>
              </Link>
            ))}
          </div>

          <Link
            to={`${basePath}/notifications`}
            onClick={() => setOpen(false)}
            className="flex items-center justify-center gap-1 border-t border-gray-100 px-3 py-2.5 text-sm font-medium text-red-700 transition-colors hover:bg-red-50"
          >
            Lihat Semua Notifikasi
            <ChevronRight className="h-4 w-4" />
          </Link>
        </div>
      ) : null}
    </div>
  )
}
