import { createBrowserRouter } from 'react-router-dom'
import { GuestRoute, PrivateRoute, UserRoute, AdminRoute, LegalRoute, LegalAURoute, ExternalRoute } from './guards'

import PublicLayout from '@/layouts/PublicLayout'
import DashboardLayout from '@/layouts/DashboardLayout'
import AdminLayout from '@/layouts/AdminLayout'
import LegalLayout from '@/layouts/LegalLayout'
import ExternalLayout from '@/layouts/ExternalLayout'
import LegalAULayout from '@/layouts/LegalAULayout'

import HomePage from '@/pages/public/HomePage'
import AktaPerusahaanPage from '@/pages/public/AktaPerusahaanPage'
import ComingSoonPage from '@/pages/public/ComingSoonPage'
import LoginPage from '@/pages/auth/LoginPage'

import UserDashboardPage from '@/pages/dashboard/UserDashboardPage'
import SettingsPage from '@/pages/settings/SettingsPage'
import NotificationListPage from '@/components/common/NotificationListPage'
import LegalOpinionListPage from '@/pages/dashboard/legal-opinions/LegalOpinionListPage'
import LegalOpinionFormPage from '@/pages/dashboard/legal-opinions/LegalOpinionFormPage'
import LegalOpinionDetailPage from '@/pages/dashboard/legal-opinions/LegalOpinionDetailPage'
import ReviewDocumentListPage from '@/pages/dashboard/review-documents/ReviewDocumentListPage'
import ReviewDocumentFormPage from '@/pages/dashboard/review-documents/ReviewDocumentFormPage'
import ReviewDocumentDetailPage from '@/pages/dashboard/review-documents/ReviewDocumentDetailPage'

import AdminDashboardPage from '@/pages/admin/AdminDashboardPage'
import UserManagementPage from '@/pages/admin/users/UserManagementPage'
import AdminLegalOpinionListPage from '@/pages/admin/legal-opinions/AdminLegalOpinionListPage'
import AdminLegalOpinionDetailPage from '@/pages/admin/legal-opinions/AdminLegalOpinionDetailPage'
import AdminReviewDocumentListPage from '@/pages/admin/review-documents/AdminReviewDocumentListPage'
import AdminReviewDocumentDetailPage from '@/pages/admin/review-documents/AdminReviewDocumentDetailPage'
import AdminLegalCaseListPage from '@/pages/admin/legal-cases/AdminLegalCaseListPage'
import AdminLegalCaseDetailPage from '@/pages/admin/legal-cases/AdminLegalCaseDetailPage'
import AuditLogPage from '@/pages/admin/AuditLogPage'

import CompanyManagementPage from '@/pages/admin/companies/CompanyManagementPage'
import PurposeTypeManagementPage from '@/pages/admin/purpose-types/PurposeTypeManagementPage'
import DocumentTypeManagementPage from '@/pages/admin/document-types/DocumentTypeManagementPage'
import CaseTypeManagementPage from '@/pages/admin/case-types/CaseTypeManagementPage'
import CaseCategoryManagementPage from '@/pages/admin/case-categories/CaseCategoryManagementPage'
import RegencyManagementPage from '@/pages/admin/regencies/RegencyManagementPage'
import CedantManagementPage from '@/pages/admin/cedants/CedantManagementPage'
import DivisionManagementPage from '@/pages/admin/divisions/DivisionManagementPage'

import MaterialManagementPage from '@/pages/admin/materials/MaterialManagementPage'
import DashboardMaterialManagementPage from '@/pages/dashboard/materials/LegalMaterialManagementPage'
import ExternalMaterialManagementPage from '@/pages/external/materials/LegalMaterialManagementPage'
import MaterialListingPage from '@/pages/public/MaterialListingPage'
import MaterialFormPage from '@/pages/materials/MaterialFormPage'
import ReportPage from '@/pages/reports/ReportPage'

