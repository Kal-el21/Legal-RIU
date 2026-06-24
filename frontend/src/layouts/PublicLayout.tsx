import { Outlet } from 'react-router-dom'
import Navbar from '@/components/common/Navbar'
import Footer from '@/components/common/Footer'
import { useAuthStore } from '@/store/auth.store'

export default function PublicLayout() {
  const { isAuthenticated } = useAuthStore()

  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <main className="flex-1">
        <Outlet />
      </main>
      {isAuthenticated && <Footer />}
    </div>
  )
}
