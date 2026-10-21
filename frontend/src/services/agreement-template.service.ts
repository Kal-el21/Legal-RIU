import api from './api'

// Satu isian di definisi jenis perjanjian; name disimpan huruf kecil dan ditulis
// huruf besar di Word ({{NAMA_MITRA}}).
export interface AgreementFieldDef {
  name: string
  label: string
  type: string
  required: boolean
  // Pilihan tipe select, atau saran isian tipe text.
  options?: string[]
  // Sub-field tipe group (mis. termin), diisi sekali per baris.
  fields?: AgreementFieldDef[]
  // Kode field uang yang dibagi ke baris grup menurut persennya.
  split_of?: string
  // Field uang yang memilih penandatangan Pihak Pertama (PR 11); paling banyak satu.
  primary_value?: boolean
}

export interface AgreementSectionDef {
  title: string
  fields: AgreementFieldDef[]
}

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
  fields?: AgreementSectionDef[]
  note?: string
  effective_date?: string
  activated_at?: string
  created_at?: string
  // Kosong untuk template bawaan aplikasi (dipasang sistem).
  uploaded_by?: string
  uploader?: { full_name?: string }
  running_documents: number
}

export interface PlaceholderReference {
  token: string
  group: string
  description: string
}

// Hasil pencocokan DOCX dengan field jenis dokumen saat versi dibuat.
export interface TemplateValidation {
  placeholders: string[]
  unknown_placeholders: string[]
  unused_placeholders: string[]
  warnings: string[]
}

export interface UploadTemplatePayload {
  code: string
  name?: string
  note: string
  effective_date?: string
}

type VersiBaru = { template: AgreementTemplate; validation: TemplateValidation }

const data = <T>(r: { data: { data: T } }) => r.data.data

// Template perjanjian hanya dikelola Admin (keputusan T1).
export const agreementTemplateService = {
  list: (code: string) =>
    api.get('/admin/agreement-templates', { params: { code } }).then((r) => data<AgreementTemplate[]>(r)),

  placeholders: (code: string) =>
    api.get('/admin/agreement-templates/placeholders', { params: { code } }).then((r) => data<PlaceholderReference[]>(r)),

  upload: (file: File, payload: UploadTemplatePayload) => {
    const form = new FormData()
    form.append('file', file)
    form.append('code', payload.code)
    form.append('note', payload.note)
    if (payload.name) form.append('name', payload.name)
    if (payload.effective_date) form.append('effective_date', payload.effective_date)
    return api.post('/admin/agreement-templates', form).then((r) => data<VersiBaru>(r))
  },

  // "Simpan sebagai versi baru": DOCX disalin dari versi terakhir.
  newVersion: (payload: UploadTemplatePayload) =>
    api.post('/admin/agreement-templates/new-version', payload).then((r) => data<VersiBaru>(r)),

  activate: (id: string) =>
    api.post(`/admin/agreement-templates/${id}/activate`).then((r) => data<AgreementTemplate>(r)),

  remove: (id: string) => api.delete(`/admin/agreement-templates/${id}`),

  fileUrl: (id: string, kind: 'preview' | 'download') =>
    `${api.defaults.baseURL}/admin/agreement-templates/${id}/${kind}`,
}