import External_OpinionListPage from '@/pages/external/legal-opinions/External_OpinionListPage'
import External_ReviewDocumentListPage from '@/pages/external/review-documents/External_ReviewDocumentListPage'
import LegalDashboardPage from '@/pages/legal/LegalDashboardPage'
import LegalLegalOpinionListPage from '@/pages/legal/legal-opinions/LegalOpinionListPage'
import LegalLegalOpinionDetailPage from '@/pages/legal/legal-opinions/LegalOpinionDetailPage'
import LegalReviewDocumentListPage from '@/pages/legal/review-documents/ReviewDocumentListPage'
import LegalReviewDocumentDetailPage from '@/pages/legal/review-documents/ReviewDocumentDetailPage'
import LegalMaterialManagementPage from '@/pages/legal/materials/LegalMaterialManagementPage'
import LegalMaterialDetailPage from '@/pages/legal/materials/LegalMaterialDetailPage'
import AgreementDocumentList from '@/components/shared/AgreementDocumentList'
import AgreementDocumentForm from '@/components/shared/AgreementDocumentForm'
import AgreementDocumentDetail from '@/components/shared/AgreementDocumentDetail'
import AgreementCompanyMasterPage from '@/pages/admin/agreement-company-master/AgreementCompanyMasterPage'

import LegalAUCaseListPage from '@/pages/legal-au/legal-cases/LegalAUCaseListPage'
import LegalAUCaseDetailPage from '@/pages/legal-au/legal-cases/LegalAUCaseDetailPage'
import LegalAUCaseFormPage from '@/pages/legal-au/legal-cases/LegalAUCaseFormPage'
import LegalAU_OpinionListPage from '@/pages/legal-au/legal-opinions/LegalAU_OpinionListPage'
import LegalAU_ReviewDocumentListPage from '@/pages/legal-au/review-documents/LegalAU_ReviewDocumentListPage'
import LegalAUMaterialManagementPage from '@/pages/legal-au/materials/LegalAUMaterialManagementPage'

const notFoundElement = (
  <div className="flex items-center justify-center min-h-[60vh]">
    <h1 className="text-2xl font-semibold text-gray-400">404 - Halaman tidak ditemukan</h1>
  </div>
)

