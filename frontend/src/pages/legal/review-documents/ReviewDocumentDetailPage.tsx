import { useParams } from 'react-router-dom'
import { useState, useMemo } from 'react'
import { Upload } from 'lucide-react'
import { useDocumentReview, useLegalUpdateDocumentReviewStatus } from '@/hooks/useDocumentReview'
import { useAuthStore } from '@/store/auth.store'
import { useQueryClient, useMutation } from '@tanstack/react-query'
import SharedReviewDocumentDetailPage from '@/components/shared/ReviewDocumentDetailPage'
import { validateFile } from '@/lib/utils'
import { documentReviewService } from '@/services/document-review.service'

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

export default function LegalReviewDocumentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const qc = useQueryClient()
  const hasPermission = useAuthStore((state) => state.hasPermission)

  const { data: dr } = useDocumentReview(id!)
  const updateStatus = useLegalUpdateDocumentReviewStatus()
  const canUpdateStatus = hasPermission('document_review.update_status.all')
  const canUploadResult = hasPermission('document_review.upload_result.all')

  const nextOptions = useMemo(() => {
    return NEXT_STATUSES[dr?.status ?? ''] ?? []
  }, [dr?.status])

  const [newStatus, setNewStatus] = useState('')
  const [adminNote, setAdminNote] = useState('')
  const [resultFile, setResultFile] = useState<File | null>(null)
  const [resultNotes, setResultNotes] = useState('')
  const [fileError, setFileError] = useState('')

  const uploadResultMutation = useMutation({
    mutationFn: ({ file, notes }: { file: File; notes: string }) =>
      documentReviewService.legalUploadResult(id!, file, notes),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['document-reviews', 'detail', id] })
      setResultFile(null)
      setResultNotes('')
    },
  })

  const handleStatusUpdate = async () => {
    if (!newStatus || !id) return
    await updateStatus.mutateAsync({ id, status: newStatus, admin_note: adminNote })
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

  const handleUploadResult = async () => {
    if (!resultFile) return
    await uploadResultMutation.mutateAsync({ file: resultFile, notes: resultNotes })
  }

  const actionCards = (
    <>
      {canUpdateStatus && nextOptions.length > 0 && (
        <div className="bg-white rounded-2xl border border-gray-100 p-6">
          <h3 className="text-sm font-semibold mb-4" style={{ color: '#0B2545' }}>Ubah Status</h3>
          <div className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-xs text-gray-500">Status Baru</label>
              <select value={newStatus} onChange={(e) => setNewStatus(e.target.value)}
                className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm">
                <option value="">Pilih status...</option>
                {nextOptions.map((s) => (
                  <option key={s} value={s}>{STATUS_LABELS[s] ?? s}</option>
                ))}
              </select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs text-gray-500">Catatan Admin (opsional)</label>
              <textarea
                value={adminNote}
                onChange={(e) => setAdminNote(e.target.value)}
                placeholder="Tambahkan catatan untuk pemohon..."
                rows={3}
                className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm"
              />
            </div>
            {updateStatus.isError && <p className="text-xs text-red-500">{(updateStatus.error as Error)?.message}</p>}
            {updateStatus.isSuccess && <p className="text-xs text-green-600">Status berhasil diubah!</p>}
            <button
              onClick={handleStatusUpdate}
              disabled={!newStatus || updateStatus.isPending}
              className="w-full text-white text-sm py-2.5 rounded-lg disabled:opacity-50"
              style={{ background: '#0B2545' }}
            >
              {updateStatus.isPending ? 'Menyimpan...' : 'Simpan Perubahan'}
            </button>
          </div>
        </div>
      )}

      {canUploadResult && (
        <div className="bg-white rounded-2xl border border-gray-100 p-6">
          <h3 className="text-sm font-semibold mb-4" style={{ color: '#0B2545' }}>Upload Hasil Review</h3>
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
              <label className="text-xs text-gray-500">Catatan Hasil (opsional)</label>
              <textarea
                value={resultNotes}
                onChange={(e) => setResultNotes(e.target.value)}
                placeholder="Catatan hasil review..."
                rows={2}
                className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm"
              />
            </div>
            {uploadResultMutation.isSuccess && <p className="text-xs text-green-600">File berhasil diupload!</p>}
            <button
              onClick={handleUploadResult}
              disabled={!resultFile || uploadResultMutation.isPending}
              className="w-full text-white text-sm py-2.5 rounded-lg disabled:opacity-50"
              style={{ background: '#C8102E' }}
            >
              {uploadResultMutation.isPending ? 'Mengupload...' : 'Upload Hasil'}
            </button>
          </div>
        </div>
      )}
    </>
  )

  return (
    <SharedReviewDocumentDetailPage>
      {actionCards}
    </SharedReviewDocumentDetailPage>
  )
}
