import { Link } from 'react-router-dom'
import { DropdownMenu as MenuPrimitive } from 'radix-ui'
import { Bell, CheckCheck, ChevronRight, Loader2 } from 'lucide-react'
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import NotificationRow from '@/components/common/NotificationRow'
import { useMarkAllNotificationsRead, useMarkNotificationRead, useNotifications } from '@/hooks/useNotifications'
import { useAuthStore } from '@/store/auth.store'
import { cn } from '@/lib/utils'
import { getNotificationBasePath, getReminderDetailPath, getReminderRowTone } from '@/lib/reminders'

// Item menu memakai primitif Radix langsung: DropdownMenuItem bawaan shadcn
// memaksa warna semua teks di dalamnya saat difokus, yang merusak badge.
const actionClass =
  'flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-navy outline-none select-none hover:bg-gray-100 focus:bg-gray-100 data-disabled:pointer-events-none data-disabled:opacity-50'

export default function NotificationDropdown() {
  const user = useAuthStore((state) => state.user)
  const { data, isLoading, isError, refetch } = useNotifications({ page: 1, limit: 5 })
  const markRead = useMarkNotificationRead()
  const markAllRead = useMarkAllNotificationsRead()

  const items = data?.items ?? []
  const unreadTotal = data?.unread_total ?? 0
  const countLabel = unreadTotal > 99 ? '99+' : String(unreadTotal)
  // Data lama tetap ditampilkan kalau hanya pemuatan ulangnya yang gagal
  const failed = isError && !data

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="relative inline-flex h-9 w-9 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-800"
          aria-label={unreadTotal > 0 ? `Notifikasi, ${unreadTotal} belum dibaca` : 'Notifikasi'}
        >
          <Bell className="h-5 w-5" />
          {unreadTotal > 0 ? (
            <span className="absolute -right-1 -top-1 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-red-600 px-1 text-[10px] font-bold text-white ring-2 ring-white">
              {countLabel}
            </span>
          ) : null}
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" sideOffset={8} className="w-80 max-w-[calc(100vw-2rem)] overflow-hidden p-0">
        <div className="flex items-center justify-between gap-2 border-b border-gray-100 px-3 py-2.5">
          <div>
            <p className="text-sm font-semibold text-gray-900">Notifikasi</p>
            <p className="text-xs text-gray-500">
              {unreadTotal > 0 ? `${unreadTotal} belum dibaca` : 'Semua sudah dibaca'}
            </p>
          </div>
          <MenuPrimitive.Item
            className={actionClass}
            disabled={unreadTotal === 0 || markAllRead.isPending}
            // preventDefault: menu tetap terbuka supaya hasilnya terlihat
            onSelect={(event) => { event.preventDefault(); markAllRead.mutate() }}
          >
            <CheckCheck className="h-3.5 w-3.5" />
            Tandai semua terbaca
          </MenuPrimitive.Item>
        </div>

        <div className="max-h-80 overflow-y-auto">
          {isLoading ? (
            <div className="flex items-center justify-center gap-2 px-3 py-7 text-sm text-gray-500">
              <Loader2 className="h-4 w-4 animate-spin" />
              Memuat notifikasi
            </div>
          ) : null}

          {failed ? (
            <div className="flex flex-col items-center gap-2 px-3 py-7 text-center">
              <p className="text-sm font-medium text-gray-700">Gagal memuat notifikasi</p>
              <MenuPrimitive.Item
                className={actionClass}
                onSelect={(event) => { event.preventDefault(); refetch() }}
              >
                Coba lagi
              </MenuPrimitive.Item>
            </div>
          ) : null}

          {!isLoading && !failed && items.length === 0 ? (
            <div className="px-3 py-7 text-center">
              <p className="text-sm font-medium text-gray-700">Tidak ada notifikasi</p>
              <p className="mt-1 text-xs text-gray-400">Belum ada kabar baru tentang pengajuan.</p>
            </div>
          ) : null}

          {items.map((item) => (
            <MenuPrimitive.Item
              key={`${item.submission_type}-${item.id}`}
              asChild
              onSelect={() => {
                if (!item.is_read) markRead.mutate({ submission_type: item.submission_type, submission_id: item.id })
              }}
            >
              <Link
                to={getReminderDetailPath(item, user?.role)}
                className={cn(
                  'block border-b border-gray-50 px-3 py-2.5 outline-none transition-colors last:border-b-0',
                  getReminderRowTone(item),
                )}
              >
                <NotificationRow item={item} compact />
              </Link>
            </MenuPrimitive.Item>
          ))}
        </div>

        <MenuPrimitive.Item asChild>
          <Link
            to={`${getNotificationBasePath(user?.role)}/notifications`}
            className="flex items-center justify-center gap-1 border-t border-gray-100 px-3 py-2.5 text-sm font-medium text-red-700 outline-none transition-colors hover:bg-red-50 focus:bg-red-50"
          >
            Lihat Semua Notifikasi
            <ChevronRight className="h-4 w-4" />
          </Link>
        </MenuPrimitive.Item>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
