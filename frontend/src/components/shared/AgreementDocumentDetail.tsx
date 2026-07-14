import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { agreementService, type AgreementDocument } from '@/services/agreement-document.service'
import { Button } from '@/components/ui/button'

interface Props { apiBase?: string; approver?: boolean }

export default function AgreementDocumentDetail({ apiBase = '', approver = false }: Props) {
  const { id } = useParams()
  const [document, setDocument] = useState<AgreementDocument>()
  const [note, setNote] = useState('')
  const [meta, setMeta] = useState<Record<string, string>>({})
  const [error, setError] = useState('')
  const [previewUrl, setPreviewUrl] = useState('')
  const [previewError, setPreviewError] = useState('')
  const [previewRevision, setPreviewRevision] = useState(0)

  const load = useCallback(async () => {
    if (!id) return
    try {
      const value = await agreementService.get(id, apiBase)
      setDocument(value)
      setMeta({
        agreement_number: value.agreement_number,
        signing_place: String(value.form_data.tempat_ttd || ''),
        signing_date: String(value.form_data.tanggal_ttd || ''),
        party_one_signatory_name: String(value.form_data.pihak_pertama_pejabat || ''),
        party_one_signatory_position: String(value.form_data.pihak_pertama_jabatan || ''),
      })
      setError('')
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Gagal memuat pengajuan')
    }
  }, [apiBase, id])

  useEffect(() => { void load() }, [load])

  useEffect(() => {
    if (!id) return

    let active = true
    let objectUrl = ''
    setPreviewUrl('')
    setPreviewError('')

    void agreementService.preview(apiBase, id)
      .then((blob) => {
        if (!active) return
        objectUrl = URL.createObjectURL(blob)
        setPreviewUrl(objectUrl)
      })
      .catch((reason: unknown) => {
        if (!active) return
        setPreviewError(reason instanceof Error ? reason.message : 'Gagal memuat preview PDF')
      })

    return () => {
      active = false
      if (objectUrl) URL.revokeObjectURL(objectUrl)
    }
  }, [apiBase, id, previewRevision])

  const changeStatus = async (status: string) => {
    if (!id) return
    try {
      if (status === 'COMPLETED') await agreementService.meta(apiBase, id, meta)
      await agreementService.status(apiBase, id, status, note)
      await load()
      setPreviewRevision((value) => value + 1)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Gagal memproses pengajuan')
    }
  }

  if (error && !document) return <div className="p-6 text-red-600">{error}</div>
  if (!document) return <p className="p-6">Memuat...</p>

  return <div className="p-6 max-w-6xl mx-auto">
    <h1 className="text-2xl font-bold text-[#0B2545]">{document.ticket_number}</h1>
    <p className="mb-5">{document.status} · {document.agreement_number}</p>
    {error && <div className="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</div>}
    <div className="grid lg:grid-cols-2 gap-5">
      <div className="bg-white border rounded-xl p-5">
        <h2 className="font-semibold mb-3">Data Pengajuan</h2>
        {Object.entries(document.form_data).map(([key, value]) => <div key={key} className="grid grid-cols-2 text-sm py-1 border-b"><span className="text-gray-500">{key.split('_').join(' ')}</span><span>{String(value)}</span></div>)}
        <h2 className="font-semibold mt-5">Attachment</h2>
        {document.attachments?.map((attachment) => <a key={attachment.id} className="block text-red-600" href={`${agreementService.fileUrl(apiBase, id!, 'preview').replace('/preview', `/attachments/${attachment.id}`)}`}>{attachment.file_name}</a>)}
      </div>
      <div>
        {approver && <div className="bg-white border rounded-xl p-5 mb-4">
          <h2 className="font-semibold mb-3">Data Approver</h2>
          {Object.entries(meta).map(([key, value]) => <label className="block mb-2" key={key}><span className="text-xs">{key.split('_').join(' ')}</span><input type={key === 'signing_date' ? 'date' : 'text'} className="w-full border rounded p-2" value={value} onChange={(event) => setMeta({ ...meta, [key]: event.target.value })} /></label>)}
          <textarea className="w-full border rounded p-2" placeholder="Catatan revisi/penolakan" value={note} onChange={(event) => setNote(event.target.value)} />
          <div className="flex gap-2 flex-wrap mt-3"><Button onClick={() => void changeStatus('UNDER_REVIEW')}>Mulai Review</Button><Button onClick={() => void changeStatus('NEED_REVISION')}>Kembalikan</Button><Button onClick={() => void changeStatus('REJECTED')}>Tolak</Button><Button onClick={() => void changeStatus('COMPLETED')}>Approve</Button></div>
        </div>}
        {previewError && <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">Preview PDF gagal dimuat: {previewError}</div>}
        {!previewError && !previewUrl && <div className="h-[650px] rounded-xl border bg-gray-50 flex items-center justify-center text-sm text-gray-500">Menyiapkan preview PDF...</div>}
        {previewUrl && <iframe title="preview" className="w-full h-[650px] border rounded-xl" src={previewUrl} />}
        {document.status === 'COMPLETED' && <div className="flex gap-2 mt-3"><a href={agreementService.fileUrl(apiBase, id!, 'pdf')}><Button>Download PDF</Button></a>{approver && <a href={agreementService.fileUrl(apiBase, id!, 'docx')}><Button>Download DOCX</Button></a>}</div>}
      </div>
    </div>
  </div>
}
