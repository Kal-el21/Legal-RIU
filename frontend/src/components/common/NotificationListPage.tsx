import { Link, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Bell, CheckCheck, ChevronLeft, ChevronRight, FileSearch, FileText, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import StatusBadge from '@/components/common/StatusBadge'
import WarningBadge from '@/components/common/WarningBadge'
import { dashboardService } from '@/services/dashboard.service'
import { useAuthStore } from '@/store/auth.store'
import { cn, formatDate } from '@/lib/utils'
import { getReminderDetailPath, getReminderTypeLabel } from '@/lib/reminders'
import type { ReminderItem, SubmissionStatus } from '@/types'

const PAGE_LIMIT = 10

function parsePage(value: string | null) {
  const page = Number(value)
  return Number.isFinite(page) && page > 0 ? Math.floor(page) : 1
}

function NotificationIcon({ item }: { item: ReminderItem }) {
  const Icon = item.submission_type === 'document_review' ? FileSearch : FileText
  const color = item.warning_level === 'RED' ? 'text-red-600 bg-red-50' : 'text-amber-600 bg-amber-50'

  return (
    <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg ${color}`}>
      <Icon className="h-5 w-5" />
    </div>
  )
}

export default function NotificationListPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const user = useAuthStore((state) => state.user)
  const queryClient = useQueryClient()
  const page = parsePage(searchParams.get('page'))

  const { data, isLoading } = useQuery({
    queryKey: ['notifications', 'list', page],
    queryFn: () => dashboardService.getReminders({ page, limit: PAGE_LIMIT }),
    enabled: Boolean(user),
  })

  const markRead = useMutation({
    mutationFn: dashboardService.markReminderRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['dashboard'] })
    },
  })

  const markAllRead = useMutation({
    mutationFn: dashboardService.markAllRemindersRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] })
      queryClient.invalidateQueries({ queryKey: ['dashboard'] })
    },
  })

  const items = data?.items ?? []
  const total = data?.total ?? 0
  const unreadTotal = data?.unread_total ?? items.filter((item) => !item.is_read).length
  const totalPages = data?.total_pages ?? 0
  const canGoPrev = page > 1
  const canGoNext = totalPages > 0 && page < totalPages

  const setPage = (nextPage: number) => {
    const next = new URLSearchParams(searchParams)
    if (nextPage <= 1) {
      next.delete('page')
    } else {
      next.set('page', String(nextPage))
    }
    setSearchParams(next)
  }

  const handleItemClick = (item: ReminderItem) => {
    if (item.is_read) return
    markRead.mutate({
      submission_type: item.submission_type,
      submission_id: item.id,
    })
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6 p-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <Bell className="h-5 w-5 text-red-700" />
            <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Notifikasi</h1>
          </div>
          <p className="mt-1 text-sm text-gray-500">Daftar pengajuan yang perlu perhatian berdasarkan batas waktu.</p>
        </div>
        <div className="flex flex-col items-start gap-2 sm:items-end">
          <div className="text-sm text-gray-500">
            {unreadTotal} belum dibaca dari {total}
          </div>
          <Button
            type="button"
            variant="outline"
            onClick={() => markAllRead.mutate()}
            disabled={unreadTotal === 0 || markAllRead.isPending}
          >
            {markAllRead.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <CheckCheck className="h-4 w-4" />}
            Tandai semua terbaca
          </Button>
        </div>
      </div>

      <div className="overflow-hidden rounded-lg border border-gray-100 bg-white">
        {isLoading ? (
          <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-gray-500">
            <Loader2 className="h-4 w-4 animate-spin" />
            Memuat notifikasi
          </div>
        ) : null}

        {!isLoading && items.length === 0 ? (
          <div className="px-6 py-16 text-center">
            <p className="text-sm font-medium text-gray-700">Tidak ada notifikasi</p>
            <p className="mt-1 text-sm text-gray-400">Belum ada pengajuan yang melewati batas peringatan.</p>
          </div>
        ) : null}

        {!isLoading && items.length > 0 ? (
          <div className="divide-y divide-gray-100">
            {items.map((item) => (
              <Link
                key={`${item.submission_type}-${item.id}`}
                to={getReminderDetailPath(item, user?.role)}
                onClick={() => handleItemClick(item)}
                className={cn(
                  'flex gap-4 px-5 py-4 transition-colors hover:bg-gray-50',
                  !item.is_read && 'bg-red-50/30',
                )}
              >
                <NotificationIcon item={item} />
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    {!item.is_read ? <span className="h-2 w-2 rounded-full bg-red-500" /> : null}
                    <span className="text-xs font-medium text-gray-500">{getReminderTypeLabel(item)}</span>
                    <span className="text-xs text-gray-300">-</span>
                    <span className="text-xs font-medium text-gray-500">{item.ticket_number}</span>
                    {item.warning_level !== 'NONE' ? <WarningBadge level={item.warning_level} /> : null}
                  </div>
                  <p className={cn(
                    'mt-1 text-sm text-gray-900',
                    item.is_read ? 'font-medium' : 'font-semibold',
                  )}>{item.title}</p>
                  <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-gray-500">
                    <StatusBadge status={item.status as SubmissionStatus} />
                    <span>{item.days_since_submission} hari sejak pengajuan</span>
                    <span className="text-gray-300">-</span>
                    <span>{item.days_since_last_update} hari sejak update terakhir</span>
                  </div>
                  <p className="mt-2 text-xs text-gray-400">Tanggal pengajuan: {formatDate(item.submitted_at)}</p>
                </div>
                <ChevronRight className="mt-2 h-4 w-4 shrink-0 text-gray-300" />
              </Link>
            ))}
          </div>
        ) : null}
      </div>

      {totalPages > 1 ? (
        <div className="flex items-center justify-between">
          <Button
            type="button"
            variant="outline"
            onClick={() => setPage(page - 1)}
            disabled={!canGoPrev}
          >
            <ChevronLeft className="h-4 w-4" />
            Sebelumnya
          </Button>
          <p className="text-sm text-gray-500">
            Halaman {page} dari {totalPages}
          </p>
          <Button
            type="button"
            variant="outline"
            onClick={() => setPage(page + 1)}
            disabled={!canGoNext}
          >
            Berikutnya
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      ) : null}
    </div>
  )
}
