import axios, {
  AxiosHeaders,
  type AxiosError,
  type InternalAxiosRequestConfig,
} from 'axios'
import { getRoleHome } from '@/routes/role-home'
import { useAuthStore } from '@/store/auth.store'
import type { ApiResponse, AuthResponse, User } from '@/types'

const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
})

function clearLegacyAuthStorage() {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
}

interface RetriableRequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean
  _csrfRetry?: boolean
}

let refreshPromise: Promise<AuthResponse> | null = null
let csrfToken: string | null = null

function refreshSession() {
  if (!refreshPromise) {
    refreshPromise = axios
      .post<ApiResponse<AuthResponse>>('/api/v1/auth/refresh', {}) // No body needed - cookie sent automatically
      .then((res) => res.data.data!)
      .finally(() => {
        refreshPromise = null
      })
  }
  return refreshPromise
}

// Pesan 403 dari pemeriksaan peran di backend (RoleMiddleware di
// backend/internal/middleware/auth.middleware.go).
const ROLE_FORBIDDEN_MESSAGE = 'Anda tidak memiliki akses ke halaman ini'

let roleCheckPromise: Promise<void> | null = null

// Peran user bisa diganti admin saat ia sedang login. Backend langsung menolak
// alamat peran lamanya, sedangkan peran yang tersimpan di browser baru
// diperbarui saat halaman dimuat ulang. Peran dibaca ulang, dan kalau memang
// berubah, user dipindah ke halaman awal peran barunya dengan muat ulang penuh
// supaya menu dan izinnya ikut diperbarui.
function syncRoleAfterForbidden() {
  if (!roleCheckPromise) {
    roleCheckPromise = axios
      .get<ApiResponse<User>>('/api/v1/auth/me')
      .then((res) => {
        const role = res.data.data?.role
        if (role && role !== useAuthStore.getState().user?.role) {
          window.location.href = getRoleHome(role)
        }
      })
      .catch(() => {
        // Sesi yang tidak sah ditangani penanganan 401 di bawah.
      })
      .finally(() => {
        roleCheckPromise = null
      })
  }
  return roleCheckPromise
}

type HeaderBag = { get?: (name: string) => unknown } | Record<string, unknown>

function getHeader(headers: HeaderBag | undefined, name: string) {
  const getter = (headers as { get?: (name: string) => unknown } | undefined)?.get

  if (typeof getter === 'function') {
    return getter.call(headers, name)
  }

  if (headers && typeof headers === 'object' && name in headers) {
    return (headers as Record<string, unknown>)[name]
  }

  return undefined
}

api.interceptors.request.use((config) => {
  const headers = AxiosHeaders.from(config.headers)
  const isFormData = config.data instanceof FormData
  const method = (config.method ?? 'GET').toUpperCase()

  // Penting untuk multipart/form-data.
  // Browser wajib men-set Content-Type termasuk boundary.
  if (isFormData) {
    headers.delete('Content-Type')
  } else if (config.data !== undefined) {
    headers.set('Content-Type', 'application/json')
  }

  // Backend membutuhkan X-CSRF-Token untuk non-GET.
  if (csrfToken && !['GET', 'HEAD', 'OPTIONS'].includes(method)) {
    headers.set('X-CSRF-Token', csrfToken)
  }

  config.headers = headers

  return config
})

api.interceptors.response.use(
  (response) => {
    const csrfHeader =
      getHeader(response.headers, 'x-csrf-token') ||
      getHeader(response.headers, 'X-CSRF-Token')

    if (typeof csrfHeader === 'string') {
      csrfToken = csrfHeader
    }

    return response
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as RetriableRequestConfig | undefined
    const isAuthRefresh = originalRequest?.url?.includes('/auth/refresh')
    const isAuthLogin = originalRequest?.url?.includes('/auth/login')
    const skipAuthRedirect =
      getHeader(originalRequest?.headers, 'X-Skip-Auth-Redirect') === 'true'

    if (
      error.response?.status === 401 &&
      originalRequest &&
      !originalRequest._retry &&
      !isAuthRefresh &&
      !isAuthLogin
    ) {
      try {
        originalRequest._retry = true

        const refreshed = await refreshSession()
        useAuthStore.getState().setAuth(refreshed.user, refreshed.permissions ?? [])

        return api(originalRequest)
      } catch {
        // Lanjut ke logout di bawah.
      }
    }

    if (error.response?.status === 401) {
      if (skipAuthRedirect) {
        return Promise.reject(error)
      }

      useAuthStore.getState().logout()
      clearLegacyAuthStorage()

      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }

    const responseData = error.response?.data as ApiResponse<unknown> | undefined

    if (
      error.response?.status === 403 &&
      originalRequest &&
      !originalRequest._csrfRetry
    ) {
      const msg = String(responseData?.message ?? '')
      if (msg.includes('CSRF token invalid')) {
        try {
          originalRequest._csrfRetry = true
          await api.get('/agreement-document-types')
          return api(originalRequest)
        } catch {
          // Fall through to reject with original error message.
        }
      }
    }

    if (error.response?.status === 403 && responseData?.message === ROLE_FORBIDDEN_MESSAGE) {
      await syncRoleAfterForbidden()
    }

    if (responseData?.message) {
      error.message = responseData.message
    }

    return Promise.reject(error)
  }
)

export default api
