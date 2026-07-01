import api from './api'
import type { ApiResponse, NotificationSetting } from '@/types'

export const notificationSettingService = {
  getAll: async (): Promise<NotificationSetting[]> => {
    const res = await api.get<ApiResponse<NotificationSetting[]>>('/admin/notification-settings')
    return res.data.data!
  },

  update: async (id: string, data: { days_threshold: number; is_active?: boolean }): Promise<NotificationSetting> => {
    const res = await api.put<ApiResponse<NotificationSetting>>(`/admin/notification-settings/${id}`, data)
    return res.data.data!
  },
}
