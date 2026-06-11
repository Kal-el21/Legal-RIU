import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/store/auth.store'
import type { ApiResponse, AuthResponse } from '@/types'

const api = axios.create({
  baseURL: '/api/v1',
})

function clearLegacyAuthStorage() {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
}

interface RetriableRequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean
}

let refreshPromise: Promise<AuthResponse> | null = null
let csrfToken: string | null = null

function refreshSession(refreshToken: string) {
  if (!refreshPromise) {
    refreshPromise = axios
      .post<ApiResponse<AuthResponse>>('/auth/refresh', { refresh_token: refreshToken })
      .then((res) => res.data.data!)
      .finally(() => {
        refreshPromise = null
      })
  }
  return refreshPromise
}

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token
  if (token) {
    config.headers = config.headers ?? {}
    ;(config.headers as Record<string, string>).Authorization = `Bearer ${token}`
  }
  // Don't set Content-Type for FormData - browser sets it with boundary
  if (!config.data || !(config.data instanceof FormData)) {
    ;(config.headers as Record<string, string>)['Content-Type'] = 'application/json'
  }
  // Attach CSRF token for non-GET requests
  if (csrfToken && !['GET', 'HEAD', 'OPTIONS'].includes(config.method?.toUpperCase() || 'GET')) {
    ;(config.headers as Record<string, string>)['X-CSRF-Token'] = csrfToken
  }
  return config
})

api.interceptors.response.use(
  (response) => {
    // Extract CSRF token from response headers (axios puts it in 'x-csrf-token')
    const csrfHeader = response.headers['x-csrf-token'] || response.headers['X-CSRF-Token']
    if (csrfHeader && typeof csrfHeader === 'string') {
      csrfToken = csrfHeader
    }
    return response
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as RetriableRequestConfig | undefined
    const isAuthRefresh = originalRequest?.url?.includes('/auth/refresh')
    const isAuthLogin = originalRequest?.url?.includes('/auth/login')

    if (error.response?.status === 401 && originalRequest && !originalRequest._retry && !isAuthRefresh && !isAuthLogin) {
      const refreshToken = useAuthStore.getState().refreshToken
      if (refreshToken) {
        try {
          originalRequest._retry = true
          const refreshed = await refreshSession(refreshToken)
          const token = refreshed.access_token ?? refreshed.token
          useAuthStore.getState().setAuth(token, refreshed.refresh_token, refreshed.user)
          originalRequest.headers = originalRequest.headers ?? {}
          ;(originalRequest.headers as Record<string, string>).Authorization = `Bearer ${token}`
          return api(originalRequest)
        } catch {
          // Fall through to logout below.
        }
      }
    }

    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      clearLegacyAuthStorage()

      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

export default api