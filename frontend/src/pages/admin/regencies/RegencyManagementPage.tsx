import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, Search, Edit, Trash2, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useRegencies, useImportRegencies } from '@/hooks/useLegalCase'
import { legalCaseService } from '@/services/legal-case.service'
import ImportCard from '@/components/common/ImportCard'
import api from '@/services/api'
import { useQueryClient } from '@tanstack/react-query'
import type { Regency } from '@/types'

const schema = z.object({
  name: z.string().min(1, 'Wajib diisi'),
  province: z.string().min(1, 'Wajib diisi'),
  type: z.string().min(1, 'Wajib diisi'),
})

type FormData = z.infer<typeof schema>

function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: React.ReactNode }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-md">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <h2 className="text-base font-semibold" style={{ color: '#0B2545' }}>{title}</h2>
          <button onClick={onClose} className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors">
            <X className="w-4 h-4 text-gray-500" />
          </button>
        </div>
        <div className="px-6 py-5">{children}</div>
      </div>
    </div>
  )
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label className="text-sm font-medium text-gray-700">{label}</Label>
      {children}
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}

export default function RegencyManagementPage() {
  const [search, setSearch] = useState('')
  const [modal, setModal] = useState<'create' | 'edit' | null>(null)
  const [selected, setSelected] = useState<Regency | null>(null)

  const { data: regenciesData, isLoading } = useRegencies({ limit: 500 })
  const importMutation = useImportRegencies()
  const queryClient = useQueryClient()

  const form = useForm<FormData>({ resolver: zodResolver(schema), defaultValues: { name: '', province: '', type: 'kabupaten' } })

  const openEdit = (item: Regency) => {
    setSelected(item)
    form.reset({ name: item.name, province: item.province, type: item.type })
    setModal('edit')
  }

  const closeModal = () => { setModal(null); setSelected(null); form.reset({ name: '', province: '', type: 'kabupaten' }) }

  const onCreateSubmit = async (data: FormData) => {
    await api.post('/admin/regencies', data)
    form.reset()
    closeModal()
    window.location.reload()
  }

  const onEditSubmit = async (data: FormData) => {
    if (!selected) return
    await api.put(`/admin/regencies/${selected.id}`, data)
    closeModal()
    window.location.reload()
  }

  const handleDelete = async (item: Regency) => {
    if (!confirm(`Hapus ${item.name}?`)) return
    await api.delete(`/admin/regencies/${item.id}`)
    queryClient.invalidateQueries({ queryKey: ['legal-cases', 'regencies'] })
  }

  const filtered = (regenciesData ?? []).filter((r) => r.name.toLowerCase().includes(search.toLowerCase()) || r.province.toLowerCase().includes(search.toLowerCase()))

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Kabupaten/Kota</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola data kabupaten/kota</p>
        </div>
        <Button onClick={() => setModal('create')} className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }}>
          <Plus className="w-4 h-4" /> Tambah
        </Button>
      </div>

      <div className="relative mb-6 max-w-xs">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <Input placeholder="Cari..." className="pl-9" value={search} onChange={(e) => { setSearch(e.target.value) }} />
      </div>

      <ImportCard
        title="Impor dari Excel"
        onImport={(file) => importMutation.mutateAsync(file)}
        onDownloadTemplate={() => legalCaseService.downloadRegencyTemplate()}
      />

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden mt-4">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !filtered.length ? (
          <div className="p-16 text-center"><p className="font-medium text-gray-500">Belum ada data</p></div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Nama</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Provinsi</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Tipe</th>
                <th className="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 uppercase tracking-wide">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {filtered.map((item) => (
                <tr key={item.id} className="hover:bg-gray-50/50 transition-colors">
                  <td className="px-6 py-4 text-sm text-gray-900">{item.name}</td>
                  <td className="px-6 py-4 text-sm text-gray-700">{item.province}</td>
                  <td className="px-6 py-4 text-sm text-gray-700">{item.type}</td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1 justify-end">
                      <button onClick={() => openEdit(item)} title="Edit" className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors text-gray-400 hover:text-gray-700">
                        <Edit className="w-4 h-4" />
                      </button>
                      <button onClick={() => handleDelete(item)} title="Hapus" className="p-1.5 rounded-lg hover:bg-red-50 transition-colors text-gray-400 hover:text-red-600">
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {modal === 'create' && (
        <Modal title="Tambah Kabupaten/Kota" onClose={closeModal}>
          <form onSubmit={form.handleSubmit(onCreateSubmit)} className="space-y-4">
            <Field label="Nama" error={form.formState.errors.name?.message}>
              <Input {...form.register('name')} placeholder="Nama kabupaten/kota" />
            </Field>
            <Field label="Provinsi" error={form.formState.errors.province?.message}>
              <Input {...form.register('province')} placeholder="Provinsi" />
            </Field>
            <Field label="Tipe" error={form.formState.errors.type?.message}>
              <Input {...form.register('type')} placeholder="kabupaten/kota" />
            </Field>
            <div className="flex gap-2 pt-2">
              <Button type="button" variant="outline" className="flex-1" onClick={closeModal}>Batal</Button>
              <Button type="submit" className="flex-1 text-white" style={{ background: '#C8102E' }}>
                Buat
              </Button>
            </div>
          </form>
        </Modal>
      )}

      {modal === 'edit' && selected && (
        <Modal title={`Edit — ${selected.name}`} onClose={closeModal}>
          <form onSubmit={form.handleSubmit(onEditSubmit)} className="space-y-4">
            <Field label="Nama" error={form.formState.errors.name?.message}>
              <Input {...form.register('name')} />
            </Field>
            <Field label="Provinsi" error={form.formState.errors.province?.message}>
              <Input {...form.register('province')} />
            </Field>
            <Field label="Tipe" error={form.formState.errors.type?.message}>
              <Input {...form.register('type')} />
            </Field>
            <div className="flex gap-2 pt-2">
              <Button type="button" variant="outline" className="flex-1" onClick={closeModal}>Batal</Button>
              <Button type="submit" className="flex-1 text-white" style={{ background: '#0B2545' }}>
                Simpan
              </Button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
