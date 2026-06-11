import { useState } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, Search, Edit, KeyRound, ToggleLeft, ToggleRight, X, Users, Trash2, Shield } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useUsers, useCreateUser, useUpdateUser, useUpdateUserStatus, useResetPassword, useDeleteUser } from '@/hooks/useUser'
import { formatDate } from '@/lib/utils'
import type { User } from '@/types'

const DIVISIONS = [
  'Underwriting',
  'Claims',
  'IT',
  'Finance',
  'HR',
  'Legal',
  'Marketing',
  'Operations',
  'Risk Management',
  'Reinsurance',
  'Actuarial',
  'Corporate',
  'Lainnya',
]

// ── Schemas ───────────────────────────────────────────────────────────────────

const createSchema = z.object({
  full_name: z.string().min(1, 'Wajib diisi'),
  email: z.string().email('Email tidak valid'),
  password: z.string().min(8, 'Minimal 8 karakter'),
  position: z.string().min(1, 'Wajib diisi'),
  division: z.string().min(1, 'Wajib diisi'),
  role: z.enum(['USER', 'ADMIN']),
})

const editSchema = z.object({
  full_name: z.string().min(1, 'Wajib diisi'),
  position: z.string().min(1, 'Wajib diisi'),
  division: z.string().min(1, 'Wajib diisi'),
  role: z.enum(['USER', 'ADMIN']),
})

const resetSchema = z.object({
  new_password: z.string().min(8, 'Minimal 8 karakter'),
  confirm_password: z.string(),
}).refine((d) => d.new_password === d.confirm_password, {
  message: 'Password tidak cocok',
  path: ['confirm_password'],
})

type CreateForm = z.infer<typeof createSchema>
type EditForm = z.infer<typeof editSchema>
type ResetForm = z.infer<typeof resetSchema>

