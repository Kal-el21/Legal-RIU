import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/store/auth.store'
import { getRoleHome } from '@/routes/role-home'

interface PermissionGateProps {
  permission: string
  children: React.ReactNode
  redirectTo?: string
}

export default function PermissionGate({
  permission,
  children,
  redirectTo,
}: PermissionGateProps) {
  const navigate = useNavigate()
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const role = useAuthStore((state) => state.user?.role)

  useEffect(() => {
    if (!hasPermission(permission)) {
      navigate(redirectTo || getRoleHome(role), { replace: true })
    }
  }, [hasPermission, navigate, permission, redirectTo, role])

  return <>{children}</>
}
