import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  setAuth: (user: User) => void
  updateUser: (user: User) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,

      setAuth: (user: User) => {
        set({ user, isAuthenticated: true })
      },

      updateUser: (user: User) => {
        set({ user })
      },

      logout: () => {
        set({ user: null, isAuthenticated: false })
      },
    }),
    {
      name: 'legal-riu-auth',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)