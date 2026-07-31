import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '@/store/auth.store'
import { getRoleHome } from './role-home'

// Redirect to login if not authenticated
export function PrivateRoute() {
  const { isAuthenticated } = useAuthStore()
  return isAuthenticated ? <Outlet /> : <Navigate to="/login" replace />
}

// Redirect to dashboard if not admin
export function AdminRoute() {
  const { isAuthenticated, user } = useAuthStore()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  if (user?.role !== 'ADMIN') return <Navigate to={getRoleHome(user?.role)} replace />
  return <Outlet />
}

// Legal route - for LEGAL role users
export function LegalRoute() {
  const { isAuthenticated, user } = useAuthStore()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  if (user?.role !== 'LEGAL') return <Navigate to={getRoleHome(user?.role)} replace />
  return <Outlet />
}

// Legal AU route - for LEGAL_AU role users
export function LegalAURoute() {
  const { isAuthenticated, user } = useAuthStore()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  if (user?.role !== 'LEGAL_AU') return <Navigate to={getRoleHome(user?.role)} replace />
  return <Outlet />
}

// External user route - for EXTERNAL role users
export function ExternalRoute() {
  const { isAuthenticated, user } = useAuthStore()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  if (user?.role !== 'EXTERNAL') return <Navigate to={getRoleHome(user?.role)} replace />
  return <Outlet />
}

// User route - for regular USER role users
export function UserRoute() {
  const { isAuthenticated, user } = useAuthStore()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  if (user?.role !== 'USER') return <Navigate to={getRoleHome(user?.role)} replace />
  return <Outlet />
}

// Redirect to dashboard if already logged in
export function GuestRoute() {
  const { isAuthenticated, user } = useAuthStore()
  if (!isAuthenticated) return <Outlet />
  return <Navigate to={getRoleHome(user?.role)} replace />
}
