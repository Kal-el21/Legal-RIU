import { useParams, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { ArrowLeft, FileText, MessageSquare, Upload, Download, CheckCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import StatusBadge from '@/components/common/StatusBadge'
import { useDocumentReview, useAdminUpdateDocumentReviewStatus } from '@/hooks/useDocumentReview'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { documentReviewService } from '@/services/document-review.service'
import { formatDateTime, formatFileSize, validateFile } from '@/lib/utils'
import type { SubmissionStatus } from '@/types'

const NEXT_STATUSES: Record<string, string[]> = {
  SUBMITTED: ['UNDER_REVIEW'],
  UNDER_REVIEW: ['NEED_REVISION', 'REJECTED', 'COMPLETED'],
  NEED_REVISION: ['UNDER_REVIEW'],
  REJECTED: ['UNDER_REVIEW'],
  RESUBMITTED: ['UNDER_REVIEW'],
  COMPLETED: [],
}

const STATUS_LABELS: Record<string, string> = {
  UNDER_REVIEW: 'Mulai Review',
  NEED_REVISION: 'Perlu Revisi',
  REJECTED: 'Tolak',
  COMPLETED: 'Selesai',
}

export default function AdminReviewDocumentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const { data: dr, isLoading } = useDocumentReview(id!)
  const updateStatus = useAdminUpdateDocumentReviewStatus()

  const [newStatus, setNewStatus] = useState('')
  const [adminNote, setAdminNote] = useState('')
  const [resultFile, setResultFile] = useState<File | null>(null)
  const [resultNotes, setResultNotes] = useState('')
  const [fileError, setFileError] = useState('')

  const uploadResultMutation = useMutation({
    mutationFn: ({ file, notes }: { file: File; notes: string }) =>
      documentReviewService.adminUploadResult(id!, file, notes),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['document-reviews', 'detail', id] })
      setResultFile(null)
      setResultNotes('')
    },
  })

  const handleStatusUpdate = async () => {
    if (!newStatus) return
    await updateStatus.mutateAsync({ id: id!, status: newStatus, admin_note: adminNote })
    setNewStatus('')
    setAdminNote('')
  }

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const err = await validateFile(file)
    if (err) { setFileError(err); return }
    setFileError('')
    setResultFile(file)
  }

  const handleDownload = async (filePath: string, fileName: string) => {
    const url = await documentReviewService.getPresignedURL(filePath)
    const a = document.createElement('a')
    a.href = url; a.download = fileName; a.target = '_blank'; a.click()
  }

  if (isLoading) return <div className="p-12 text-center text-gray-400">Memuat data...</div>
  if (!dr) return <div className="p-12 text-center text-gray-500">Pengajuan tidak ditemukan</div>

  const status = dr.status as SubmissionStatus
  const nextOptions = NEXT_STATUSES[dr.status] ?? []

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="flex items-start gap-3 mb-8">
        <button onClick={() => navigate(-1)} className="p-2 rounded-lg hover:bg-gray-100 mt-0.5">
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </button>
        <div className="flex-1">
          <div className="flex items-center gap-3 flex-wrap">
            <span className="text-xs font-mono font-medium bg-gray-100 text-gray-600 px-2 py-1 rounded">
              {dr.ticket_number}
            </span>
            <StatusBadge status={status} />
          </div>
          <h1 className="text-xl font-bold mt-2" style={{ color: '#0B2545' }}>{dr.document_name}</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Oleh: {dr.requestor_name} · {dr.requestor_division} · {formatDateTime(dr.created_at)}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-5">
          <Card title="Informasi Pemohon">
            <Grid>
              <Info label="Nama Lengkap" value={dr.requestor_name} />
              <Info label="Posisi Jabatan" value={dr.requestor_position} />
              <Info label="Divisi" value={dr.requestor_division} />
              <Info label="Email" value={dr.requestor_email} />
              <Info label="WhatsApp" value={dr.requestor_phone} />
            </Grid>
          </Card>

          <Card title="Detail Dokumen">
            <Grid>
              <Info label="Nama Dokumen" value={dr.document_name} />
              <Info label="Jenis Dokumen" value={dr.document_type === 'Lain-Lain' && dr.document_type_other ? dr.document_type_other : dr.document_type} />
              <Info label="Pihak Kedua" value={dr.second_party} />
              {dr.third_party && <Info label="Pihak Ketiga" value={dr.third_party} />}
            </Grid>
            {dr.additional_note && (
              <div className="mt-4 pt-4 border-t border-gray-100">
                <p className="text-xs text-gray-400 mb-1">Keterangan Tambahan</p>
                <p className="text-sm text-gray-800 whitespace-pre-wrap">{dr.additional_note}</p>
              </div>
            )}
          </Card>

          {(dr.attachments?.length ?? 0) > 0 && (
            <Card title="Draft Perjanjian">
              <div className="space-y-2">
                {dr.attachments?.map((att) => (
                  <div key={att.id} className="flex items-center gap-3 p-3 rounded-lg bg-gray-50 border border-gray-100">
                    <FileText className="w-4 h-4 text-gray-400 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-gray-700 truncate">{att.file_name}</p>
                      <p className="text-xs text-gray-400">{formatFileSize(att.file_size)} · Round {att.upload_round}</p>
                    </div>
                    <button onClick={() => handleDownload(att.file_path, att.file_name)}
                      className="p-1.5 rounded-lg hover:bg-gray-200 text-gray-400 hover:text-gray-600">
                      <Download className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            </Card>
          )}

          {(dr.results?.length ?? 0) > 0 && (
            <Card title="Hasil Review Terupload">
              <div className="space-y-2">
                {dr.results?.map((res) => (
                  <div key={res.id} className="flex items-center gap-3 p-3 rounded-lg bg-green-50 border border-green-100">
                    <CheckCircle className="w-4 h-4 text-green-500 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-gray-700 truncate">{res.file_name}</p>
                      {res.notes && <p className="text-xs text-gray-500 mt-0.5">{res.notes}</p>}
                    </div>
                    <button onClick={() => handleDownload(res.file_path, res.file_name)}
                      className="p-1.5 rounded-lg hover:bg-green-200 text-green-500">
                      <Download className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            </Card>
          )}
        </div>

        {/* Right — actions */}
        <div className="space-y-5">
          {dr.admin_note && (
            <Card title="Catatan Sebelumnya">
              <div className="flex items-start gap-2">
                <MessageSquare className="w-4 h-4 text-orange-400 mt-0.5 flex-shrink-0" />
                <p className="text-sm text-gray-700">{dr.admin_note}</p>
              </div>
            </Card>
          )}

          {nextOptions.length > 0 && (
            <Card title="Ubah Status">
              <div className="space-y-4">
                <div className="space-y-1.5">
                  <Label className="text-xs text-gray-500">Status Baru</Label>
                  <Select onValueChange={setNewStatus} value={newStatus}>
                    <SelectTrigger><SelectValue placeholder="Pilih status..." /></SelectTrigger>
                    <SelectContent>
                      {nextOptions.map((s) => (
                        <SelectItem key={s} value={s}>{STATUS_LABELS[s] ?? s}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5">
                  <Label className="text-xs text-gray-500">Catatan Admin (opsional)</Label>
                  <Textarea value={adminNote} onChange={(e) => setAdminNote(e.target.value)}
                    placeholder="Tambahkan catatan untuk pemohon..." rows={3} />
                </div>
                {updateStatus.isError && <p className="text-xs text-red-500">{(updateStatus.error as Error)?.message}</p>}
                {updateStatus.isSuccess && <p className="text-xs text-green-600">Status berhasil diubah!</p>}
                <Button onClick={handleStatusUpdate} disabled={!newStatus || updateStatus.isPending}
                  className="w-full text-white" style={{ background: '#0B2545' }}>
                  {updateStatus.isPending ? 'Menyimpan...' : 'Simpan Perubahan'}
                </Button>
              </div>
            </Card>
          )}

          <Card title="Upload Hasil Review">
            <div className="space-y-3">
              <label className="flex flex-col items-center gap-2 p-5 border-2 border-dashed border-gray-200 rounded-xl cursor-pointer hover:bg-gray-50 transition-colors">
                <Upload className="w-5 h-5 text-gray-400" />
                <p className="text-xs text-gray-500 text-center">
                  {resultFile ? resultFile.name : 'Klik untuk pilih file hasil review'}
                </p>
                <input type="file" accept=".pdf,.doc,.docx" className="hidden" onChange={handleFileChange} />
              </label>
              {fileError && <p className="text-xs text-red-500">{fileError}</p>}
              <div className="space-y-1.5">
                <Label className="text-xs text-gray-500">Catatan Hasil (opsional)</Label>
                <Textarea value={resultNotes} onChange={(e) => setResultNotes(e.target.value)}
                  placeholder="Catatan hasil review..." rows={2} />
              </div>
              {uploadResultMutation.isSuccess && <p className="text-xs text-green-600">File berhasil diupload!</p>}
              <Button onClick={() => uploadResultMutation.mutateAsync({ file: resultFile!, notes: resultNotes })}
                disabled={!resultFile || uploadResultMutation.isPending}
                className="w-full text-white" style={{ background: '#C8102E' }}>
                {uploadResultMutation.isPending ? 'Mengupload...' : 'Upload Hasil'}
              </Button>
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}

function Card({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-white rounded-2xl border border-gray-100 p-5">
      <h3 className="text-sm font-semibold mb-4" style={{ color: '#0B2545' }}>{title}</h3>
      {children}
    </div>
  )
}
function Grid({ children }: { children: React.ReactNode }) {
  return <div className="grid grid-cols-2 gap-3">{children}</div>
}
function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs text-gray-400">{label}</p>
      <p className="text-sm text-gray-800 font-medium mt-0.5">{value}</p>
    </div>
  )
}