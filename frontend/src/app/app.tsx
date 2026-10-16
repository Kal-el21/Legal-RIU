import { useEffect } from 'react'
import { RouterProvider } from 'react-router-dom'
import { MutationCache, QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { isAxiosError } from 'axios'
import { Toaster, toast } from 'sonner'
import { router } from '@/routes'
import api from '@/services/api'
import { useAuthStore } from '@/store/auth.store'
import type { ApiResponse, User } from '@/types'

declare module '@tanstack/react-query' {
  interface Register {
    // success: pesan notifikasi saat aksi berhasil; tanpa pesan, tidak ada notifikasi.
    // errorToast: false untuk aksi yang kegagalannya sudah ditampilkan sendiri (mis. login).
    mutationMeta: { success?: string; errorToast?: boolean }
  }
}

const queryClient = new QueryClient({
  // Hasil semua aksi diberitahukan dari satu tempat. Pesan gagal berasal dari server,
  // yang dipindahkan ke error.message oleh services/api.ts.
  mutationCache: new MutationCache({
    onSuccess: (_data, _variables, _context, mutation) => {
      if (mutation.meta?.success) toast.success(mutation.meta.success)
    },
    onError: (error, _variables, _context, mutation) => {
      if (mutation.meta?.errorToast === false) return
      // Sesi habis: services/api.ts sudah mengarahkan ke halaman login.
      if (isAxiosError(error) && error.response?.status === 401) return
      const message = isAxiosError(error) && !error.response
        ? 'Tidak dapat terhubung ke server. Periksa koneksi lalu coba lagi.'
        : error.message || 'Terjadi kesalahan, coba lagi.'
      toast.error(message, { duration: 8000 })
    },
  }),
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 1000 * 60 * 5, // 5 minutes
      refetchOnWindowFocus: false,
    },
  },
})

export default function App() {
  useEffect(() => {
    api
      .get<ApiResponse<User>>('/auth/me', {
        headers: {
          'X-Skip-Auth-Redirect': 'true',
        },
      })
      .then((res) => {
        const user = res.data.data!
        api
          .get<ApiResponse<string[]>>('/auth/permissions', {
            headers: {
              'X-Skip-Auth-Redirect': 'true',
            },
          })
          .then((permissionRes) => {
            useAuthStore.getState().setAuth(user, permissionRes.data.data ?? [])
          })
          .catch(() => {
            useAuthStore.getState().setAuth(user)
          })
      })
      .catch(() => {
        // Ignore. User belum login atau session expired.
      })
  }, [])

  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Toaster position="top-right" richColors closeButton />
    </QueryClientProvider>
  )
}
