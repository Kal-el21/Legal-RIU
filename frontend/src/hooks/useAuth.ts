import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { authService } from '@/services/auth.service'
import { useAuthStore } from '@/store/auth.store'
import type { LoginRequest } from '@/types'

export function useLogin() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)

  return useMutation({
    // Login gagal sudah ditampilkan di form login, tidak perlu notifikasi lagi.
    meta: { errorToast: false },
    // Isian ber-"@" adalah email akun lokal; selain itu username akun kantor (LDAP).
    mutationFn: (data: LoginRequest) =>
      data.email.includes('@')
        ? authService.login(data)
        : authService.ldapLogin({ username: data.email, password: data.password }),
    onSuccess: (res) => {
      setAuth(res.user, res.permissions ?? [])
      if (res.user.role === 'ADMIN') {
        navigate('/admin')
      } else if (res.user.role === 'LEGAL') {
        navigate('/legal')
      } else if (res.user.role === 'LEGAL_AU') {
        navigate('/legal-au')
      } else if (res.user.role === 'EXTERNAL') {
        navigate('/external/legal-cases')
      } else {
        navigate('/dashboard')
      }
    },
  })
}

export function useLogout() {
  const navigate = useNavigate()
  const logout = useAuthStore((s) => s.logout)

  return () => {
    const finish = () => {
      logout()
      navigate('/login')
    }

    authService.logout().finally(finish)
  }
}

export function useCurrentUser() {
  return useAuthStore((s) => s.user)
}
