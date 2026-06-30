import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search, Filter, X, History } from 'lucide-react'
import { auditLogService } from '@/services/audit-log.service'
import { cn } from '@/lib/utils'
import type { AuditLog, AuditAction } from '@/types'

const ACTION_COLORS: Record<AuditAction, { bg: string; text: string }> = {
  STATUS_CHANGE: { bg: '#FEF3C7', text: '#92400E' },
  FILE_UPLOAD:   { bg: '#D1FAE5', text: '#065F46' },
  USER_UPDATE:   { bg: '#DBEAFE', text: '#1E40AF' },
  LOGIN:         { bg: '#E0E7FF', text: '#3730A3' },
  LOGOUT:        { bg: '#EDE9FE', text: '#5B21B6' },
  DELETE:        { bg: '#FEE2E2', text: '#991B1B' },
  FILE_DELETE:   { bg: '#FEE2E2', text: '#991B1B' },
}

function Badge({ action, className }: { action: AuditAction; className?: string }) {
  const colors = ACTION_COLORS[action] || { bg: '#F3F4F6', text: '#374151' }
  return (
    <span className={cn('inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium', className)}
      style={{ background: colors.bg, color: colors.text }}>
      {action.replace('_', ' ')}
    </span>
  )
}

export default function AuditLogPage() {
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState({
    action: '',
    entity_type: '',
    search: '',
    date_from: '',
    date_to: '',
  })
  const [showFilters, setShowFilters] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs', page, filters],
    queryFn: () => auditLogService.getAll({
      ...filters,
      page,
      limit: 20,
    }),
  })

  const items = data?.items || []
  const totalPages = data?.total_pages || 1
  const total = data?.total || 0

  const updateFilter = (key: string, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }))
    setPage(1)
  }

  const clearFilters = () => {
    setFilters({ action: '', entity_type: '', search: '', date_from: '', date_to: '' })
    setPage(1)
  }

  const hasActiveFilters = useMemo(() => {
    return Object.values(filters).some(v => v !== '')
  }, [filters])

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Audit Log</h1>
          <p className="text-sm text-gray-500 mt-0.5">Riwayat aktivitas dan perubahan sistem</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={cn('flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium transition-colors')}
            style={{ background: showFilters ? '#0B2545' : '#F3F4F6', color: showFilters ? 'white' : '#374151' }}
          >
            <Filter className="w-4 h-4" />
            Filter
          </button>
        </div>
      </div>

      {showFilters && (
        <div className="bg-white rounded-2xl border border-gray-100 p-5 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold" style={{ color: '#0B2545' }}>Filter Pencarian</h3>
            {hasActiveFilters && (
              <button onClick={clearFilters} className="flex items-center gap-1 text-xs text-red-500 hover:text-red-700">
                <X className="w-3 h-3" /> Hapus filter
              </button>
            )}
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Aksi</label>
              <select
                value={filters.action}
                onChange={(e) => updateFilter('action', e.target.value)}
                className="w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-red-500/20"
                style={{ borderColor: '#E5E7EB' }}
              >
                <option value="">Semua aksi</option>
                <option value="STATUS_CHANGE">Status Change</option>
                <option value="FILE_UPLOAD">File Upload</option>
                <option value="USER_UPDATE">User Update</option>
                <option value="LOGIN">Login</option>
                <option value="LOGOUT">Logout</option>
                <option value="DELETE">Delete</option>
                <option value="FILE_DELETE">File Delete</option>
              </select>
            </div>
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Entity Type</label>
              <input
                type="text"
                value={filters.entity_type}
                onChange={(e) => updateFilter('entity_type', e.target.value)}
                placeholder="Contoh: legal_opinion"
                className="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-red-500/20"
                style={{ borderColor: '#E5E7EB' }}
              />
            </div>
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Dari Tanggal</label>
              <input
                type="date"
                value={filters.date_from}
                onChange={(e) => updateFilter('date_from', e.target.value)}
                className="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-red-500/20"
                style={{ borderColor: '#E5E7EB' }}
              />
            </div>
            <div>
              <label className="text-xs text-gray-500 mb-1 block">Sampai Tanggal</label>
              <input
                type="date"
                value={filters.date_to}
                onChange={(e) => updateFilter('date_to', e.target.value)}
                className="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-red-500/20"
                style={{ borderColor: '#E5E7EB' }}
              />
            </div>
          </div>
          <div className="pt-1">
            <div className="relative">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                value={filters.search}
                onChange={(e) => updateFilter('search', e.target.value)}
                placeholder="Cari deskripsi, old value, atau new value..."
                className="w-full pl-9 pr-4 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-red-500/20"
                style={{ borderColor: '#E5E7EB' }}
              />
            </div>
          </div>
        </div>
      )}

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-50" style={{ background: '#F9FAFB' }}>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wider">Timestamp</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wider">User</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wider">Aksi</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wider">Entity</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wider">Perubahan</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wider">IP Address</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center">
                    <div className="flex flex-col items-center gap-2">
                      <div className="w-8 h-8 rounded-full border-2 border-red-500 border-t-transparent animate-spin" />
                      <p className="text-sm text-gray-400">Memuat data...</p>
                    </div>
                  </td>
                </tr>
              ) : items.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center">
                    <History className="w-10 h-10 text-gray-300 mx-auto mb-2" />
                    <p className="text-sm text-gray-400">Belum ada aktivitas audit log</p>
                  </td>
                </tr>
              ) : (
                items.map((log: AuditLog) => (
                  <tr key={log.id} className="hover:bg-gray-50/50 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">
                        {new Date(log.created_at).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })}
                      </div>
                      <div className="text-xs text-gray-400">
                        {new Date(log.created_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-900">{log.user?.full_name || '-'}</div>
                      <div className="text-xs text-gray-400">{log.user?.email || '-'}</div>
                    </td>
                    <td className="px-6 py-4">
                      <Badge action={log.action} />
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-600">{log.entity_type}</div>
                      <div className="text-xs text-gray-400 font-mono">{log.entity_id.slice(0, 8)}...</div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-xs text-gray-600 max-w-[200px]">
                        {log.old_value && (
                          <span className="line-through text-gray-400 mr-1">{log.old_value.length > 30 ? log.old_value.slice(0, 30) + '...' : log.old_value}</span>
                        )}
                        {(log.old_value && log.new_value) && (
                          <span className="text-gray-400 mx-1">→</span>
                        )}
                        {log.new_value && (
                          <span className="text-gray-900 font-medium">{log.new_value.length > 30 ? log.new_value.slice(0, 30) + '...' : log.new_value}</span>
                        )}
                        {!log.old_value && !log.new_value && log.description && (
                          <span className="text-gray-600">{log.description.length > 50 ? log.description.slice(0, 50) + '...' : log.description}</span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-xs text-gray-500 font-mono">{log.ip_address}</span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {totalPages > 1 && (
          <div className="px-6 py-4 border-t border-gray-50 flex items-center justify-between">
            <p className="text-xs text-gray-500">
              Menampilkan {(page - 1) * 20 + 1} - {Math.min(page * 20, total)} dari {total} aktivitas
            </p>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1.5 rounded-lg text-xs font-medium border border-gray-200 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-50"
                style={{ color: '#0B2545' }}
              >
                Sebelumnya
              </button>
              <span className="text-xs text-gray-500">
                Halaman {page} dari {totalPages}
              </span>
              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-3 py-1.5 rounded-lg text-xs font-medium border border-gray-200 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-50"
                style={{ color: '#0B2545' }}
              >
                Selanjutnya
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
