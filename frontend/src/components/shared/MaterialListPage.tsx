import { useState } from 'react'
import { Plus, Search, Edit, Trash2, X } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import RichTextEditor from '@/components/common/RichTextEditor'
import { useMaterials, useCreateMaterial, useUpdateMaterial, useDeleteMaterial } from '@/hooks/useMaterial'
import type { LegalMaterial } from '@/types'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useAuthStore } from '@/store/auth.store'

const schema = z.object({
  title: z.string().min(1, 'Judul wajib diisi'),
  excerpt: z.string().optional(),
  content: z.string().min(1, 'Konten wajib diisi'),
})

type FormData = z.infer<typeof schema>

function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: React.ReactNode }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-2xl">
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

interface SharedMaterialListPageProps {
  role: 'ADMIN' | 'LEGAL' | 'LEGAL_AU'
  basePath: string
  title: string
  description: string
  showCreateButton?: boolean
  showEditButton?: boolean
  navigateToDetail?: (id: string) => string
}

export default function SharedMaterialListPage({
  role: _role,
  basePath,
  title,
  description,
  showCreateButton = true,
  showEditButton = true,
  navigateToDetail,
}: SharedMaterialListPageProps) {
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [modal, setModal] = useState<'create' | 'edit' | null>(null)
  const [selected, setSelected] = useState<LegalMaterial | null>(null)

  const defaultNavigateToDetail = (id: string) => `${basePath}/${id}`
  const getNavigateToDetail = navigateToDetail || defaultNavigateToDetail

  const { data, isLoading } = useMaterials()
  const createMutation = useCreateMaterial()
  const updateMutation = useUpdateMaterial()
  const deleteMutation = useDeleteMaterial()

  const hasPermission = useAuthStore((state) => state.hasPermission)
  const canManageMaterials = hasPermission('legal_material.manage')

  const form = useForm<FormData>({ resolver: zodResolver(schema), defaultValues: { title: '', excerpt: '', content: '' } })

  const openCreate = () => {
    setSelected(null)
    form.reset({ title: '', excerpt: '', content: '' })
    setModal('create')
  }

  const openEdit = (item: LegalMaterial) => {
    setSelected(item)
    form.reset({ title: item.title, excerpt: item.excerpt || '', content: item.content })
    setModal('edit')
  }

  const closeModal = () => { setModal(null); setSelected(null); form.reset({ title: '', excerpt: '', content: '' }) }

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

  const handleDelete = async (item: LegalMaterial) => {
    if (!confirm(`Hapus materi "${item.title}"?`)) return
    await deleteMutation.mutateAsync(item.id)
  }

  const filtered = data?.items?.filter((m) => m.title.toLowerCase().includes(search.toLowerCase())) ?? data?.items ?? []

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>{title}</h1>
          <p className="text-sm text-gray-500 mt-0.5">{description}</p>
        </div>
        {showCreateButton && canManageMaterials && (
          <Button onClick={openCreate} className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }}>
            <Plus className="w-4 h-4" /> Tambah Materi
          </Button>
        )}
      </div>

      <div className="relative mb-6 max-w-xs">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <Input placeholder="Cari judul..." className="pl-9" value={search} onChange={(e) => { setSearch(e.target.value) }} />
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !filtered.length ? (
          <div className="p-16 text-center">
            <p className="font-medium text-gray-500">Belum ada materi</p>
          </div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Judul</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Excerpt</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Dibuat</th>
                <th className="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 uppercase tracking-wide">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {filtered.map((item) => (
                <tr key={item.id} className="hover:bg-gray-50/50 transition-colors">
                  <td className="px-6 py-4 text-sm text-gray-900">{item.title}</td>
                  <td className="px-6 py-4 text-sm text-gray-700 max-w-xs truncate">{item.excerpt}</td>
                  <td className="px-6 py-4 text-sm text-gray-500">{new Date(item.created_at).toLocaleDateString('id-ID')}</td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1 justify-end">
                      {showEditButton && canManageMaterials && (
                        <button onClick={() => { openEdit(item); navigate(getNavigateToDetail(item.id)) }} title="Edit" className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors text-gray-400 hover:text-gray-700">
                          <Edit className="w-4 h-4" />
                        </button>
                      )}
                      {!showEditButton && (
                        <button onClick={() => navigate(getNavigateToDetail(item.id))} title="Lihat Detail" className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors text-gray-400 hover:text-gray-700">
                          <Edit className="w-4 h-4" />
                        </button>
                      )}
                      {canManageMaterials && (
                        <button onClick={() => handleDelete(item)} title="Hapus" className="p-1.5 rounded-lg hover:bg-red-50 transition-colors text-gray-400 hover:text-red-600">
                          <Trash2 className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {modal === 'create' && (
        <Modal title="Tambah Materi" onClose={closeModal}>
          <form onSubmit={form.handleSubmit(onCreateSubmit)} className="space-y-4">
            <Field label="Judul" error={form.formState.errors.title?.message}>
              <Input {...form.register('title')} placeholder="Judul materi" />
            </Field>
            <Field label="Excerpt" error={form.formState.errors.excerpt?.message}>
              <Input {...form.register('excerpt')} placeholder="Ringkasan singkat" />
            </Field>
            <Field label="Konten" error={form.formState.errors.content?.message}>
              <RichTextEditor value={form.watch('content') || ''} onChange={(v) => form.setValue('content', v)} />
              <input type="hidden" {...form.register('content')} />
            </Field>
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

      {canManageMaterials && modal === 'edit' && selected && (
        <Modal title={`Edit — ${selected.title}`} onClose={closeModal}>
          <form onSubmit={form.handleSubmit(onEditSubmit)} className="space-y-4">
            <Field label="Judul" error={form.formState.errors.title?.message}>
              <Input {...form.register('title')} />
            </Field>
            <Field label="Excerpt" error={form.formState.errors.excerpt?.message}>
              <Input {...form.register('excerpt')} />
            </Field>
            <Field label="Konten" error={form.formState.errors.content?.message}>
              <RichTextEditor value={form.watch('content') || ''} onChange={(v) => form.setValue('content', v)} />
              <input type="hidden" {...form.register('content')} />
            </Field>
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
