import { createBrowserRouter } from 'react-router-dom'
import { GuestRoute, PrivateRoute, AdminRoute } from './guards'

import PublicLayout from '@/layouts/PublicLayout'
import DashboardLayout from '@/layouts/DashboardLayout'
import HomePage from '@/pages/public/HomePage'
import AktaPerusahaanPage from '@/pages/public/AktaPerusahaanPage'
import ComingSoonPage from '@/pages/public/ComingSoonPage'
import LoginPage from '@/pages/auth/LoginPage'
import LegalOpinionListPage from '@/pages/dashboard/legal-opinions/LegalOpinionListPage'
import LegalOpinionFormPage from '@/pages/dashboard/legal-opinions/LegalOpinionFormPage'
import LegalOpinionDetailPage from '@/pages/dashboard/legal-opinions/LegalOpinionDetailPage'

const Placeholder = ({ title }: { title: string }) => (
  <div className="flex items-center justify-center min-h-[60vh]">
    <h1 className="text-2xl font-semibold text-gray-400">{title} — Coming Soon</h1>
  </div>
)

export const router = createBrowserRouter([
  // ─── Public ───────────────────────────────────────────────────────────────
  {
    element: <PublicLayout />,
    children: [
      { path: '/', element: <HomePage /> },
      { path: '/akta-perusahaan', element: <AktaPerusahaanPage /> },
      { path: '/asset-perusahaan', element: <ComingSoonPage title="Asset Perusahaan" /> },
      { path: '/sk-sop-legal', element: <ComingSoonPage title="SK SOP Legal" /> },
      { path: '/materi-legal', element: <ComingSoonPage title="Materi Legal" /> },
      { path: '/profil-legal', element: <ComingSoonPage title="Profil Legal" /> },
    ],
  },

  // ─── Guest only ───────────────────────────────────────────────────────────
  {
    element: <GuestRoute />,
    children: [
      { path: '/login', element: <LoginPage /> },
    ],
  },

  // ─── User dashboard ───────────────────────────────────────────────────────
  {
    element: <PrivateRoute />,
    children: [
      {
        element: <DashboardLayout />,
        children: [
          { path: '/dashboard', element: <Placeholder title="Dashboard User" /> },
          { path: '/dashboard/legal-opinions', element: <LegalOpinionListPage /> },
          { path: '/dashboard/legal-opinions/new', element: <LegalOpinionFormPage /> },
          { path: '/dashboard/legal-opinions/:id', element: <LegalOpinionDetailPage /> },
          { path: '/dashboard/legal-opinions/:id/edit', element: <LegalOpinionFormPage /> },
          { path: '/dashboard/review-documents', element: <Placeholder title="Review Dokumen" /> },
          { path: '/dashboard/review-documents/new', element: <Placeholder title="Buat Review Dokumen" /> },
          { path: '/dashboard/review-documents/:id', element: <Placeholder title="Detail Review Dokumen" /> },
        ],
      },
    ],
  },

  // ─── Admin ────────────────────────────────────────────────────────────────
  {
    element: <AdminRoute />,
    children: [
      { path: '/admin', element: <Placeholder title="Dashboard Admin" /> },
      { path: '/admin/legal-opinions', element: <Placeholder title="Manage Legal Opinion" /> },
      { path: '/admin/legal-opinions/:id', element: <Placeholder title="Detail Legal Opinion Admin" /> },
      { path: '/admin/review-documents', element: <Placeholder title="Manage Review Dokumen" /> },
      { path: '/admin/review-documents/:id', element: <Placeholder title="Detail Review Dokumen Admin" /> },
      { path: '/admin/users', element: <Placeholder title="User Management" /> },
    ],
  },

  { path: '*', element: <Placeholder title="404 — Halaman tidak ditemukan" /> },
])