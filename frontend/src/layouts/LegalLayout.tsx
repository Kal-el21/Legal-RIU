import { Scale, LayoutDashboard, FileText, FileSearch, ScrollText, BarChart3, Files } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { dashboardService } from '@/services/dashboard.service'
import PanelLayout, { type MenuItem } from '@/layouts/PanelLayout'

const NAV: MenuItem[] = [
  { label: 'Dashboard', href: '/legal', icon: LayoutDashboard, exact: true, permissions: ['dashboard.legal.view'] },
  { label: 'Laporan', href: '/legal/reports', icon: BarChart3, permissions: ['report.legal_case.view', 'report.legal_opinion.view', 'report.document_review.view'] },
  { label: 'Legal Opinion', href: '/legal/legal-opinions', icon: FileText, permissions: ['legal_opinion.view.all'] },
  { label: 'Review Dokumen', href: '/legal/review-documents', icon: FileSearch, permissions: ['document_review.view.all'] },
  { label: 'Dokumen Perjanjian', href: '/legal/agreement-documents', icon: FileText, permissions: ['agreement_document.view.all'] },
  { label: 'Dokumen Repository', href: '/legal/repository-documents', icon: Files, permissions: ['document_repository.view'] },
  { label: 'Manajemen Kasus', href: '/legal/legal-cases', icon: Scale, permissions: ['case_management.view'] },
  { label: 'Audit Log', href: '/legal/audit-logs', icon: ScrollText, permissions: ['audit_log.view'] },
]

export default function LegalLayout() {
  const { data: reminders } = useQuery({
    queryKey: ['dashboard', 'legal', 'reminders'],
    queryFn: () => dashboardService.getReminders(),
  })

  const reminderCount = (reminders?.yellow?.length ?? 0) + (reminders?.red?.length ?? 0)

  // Jumlah pengingat ditempel di ikon Dashboard
  const nav = NAV.map((item) => (item.exact ? { ...item, badge: reminderCount } : item))

  return <PanelLayout panelLabel="Legal Panel" settingsPath="/legal/settings" nav={nav} />
}
