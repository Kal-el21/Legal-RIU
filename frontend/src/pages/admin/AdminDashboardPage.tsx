import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Users, FileText, FileSearch, Clock, AlertCircle, RefreshCw, ArrowRight, AlertTriangle } from 'lucide-react'
import { dashboardService } from '@/services/dashboard.service'
import StatusBadge from '@/components/common/StatusBadge'
import WarningBadge from '@/components/common/WarningBadge'
import { formatDate } from '@/lib/utils'
import type { SubmissionStatus } from '@/types'

function StatCard({ icon: Icon, label, value, color, bg }: {
  icon: React.ElementType; label: string; value: number | undefined; color: string; bg: string
}) {
  return (
    <div className="bg-white rounded-2xl border border-gray-100 p-6 flex items-center gap-4">
      <div className="w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0" style={{ background: bg }}>
        <Icon className="w-5 h-5" style={{ color }} />
      </div>
      <div>
        <p className="text-2xl font-bold" style={{ color: '#0B2545' }}>{value ?? 0}</p>
        <p className="text-xs text-gray-500 mt-0.5">{label}</p>
      </div>
    </div>
  )
}

function ReminderItemRow({ item }: { item: {
  id: string
  submission_type: string
  ticket_number: string
  title: string
  days_since_submission: number
  days_since_last_update: number
  warning_level: string
  warning_color: string
} }) {
  const href = item.submission_type === 'legal_opinion'
    ? `/admin/legal-opinions/${item.id}`
    : `/admin/review-documents/${item.id}`

  return (
    <Link to={href} className="flex items-center gap-3 px-6 py-3.5 hover:bg-gray-50/50 transition-colors">
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-gray-900 truncate">{item.ticket_number} — {item.title}</p>
        <p className="text-xs text-gray-400 mt-0.5">
          {item.days_since_submission} hari sejak pengajuan · {item.days_since_last_update} hari sejak update terakhir
        </p>
      </div>
      <WarningBadge level={item.warning_level as 'YELLOW' | 'RED'} />
    </Link>
  )
}

function RemindersSection() {
  const { data: reminders } = useQuery({
    queryKey: ['dashboard', 'admin', 'reminders'],
    queryFn: () => dashboardService.getReminders(),
  })

  const yellow = reminders?.yellow ?? []
  const red = reminders?.red ?? []

  if (!yellow.length && !red.length) return null

  return (
    <div className="space-y-4">
      {red.length > 0 && (
        <div className="bg-white rounded-2xl border border-red-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-red-50 flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-red-500" />
            <h2 className="text-sm font-semibold text-red-700">Terlambat</h2>
            <span className="text-xs text-red-500">{red.length}</span>
          </div>
          <div className="divide-y divide-red-50">
            {red.map((item) => <ReminderItemRow key={item.id} item={item} />)}
          </div>
        </div>
      )}

      {yellow.length > 0 && (
        <div className="bg-white rounded-2xl border border-amber-100 overflow-hidden">
          <div className="px-6 py-4 border-b border-amber-50 flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-amber-500" />
            <h2 className="text-sm font-semibold text-amber-700">Perlu Perhatian</h2>
            <span className="text-xs text-amber-600">{yellow.length}</span>
          </div>
          <div className="divide-y divide-amber-50">
            {yellow.map((item) => <ReminderItemRow key={item.id} item={item} />)}
          </div>
        </div>
      )}
    </div>
  )
}

export default function AdminDashboardPage() {
  const { data: stats } = useQuery({
    queryKey: ['dashboard', 'admin', 'stats'],
    queryFn: dashboardService.getAdminStats,
  })

  const { data: recent } = useQuery({
    queryKey: ['dashboard', 'admin', 'recent'],
    queryFn: dashboardService.getAdminRecent,
  })

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Dashboard Admin</h1>
        <p className="text-sm text-gray-500 mt-0.5">Overview seluruh aktivitas Legal RIU Portal</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
        <StatCard icon={Users} label="Total User Aktif" value={stats?.total_users} color="#0B2545" bg="#EFF6FF" />
        <StatCard icon={FileText} label="Total Legal Opinion" value={stats?.total_legal_opinions} color="#C8102E" bg="#FEF2F2" />
        <StatCard icon={FileSearch} label="Total Review Dokumen" value={stats?.total_document_reviews} color="#7C3AED" bg="#F5F3FF" />
        <StatCard icon={Clock} label="Pending Review" value={stats?.pending_review} color="#D97706" bg="#FFFBEB" />
        <StatCard icon={AlertCircle} label="Need Revision" value={stats?.need_revision} color="#EA580C" bg="#FFF7ED" />
        <StatCard icon={RefreshCw} label="Resubmitted" value={stats?.resubmitted} color="#0891B2" bg="#ECFEFF" />
      </div>

      {/* Reminders */}
      <RemindersSection />

      {/* Recent submissions */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Legal Opinion */}
        <div className="bg-white rounded-2xl border border-gray-100">
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-50">
            <div className="flex items-center gap-2">
              <FileText className="w-4 h-4" style={{ color: '#C8102E' }} />
              <h2 className="text-sm font-semibold" style={{ color: '#0B2545' }}>Legal Opinion Terbaru</h2>
            </div>
            <Link to="/admin/legal-opinions" className="text-xs font-medium flex items-center gap-1 hover:underline" style={{ color: '#C8102E' }}>
              Kelola <ArrowRight className="w-3 h-3" />
            </Link>
          </div>
          <div className="divide-y divide-gray-50">
            {!recent?.legal_opinions?.length ? (
              <div className="px-6 py-10 text-center"><p className="text-sm text-gray-400">Belum ada pengajuan</p></div>
            ) : recent.legal_opinions.map((lo) => (
              <Link key={lo.id} to={`/admin/legal-opinions/${lo.id}`}
                className="flex items-center gap-3 px-6 py-3.5 hover:bg-gray-50/50 transition-colors">
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900 truncate">{lo.title}</p>
                  <p className="text-xs text-gray-400 mt-0.5">
                    {lo.ticket_number} · {lo.user?.full_name ?? '-'} · {formatDate(lo.created_at)}
                  </p>
                </div>
                <StatusBadge status={lo.status as SubmissionStatus} />
              </Link>
            ))}
          </div>
        </div>

        {/* Review Dokumen */}
        <div className="bg-white rounded-2xl border border-gray-100">
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-50">
            <div className="flex items-center gap-2">
              <FileSearch className="w-4 h-4" style={{ color: '#0B2545' }} />
              <h2 className="text-sm font-semibold" style={{ color: '#0B2545' }}>Review Dokumen Terbaru</h2>
            </div>
            <Link to="/admin/review-documents" className="text-xs font-medium flex items-center gap-1 hover:underline" style={{ color: '#C8102E' }}>
              Kelola <ArrowRight className="w-3 h-3" />
            </Link>
          </div>
          <div className="divide-y divide-gray-50">
            {!recent?.document_reviews?.length ? (
              <div className="px-6 py-10 text-center"><p className="text-sm text-gray-400">Belum ada pengajuan</p></div>
            ) : recent.document_reviews.map((dr) => (
              <Link key={dr.id} to={`/admin/review-documents/${dr.id}`}
                className="flex items-center gap-3 px-6 py-3.5 hover:bg-gray-50/50 transition-colors">
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900 truncate">{dr.document_name}</p>
                  <p className="text-xs text-gray-400 mt-0.5">
                    {dr.ticket_number} · {dr.user?.full_name ?? '-'} · {formatDate(dr.created_at)}
                  </p>
                </div>
                <StatusBadge status={dr.status as SubmissionStatus} />
              </Link>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
