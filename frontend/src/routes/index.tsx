import { createBrowserRouter } from 'react-router-dom'
import { GuestRoute, PrivateRoute, AdminRoute } from './guards'

import PublicLayout from '@/layouts/PublicLayout'
import DashboardLayout from '@/layouts/DashboardLayout'
import AdminLayout from '@/layouts/AdminLayout'

import HomePage from '@/pages/public/HomePage'
import AktaPerusahaanPage from '@/pages/public/AktaPerusahaanPage'
import ComingSoonPage from '@/pages/public/ComingSoonPage'
import LoginPage from '@/pages/auth/LoginPage'

import UserDashboardPage from '@/pages/dashboard/UserDashboardPage'
import SettingsPage from '@/pages/settings/SettingsPage'
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
          { path: '/admin/legal-opinions', element: <AdminLegalOpinionListPage /> },
          { path: '/admin/legal-opinions/:id', element: <AdminLegalOpinionDetailPage /> },
          { path: '/admin/review-documents', element: <AdminReviewDocumentListPage /> },
          { path: '/admin/review-documents/:id', element: <AdminReviewDocumentDetailPage /> },
          { path: '/admin/users', element: <UserManagementPage /> },
        ],
      },
    ],
  },

  { path: '*', element: notFoundElement },
])
