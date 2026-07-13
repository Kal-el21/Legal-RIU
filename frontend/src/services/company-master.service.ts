import api from './api'
import type { CompanyMaster, CompanyMasterTemplate, TemplateFieldPosition } from '@/types'

export const companyMasterService = {
  getAll: async () => {
    const res = await api.get('/admin/company-masters')
    return res.data.data as CompanyMaster[]
  },

  getByID: async (id: string) => {
    const res = await api.get(`/admin/company-masters/${id}`)
    return res.data.data as CompanyMaster
  },

  create: async (data: Partial<CompanyMaster>) => {
    const res = await api.post('/admin/company-masters', data)
    return res.data.data as CompanyMaster
  },

  update: async (id: string, data: Partial<CompanyMaster>) => {
    const res = await api.put(`/admin/company-masters/${id}`, data)
    return res.data.data as CompanyMaster
  },

  delete: async (id: string) => {
    await api.delete(`/admin/company-masters/${id}`)
  },

  uploadTemplate: async (version: string, file: File): Promise<CompanyMasterTemplate> => {
    const formData = new FormData()
    formData.append('template', file, file.name)
    formData.append('version', version)
    const res = await api.post('/admin/company-masters/template/upload', formData)
    return res.data.data as CompanyMasterTemplate
  },

  getActiveTemplate: async (): Promise<CompanyMasterTemplate | undefined> => {
    const res = await api.get('/admin/company-masters/template/active')
    return res.data.data as CompanyMasterTemplate
  },

  getTemplate: async (version: string) => {
    const res = await api.get(`/admin/company-masters/template/${version}`)
    return res.data.data as CompanyMasterTemplate
  },

  deleteTemplate: async (version: string) => {
    await api.delete(`/admin/company-masters/template/${version}`)
  },

  getFieldPositions: async (version: string): Promise<TemplateFieldPosition[]> => {
    const res = await api.get(`/admin/templates/${version}/field-positions`)
    return res.data.data as TemplateFieldPosition[]
  },

  saveFieldPositions: async (version: string, positions: TemplateFieldPosition[]) => {
    await api.put(`/admin/templates/${version}/field-positions`, positions)
  },

  getTemplatePreview: async (version: string, page: number = 1): Promise<string> => {
    const res = await api.get(`/admin/templates/${version}/preview`, {
      responseType: 'blob',
      params: { page },
    })
    return URL.createObjectURL(res.data)
  },
}
