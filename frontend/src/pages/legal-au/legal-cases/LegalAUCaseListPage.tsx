import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Search, Plus, Eye } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useLegalCases } from '@/hooks/useLegalCase'
import { useAuthStore } from '@/store/auth.store'
import { formatDate } from '@/lib/utils'
import type { LegalCase } from '@/types'

export default function LegalAUCaseListPage() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const { data, isLoading } = useLegalCases({ page, limit: 10, search })
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const canCreateCase = hasPermission('case_management.create')

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Manajemen Kasus</h1>
          <p className="text-sm text-gray-500 mt-0.5">Daftar kasus hukum perusahaan Anda</p>
        </div>
        {canCreateCase && (
          <Link to="/legal-au/cases/new">
            <Button className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }}>
              <Plus className="w-4 h-4" /> Tambah Kasus
            </Button>
          </Link>
        )}
      </div>

      <div className="relative mb-6 max-w-xs">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <Input placeholder="Cari nama kasus..." className="pl-9" value={search} onChange={(e) => { setSearch(e.target.value); setPage(1) }} />
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !data?.items?.length ? (
          <div className="p-16 text-center">
            <p className="font-medium text-gray-500">Belum ada kasus</p>
          </div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="w-32 px-6 py-3.5 text-left text-xs font-semibold uppercase tracking-wide text-gray-500">Ticket Number</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold uppercase tracking-wide text-gray-500">Nama Kasus</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold uppercase tracking-wide text-gray-500">Jenis</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold uppercase tracking-wide text-gray-500">Status</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold uppercase tracking-wide text-gray-500">Tanggal</th>
                <th className="px-6 py-3.5 text-right text-xs font-semibold uppercase tracking-wide text-gray-500">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {data.items.map((item: LegalCase) => (
                <tr key={item.id} className="hover:bg-gray-50/50 transition-colors">
                  <td className="px-6 py-4">
                    <span className="text-xs font-mono font-medium text-gray-600 bg-gray-100 px-2 py-1 rounded">
                      {item.ticket_number || '-'}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm font-medium text-gray-900">{item.case_name}</p>
                    <p className="text-xs text-gray-400">{item.company?.name}</p>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-700">{item.case_type?.label}</td>
                  <td className="px-6 py-4">
                    <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-700">
                      {item.current_status || '-'}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500">{formatDate(item.case_date)}</td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1 justify-end">
                      <Link to={`/legal-au/cases/${item.id}`}>
                        <button title="Lihat" className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors text-gray-400 hover:text-gray-700">
                          <Eye className="w-4 h-4" />
                        </button>
                      </Link>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {data && data.total_pages > 1 && (
        <div className="px-6 py-4 border-t border-gray-100 flex items-center justify-between">
          <p className="text-sm text-gray-500">
            Menampilkan {((page - 1) * 10) + 1}–{Math.min(page * 10, data.total)} dari {data.total} kasus
          </p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" disabled={page === 1} onClick={() => setPage(p => p - 1)}>Sebelumnya</Button>
            <Button variant="outline" size="sm" disabled={page === data.total_pages} onClick={() => setPage(p => p + 1)}>Berikutnya</Button>
          </div>
        </div>
      )}
    </div>
  )
}
