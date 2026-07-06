import { useQuery } from '@tanstack/react-query'
import { reportService } from '@/services/report.service'
import { useAuthStore } from '@/store/auth.store'
import type { ReportFeature, ReportGroupBy } from '@/types'

const PERMISSIONS: Record<ReportFeature, string> = {
  'legal-cases': 'report.legal_case.view',
  'legal-opinions': 'report.legal_opinion.view',
  'document-reviews': 'report.document_review.view',
}

export function useReport(feature: ReportFeature, groupBy: ReportGroupBy, dateFrom?: string, dateTo?: string) {
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const permission = PERMISSIONS[feature]

  return useQuery({
    queryKey: ['report', feature, groupBy, dateFrom, dateTo],
    queryFn: () => reportService.getReport(feature, groupBy, dateFrom, dateTo),
    enabled: !!feature && !!groupBy && hasPermission(permission),
  })
}
