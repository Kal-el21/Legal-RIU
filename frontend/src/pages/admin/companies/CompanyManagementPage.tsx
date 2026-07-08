import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, Search, Edit, Trash2, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useCompanies, useCreateCompany, useUpdateCompany, useDeleteCompany, useImportCompanies } from '@/hooks/useLegalCase'
import { legalCaseService } from '@/services/legal-case.service'
import ImportCard from '@/components/common/ImportCard'
import type { Company } from '@/types'

const schema = z.object({
  name: z.string().min(1, 'Wajib diisi'),
  email_domain: z.string().min(1, 'Wajib diisi'),
  is_internal: z.boolean(),
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

export default function CompanyManagementPage() {
  const [search, setSearch] = useState('')
  const [modal, setModal] = useState<'create' | 'edit' | null>(null)
  const [selected, setSelected] = useState<Company | null>(null)

  const { data, isLoading } = useCompanies()
  const createMutation = useCreateCompany()
  const updateMutation = useUpdateCompany()
  const deleteMutation = useDeleteCompany()
  const importMutation = useImportCompanies()

  const form = useForm<FormData>({ resolver: zodResolver(schema), defaultValues: { name: '', email_domain: '', is_internal: true } })

  const openEdit = (company: Company) => {
    setSelected(company)
    form.reset({ name: company.name, email_domain: company.email_domain, is_internal: company.is_internal })
    setModal('edit')
  }

  const closeModal = () => { setModal(null); setSelected(null); form.reset({ name: '', email_domain: '', is_internal: true }) }

  const onCreateSubmit = async (data: FormData) => {
    await createMutation.mutateAsync(data)
    form.reset()
    closeModal()
  }

  const onEditSubmit = async (data: FormData) => {
    if (!selected) return
    await updateMutation.mutateAsync({ id: selected.id, data })
    closeModal()
  }

  const handleDelete = async (company: Company) => {
    if (!confirm(`Hapus perusahaan ${company.name}?`)) return
    await deleteMutation.mutateAsync(company.id)
  }

  const filtered = data?.filter((c) => c.name.toLowerCase().includes(search.toLowerCase()) || c.email_domain.toLowerCase().includes(search.toLowerCase())) ?? data ?? []

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Perusahaan</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola data perusahaan</p>
        </div>
        <Button onClick={() => setModal('create')} className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }}>
          <Plus className="w-4 h-4" /> Tambah Perusahaan
        </Button>
      </div>

      <div className="relative mb-6 max-w-xs">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <Input placeholder="Cari perusahaan..." className="pl-9" value={search} onChange={(e) => { setSearch(e.target.value) }} />
      </div>

      <ImportCard
        title="Impor dari Excel"
        onImport={(file) => importMutation.mutateAsync(file)}
        onDownloadTemplate={() => legalCaseService.downloadCompanyTemplate()}
      />

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden mt-4">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !filtered.length ? (
          <div className="p-16 text-center">
            <p className="font-medium text-gray-500">Belum ada perusahaan</p>
          </div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Nama</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Domain Email</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Internal</th>
                <th className="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 uppercase tracking-wide">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {filtered.map((company) => (
                <tr key={company.id} className="hover:bg-gray-50/50 transition-colors">
                  <td className="px-6 py-4 text-sm text-gray-900">{company.name}</td>
                  <td className="px-6 py-4 text-sm text-gray-700">{company.email_domain}</td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium ${company.is_internal ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                      {company.is_internal ? 'Ya' : 'Tidak'}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1 justify-end">
                      <button onClick={() => openEdit(company)} title="Edit" className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors text-gray-400 hover:text-gray-700">
                        <Edit className="w-4 h-4" />
                      </button>
                      <button onClick={() => handleDelete(company)} title="Hapus" className="p-1.5 rounded-lg hover:bg-red-50 transition-colors text-gray-400 hover:text-red-600">
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
        <Modal title="Tambah Perusahaan" onClose={closeModal}>
          <form onSubmit={form.handleSubmit(onCreateSubmit)} className="space-y-4">
            <Field label="Nama Perusahaan" error={form.formState.errors.name?.message}>
              <Input {...form.register('name')} placeholder="Nama perusahaan" />
            </Field>
            <Field label="Domain Email" error={form.formState.errors.email_domain?.message}>
              <Input {...form.register('email_domain')} placeholder="contoh.co.id" />
            </Field>
            <div className="flex items-center gap-2">
              <input type="checkbox" id="is_internal" {...form.register('is_internal')} className="rounded border-gray-300" />
              <Label htmlFor="is_internal" className="text-sm text-gray-700">Perusahaan Internal</Label>
            </div>
            {createMutation.isError && <p className="text-xs text-red-500">{(createMutation.error as Error)?.message}</p>}
            <div className="flex gap-2 pt-2">
              <Button type="button" variant="outline" className="flex-1" onClick={closeModal}>Batal</Button>
              <Button type="submit" disabled={createMutation.isPending} className="flex-1 text-white" style={{ background: '#C8102E' }}>
                {createMutation.isPending ? 'Menyimpan...' : 'Buat'}
              </Button>
            </div>
          </form>
        </Modal>
      )}

      {modal === 'edit' && selected && (
        <Modal title={`Edit — ${selected.name}`} onClose={closeModal}>
          <form onSubmit={form.handleSubmit(onEditSubmit)} className="space-y-4">
            <Field label="Nama Perusahaan" error={form.formState.errors.name?.message}>
              <Input {...form.register('name')} />
            </Field>
            <Field label="Domain Email" error={form.formState.errors.email_domain?.message}>
              <Input {...form.register('email_domain')} />
            </Field>
            <div className="flex items-center gap-2">
              <input type="checkbox" id="is_internal" {...form.register('is_internal')} className="rounded border-gray-300" />
              <Label htmlFor="is_internal" className="text-sm text-gray-700">Perusahaan Internal</Label>
            </div>
            {updateMutation.isError && <p className="text-xs text-red-500">{(updateMutation.error as Error)?.message}</p>}
            <div className="flex gap-2 pt-2">
              <Button type="button" variant="outline" className="flex-1" onClick={closeModal}>Batal</Button>
              <Button type="submit" disabled={updateMutation.isPending} className="flex-1 text-white" style={{ background: '#0B2545' }}>
                {updateMutation.isPending ? 'Menyimpan...' : 'Simpan'}
              </Button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
