import { Scale, FileText, FileSearch, ScrollText, BarChart3, Users, Files } from 'lucide-react'
import PanelLayout, { type MenuItem } from '@/layouts/PanelLayout'

const NAV: MenuItem[] = [
  { label: 'Laporan', href: '/external/reports', icon: BarChart3, permissions: ['report.legal_case.view', 'report.legal_opinion.view', 'report.document_review.view'] },
  { label: 'Legal Opinion', href: '/external/legal-opinions', icon: FileText, permissions: ['legal_opinion.view.own'] },
  { label: 'Review Dokumen', href: '/external/review-documents', icon: FileSearch, permissions: ['document_review.view.own'] },
  { label: 'Dokumen Repository', href: '/external/repository-documents', icon: Files, permissions: ['document_repository.view'] },
  { label: 'Manajemen Kasus', href: '/external/legal-cases', icon: Scale, permissions: ['case_management.view'] },
  { label: 'Materi Legal', href: '/external/materials', icon: FileText, permissions: ['legal_material.view'] },
  { label: 'Audit Log', href: '/external/audit-logs', icon: ScrollText, permissions: ['audit_log.view'] },
  { label: 'User Management', href: '/external/users', icon: Users, permissions: ['user_management.view'] },
]

export default function ExternalLayout() {
  return <PanelLayout panelLabel="External Panel" settingsPath="/external/settings" nav={NAV} />
}
