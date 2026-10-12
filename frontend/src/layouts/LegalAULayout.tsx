import { Scale, LayoutDashboard, FileText, FileSearch, ScrollText, BarChart3, Users, Files, BookOpen } from 'lucide-react'
import PanelLayout, { type MenuItem } from '@/layouts/PanelLayout'

const NAV: MenuItem[] = [
  { label: 'Dashboard', href: '/legal-au', icon: LayoutDashboard, exact: true, permissions: ['dashboard.legal.view'] },
  { label: 'Laporan', href: '/legal-au/reports', icon: BarChart3, permissions: ['report.legal_case.view', 'report.legal_opinion.view', 'report.document_review.view'] },
  { label: 'Legal Opinion', href: '/legal-au/legal-opinions', icon: FileText, permissions: ['legal_opinion.view.all', 'legal_opinion.view.own'] },
  { label: 'Review Dokumen', href: '/legal-au/review-documents', icon: FileSearch, permissions: ['document_review.view.all', 'document_review.view.own'] },
  { label: 'Dokumen Repository', href: '/legal-au/repository-documents', icon: Files, permissions: ['document_repository.view'] },
  { label: 'Manajemen Kasus', href: '/legal-au/cases', icon: Scale, permissions: ['case_management.view'] },
  { label: 'Materi Legal', href: '/legal-au/materials', icon: BookOpen, permissions: ['legal_material.view'] },
  { label: 'Manajemen Pengguna', href: '/legal-au/users', icon: Users, permissions: ['user_management.view'] },
  { label: 'Audit Log', href: '/legal-au/audit-logs', icon: ScrollText, permissions: ['audit_log.view'] },
]

export default function LegalAULayout() {
  return <PanelLayout panelLabel="Legal AU Panel" settingsPath="/legal-au/settings" nav={NAV} />
}
