import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { FileText, FileSearch, Clock, AlertCircle, CheckCircle, Plus, ArrowRight } from 'lucide-react'
import { dashboardService } from '@/services/dashboard.service'
import { useAuthStore } from '@/store/auth.store'
import StatusBadge from '@/components/common/StatusBadge'
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

export default function UserDashboardPage() {
  const user = useAuthStore((s) => s.user)

  const { data: stats } = useQuery({
    queryKey: ['dashboard', 'user', 'stats'],
    queryFn: dashboardService.getUserStats,
  })

  const { data: recent } = useQuery({
    queryKey: ['dashboard', 'user', 'recent'],
    queryFn: dashboardService.getUserRecent,
  })

  const hour = new Date().getHours()
  const greeting = hour < 12 ? 'Selamat Pagi' : hour < 15 ? 'Selamat Siang' : hour < 18 ? 'Selamat Sore' : 'Selamat Malam'

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-8">
      {/* Header */}
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <p className="text-sm text-gray-500">{greeting},</p>
          <h1 className="text-2xl font-bold mt-0.5" style={{ color: '#0B2545' }}>{user?.full_name}</h1>
          <p className="text-sm text-gray-400 mt-0.5">{user?.position} · {user?.division}</p>
        </div>
        <div className="flex gap-3">
          <Link to="/dashboard/legal-opinions/new"
            className="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium text-white transition-all hover:opacity-90"
            style={{ background: '#C8102E' }}>
            <Plus className="w-4 h-4" /> Legal Opinion
          </Link>
          <Link to="/dashboard/review-documents/new"
            className="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium text-white transition-all hover:opacity-90"
            style={{ background: '#0B2545' }}>
            <Plus className="w-4 h-4" /> Review Dokumen
          </Link>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-5 gap-4">
        <StatCard icon={FileText} label="Legal Opinion" value={stats?.total_legal_opinions} color="#C8102E" bg="#FEF2F2" />
        <StatCard icon={FileSearch} label="Review Dokumen" value={stats?.total_document_reviews} color="#0B2545" bg="#EFF6FF" />
        <StatCard icon={Clock} label="Pending" value={stats?.pending} color="#D97706" bg="#FFFBEB" />
        <StatCard icon={AlertCircle} label="Perlu Revisi" value={stats?.need_revision} color="#EA580C" bg="#FFF7ED" />
        <StatCard icon={CheckCircle} label="Selesai" value={stats?.completed} color="#16A34A" bg="#F0FDF4" />
      </div>

      {/* Recent activities */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Legal Opinion recent */}
        <div className="bg-white rounded-2xl border border-gray-100">
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-50">
            <div className="flex items-center gap-2">
              <FileText className="w-4 h-4" style={{ color: '#C8102E' }} />
              <h2 className="text-sm font-semibold" style={{ color: '#0B2545' }}>Legal Opinion Terbaru</h2>
            </div>
            <Link to="/dashboard/legal-opinions" className="text-xs font-medium flex items-center gap-1 hover:underline" style={{ color: '#C8102E' }}>
              Lihat semua <ArrowRight className="w-3 h-3" />
            </Link>
          </div>
          <div className="divide-y divide-gray-50">
            {!recent?.legal_opinions?.length ? (
              <div className="px-6 py-10 text-center">
                <p className="text-sm text-gray-400">Belum ada pengajuan</p>
              </div>
            ) : recent.legal_opinions.map((lo) => (
              <Link key={lo.id} to={`/dashboard/legal-opinions/${lo.id}`}
                className="flex items-center gap-3 px-6 py-3.5 hover:bg-gray-50/50 transition-colors">
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900 truncate">{lo.title}</p>
                  <p className="text-xs text-gray-400 mt-0.5">{lo.ticket_number} · {formatDate(lo.created_at)}</p>
                </div>
                <StatusBadge status={lo.status as SubmissionStatus} />
              </Link>
            ))}
          </div>
        </div>

        {/* Review Dokumen recent */}
        <div className="bg-white rounded-2xl border border-gray-100">
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-50">
            <div className="flex items-center gap-2">
              <FileSearch className="w-4 h-4" style={{ color: '#0B2545' }} />
              <h2 className="text-sm font-semibold" style={{ color: '#0B2545' }}>Review Dokumen Terbaru</h2>
            </div>
            <Link to="/dashboard/review-documents" className="text-xs font-medium flex items-center gap-1 hover:underline" style={{ color: '#C8102E' }}>
              Lihat semua <ArrowRight className="w-3 h-3" />
            </Link>
          </div>
          <div className="divide-y divide-gray-50">
            {!recent?.document_reviews?.length ? (
              <div className="px-6 py-10 text-center">
                <p className="text-sm text-gray-400">Belum ada pengajuan</p>
              </div>
            ) : recent.document_reviews.map((dr) => (
              <Link key={dr.id} to={`/dashboard/review-documents/${dr.id}`}
                className="flex items-center gap-3 px-6 py-3.5 hover:bg-gray-50/50 transition-colors">
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900 truncate">{dr.document_name}</p>
                  <p className="text-xs text-gray-400 mt-0.5">{dr.ticket_number} · {formatDate(dr.created_at)}</p>
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