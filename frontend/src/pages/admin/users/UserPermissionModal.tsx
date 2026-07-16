import { useMemo, useState } from 'react'
import { RotateCcw, Save, ShieldCheck, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useUpdateUserPermissions, useUserPermissions } from '@/hooks/usePermission'
import { cn } from '@/lib/utils'
import type { Permission, PermissionEffect, User } from '@/types'

type OverrideState = 'INHERIT' | PermissionEffect

interface UserPermissionModalProps {
  user: User
  onClose: () => void
}

const FEATURE_LABELS: Record<string, string> = {
  dashboard: 'Dashboard',
  legal_opinion: 'Legal Opinion',
  document_review: 'Review Dokumen',
  case_management: 'Case Management',
  user_management: 'User Management',
  audit_log: 'Audit Log',
  master_data: 'Master Data',
  notification_setting: 'Notifikasi',
  legal_material: 'Materi Legal',
}

function groupPermissions(permissions: Permission[]) {
  return permissions.reduce<Record<string, Permission[]>>((acc, permission) => {
    acc[permission.feature] = acc[permission.feature] ?? []
    acc[permission.feature].push(permission)
    return acc
  }, {})
}

export default function UserPermissionModal({ user, onClose }: UserPermissionModalProps) {
  const { data: access, isLoading } = useUserPermissions(user.id)
  const updateMutation = useUpdateUserPermissions(user.id)
  const [overrideEdits, setOverrideEdits] = useState<Record<string, OverrideState>>({})

  const overrides = useMemo(() => {
    const base: Record<string, OverrideState> = {}
    if (!access) return base
    access.overrides.forEach((item) => {
      base[item.code] = item.effect
    })
    return { ...base, ...overrideEdits }
  }, [access, overrideEdits])

  const grouped = useMemo(() => groupPermissions(access?.permissions ?? []), [access?.permissions])
  const rolePermissionSet = useMemo(() => new Set(access?.role_permissions ?? []), [access?.role_permissions])
  const effectivePermissionSet = useMemo(() => new Set(access?.effective_permissions ?? []), [access?.effective_permissions])

  const setOverride = (code: string, effect: OverrideState) => {
    setOverrideEdits((prev) => {
      const next = { ...prev }
      if (effect === 'INHERIT') {
        delete next[code]
      } else {
        next[code] = effect
      }
      return next
    })
  }

  const handleSave = async () => {
    const payload = Object.entries(overrides).map(([code, effect]) => ({
      code,
      effect: effect as PermissionEffect,
    }))
    await updateMutation.mutateAsync(payload)
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40">
      <div className="bg-white rounded-xl shadow-2xl w-full max-w-5xl max-h-[90vh] flex flex-col">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <div className="min-w-0">
            <h2 className="text-base font-semibold truncate" style={{ color: '#0B2545' }}>
              Permission - {user.full_name}
            </h2>
            <p className="text-xs text-gray-500">{user.role} / {user.email}</p>
          </div>
          <button onClick={onClose} className="p-1.5 rounded-lg hover:bg-gray-100 transition-colors">
            <X className="w-4 h-4 text-gray-500" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-5">
          {isLoading ? (
            <div className="py-16 text-center text-sm text-gray-400">Memuat permission...</div>
          ) : (
            <div className="space-y-5">
              {Object.entries(grouped).map(([feature, permissions]) => (
                <section key={feature} className="border border-gray-100 rounded-lg overflow-hidden">
                  <div className="px-4 py-3 bg-gray-50 flex items-center justify-between">
                    <h3 className="text-sm font-semibold text-gray-700">
                      {FEATURE_LABELS[feature] ?? feature}
                    </h3>
                    <span className="text-xs text-gray-400">{permissions.length} akses</span>
                  </div>

                  <div className="divide-y divide-gray-50">
                    {permissions.map((permission) => {
                      const defaultAllowed = rolePermissionSet.has(permission.code)
                      const effectiveAllowed = effectivePermissionSet.has(permission.code)
                      const override = overrides[permission.code] ?? 'INHERIT'

                      return (
                        <div key={permission.code} className="grid grid-cols-[1fr_auto] gap-4 px-4 py-3 items-center">
                          <div className="min-w-0">
                            <div className="flex items-center gap-2 flex-wrap">
                              <p className="text-sm font-medium text-gray-800">{permission.label}</p>
                              <span className={cn(
                                'inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium',
                                effectiveAllowed ? 'bg-green-50 text-green-700' : 'bg-gray-100 text-gray-500'
                              )}>
                                {effectiveAllowed ? 'Aktif' : 'Tidak aktif'}
                              </span>
                              {defaultAllowed && (
                                <span className="inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium bg-blue-50 text-blue-700">
                                  Default role
                                </span>
                              )}
                            </div>
                            <p className="text-xs text-gray-400 mt-0.5">{permission.code}</p>
                          </div>

                          <div className="inline-flex rounded-lg border border-gray-200 overflow-hidden">
                            {(['INHERIT', 'ALLOW', 'DENY'] as OverrideState[]).map((effect) => (
                              <button
                                key={effect}
                                type="button"
                                onClick={() => setOverride(permission.code, effect)}
                                className={cn(
                                  'px-3 py-1.5 text-xs font-medium transition-colors border-l first:border-l-0 border-gray-200',
                                  override === effect
                                    ? effect === 'ALLOW'
                                      ? 'bg-green-600 text-white'
                                      : effect === 'DENY'
                                        ? 'bg-red-600 text-white'
                                        : 'bg-gray-800 text-white'
                                    : 'bg-white text-gray-500 hover:bg-gray-50'
                                )}
                              >
                                {effect === 'INHERIT' ? 'Default' : effect === 'ALLOW' ? 'Allow' : 'Deny'}
                              </button>
                            ))}
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </section>
              ))}
            </div>
          )}
        </div>

        <div className="flex items-center justify-between gap-3 px-6 py-4 border-t border-gray-100">
          <Button type="button" variant="outline" onClick={() => setOverrideEdits({})} className="gap-2">
            <RotateCcw className="w-4 h-4" />
            Reset Default
          </Button>
          <div className="flex gap-2">
            <Button type="button" variant="outline" onClick={onClose}>Batal</Button>
            <Button
              type="button"
              disabled={updateMutation.isPending}
              onClick={handleSave}
              className="gap-2 text-white"
              style={{ background: '#C8102E' }}
            >
              {updateMutation.isPending ? <ShieldCheck className="w-4 h-4 animate-pulse" /> : <Save className="w-4 h-4" />}
              Simpan
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
