import api from './api'
import { useAuthStore } from '@/store/auth.store'
import type { ApiResponse, ReportChartResponse, ReportFeature } from '@/types'

function getReportEndpoint(feature: ReportFeature): string {
  const role = useAuthStore.getState().user?.role
  const pathname = typeof window === 'undefined' ? '' : window.location.pathname
  if (role === 'ADMIN' || pathname.startsWith('/admin')) return `/admin/reports/${feature}`
  if (role === 'LEGAL' || pathname.startsWith('/legal')) return `/legal/reports/${feature}`
  return `/reports/${feature}`
}

export const reportService = {
  getReport: (feature: ReportFeature, groupBy: string, dateFrom?: string, dateTo?: string) =>
    api.get<ApiResponse<ReportChartResponse>>(getReportEndpoint(feature), { params: { group_by: groupBy, date_from: dateFrom, date_to: dateTo } }).then(res => res.data.data!)
}
