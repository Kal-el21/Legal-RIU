import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation } from '@tanstack/react-query'
import { User, KeyRound, CheckCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuthStore } from '@/store/auth.store'
import { authService } from '@/services/auth.service'

const schema = z.object({
  current_password: z.string().min(1, 'Wajib diisi'),
  new_password: z.string().min(8, 'Minimal 8 karakter'),
  confirm_password: z.string(),
}).refine((d) => d.new_password === d.confirm_password, {
  message: 'Password tidak cocok',
  path: ['confirm_password'],
})

type FormData = z.infer<typeof schema>

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label className="text-sm font-medium text-gray-700">{label}</Label>
      {children}
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}

export default function ProfilePage() {
  const user = useAuthStore((s) => s.user)

  const { register, handleSubmit, reset, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  const mutation = useMutation({
    mutationFn: (data: { current_password: string; new_password: string }) =>
      authService.changePassword(data),
    onSuccess: () => reset(),
  })

  const onSubmit = (data: FormData) => {
    mutation.mutate({ current_password: data.current_password, new_password: data.new_password })
  }

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Profil Saya</h1>
        <p className="text-sm text-gray-500 mt-0.5">Informasi akun dan keamanan</p>
      </div>

      {/* Info */}
      <div className="bg-white rounded-2xl border border-gray-100 p-6">
        <div className="flex items-center gap-4 mb-6">
          <div className="w-16 h-16 rounded-2xl flex items-center justify-center text-white text-2xl font-bold flex-shrink-0"
            style={{ background: '#0B2545' }}>
            {user?.full_name?.charAt(0).toUpperCase()}
          </div>
          <div>
            <p className="text-lg font-bold" style={{ color: '#0B2545' }}>{user?.full_name}</p>
            <p className="text-sm text-gray-500">{user?.email}</p>
            <span className={`inline-flex items-center mt-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${
              user?.role === 'ADMIN' ? 'bg-red-100 text-red-700' : 'bg-blue-100 text-blue-700'
            }`}>
              {user?.role}
            </span>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-4 border-t border-gray-100">
          {[
            { icon: User, label: 'Posisi Jabatan', value: user?.position },
            { icon: User, label: 'Divisi', value: user?.division },
          ].map((item) => (
            <div key={item.label} className="flex items-start gap-3 p-3 rounded-xl bg-gray-50">
              <item.icon className="w-4 h-4 text-gray-400 mt-0.5 flex-shrink-0" />
              <div>
                <p className="text-xs text-gray-500">{item.label}</p>
                <p className="text-sm font-medium text-gray-800 mt-0.5">{item.value ?? '-'}</p>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Change password */}
      <div className="bg-white rounded-2xl border border-gray-100 p-6">
        <div className="flex items-center gap-2 mb-5">
          <KeyRound className="w-4 h-4" style={{ color: '#C8102E' }} />
          <h2 className="text-base font-semibold" style={{ color: '#0B2545' }}>Ganti Password</h2>
        </div>

        {mutation.isSuccess && (
          <div className="flex items-center gap-2 p-3 rounded-xl bg-green-50 border border-green-200 mb-5">
            <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0" />
            <p className="text-sm text-green-700">Password berhasil diubah!</p>
          </div>
        )}

        {mutation.isError && (
          <div className="p-3 rounded-xl bg-red-50 border border-red-200 mb-5">
            <p className="text-sm text-red-600">{(mutation.error as Error)?.message ?? 'Gagal mengubah password'}</p>
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Field label="Password Saat Ini" error={errors.current_password?.message}>
            <Input {...register('current_password')} type="password" placeholder="Masukkan password saat ini" />
          </Field>
          <Field label="Password Baru" error={errors.new_password?.message}>
            <Input {...register('new_password')} type="password" placeholder="Min. 8 karakter" />
          </Field>
          <Field label="Konfirmasi Password Baru" error={errors.confirm_password?.message}>
            <Input {...register('confirm_password')} type="password" placeholder="Ulangi password baru" />
          </Field>
          <div className="pt-2">
            <Button type="submit" disabled={mutation.isPending} className="text-white" style={{ background: '#C8102E' }}>
              {mutation.isPending ? 'Menyimpan...' : 'Ubah Password'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}