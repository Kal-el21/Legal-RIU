import { useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, FileText, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { formatDateTime } from '@/lib/utils'
import { useAuthStore } from '@/store/auth.store'

export interface DraftItem {
  id: string
  judul: string
  diubah: string
}

interface DraftListPageProps {
  judul: string
  kembali: string
  izin: string // create.own modulnya; draf hanya untuk yang boleh membuat pengajuan
  items?: DraftItem[]
  isLoading: boolean
  lanjutkan: (id: string) => string
  onHapus: (id: string) => void
}

// Daftar draf milik sendiri untuk satu fitur pengajuan (PR Draf). Data dan aksi hapus
// berasal dari halaman pembungkus tiap modul.
export default function DraftListPage({ judul, kembali, izin, items, isLoading, lanjutkan, onHapus }: DraftListPageProps) {
  const navigate = useNavigate()
  const boleh = useAuthStore((s) => s.hasPermission(izin))

  useEffect(() => {
    if (!boleh) navigate(kembali, { replace: true })
  }, [boleh, kembali, navigate])

  const hapus = (item: DraftItem) => {
    if (confirm(`Hapus draf "${item.judul}"? Isian dan lampirannya ikut terhapus permanen.`)) onHapus(item.id)
  }

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <Button variant="outline" size="icon" onClick={() => navigate(kembali)} aria-label="Kembali" title="Kembali">
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </Button>
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>{judul}</h1>
          <p className="text-sm text-gray-500 mt-0.5">Pengajuan yang belum dikirim. Hanya Anda yang dapat melihatnya.</p>
        </div>
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !items?.length ? (
          <div className="p-16 text-center">
            <div className="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <FileText className="w-7 h-7 text-gray-400" />
            </div>
            <p className="font-medium text-gray-500">Belum ada draf</p>
          </div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Judul</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Terakhir diubah</th>
                <th className="px-6 py-3.5" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {items.map((item) => (
                <tr key={item.id} className="hover:bg-gray-50/50 transition-colors">
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-700 max-w-md truncate">{item.judul}</p>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-500">{formatDateTime(item.diubah)}</p>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center justify-end gap-2">
                      <Link to={lanjutkan(item.id)}>
                        <Button variant="outline" size="sm">Lanjutkan</Button>
                      </Link>
                      <Button variant="outline" size="icon-sm" onClick={() => hapus(item)} aria-label="Hapus draf" title="Hapus draf"
                        className="text-red-600 hover:bg-red-50 hover:text-red-700">
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
