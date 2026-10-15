import { Scale, LayoutDashboard, FileText, FileSearch, Users, ScrollText, Building2, Target, Briefcase, MapPin, UserCog, FolderOpen, BarChart3, Files, Handshake, BookOpen, FileType, LayoutTemplate, Network } from 'lucide-react'
import PanelLayout, { type MenuItem, type MenuGroup } from '@/layouts/PanelLayout'

const NAV: (MenuItem | MenuGroup)[] = [
  { label: 'Dashboard', href: '/admin', icon: LayoutDashboard, exact: true, permissions: ['dashboard.admin.view'] },
  { label: 'Laporan', href: '/admin/reports', icon: BarChart3, permissions: ['report.legal_case.view', 'report.legal_opinion.view', 'report.document_review.view'] },
  { label: 'Legal Opinion', href: '/admin/legal-opinions', icon: FileText, permissions: ['legal_opinion.view.all'] },
  { label: 'Review Dokumen', href: '/admin/review-documents', icon: FileSearch, permissions: ['document_review.view.all'] },
  { label: 'Dokumen Perjanjian', href: '/admin/agreement-documents', icon: Handshake, permissions: ['agreement_document.view.all'] },
  { label: 'Dokumen Repository', href: '/admin/repository-documents', icon: Files, permissions: ['document_repository.view'] },
  { label: 'Manajemen Kasus', href: '/admin/legal-cases', icon: Scale, permissions: ['case_management.view'] },
  { label: 'Audit Log', href: '/admin/audit-logs', icon: ScrollText, permissions: ['audit_log.view'] },
  {
    label: 'Data Master',
    icon: FolderOpen,
    children: [
      { label: 'Manajemen Pengguna', href: '/admin/users', icon: Users, permissions: ['user_management.view'] },
      { label: 'Perusahaan', href: '/admin/companies', icon: Building2, permissions: ['master_data.manage'] },
      { label: 'Tujuan Pembuatan', href: '/admin/purpose-types', icon: Target, permissions: ['master_data.manage'] },
      { label: 'Jenis Dokumen', href: '/admin/document-types', icon: FileType, permissions: ['master_data.manage'] },
      { label: 'Jenis Kasus', href: '/admin/case-types', icon: Briefcase, permissions: ['master_data.manage'] },
      { label: 'Kategori', href: '/admin/case-categories', icon: FolderOpen, permissions: ['master_data.manage'] },
      { label: 'Kabupaten/Kota', href: '/admin/regencies', icon: MapPin, permissions: ['master_data.manage'] },
      { label: 'Cedant', href: '/admin/cedants', icon: UserCog, permissions: ['master_data.manage'] },
      { label: 'Divisi', href: '/admin/divisions', icon: Network, permissions: ['master_data.manage'] },
      { label: 'Pihak Pertama', href: '/admin/agreement-company-master', icon: Building2, permissions: ['agreement_company_master.view.all'] },
      { label: 'Template Perjanjian', href: '/admin/agreement-templates', icon: LayoutTemplate, permissions: ['agreement_template.view.all'] },
    ],
  },
  { label: 'Materi Legal', href: '/admin/materials', icon: BookOpen, permissions: ['legal_material.view'] },
]

export default function AdminLayout() {
  return <PanelLayout panelLabel="Admin Panel" settingsPath="/admin/settings" nav={NAV} />
}