// ── Modal wrapper ─────────────────────────────────────────────────────────────

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

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function UserManagementPage() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [modal, setModal] = useState<'create' | 'edit' | 'reset' | null>(null)
  const [selected, setSelected] = useState<User | null>(null)

  const { data, isLoading } = useUsers({ page, limit: 10, search })
  const createMutation = useCreateUser()
  const updateMutation = useUpdateUser()
  const statusMutation = useUpdateUserStatus()
  const resetMutation = useResetPassword()
  const deleteMutation = useDeleteUser()

  const createForm = useForm<CreateForm>({ resolver: zodResolver(createSchema), defaultValues: { role: 'USER' } })
  const editForm = useForm<EditForm>({ resolver: zodResolver(editSchema) })
  const resetForm = useForm<ResetForm>({ resolver: zodResolver(resetSchema) })

  const openEdit = (user: User) => {
    setSelected(user)
    editForm.reset({ full_name: user.full_name, position: user.position, division: user.division, role: user.role })
    setModal('edit')
  }

  const openReset = (user: User) => {
    setSelected(user)
    resetForm.reset()
    setModal('reset')
  }

  const closeModal = () => { setModal(null); setSelected(null); createForm.reset({ role: 'USER' }) }

  const onCreateSubmit = async (data: CreateForm) => {
    await createMutation.mutateAsync(data)
    createForm.reset()
    closeModal()
  }

  const onEditSubmit = async (data: EditForm) => {
    if (!selected) return
    await updateMutation.mutateAsync({ id: selected.id, data })
    closeModal()
  }

  const onResetSubmit = async (data: ResetForm) => {
    if (!selected) return
    await resetMutation.mutateAsync({ id: selected.id, password: data.new_password })
    closeModal()
  }

  const toggleStatus = async (user: User) => {
    const next = user.status === 'ACTIVE' ? 'INACTIVE' : 'ACTIVE'
    if (!confirm(`${next === 'INACTIVE' ? 'Nonaktifkan' : 'Aktifkan'} user ${user.full_name}?`)) return
    await statusMutation.mutateAsync({ id: user.id, status: next })
  }

  const handleDelete = async (user: User) => {
    if (!confirm(`Hapus user ${user.full_name}?`)) return
    await deleteMutation.mutateAsync(user.id)
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>User Management</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola akun pengguna Legal RIU Portal</p>
        </div>
        <Button onClick={() => setModal('create')} className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }}>
          <Plus className="w-4 h-4" /> Tambah User
        </Button>
      </div>

      {/* Search */}
      <div className="relative mb-6 max-w-xs">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <Input
          placeholder="Cari nama atau email..."
          className="pl-9"
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }}
        />
      </div>

      {/* Table */}
      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !data?.items?.length ? (
          <div className="p-16 text-center">
            <div className="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mx-auto mb-4">
              <Users className="w-7 h-7 text-gray-400" />
            </div>
            <p className="font-medium text-gray-500">Belum ada user</p>
          </div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Nama</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Jabatan & Divisi</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Role</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Status</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Bergabung</th>
                <th className="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 uppercase tracking-wide">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {data.items.map((user) => (
                <tr key={user.id} className="hover:bg-gray-50/50 transition-colors">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold flex-shrink-0"
                        style={{ background: user.role === 'ADMIN' ? '#C8102E' : '#0B2545' }}>
                        {user.full_name.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-900">{user.full_name}</p>
                        <p className="text-xs text-gray-400">{user.email}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-700">{user.position}</p>
                    <p className="text-xs text-gray-400">{user.division}</p>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium ${
                      user.role === 'ADMIN' ? 'bg-red-100 text-red-700' : 'bg-blue-100 text-blue-700'
                    }`}>
                      {user.role === 'ADMIN' ? <Shield className="w-3 h-3 mr-1" /> : null}
                      {user.role}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium ${
                      user.status === 'ACTIVE' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
                    }`}>
                      {user.status === 'ACTIVE' ? 'Aktif' : 'Nonaktif'}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-500">{formatDate(user.created_at)}</p>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-1 justify-end">
                      <button onClick={() => openEdit(user)} title="Edit"
                        className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors text-gray-400 hover:text-gray-700">
                        <Edit className="w-4 h-4" />
                      </button>
                      <button onClick={() => openReset(user)} title="Reset Password"
                        className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors text-gray-400 hover:text-gray-700">
                        <KeyRound className="w-4 h-4" />
                      </button>
                      <button onClick={() => toggleStatus(user)} title={user.status === 'ACTIVE' ? 'Nonaktifkan' : 'Aktifkan'}
                        className={`p-1.5 rounded-lg hover:bg-gray-100 transition-colors ${
                          user.status === 'ACTIVE' ? 'text-green-500 hover:text-green-700' : 'text-gray-400 hover:text-gray-600'
                        }`}>
                        {user.status === 'ACTIVE' ? <ToggleRight className="w-4 h-4" /> : <ToggleLeft className="w-4 h-4" />}
                      </button>
                      <button onClick={() => handleDelete(user)} title="Hapus"
                        className="p-1.5 rounded-lg hover:bg-red-50 transition-colors text-gray-400 hover:text-red-600">
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {data && data.total_pages > 1 && (
          <div className="px-6 py-4 border-t border-gray-100 flex items-center justify-between">
            <p className="text-sm text-gray-500">
              Menampilkan {((page - 1) * 10) + 1}–{Math.min(page * 10, data.total)} dari {data.total} user
            </p>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" disabled={page === 1} onClick={() => setPage(p => p - 1)}>Sebelumnya</Button>
              <Button variant="outline" size="sm" disabled={page === data.total_pages} onClick={() => setPage(p => p + 1)}>Berikutnya</Button>
            </div>
          </div>
        )}
      </div>

      {/* ── Create Modal ─────────────────────────────────────────────────── */}
      {modal === 'create' && (
        <Modal title="Tambah User Baru" onClose={closeModal}>
          <form onSubmit={createForm.handleSubmit(onCreateSubmit)} className="space-y-4">
            <Field label="Nama Lengkap" error={createForm.formState.errors.full_name?.message}>
              <Input {...createForm.register('full_name')} placeholder="Nama lengkap" />
            </Field>
            <Field label="Email Kantor" error={createForm.formState.errors.email?.message}>
              <Input {...createForm.register('email')} type="email" placeholder="email@indonesiare.co.id" />
            </Field>
            <Field label="Password" error={createForm.formState.errors.password?.message}>
              <Input {...createForm.register('password')} type="password" placeholder="Min. 8 karakter" />
            </Field>
<div className="grid grid-cols-2 gap-4">
               <Field label="Posisi Jabatan" error={createForm.formState.errors.position?.message}>
                 <Input {...createForm.register('position')} placeholder="Jabatan" />
               </Field>
               <Field label="Divisi" error={createForm.formState.errors.division?.message}>
                 <Controller
                   name="division"
                   control={createForm.control}
                   render={({ field }) => (
                     <Select onValueChange={field.onChange} value={field.value}>
                       <SelectTrigger><SelectValue placeholder="Pilih divisi" /></SelectTrigger>
                       <SelectContent>
                         {DIVISIONS.map((d) => <SelectItem key={d} value={d}>{d}</SelectItem>)}
                       </SelectContent>
                     </Select>
                   )}
                 />
               </Field>
             </div>
            <Field label="Role" error={createForm.formState.errors.role?.message}>
              <Controller
                name="role"
                control={createForm.control}
                render={({ field }) => (
                  <Select onValueChange={field.onChange} value={field.value}>
                    <SelectTrigger><SelectValue placeholder="Pilih role" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="USER">User</SelectItem>
                      <SelectItem value="ADMIN">Admin</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
            </Field>
            {createMutation.isError && createMutation.error && (
              <p className="text-xs text-red-500">
                {(createMutation.error as any)?.response?.data?.message || (createMutation.error as Error)?.message}
              </p>
            )}
            <div className="flex gap-2 pt-2">
              <Button type="button" variant="outline" className="flex-1" onClick={closeModal}>Batal</Button>
              <Button type="submit" disabled={createMutation.isPending} className="flex-1 text-white" style={{ background: '#C8102E' }}>
                {createMutation.isPending ? 'Menyimpan...' : 'Buat User'}
              </Button>
            </div>
          </form>
        </Modal>
      )}

      {/* ── Edit Modal ───────────────────────────────────────────────────── */}
      {modal === 'edit' && selected && (
        <Modal title={`Edit — ${selected.full_name}`} onClose={closeModal}>
          <form onSubmit={editForm.handleSubmit(onEditSubmit)} className="space-y-4">
            <Field label="Nama Lengkap" error={editForm.formState.errors.full_name?.message}>
              <Input {...editForm.register('full_name')} />
            </Field>
            <Field label="Posisi Jabatan" error={editForm.formState.errors.position?.message}>
              <Input {...editForm.register('position')} />
            </Field>
<Field label="Divisi" error={editForm.formState.errors.division?.message}>
               <Controller
                 name="division"
                 control={editForm.control}
                 render={({ field }) => (
                   <Select onValueChange={field.onChange} value={field.value}>
                     <SelectTrigger><SelectValue placeholder="Pilih divisi" /></SelectTrigger>
                     <SelectContent>
                       {DIVISIONS.map((d) => <SelectItem key={d} value={d}>{d}</SelectItem>)}
                     </SelectContent>
                   </Select>
                 )}
               />
             </Field>
            <Field label="Role" error={editForm.formState.errors.role?.message}>
              <Controller
                name="role"
                control={editForm.control}
                render={({ field }) => (
                  <Select onValueChange={field.onChange} value={field.value}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="USER">User</SelectItem>
                      <SelectItem value="ADMIN">Admin</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              />
            </Field>
            {updateMutation.isError && (
              <p className="text-xs text-red-500">{(updateMutation.error as Error)?.message}</p>
            )}
            <div className="flex gap-2 pt-2">
              <Button type="button" variant="outline" className="flex-1" onClick={closeModal}>Batal</Button>
              <Button type="submit" disabled={updateMutation.isPending} className="flex-1 text-white" style={{ background: '#0B2545' }}>
                {updateMutation.isPending ? 'Menyimpan...' : 'Simpan'}
              </Button>
            </div>
          </form>
        </Modal>
      )}

      {/* ── Reset Password Modal ─────────────────────────────────────────── */}
      {modal === 'reset' && selected && (
        <Modal title={`Reset Password — ${selected.full_name}`} onClose={closeModal}>
          <form onSubmit={resetForm.handleSubmit(onResetSubmit)} className="space-y-4">
            <p className="text-sm text-gray-500">Password baru untuk <strong>{selected.email}</strong></p>
            <Field label="Password Baru" error={resetForm.formState.errors.new_password?.message}>
              <Input {...resetForm.register('new_password')} type="password" placeholder="Min. 8 karakter" />
            </Field>
            <Field label="Konfirmasi Password" error={resetForm.formState.errors.confirm_password?.message}>
              <Input {...resetForm.register('confirm_password')} type="password" placeholder="Ulangi password" />
            </Field>
            {resetMutation.isError && (
              <p className="text-xs text-red-500">{(resetMutation.error as Error)?.message}</p>
            )}
            {resetMutation.isSuccess && (
              <p className="text-xs text-green-600">Password berhasil direset!</p>
            )}
            <div className="flex gap-2 pt-2">
              <Button type="button" variant="outline" className="flex-1" onClick={closeModal}>Batal</Button>
              <Button type="submit" disabled={resetMutation.isPending} className="flex-1 text-white" style={{ background: '#C8102E' }}>
                {resetMutation.isPending ? 'Mereset...' : 'Reset Password'}
              </Button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}