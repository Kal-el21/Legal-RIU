import { Link, useSearchParams } from 'react-router-dom'
import { Bell, CheckCheck, ChevronLeft, ChevronRight, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import NotificationRow from '@/components/common/NotificationRow'
import { useMarkAllNotificationsRead, useMarkNotificationRead, useNotifications } from '@/hooks/useNotifications'
import { useAuthStore } from '@/store/auth.store'
import { cn } from '@/lib/utils'
import { getReminderDetailPath, getReminderRowTone } from '@/lib/reminders'

const PAGE_LIMIT = 10

function parsePage(value: string | null) {
  const page = Number(value)
  return Number.isFinite(page) && page > 0 ? Math.floor(page) : 1
}

export default function NotificationListPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const user = useAuthStore((state) => state.user)
  const page = parsePage(searchParams.get('page'))
  const unreadOnly = searchParams.get('filter') === 'unread'

  const { data, isLoading, isError, refetch } = useNotifications({
    page,
    limit: PAGE_LIMIT,
    unread_only: unreadOnly || undefined,
  })
  const markRead = useMarkNotificationRead()
  const markAllRead = useMarkAllNotificationsRead()

  const items = data?.items ?? []
  const total = data?.total ?? 0
  const unreadTotal = data?.unread_total ?? 0
  const totalPages = data?.total_pages ?? 0
  const canGoPrev = page > 1
  const canGoNext = totalPages > 0 && page < totalPages
  // Data lama tetap ditampilkan kalau hanya pemuatan ulangnya yang gagal
  const failed = isError && !data

  const setPage = (nextPage: number) => {
    const next = new URLSearchParams(searchParams)
    if (nextPage <= 1) {
      next.delete('page')
    } else {
      next.set('page', String(nextPage))
    }
    setSearchParams(next)
  }

  const toggleUnreadOnly = () => {
    // Kembali ke halaman pertama: jumlah halaman berubah saat saringan berganti
    setSearchParams(unreadOnly ? {} : { filter: 'unread' })
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6 p-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <Bell className="h-5 w-5 text-red-700" />
            <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Notifikasi</h1>
          </div>
          <p className="mt-1 text-sm text-gray-500">Kabar terbaru tentang pengajuan: perubahan status dan yang menunggu ditindak.</p>
        </div>
        <div className="flex flex-col items-start gap-2 sm:items-end">
          <div className="text-sm text-gray-500">
            {unreadOnly ? `${total} belum dibaca` : `${unreadTotal} belum dibaca dari ${total}`}
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              variant={unreadOnly ? 'default' : 'outline'}
              aria-pressed={unreadOnly}
              onClick={toggleUnreadOnly}
            >
              Belum dibaca
            </Button>
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
      </div>

      <div className="overflow-hidden rounded-lg border border-gray-100 bg-white">
        {isLoading ? (
          <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-gray-500">
            <Loader2 className="h-4 w-4 animate-spin" />
            Memuat notifikasi
          </div>
        ) : null}

        {failed ? (
          <div className="flex flex-col items-center gap-3 px-6 py-16 text-center">
            <div>
              <p className="text-sm font-medium text-gray-700">Gagal memuat notifikasi</p>
              <p className="mt-1 text-sm text-gray-400">Periksa sambungan Anda, lalu coba lagi.</p>
            </div>
            <Button type="button" variant="outline" onClick={() => refetch()}>Coba lagi</Button>
          </div>
        ) : null}

        {!isLoading && !failed && items.length === 0 ? (
          <div className="px-6 py-16 text-center">
            <p className="text-sm font-medium text-gray-700">
              {unreadOnly ? 'Semua notifikasi sudah dibaca' : 'Tidak ada notifikasi'}
            </p>
            <p className="mt-1 text-sm text-gray-400">
              {unreadOnly ? 'Matikan saringan untuk melihat yang sudah dibaca.' : 'Belum ada kabar baru tentang pengajuan.'}
            </p>
          </div>
        ) : null}

        {items.length > 0 ? (
          <div className="divide-y divide-gray-100">
            {items.map((item) => (
              <Link
                key={`${item.submission_type}-${item.id}`}
                to={getReminderDetailPath(item, user?.role)}
                onClick={() => {
                  if (!item.is_read) markRead.mutate({ submission_type: item.submission_type, submission_id: item.id })
                }}
                className={cn('flex items-center gap-4 px-5 py-4 transition-colors', getReminderRowTone(item))}
              >
                <NotificationRow item={item} />
                <ChevronRight className="h-4 w-4 shrink-0 text-gray-300" />
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
