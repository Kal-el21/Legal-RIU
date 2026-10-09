import { Link } from 'react-router-dom'
import { useAuthStore } from '@/store/auth.store'
import { getRoleHome } from '@/routes/role-home'

export default function NotFoundPage() {
  const { isAuthenticated, user } = useAuthStore()

  return (
    <div className="flex flex-col items-center justify-center gap-6 min-h-[60vh]">
      <h1 className="text-2xl font-semibold text-gray-400">404 - Halaman tidak ditemukan</h1>
      <Link to={isAuthenticated ? getRoleHome(user?.role) : '/'}
        className="inline-flex items-center gap-2 px-6 py-3 rounded-xl text-sm font-semibold text-white transition-all hover:opacity-90"
        style={{ background: '#C8102E' }}>
        {isAuthenticated ? 'Ke Halaman Awal' : 'Ke Beranda'}
      </Link>
    </div>
  )
}
