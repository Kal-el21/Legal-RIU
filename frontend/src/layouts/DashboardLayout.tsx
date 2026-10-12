import { Scale, LayoutDashboard, FileText, FileSearch, ScrollText, Files, Handshake } from 'lucide-react'
import PanelLayout, { type MenuItem } from '@/layouts/PanelLayout'

const NAV: MenuItem[] = [
  { label: 'Dashboard', href: '/dashboard', icon: LayoutDashboard, exact: true, permissions: ['dashboard.user.view'] },
  { label: 'Legal Opinion', href: '/dashboard/legal-opinions', icon: FileText, permissions: ['legal_opinion.view.own', 'legal_opinion.view.all'] },
  { label: 'Review Dokumen', href: '/dashboard/review-documents', icon: FileSearch, permissions: ['document_review.view.own', 'document_review.view.all'] },
  { label: 'Dokumen Perjanjian', href: '/dashboard/agreement-documents', icon: Handshake, permissions: ['agreement_document.view.own'] },
  { label: 'Dokumen Repository', href: '/dashboard/repository-documents', icon: Files, permissions: ['document_repository.view'] },
  { label: 'Manajemen Kasus', href: '/dashboard/legal-cases', icon: Scale, permissions: ['case_management.view'] },
  { label: 'Audit Log', href: '/dashboard/audit-logs', icon: ScrollText, permissions: ['audit_log.view'] },
]

export default function DashboardLayout() {
  return <PanelLayout panelLabel="Indonesia Re" settingsPath="/dashboard/settings" nav={NAV} />
}
