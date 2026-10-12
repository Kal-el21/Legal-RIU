import { FileSearch, FileText } from 'lucide-react'
import WarningBadge from '@/components/common/WarningBadge'
import { cn, formatDate } from '@/lib/utils'
import { getReminderMessage, getReminderTypeLabel, reminderNeedsAction } from '@/lib/reminders'
import type { ReminderItem } from '@/types'

// Isi satu baris notifikasi, dipakai lonceng (compact) dan halaman daftar.
// Tautan pembungkusnya disediakan pemakai.
export default function NotificationRow({ item, compact }: { item: ReminderItem; compact?: boolean }) {
  const Icon = item.submission_type === 'document_review' ? FileSearch : FileText

  return (
    <div className={cn('flex w-full items-start text-left', compact ? 'gap-2.5' : 'gap-4')}>
      {compact ? null : (
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-600">
          <Icon className="h-5 w-5" />
        </div>
      )}
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-x-1.5 gap-y-1 text-xs text-gray-500">
          {item.is_read ? null : (
            <>
              <span className="h-2 w-2 shrink-0 rounded-full bg-[#0B2545]" />
              <span className="sr-only">Belum dibaca.</span>
            </>
          )}
          <span className="font-medium">{item.ticket_number}</span>
          <span className="text-gray-300">-</span>
          <span>{getReminderTypeLabel(item)}</span>
        </div>
        <p className={cn('mt-0.5 text-sm text-gray-900', compact && 'line-clamp-2', item.is_read ? 'font-medium' : 'font-semibold')}>
          {item.title}
        </p>
        <p className="mt-0.5 text-xs text-gray-600">{getReminderMessage(item)}</p>
        {item.note ? <p className="mt-0.5 line-clamp-2 text-xs text-gray-500">Catatan: {item.note}</p> : null}
        <div className="mt-1.5 flex flex-wrap items-center gap-1.5 text-xs text-gray-400">
          {item.warning_level !== 'NONE' ? (
            <>
              <WarningBadge level={item.warning_level} />
              <span>Menunggu {item.days_waiting} hari</span>
              <span className="text-gray-300">-</span>
            </>
          ) : reminderNeedsAction(item) ? (
            <>
              <span className="inline-flex items-center rounded-full bg-[#0B2545]/10 px-2.5 py-1 text-xs font-medium text-[#0B2545]">
                Perlu tindakan
              </span>
              <span className="text-gray-300">-</span>
            </>
          ) : null}
          <span>{formatDate(item.event_at)}</span>
        </div>
      </div>
    </div>
  )
}
