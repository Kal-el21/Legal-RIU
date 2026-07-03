import { useEffect } from 'react'
import { RouterProvider } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { router } from '@/routes'
import api from '@/services/api'
import { useAuthStore } from '@/store/auth.store'
import type { ApiResponse, User } from '@/types'

const queryClient = new QueryClient({
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
    </QueryClientProvider>
  )
}
