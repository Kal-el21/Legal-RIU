import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types'

interface AuthState {
  user: User | null
  permissions: string[]
  isAuthenticated: boolean
  setAuth: (user: User, permissions?: string[]) => void
  setPermissions: (permissions: string[]) => void
  hasPermission: (...codes: string[]) => boolean
  updateUser: (user: User) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      permissions: [],
      isAuthenticated: false,

      setAuth: (user: User, permissions: string[] = []) => {
        set({ user, permissions, isAuthenticated: true })
      },

      setPermissions: (permissions: string[]) => {
        set({ permissions })
      },

      hasPermission: (...codes: string[]): boolean => {
        const state = get()
        if (!codes.length) return true
        if (state.user?.role === 'ADMIN') return true
        return codes.some((code) => state.permissions.includes(code))
      },

      updateUser: (user: User) => {
        set({ user })
      },

      logout: () => {
        set({ user: null, permissions: [], isAuthenticated: false })
      },
    }),
    {
      name: 'legal-riu-auth',
      partialize: (state) => ({
        user: state.user,
        permissions: state.permissions,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)
