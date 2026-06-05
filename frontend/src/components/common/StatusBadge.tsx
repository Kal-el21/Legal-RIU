import { cn, getStatusLabel, getStatusColor } from '@/lib/utils'
import type { SubmissionStatus } from '@/types'

interface StatusBadgeProps {
  status: SubmissionStatus
  className?: string
}

export default function StatusBadge({ status, className }: StatusBadgeProps) {
  return (
    <span className={cn('inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium', getStatusColor(status), className)}>
      {getStatusLabel(status)}
    </span>
  )
}