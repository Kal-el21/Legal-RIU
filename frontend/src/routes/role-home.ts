import type { UserRole } from '@/types'

export function getRoleHome(role?: UserRole) {
  if (role === 'ADMIN') return '/admin'
  if (role === 'LEGAL') return '/legal'
  if (role === 'LEGAL_AU') return '/legal-au/cases'
  if (role === 'EXTERNAL') return '/external/legal-cases'
  return '/dashboard'
}
