import { cn } from '@/lib/utils'

interface WarningBadgeProps {
  level: 'YELLOW' | 'RED'
  className?: string
}

export default function WarningBadge({ level, className }: WarningBadgeProps) {
  const base = 'inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium'
  const variant = level === 'YELLOW' ? 'bg-amber-100 text-amber-700' : 'bg-red-100 text-red-700'

  return (
    <span className={cn(base, variant, className)}>
      {level === 'YELLOW' ? 'Perlu Perhatian' : 'Terlambat'}
    </span>
  )
}