export const router = createBrowserRouter([
  // ─── Public ───────────────────────────────────────────────────────────────
  {
    element: <PublicLayout />,
    children: [
      { path: '/', element: <HomePage /> },
    ],
  },

  // ─── Authenticated public pages ──────────────────────────────────────────
  {
    element: <PrivateRoute />,
    children: [
      {
        element: <PublicLayout />,
        children: [
          { path: '/akta-perusahaan', element: <AktaPerusahaanPage /> },
          { path: '/asset-perusahaan', element: <ComingSoonPage title="Asset Perusahaan" /> },
          { path: '/sk-sop-legal', element: <ComingSoonPage title="SK SOP Legal" /> },
          { path: '/materi-legal', element: <MaterialListingPage /> },
          { path: '/profil-legal', element: <ComingSoonPage title="Profil Legal" /> },
        ],
      },
    ],
  },

  // ─── Guest only ───────────────────────────────────────────────────────────
  {
    element: <GuestRoute />,
    children: [{ path: '/login', element: <LoginPage /> }],
  },

  // ─── User dashboard ───────────────────────────────────────────────────────
  {
    element: <UserRoute />,
    children: [
      {
        element: <DashboardLayout />,
        children: [
          { path: '/dashboard', element: <UserDashboardPage /> },
          { path: '/dashboard/settings', element: <SettingsPage /> },
          { path: '/dashboard/notifications', element: <NotificationListPage /> },
          { path: '/dashboard/legal-opinions', element: <LegalOpinionListPage /> },
          { path: '/dashboard/legal-opinions/new', element: <LegalOpinionFormPage /> },
          { path: '/dashboard/legal-opinions/:id', element: <LegalOpinionDetailPage /> },
          { path: '/dashboard/legal-opinions/:id/edit', element: <LegalOpinionFormPage /> },
          { path: '/dashboard/review-documents', element: <ReviewDocumentListPage /> },
          { path: '/dashboard/review-documents/new', element: <ReviewDocumentFormPage /> },
          { path: '/dashboard/review-documents/:id', element: <ReviewDocumentDetailPage /> },
          { path: '/dashboard/review-documents/:id/edit', element: <ReviewDocumentFormPage /> },
          { path: '/dashboard/agreement-documents', element: <AgreementDocumentList basePath="/dashboard/agreement-documents" requester /> },
          { path: '/dashboard/agreement-documents/new', element: <AgreementDocumentForm /> },
          { path: '/dashboard/agreement-documents/:id', element: <AgreementDocumentDetail /> },
          { path: '/dashboard/agreement-documents/:id/edit', element: <AgreementDocumentForm /> },
          { path: '/dashboard/legal-cases', element: <AdminLegalCaseListPage /> },
          { path: '/dashboard/legal-cases/:id', element: <AdminLegalCaseDetailPage /> },
          { path: '/dashboard/audit-logs', element: <AuditLogPage /> },
          { path: '/dashboard/reports', element: <ReportPage /> },
          { path: '/dashboard/materials', element: <DashboardMaterialManagementPage /> },
          { path: '/dashboard/materials/new', element: <MaterialFormPage /> },
          { path: '/dashboard/materials/:id', element: <MaterialFormPage /> },
          { path: '/dashboard/users', element: <UserManagementPage /> },
        ],
      },
    ],
  },

  // ─── Admin ────────────────────────────────────────────────────────────────
  {
    element: <AdminRoute />,
    children: [
      {
        element: <AdminLayout />,
        children: [
          { path: '/admin', element: <AdminDashboardPage /> },
          { path: '/admin/settings', element: <SettingsPage /> },
          { path: '/admin/notifications', element: <NotificationListPage /> },
          { path: '/admin/legal-opinions', element: <AdminLegalOpinionListPage /> },
          { path: '/admin/legal-opinions/:id', element: <AdminLegalOpinionDetailPage /> },
          { path: '/admin/review-documents', element: <AdminReviewDocumentListPage /> },
          { path: '/admin/review-documents/:id', element: <AdminReviewDocumentDetailPage /> },
          { path: '/admin/agreement-documents', element: <AgreementDocumentList basePath="/admin/agreement-documents" apiBase="/admin" /> },
          { path: '/admin/agreement-documents/:id', element: <AgreementDocumentDetail apiBase="/admin" approver /> },
          { path: '/admin/agreement-company-master', element: <AgreementCompanyMasterPage /> },
          { path: '/admin/legal-cases', element: <AdminLegalCaseListPage /> },
          { path: '/admin/legal-cases/:id', element: <AdminLegalCaseDetailPage /> },
          { path: '/admin/users', element: <UserManagementPage /> },
          { path: '/admin/audit-logs', element: <AuditLogPage /> },
          { path: '/admin/companies', element: <CompanyManagementPage /> },
          { path: '/admin/purpose-types', element: <PurposeTypeManagementPage /> },
          { path: '/admin/document-types', element: <DocumentTypeManagementPage /> },
          { path: '/admin/case-types', element: <CaseTypeManagementPage /> },
          { path: '/admin/case-categories', element: <CaseCategoryManagementPage /> },
          { path: '/admin/regencies', element: <RegencyManagementPage /> },
          { path: '/admin/cedants', element: <CedantManagementPage /> },
          { path: '/admin/divisions', element: <DivisionManagementPage /> },
          { path: '/admin/materials', element: <MaterialManagementPage /> },
          { path: '/admin/materials/new', element: <MaterialFormPage /> },
          { path: '/admin/materials/:id', element: <MaterialFormPage /> },
          { path: '/admin/reports', element: <ReportPage /> },
        ],
      },
    ],
  },

  // ─── Legal ───────────────────────────────────────────────────────────────
  {
    element: <LegalRoute />,
    children: [
      {
        element: <LegalLayout />,
        children: [
          { path: '/legal', element: <LegalDashboardPage /> },
          { path: '/legal/settings', element: <SettingsPage /> },
          { path: '/legal/notifications', element: <NotificationListPage /> },
          { path: '/legal/legal-opinions', element: <LegalLegalOpinionListPage /> },
          { path: '/legal/legal-opinions/:id', element: <LegalLegalOpinionDetailPage /> },
          { path: '/legal/review-documents', element: <LegalReviewDocumentListPage /> },
          { path: '/legal/review-documents/:id', element: <LegalReviewDocumentDetailPage /> },
          { path: '/legal/agreement-documents', element: <AgreementDocumentList basePath="/legal/agreement-documents" apiBase="/legal" /> },
          { path: '/legal/agreement-documents/:id', element: <AgreementDocumentDetail apiBase="/legal" approver /> },
          { path: '/legal/legal-cases', element: <AdminLegalCaseListPage /> },
          { path: '/legal/legal-cases/:id', element: <AdminLegalCaseDetailPage /> },
          { path: '/legal/audit-logs', element: <AuditLogPage /> },
          { path: '/legal/materials', element: <LegalMaterialManagementPage /> },
          { path: '/legal/materials/new', element: <MaterialFormPage /> },
          { path: '/legal/materials/:id', element: <LegalMaterialDetailPage /> },
          { path: '/legal/reports', element: <ReportPage /> },
        ],
      },
    ],
  },

  // ─── Legal AU ────────────────────────────────────────────────────────────
  {
    element: <LegalAURoute />,
    children: [
      {
        element: <LegalAULayout />,
        children: [
          { path: '/legal-au', element: <LegalAUCaseListPage /> },
          { path: '/legal-au/cases', element: <LegalAUCaseListPage /> },
          { path: '/legal-au/cases/new', element: <LegalAUCaseFormPage /> },
          { path: '/legal-au/cases/:id', element: <LegalAUCaseDetailPage /> },
          { path: '/legal-au/materials', element: <LegalAUMaterialManagementPage /> },
          { path: '/legal-au/materials/new', element: <MaterialFormPage /> },
          { path: '/legal-au/materials/:id', element: <MaterialFormPage /> },
          { path: '/legal-au/settings', element: <SettingsPage /> },
          { path: '/legal-au/notifications', element: <NotificationListPage /> },
          { path: '/legal-au/reports', element: <ReportPage /> },
          { path: '/legal-au/audit-logs', element: <AuditLogPage /> },
          { path: '/legal-au/legal-opinions', element: <LegalAU_OpinionListPage /> },
          { path: '/legal-au/legal-opinions/:id', element: <LegalLegalOpinionDetailPage /> },
          { path: '/legal-au/legal-opinions/new', element: <LegalOpinionFormPage /> },
          { path: '/legal-au/legal-opinions/:id/edit', element: <LegalOpinionFormPage /> },
          { path: '/legal-au/review-documents', element: <LegalAU_ReviewDocumentListPage /> },
          { path: '/legal-au/review-documents/:id', element: <LegalReviewDocumentDetailPage /> },
          { path: '/legal-au/review-documents/new', element: <ReviewDocumentFormPage /> },
          { path: '/legal-au/review-documents/:id/edit', element: <ReviewDocumentFormPage /> },
          { path: '/legal-au/users', element: <UserManagementPage /> },
        ],
      },
    ],
  },

  // ─── External ────────────────────────────────────────────────────────────
  {
    element: <ExternalRoute />,
    children: [
      {
        element: <ExternalLayout />,
        children: [
          { path: '/external/settings', element: <SettingsPage /> },
          { path: '/external/notifications', element: <NotificationListPage /> },
          { path: '/external/legal-cases', element: <AdminLegalCaseListPage /> },
          { path: '/external/legal-cases/:id', element: <AdminLegalCaseDetailPage /> },
          { path: '/external/reports', element: <ReportPage /> },
          { path: '/external/audit-logs', element: <AuditLogPage /> },
          { path: '/external/legal-opinions', element: <External_OpinionListPage /> },
          { path: '/external/legal-opinions/:id', element: <LegalLegalOpinionDetailPage /> },
          { path: '/external/legal-opinions/new', element: <LegalOpinionFormPage /> },
          { path: '/external/legal-opinions/:id/edit', element: <LegalOpinionFormPage /> },
          { path: '/external/review-documents', element: <External_ReviewDocumentListPage /> },
          { path: '/external/review-documents/:id', element: <LegalReviewDocumentDetailPage /> },
          { path: '/external/review-documents/new', element: <ReviewDocumentFormPage /> },
          { path: '/external/review-documents/:id/edit', element: <ReviewDocumentFormPage /> },
          { path: '/external/materials', element: <ExternalMaterialManagementPage /> },
          { path: '/external/materials/new', element: <MaterialFormPage /> },
          { path: '/external/materials/:id', element: <MaterialFormPage /> },
          { path: '/external/users', element: <UserManagementPage /> },
        ],
      },
    ],
  },

  { path: '*', element: notFoundElement },
])
