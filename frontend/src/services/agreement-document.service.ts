import api from './api'
import axios from 'axios'
export interface AgreementField {name:string;label:string;type:string;required:boolean}
export interface AgreementSchema {code:string;name:string;sections:{title:string;fields:AgreementField[]}[]}
export interface AgreementDocument {id:string;ticket_number:string;document_type_code:string;form_data:Record<string,unknown>;agreement_number:string;status:string;approver_note?:string;attachments?:{id:string;file_name:string}[];generated_file_name?:string;created_at?:string;user?:{full_name?:string}}
const data=<T>(r:{data:{data:T}})=>r.data.data

async function fileError(reason: unknown, fallback: string): Promise<Error> {
  if (!axios.isAxiosError(reason)) return reason instanceof Error ? reason : new Error(fallback)

  const responseData = reason.response?.data
  if (responseData instanceof Blob) {
    try {
      const payload = JSON.parse(await responseData.text()) as { message?: string }
      if (payload.message) return new Error(payload.message)
    } catch {
      // Respons bukan JSON; gunakan pesan HTTP dari Axios.
    }
  }

  return new Error(reason.message || fallback)
}

export const agreementService={
 types:()=>api.get('/agreement-document-types').then(r=>data<{code:string;name:string}[]>(r)),
 schema:(code='PKS')=>api.get(`/agreement-document-types/${code}/schema`).then(r=>data<{schema:AgreementSchema}>(r).schema),
  list:(base='',query?:Record<string,unknown>)=>api.get(`${base}/agreement-documents`, query ? {params:query} : undefined).then(r=>data<{items:AgreementDocument[]}>(r).items),
 get:(id:string,base='')=>api.get(`${base}/agreement-documents/${id}`).then(r=>data<AgreementDocument>(r)),
 create:(payload:unknown,files:File[])=>{const f=new FormData();f.append('data',JSON.stringify(payload));files.forEach(x=>f.append('attachments',x));return api.post('/agreement-documents',f).then(r=>data<AgreementDocument>(r))},
 update:(id:string,form_data:Record<string,unknown>)=>api.put(`/agreement-documents/${id}`,{form_data}),
 remove:(id:string)=>api.delete(`/agreement-documents/${id}`),
 resubmit:(id:string,files:File[])=>{const f=new FormData();files.forEach(x=>f.append('attachments',x));return api.post(`/agreement-documents/${id}/resubmit`,f)},
 meta:(base:string,id:string,payload:unknown)=>api.patch(`${base}/agreement-documents/${id}/meta`,payload),
  status:(base:string,id:string,status:string,note='')=>api.patch(`${base}/agreement-documents/${id}/status`,{status,note}),
  preview:async(base:string,id:string)=>{
    try {
      const response = await api.get<Blob>(`${base}/agreement-documents/${id}/preview`, { responseType: 'blob' })
      const contentType = String(response.headers['content-type'] || response.data.type || '')
      if (!contentType.toLowerCase().includes('application/pdf')) {
        throw new Error('Server tidak mengembalikan file PDF')
      }
      return response.data
    } catch (reason) {
      throw await fileError(reason, 'Gagal memuat preview PDF')
    }
  },
 fileUrl:(base:string,id:string,kind:'preview'|'pdf'|'docx')=>`${api.defaults.baseURL}${base}/agreement-documents/${id}/${kind}`,
 master:()=>api.get('/admin/agreement-company-master').then(r=>data<Record<string,string>>(r)),
 saveMaster:(v:Record<string,string>)=>api.put('/admin/agreement-company-master',v),
}
