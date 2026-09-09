import api from './api'

export interface AgreementTemplate {
  id: string
  code: string
  version: number
  name: string
  file_name: string
  checksum: string
  status: 'DRAFT' | 'ACTIVE' | 'ARCHIVED'
  is_legacy: boolean
  placeholders?: string[]
  note?: string
  effective_date?: string
  activated_at?: string
  created_at?: string
  uploader?: { full_name?: string }
}

export interface PlaceholderReference {
  token: string
  group: string
  description: string
}

export interface UploadTemplatePayload {
  code?: string
  name?: string
  note?: string
  effective_date?: string
}

const data = <T>(r: { data: { data: T } }) => r.data.data

export const agreementTemplateService = {
  list: (base = '/admin', code = 'PKS') =>
    api.get(`${base}/agreement-templates`, { params: { code } }).then((r) => data<AgreementTemplate[]>(r)),

  get: (base = '/admin', id: string) =>
    api.get(`${base}/agreement-templates/${id}`).then((r) => data<AgreementTemplate>(r)),

  placeholders: (base = '/admin') =>
    api.get(`${base}/agreement-templates/placeholders`).then((r) => data<PlaceholderReference[]>(r)),

  upload: (base = '/admin', file: File, payload: UploadTemplatePayload) => {
    const form = new FormData()
    form.append('file', file)
    form.append('code', payload.code || 'PKS')
    if (payload.name) form.append('name', payload.name)
    if (payload.note) form.append('note', payload.note)
    if (payload.effective_date) form.append('effective_date', payload.effective_date)
    return api.post(`${base}/agreement-templates`, form).then((r) => data<AgreementTemplate>(r))
  },

  activate: (base = '/admin', id: string) =>
    api.post(`${base}/agreement-templates/${id}/activate`).then((r) => data<AgreementTemplate>(r)),

  fileUrl: (base: string, id: string, kind: 'preview' | 'download') =>
    `${api.defaults.baseURL}${base}/agreement-templates/${id}/${kind}`,
}
