import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { dashboardService, type ReminderParams } from '@/services/dashboard.service'
import { useAuthStore } from '@/store/auth.store'

const KEYS = {
  all: ['notifications'] as const,
  list: (params: ReminderParams) => [...KEYS.all, params] as const,
}

export function useNotifications(params: ReminderParams) {
  const loggedIn = useAuthStore((state) => Boolean(state.user))
  return useQuery({
    queryKey: KEYS.list(params),
    queryFn: () => dashboardService.getReminders(params),
    staleTime: 30_000,
    enabled: loggedIn,
  })
}

// Menandai dibaca memuat ulang lonceng, halaman daftar, dan peringatan di dashboard sekaligus
export function useMarkNotificationRead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: dashboardService.markReminderRead,
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}

export function useMarkAllNotificationsRead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: dashboardService.markAllRemindersRead,
    onSuccess: () => qc.invalidateQueries({ queryKey: KEYS.all }),
  })
}
