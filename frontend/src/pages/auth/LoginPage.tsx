import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Eye, EyeOff, Scale } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useLogin } from '@/hooks/useAuth'

const loginSchema = z.object({
  email: z.string().email('Email tidak valid'),
  password: z.string().min(1, 'Password wajib diisi'),
})

type LoginForm = z.infer<typeof loginSchema>

export default function LoginPage() {
  const [showPassword, setShowPassword] = useState(false)
  const login = useLogin()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = (data: LoginForm) => {
    login.mutate(data)
  }

  return (
    <div className="min-h-screen flex">
      {/* Left panel — branding */}
      <div
        className="hidden lg:flex lg:w-1/2 flex-col justify-between p-12 relative overflow-hidden"
        style={{ background: 'linear-gradient(135deg, #0B2545 0%, #1A3A6B 60%, #C8102E 100%)' }}
      >
        {/* Background pattern */}
        <div className="absolute inset-0 opacity-10">
          <div className="absolute top-20 left-20 w-64 h-64 rounded-full border border-white" />
          <div className="absolute top-40 left-40 w-96 h-96 rounded-full border border-white" />
          <div className="absolute bottom-20 right-20 w-48 h-48 rounded-full border border-white" />
        </div>

        <div className="relative">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-xl bg-white/20 flex items-center justify-center">
              <Scale className="w-5 h-5 text-white" />
            </div>
            <span className="text-white font-semibold text-lg">Legal RIU</span>
          </div>
          <p className="text-white/60 text-sm">Indonesia Re</p>
        </div>

        <div className="relative">
          <h1 className="text-4xl font-bold text-white leading-tight mb-4">
            Portal Layanan<br />
            Hukum Digital
          </h1>
          <p className="text-white/70 text-base leading-relaxed max-w-sm">
            Kelola pengajuan Legal Opinion dan Review Dokumen dengan mudah dan efisien.
          </p>

          <div className="mt-10 grid grid-cols-3 gap-4">
            {[
              { label: 'Legal Opinion', desc: 'Kajian hukum profesional' },
              { label: 'Review Dokumen', desc: 'Tinjauan dokumen legal' },
              { label: 'Database Legal', desc: 'Regulasi terkini' },
            ].map((item) => (
              <div key={item.label} className="bg-white/10 rounded-xl p-4 backdrop-blur-sm">
                <p className="text-white text-sm font-medium">{item.label}</p>
                <p className="text-white/60 text-xs mt-1">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>

        <p className="relative text-white/40 text-xs">
          © 2025 Legal RIU — Indonesia Re. All rights reserved.
        </p>
      </div>

      {/* Right panel — form */}
      <div className="flex-1 flex items-center justify-center p-6 bg-white">
        <div className="w-full max-w-sm">
          {/* Mobile logo */}
          <div className="flex items-center gap-2 mb-8 lg:hidden">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: '#C8102E' }}>
              <Scale className="w-4 h-4 text-white" />
            </div>
            <span className="font-semibold text-gray-900">Legal RIU</span>
          </div>

          <div className="mb-8">
            <h2 className="text-2xl font-bold text-gray-900">Selamat Datang</h2>
            <p className="text-gray-500 text-sm mt-1">Masuk ke akun Legal RIU Anda</p>
          </div>

          {/* Error alert */}
          {login.isError && (
            <div className="mb-5 p-3 rounded-lg bg-red-50 border border-red-200">
              <p className="text-red-600 text-sm">
                {(login.error as Error)?.message ?? 'Login gagal, coba lagi'}
              </p>
            </div>
          )}

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
            <div className="space-y-1.5">
              <Label htmlFor="email" className="text-gray-700 text-sm font-medium">
                Email Kantor
              </Label>
              <Input
                id="email"
                type="email"
                placeholder="nama@indonesiare.co.id"
                autoComplete="email"
                {...register('email')}
                className={errors.email ? 'border-red-400 focus-visible:ring-red-400' : ''}
              />
              {errors.email && (
                <p className="text-red-500 text-xs">{errors.email.message}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="password" className="text-gray-700 text-sm font-medium">
                Password
              </Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  placeholder="Masukkan password"
                  autoComplete="current-password"
                  {...register('password')}
                  className={`pr-10 ${errors.password ? 'border-red-400 focus-visible:ring-red-400' : ''}`}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
              {errors.password && (
                <p className="text-red-500 text-xs">{errors.password.message}</p>
              )}
            </div>

            <Button
              type="submit"
              disabled={login.isPending}
              className="w-full text-white font-medium h-10"
              style={{ background: login.isPending ? '#999' : '#C8102E' }}
            >
              {login.isPending ? 'Memproses...' : 'Masuk'}
            </Button>
          </form>

          <p className="text-center text-xs text-gray-400 mt-8">
            Belum punya akun? Hubungi administrator Legal RIU.
          </p>
        </div>
      </div>
    </div>
  )
}