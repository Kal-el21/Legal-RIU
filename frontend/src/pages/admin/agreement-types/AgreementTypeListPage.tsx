import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Plus, Settings2, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAgreementTypes, useCreateAgreementType } from '@/hooks/useAgreementType'

// Jenis perjanjian yang bisa diajukan (mis. PKS). Isian dan template tiap jenis
// diatur di halaman detailnya.
export default function AgreementTypeListPage() {
  const navigate = useNavigate()
  const { data = [], isLoading } = useAgreementTypes()
  const create = useCreateAgreementType()
  const [open, setOpen] = useState(false)
  const [code, setCode] = useState('')
  const [name, setName] = useState('')

  const submit = async (event: React.FormEvent) => {
    event.preventDefault()
    const created = await create.mutateAsync({ code, name, fields: [] })
    navigate(`/admin/agreement-types/${created.code}`)
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Jenis Perjanjian</h1>
          <p className="text-sm text-gray-500 mt-0.5">Atur jenis perjanjian, isian formnya, dan template Word-nya.</p>
        </div>
        <Button onClick={() => setOpen(true)} className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }}>
          <Plus className="w-4 h-4" /> Tambah Jenis
        </Button>
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !data.length ? (
          <div className="p-16 text-center text-gray-500">Belum ada jenis perjanjian</div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Kode</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Nama</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Isian</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Status</th>
                <th className="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 uppercase tracking-wide">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {data.map((item) => (
                <tr key={item.id} className="hover:bg-gray-50/50 transition-colors">
                  <td className="px-6 py-4 text-sm font-mono text-gray-900">{item.code}</td>
                  <td className="px-6 py-4 text-sm text-gray-700">{item.name}</td>
                  <td className="px-6 py-4 text-sm text-gray-500">
                    {(item.fields ?? []).reduce((n, s) => n + s.fields.length, 0)} isian
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium ${item.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                      {item.is_active ? 'Aktif' : 'Nonaktif'}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <Button asChild variant="outline" size="sm">
                      <Link to={`/admin/agreement-types/${item.code}`}>
                        <Settings2 className="w-4 h-4" /> Atur
                      </Link>
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40">
          <form onSubmit={submit} className="bg-white rounded-2xl shadow-2xl w-full max-w-md">
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
              <h2 className="text-base font-semibold" style={{ color: '#0B2545' }}>Tambah Jenis Perjanjian</h2>
              <Button type="button" variant="ghost" size="icon-sm" onClick={() => setOpen(false)} aria-label="Tutup" title="Tutup">
                <X className="w-4 h-4 text-gray-500" />
              </Button>
            </div>
            <div className="px-6 py-5 space-y-4">
              <div className="space-y-1.5">
                <Label htmlFor="kode-jenis">Kode</Label>
                <Input id="kode-jenis" value={code} onChange={(e) => setCode(e.target.value.toUpperCase())} placeholder="mis. NDA" required />
                <p className="text-xs text-gray-400">Huruf besar, angka, garis bawah, atau tanda hubung. Tidak bisa diubah setelah dibuat.</p>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="nama-jenis">Nama</Label>
                <Input id="nama-jenis" value={name} onChange={(e) => setName(e.target.value)} placeholder="mis. Perjanjian Kerahasiaan" required />
              </div>
            </div>
            <div className="flex justify-end gap-2 px-6 pb-5">
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>Batal</Button>
              <Button type="submit" disabled={create.isPending} className="text-white" style={{ background: '#C8102E' }}>
                {create.isPending ? 'Menyimpan...' : 'Buat lalu atur isian'}
              </Button>
            </div>
          </form>
        </div>
      )}
    </div>
  )
}
