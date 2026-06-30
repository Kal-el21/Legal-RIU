import api from './api'
import type { PaginatedData, AuditLog } from '@/types'

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
    return api.get('/admin/audit-logs', { params }).then(res => res.data.data)
  },
}
