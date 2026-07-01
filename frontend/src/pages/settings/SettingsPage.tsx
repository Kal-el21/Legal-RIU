import { useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { User, Bell, Shield, CheckCircle, Eye, EyeOff, Lock, Mail, ToggleLeft, ToggleRight, Settings, AlertTriangle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useDivisions } from '@/hooks/useLegalCase'
import { useAuthStore } from '@/store/auth.store'
import { settingsService } from '@/services/auth.service'
import { authService } from '@/services/auth.service'
import { notificationSettingService } from '@/services/notification-setting.service'
import type { NotificationSetting } from '@/types'
import { cn } from '@/lib/utils'

// ── Schemas ───────────────────────────────────────────────────────────────────

const profileSchema = z.object({
  full_name: z.string().min(1, 'Wajib diisi'),
  position: z.string().min(1, 'Wajib diisi'),
  division: z.string().min(1, 'Wajib diisi'),
})

const passwordSchema = z.object({
  current_password: z.string().min(1, 'Wajib diisi'),
  new_password: z.string().min(8, 'Minimal 8 karakter'),
  confirm_password: z.string(),
}).refine((d) => d.new_password === d.confirm_password, {
  message: 'Password tidak cocok', path: ['confirm_password'],
})

const twoFASchema = z.object({
  password: z.string().min(1, 'Masukkan password untuk konfirmasi'),
})

type ProfileForm = z.infer<typeof profileSchema>
type PasswordForm = z.infer<typeof passwordSchema>
type TwoFAForm = z.infer<typeof twoFASchema>

// ── Sub-components ────────────────────────────────────────────────────────────

function Field({ label, error, hint, children }: {
  label: string; error?: string; hint?: string; children: React.ReactNode
}) {
  return (
    <div className="space-y-1.5">
      <Label className="text-sm font-medium text-gray-700">{label}</Label>
      {children}
      {hint && !error && <p className="text-xs text-gray-400">{hint}</p>}
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}

function SuccessAlert({ message }: { message: string }) {
  return (
    <div className="flex items-center gap-2 p-3 rounded-xl bg-green-50 border border-green-200">
      <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0" />
      <p className="text-sm text-green-700">{message}</p>
    </div>
  )
}

function ErrorAlert({ message }: { message: string }) {
  return (
    <div className="p-3 rounded-xl bg-red-50 border border-red-200">
      <p className="text-sm text-red-600">{message}</p>
    </div>
  )
}

// ── Profile Tab ───────────────────────────────────────────────────────────────

function ProfileTab() {
  const { user, updateUser } = useAuthStore()
  const { data: divisions = [] } = useDivisions()

  const { control, register, handleSubmit, formState: { errors } } = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      full_name: user?.full_name ?? '',
      position: user?.position ?? '',
      division: user?.division_id || user?.division || '',
    },
  })
  const divisionOptions = divisions.map((division) => ({ value: division.id, label: division.name }))

  const mutation = useMutation({
    mutationFn: (data: ProfileForm) => settingsService.updateProfile(data),
    onSuccess: (updated) => { if (updated) updateUser(updated) },
  })

  return (
    <div className="space-y-6">
      {/* Avatar section */}
      <div className="flex items-center gap-4 p-5 rounded-2xl bg-gray-50 border border-gray-100">
        <div className="w-16 h-16 rounded-2xl flex items-center justify-center text-white text-2xl font-bold flex-shrink-0"
          style={{ background: user?.role === 'ADMIN' ? '#C8102E' : '#0B2545' }}>
          {user?.full_name?.charAt(0).toUpperCase()}
        </div>
        <div>
          <p className="font-semibold text-gray-900">{user?.full_name}</p>
          <p className="text-sm text-gray-500">{user?.email}</p>
          <span className={`inline-flex items-center mt-1 px-2 py-0.5 rounded-full text-xs font-medium ${
            user?.role === 'ADMIN' ? 'bg-red-100 text-red-700' : 'bg-blue-100 text-blue-700'
          }`}>{user?.role}</span>
        </div>
      </div>

      <form onSubmit={handleSubmit((d) => mutation.mutate(d))} className="space-y-4">
        <Field label="Nama Lengkap" error={errors.full_name?.message}>
          <Input {...register('full_name')} placeholder="Nama lengkap Anda" />
        </Field>
        <Field label="Email Kantor" hint="Email tidak dapat diubah — hubungi administrator">
          <Input value={user?.email ?? ''} disabled className="bg-gray-50 text-gray-500" />
        </Field>
        <div className="grid grid-cols-2 gap-4">
          <Field label="Posisi Jabatan" error={errors.position?.message}>
            <Input {...register('position')} placeholder="Jabatan Anda" />
          </Field>
          <Field label="Divisi" error={errors.division?.message}>
            <Controller
              name="division"
              control={control}
              render={({ field }) => (
                <Select onValueChange={field.onChange} value={field.value}>
                  <SelectTrigger><SelectValue placeholder="Pilih divisi" /></SelectTrigger>
                  <SelectContent>
                    {field.value && !divisionOptions.some((division) => division.value === field.value) && (
                      <SelectItem value={field.value}>{user?.division ?? field.value}</SelectItem>
                    )}
                    {divisionOptions.map((division) => (
                      <SelectItem key={division.value} value={division.value}>{division.label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
          </Field>
        </div>

        {mutation.isSuccess && <SuccessAlert message="Profil berhasil diperbarui!" />}
        {mutation.isError && <ErrorAlert message={(mutation.error as Error)?.message ?? 'Gagal memperbarui profil'} />}

        <Button type="submit" disabled={mutation.isPending} className="text-white" style={{ background: '#C8102E' }}>
          {mutation.isPending ? 'Menyimpan...' : 'Simpan Perubahan'}
        </Button>
      </form>
    </div>
  )
}

// ── Notification Tab ──────────────────────────────────────────────────────────

function NotificationTab() {
  const { user, updateUser } = useAuthStore()
  const [emailEnabled, setEmailEnabled] = useState<boolean>(user?.email_notifications ?? true)

  const mutation = useMutation({
    mutationFn: (val: boolean) => settingsService.updateNotifications(val),
    onSuccess: (_, val) => {
      setEmailEnabled(val)
      updateUser({ ...user!, email_notifications: val })
    },
  })

  const toggle = () => mutation.mutate(!emailEnabled)

  return (
    <div className="space-y-5">
      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-1">Preferensi Notifikasi</h3>
        <p className="text-sm text-gray-500">Atur bagaimana Anda ingin menerima notifikasi dari sistem.</p>
      </div>

      {/* Email notification toggle */}
      <div className="bg-white rounded-2xl border border-gray-100 p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0" style={{ background: '#EFF6FF' }}>
              <Mail className="w-5 h-5" style={{ color: '#0B2545' }} />
            </div>
            <div>
              <p className="text-sm font-semibold text-gray-900">Notifikasi Email</p>
              <p className="text-xs text-gray-500 mt-0.5 max-w-sm">
                Terima notifikasi melalui email saat status pengajuan berubah, ada catatan dari admin, atau pengajuan selesai.
              </p>
            </div>
          </div>
          <button onClick={toggle} disabled={mutation.isPending}
            className="flex-shrink-0 transition-colors mt-0.5">
            {emailEnabled
              ? <ToggleRight className="w-8 h-8" style={{ color: '#C8102E' }} />
              : <ToggleLeft className="w-8 h-8 text-gray-400" />}
          </button>
        </div>
        <div className={`mt-3 text-xs px-3 py-2 rounded-lg inline-block ${
          emailEnabled ? 'bg-green-50 text-green-700' : 'bg-gray-100 text-gray-500'
        }`}>
          {emailEnabled ? '✓ Notifikasi email aktif' : '✗ Notifikasi email tidak aktif'}
        </div>
      </div>

      {/* Info card */}
      <div className="p-4 rounded-xl bg-blue-50 border border-blue-100">
        <p className="text-xs text-blue-700">
          <span className="font-semibold">Catatan:</span> Fitur pengiriman email akan aktif pada Phase 2. Saat ini preferensi Anda tersimpan dan akan langsung berlaku ketika fitur email diaktifkan.
        </p>
      </div>

      {mutation.isSuccess && <SuccessAlert message="Preferensi notifikasi disimpan!" />}
    </div>
  )
}

// ── Security Tab ──────────────────────────────────────────────────────────────

function SecurityTab() {
  const { user, updateUser } = useAuthStore()
  const [showCurrentPw, setShowCurrentPw] = useState(false)
  const [showNewPw, setShowNewPw] = useState(false)
  const [twoFAEnabled, setTwoFAEnabled] = useState<boolean>(user?.two_fa_enabled ?? false)
  const [showTwoFAConfirm, setShowTwoFAConfirm] = useState(false)
  const [pendingTwoFA, setPendingTwoFA] = useState<boolean | null>(null)

  const pwForm = useForm<PasswordForm>({ resolver: zodResolver(passwordSchema) })
  const twoFAForm = useForm<TwoFAForm>({ resolver: zodResolver(twoFASchema) })

  const pwMutation = useMutation({
    mutationFn: (data: PasswordForm) =>
      authService.changePassword({ current_password: data.current_password, new_password: data.new_password }),
    onSuccess: () => pwForm.reset(),
  })

  const twoFAMutation = useMutation({
    mutationFn: ({ enabled, password }: { enabled: boolean; password: string }) =>
      settingsService.toggle2FA(enabled, password),
    onSuccess: () => {
      const newVal = pendingTwoFA ?? !twoFAEnabled
      setTwoFAEnabled(newVal)
      setShowTwoFAConfirm(false)
      twoFAForm.reset()
      setPendingTwoFA(null)
      updateUser({ ...user!, two_fa_enabled: newVal })
    },
  })

  const initToggle2FA = () => {
    setPendingTwoFA(!twoFAEnabled)
    twoFAForm.reset()
    setShowTwoFAConfirm(true)
  }

  return (
    <div className="space-y-6">
      {/* Change password */}
      <div className="bg-white rounded-2xl border border-gray-100 p-5">
        <div className="flex items-center gap-3 mb-5">
          <div className="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0" style={{ background: '#FEF2F2' }}>
            <Lock className="w-5 h-5" style={{ color: '#C8102E' }} />
          </div>
          <div>
            <p className="text-sm font-semibold text-gray-900">Ganti Password</p>
            <p className="text-xs text-gray-500">Gunakan password yang kuat dan unik</p>
          </div>
        </div>

        <form onSubmit={pwForm.handleSubmit((d) => pwMutation.mutate(d))} className="space-y-4">
          <Field label="Password Saat Ini" error={pwForm.formState.errors.current_password?.message}>
            <div className="relative">
              <Input {...pwForm.register('current_password')} type={showCurrentPw ? 'text' : 'password'} placeholder="••••••••" className="pr-10" />
              <button type="button" onClick={() => setShowCurrentPw(!showCurrentPw)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600">
                {showCurrentPw ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </Field>
          <Field label="Password Baru" error={pwForm.formState.errors.new_password?.message}
            hint="Minimal 8 karakter">
            <div className="relative">
              <Input {...pwForm.register('new_password')} type={showNewPw ? 'text' : 'password'} placeholder="••••••••" className="pr-10" />
              <button type="button" onClick={() => setShowNewPw(!showNewPw)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600">
                {showNewPw ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </Field>
          <Field label="Konfirmasi Password Baru" error={pwForm.formState.errors.confirm_password?.message}>
            <Input {...pwForm.register('confirm_password')} type="password" placeholder="••••••••" />
          </Field>

          {pwMutation.isSuccess && <SuccessAlert message="Password berhasil diubah!" />}
          {pwMutation.isError && <ErrorAlert message={(pwMutation.error as Error)?.message ?? 'Gagal mengubah password'} />}

          <Button type="submit" disabled={pwMutation.isPending} className="text-white" style={{ background: '#C8102E' }}>
            {pwMutation.isPending ? 'Menyimpan...' : 'Ubah Password'}
          </Button>
        </form>
      </div>

      {/* 2FA */}
      <div className="bg-white rounded-2xl border border-gray-100 p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0"
              style={{ background: twoFAEnabled ? '#F0FDF4' : '#F8FAFC' }}>
              <Shield className="w-5 h-5" style={{ color: twoFAEnabled ? '#16A34A' : '#94A3B8' }} />
            </div>
            <div>
              <p className="text-sm font-semibold text-gray-900">Two-Step Login (2FA)</p>
              <p className="text-xs text-gray-500 mt-0.5 max-w-sm">
                Tambahkan lapisan keamanan ekstra. Setiap login akan membutuhkan kode OTP yang dikirim ke email Anda.
              </p>
              <div className={`mt-2 text-xs px-2.5 py-1 rounded-full inline-block font-medium ${
                twoFAEnabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
              }`}>
                {twoFAEnabled ? '✓ Aktif' : '✗ Tidak aktif'}
              </div>
            </div>
          </div>
          <button onClick={initToggle2FA} className="flex-shrink-0 mt-0.5 transition-colors">
            {twoFAEnabled
              ? <ToggleRight className="w-8 h-8" style={{ color: '#16A34A' }} />
              : <ToggleLeft className="w-8 h-8 text-gray-300" />}
          </button>
        </div>

        {/* Confirm with password */}
        {showTwoFAConfirm && (
          <div className="mt-5 pt-5 border-t border-gray-100">
            <p className="text-sm text-gray-700 mb-4">
              Masukkan password Anda untuk{' '}
              <strong>{pendingTwoFA ? 'mengaktifkan' : 'menonaktifkan'}</strong> Two-Step Login.
            </p>
            <form onSubmit={twoFAForm.handleSubmit((d) =>
              twoFAMutation.mutate({ enabled: pendingTwoFA!, password: d.password })
            )} className="space-y-3">
              <Field label="Password" error={twoFAForm.formState.errors.password?.message}>
                <Input {...twoFAForm.register('password')} type="password" placeholder="Masukkan password Anda" />
              </Field>
              {twoFAMutation.isError && <ErrorAlert message={(twoFAMutation.error as Error)?.message ?? 'Gagal'} />}
              {twoFAMutation.isSuccess && <SuccessAlert message={`2FA berhasil ${pendingTwoFA ? 'diaktifkan' : 'dinonaktifkan'}!`} />}
              <div className="flex gap-2">
                <Button type="button" variant="outline" size="sm" onClick={() => setShowTwoFAConfirm(false)}>Batal</Button>
                <Button type="submit" size="sm" disabled={twoFAMutation.isPending}
                  className="text-white" style={{ background: pendingTwoFA ? '#16A34A' : '#C8102E' }}>
                  {twoFAMutation.isPending ? 'Memproses...' : `${pendingTwoFA ? 'Aktifkan' : 'Nonaktifkan'} 2FA`}
                </Button>
              </div>
            </form>
          </div>
        )}
      </div>
    </div>
  )
}

// ── Main Settings Page ────────────────────────────────────────────────────────

const TABS = [
  { id: 'profile', label: 'Profil', icon: User },
  { id: 'notifications', label: 'Notifikasi', icon: Bell },
  { id: 'security', label: 'Keamanan', icon: Shield },
]

const ADMIN_TABS = [
  { id: 'profile', label: 'Profil', icon: User },
  { id: 'notifications', label: 'Notifikasi', icon: Bell },
  { id: 'security', label: 'Keamanan', icon: Shield },
  { id: 'notification-settings', label: 'Konfigurasi Notifikasi', icon: Settings },
]

function AdminNotificationSettingsTab() {
  const { user } = useAuthStore()
  const queryClient = useQueryClient()
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editValue, setEditValue] = useState<number>(0)
  const [editActive, setEditActive] = useState<boolean>(true)

  const { data: settings = [], isLoading } = useQuery({
    queryKey: ['notification-settings'],
    queryFn: notificationSettingService.getAll,
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, days_threshold, is_active }: { id: string; days_threshold: number; is_active?: boolean }) =>
      notificationSettingService.update(id, { days_threshold, is_active }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notification-settings'] })
      setEditingId(null)
    },
  })

  const startEdit = (setting: NotificationSetting) => {
    setEditingId(setting.id)
    setEditValue(setting.days_threshold)
    setEditActive(setting.is_active)
  }

  const saveEdit = (id: string) => {
    if (editValue < 1) return
    updateMutation.mutate({ id, days_threshold: editValue, is_active: editActive })
  }

  const grouped = settings.reduce<Record<string, NotificationSetting[]>>((acc, setting) => {
    if (!acc[setting.submission_type]) acc[setting.submission_type] = []
    acc[setting.submission_type].push(setting)
    return acc
  }, {})

  const label = (type: string) => {
    if (type === 'legal_opinion') return 'Legal Opinion'
    if (type === 'document_review') return 'Document Review'
    return type
  }

  if (user?.role !== 'ADMIN') return null

  return (
    <div className="space-y-5">
      <div>
        <h3 className="text-sm font-semibold text-gray-900 mb-1">Konfigurasi Notifikasi</h3>
        <p className="text-sm text-gray-500">Atur durasi threshold untuk setiap level notifikasi. Perubahan langsung berdampak ke seluruh sistem.</p>
      </div>

      {isLoading ? (
        <p className="text-sm text-gray-500">Memuat konfigurasi...</p>
      ) : (
        <div className="space-y-4">
          {Object.entries(grouped).map(([type, items]) => (
            <div key={type} className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
              <div className="px-5 py-3 border-b border-gray-50">
                <p className="text-sm font-semibold text-gray-900">{label(type)}</p>
              </div>
              <div className="divide-y divide-gray-50">
                {items.map((setting) => (
                  <div key={setting.id} className="px-5 py-3 flex items-center justify-between gap-4">
                    <div className="flex items-center gap-3">
                      <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${
                        setting.warning_level === 'YELLOW' ? 'bg-amber-100' : 'bg-red-100'
                      }`}>
                        <AlertTriangle className={`w-4 h-4 ${
                          setting.warning_level === 'YELLOW' ? 'text-amber-600' : 'text-red-600'
                        }`} />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-900">
                          {setting.warning_level === 'YELLOW' ? 'Peringatan Kuning' : 'Peringatan Merah'}
                        </p>
                        <p className="text-xs text-gray-500">
                          {setting.warning_level === 'YELLOW' ? 'Perlu perhatian' : 'Terlambat'} — aktif: {setting.is_active ? 'Ya' : 'Tidak'}
                        </p>
                      </div>
                    </div>

                    {editingId === setting.id ? (
                      <div className="flex items-center gap-2">
                        <Input
                          type="number"
                          min={1}
                          value={editValue}
                          onChange={(e) => setEditValue(parseInt(e.target.value || '0', 10))}
                          className="w-20 h-8 text-xs"
                        />
                        <span className="text-xs text-gray-500">hari</span>
                        <Button
                          size="sm"
                          disabled={editValue < 1 || updateMutation.isPending}
                          className="text-white text-xs h-8"
                          style={{ background: '#C8102E' }}
                          onClick={() => saveEdit(setting.id)}
                        >
                          Simpan
                        </Button>
                        <Button size="sm" variant="outline" className="text-xs h-8" onClick={() => setEditingId(null)}>
                          Batal
                        </Button>
                      </div>
                    ) : (
                      <div className="flex items-center gap-3">
                        <span className="text-sm font-semibold text-gray-900">{setting.days_threshold} hari</span>
                        <Button size="sm" variant="outline" className="text-xs h-8" onClick={() => startEdit(setting)}>
                          Edit
                        </Button>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {updateMutation.isSuccess && <SuccessAlert message="Konfigurasi notifikasi berhasil diperbarui!" />}
      {updateMutation.isError && <ErrorAlert message={(updateMutation.error as Error)?.message ?? 'Gagal memperbarui'} />}
    </div>
  )
}

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState('profile')
  const { user } = useAuthStore()

  const tabs = user?.role === 'ADMIN' ? ADMIN_TABS : TABS

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Pengaturan</h1>
        <p className="text-sm text-gray-500 mt-0.5">Kelola profil, notifikasi, dan keamanan akun Anda</p>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 p-1 rounded-xl bg-gray-100 mb-6 overflow-x-auto">
        {tabs.map((tab) => (
          <button key={tab.id} onClick={() => setActiveTab(tab.id)}
            className={cn(
              'flex-1 flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all whitespace-nowrap',
              activeTab === tab.id
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-500 hover:text-gray-700'
            )}>
            <tab.icon className="w-4 h-4" />
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div>
        {activeTab === 'profile' && <ProfileTab />}
        {activeTab === 'notifications' && <NotificationTab />}
        {activeTab === 'security' && <SecurityTab />}
        {activeTab === 'notification-settings' && <AdminNotificationSettingsTab />}
      </div>
    </div>
  )
}
