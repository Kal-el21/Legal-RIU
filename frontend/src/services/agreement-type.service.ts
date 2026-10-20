import api from './api'
import type { AgreementSectionDef } from './agreement-template.service'

// Jenis perjanjian (mis. PKS). fields adalah field kerja Admin; pengajuan memakai
// salinan field di versi template yang dikuncinya.
export interface AgreementType {
  id: string
  code: string
  name: string
  is_active: boolean
  fields: AgreementSectionDef[] | null
}

export interface AgreementTypePayload {
  code?: string
  name: string
  is_active?: boolean
  fields?: AgreementSectionDef[]
}

const data = <T>(r: { data: { data: T } }) => r.data.data

export const agreementTypeService = {
  list: () => api.get('/admin/agreement-types').then((r) => data<AgreementType[]>(r)),
  create: (payload: AgreementTypePayload) =>
    api.post('/admin/agreement-types', payload).then((r) => data<AgreementType>(r)),
  update: (code: string, payload: AgreementTypePayload) =>
    api.put(`/admin/agreement-types/${code}`, payload).then((r) => data<AgreementType>(r)),
}
