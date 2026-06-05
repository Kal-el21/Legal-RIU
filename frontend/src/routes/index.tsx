import { createBrowserRouter } from 'react-router-dom'
import { GuestRoute, PrivateRoute, AdminRoute } from './guards'

import PublicLayout from '@/layouts/PublicLayout'
import HomePage from '@/pages/public/HomePage'
import AktaPerusahaanPage from '@/pages/public/AktaPerusahaanPage'
import ComingSoonPage from '@/pages/public/ComingSoonPage'
import LoginPage from '@/pages/auth/LoginPage'

const Placeholder = ({ title }: { title: string }) => (
  <div className="flex items-center justify-center min-h-screen">
    <h1 className="text-2xl font-semibold text-gray-400">{title} — Coming Soon</h1>
  </div>
)

export const router = createBrowserRouter([
  // ─── Public (with Navbar + Footer) ───────────────────────────────────────
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

  // ─── User routes ──────────────────────────────────────────────────────────
  {
    element: <PrivateRoute />,
    children: [
      { path: '/dashboard', element: <Placeholder title="Dashboard User" /> },
      { path: '/dashboard/legal-opinions', element: <Placeholder title="Legal Opinion List" /> },
      { path: '/dashboard/legal-opinions/new', element: <Placeholder title="Buat Legal Opinion" /> },
      { path: '/dashboard/legal-opinions/:id', element: <Placeholder title="Detail Legal Opinion" /> },
      { path: '/dashboard/review-documents', element: <Placeholder title="Review Dokumen List" /> },
      { path: '/dashboard/review-documents/new', element: <Placeholder title="Buat Review Dokumen" /> },
      { path: '/dashboard/review-documents/:id', element: <Placeholder title="Detail Review Dokumen" /> },
    ],
  },

  // ─── Admin routes ─────────────────────────────────────────────────────────
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