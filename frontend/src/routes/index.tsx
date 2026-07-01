import { createBrowserRouter } from 'react-router-dom'
import { GuestRoute, PrivateRoute, AdminRoute, LegalRoute, ExternalRoute } from './guards'

import PublicLayout from '@/layouts/PublicLayout'
import DashboardLayout from '@/layouts/DashboardLayout'
import AdminLayout from '@/layouts/AdminLayout'
import LegalLayout from '@/layouts/LegalLayout'
import ExternalLayout from '@/layouts/ExternalLayout'

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

import LegalDashboardPage from '@/pages/legal/LegalDashboardPage'
import LegalLegalOpinionListPage from '@/pages/legal/legal-opinions/LegalOpinionListPage'
import LegalLegalOpinionDetailPage from '@/pages/legal/legal-opinions/LegalOpinionDetailPage'
import LegalReviewDocumentListPage from '@/pages/legal/review-documents/ReviewDocumentListPage'
import LegalReviewDocumentDetailPage from '@/pages/legal/review-documents/ReviewDocumentDetailPage'

import ExternalDashboardPage from '@/pages/external/ExternalDashboardPage'
import ExternalLegalOpinionListPage from '@/pages/external/legal-opinions/LegalOpinionListPage'
import ExternalLegalOpinionDetailPage from '@/pages/external/legal-opinions/LegalOpinionDetailPage'
import ExternalReviewDocumentListPage from '@/pages/external/review-documents/ReviewDocumentListPage'
import ExternalReviewDocumentDetailPage from '@/pages/external/review-documents/ReviewDocumentDetailPage'

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

  {
    element: <PrivateRoute />,
    children: [
      {
        element: <PublicLayout />,
        children: [
          { path: '/akta-perusahaan', element: <AktaPerusahaanPage /> },
          { path: '/asset-perusahaan', element: <ComingSoonPage title="Asset Perusahaan" /> },
          { path: '/sk-sop-legal', element: <ComingSoonPage title="SK SOP Legal" /> },
          { path: '/materi-legal', element: <ComingSoonPage title="Materi Legal" /> },
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
    element: <PrivateRoute />,
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
          { path: '/admin/legal-cases', element: <AdminLegalCaseListPage /> },
          { path: '/admin/legal-cases/:id', element: <AdminLegalCaseDetailPage /> },
          { path: '/admin/users', element: <UserManagementPage /> },
          { path: '/admin/audit-logs', element: <AuditLogPage /> },
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
          { path: '/external', element: <ExternalDashboardPage /> },
          { path: '/external/settings', element: <SettingsPage /> },
          { path: '/external/notifications', element: <NotificationListPage /> },
          { path: '/external/legal-opinions', element: <ExternalLegalOpinionListPage /> },
          { path: '/external/legal-opinions/:id', element: <ExternalLegalOpinionDetailPage /> },
          { path: '/external/review-documents', element: <ExternalReviewDocumentListPage /> },
          { path: '/external/review-documents/:id', element: <ExternalReviewDocumentDetailPage /> },
        ],
      },
    ],
  },

  { path: '*', element: notFoundElement },
])
