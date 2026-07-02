import api from './api'
import { useAuthStore } from '@/store/auth.store'
import type { PaginatedData, AuditLog } from '@/types'

function getAuditLogEndpoint() {
  const role = useAuthStore.getState().user?.role
  const pathname = typeof window === 'undefined' ? '' : window.location.pathname
  return role === 'LEGAL' || pathname.startsWith('/legal') ? '/legal/audit-logs' : '/admin/audit-logs'
}

export const auditLogService = {
  getAll: (params: {
    action?: string
    entity_type?: string
    user_id?: string
    date_from?: string
    date_to?: string
    search?: string
    page: number
    limit: number
  }): Promise<PaginatedData<AuditLog>> => {
    return api.get(getAuditLogEndpoint(), { params }).then(res => res.data.data)
  },
}
