import DraftListPage from '@/components/shared/DraftListPage'
import { useDeleteDocumentReview, useDocumentReviews } from '@/hooks/useDocumentReview'
import { judulDraf } from '@/lib/draf'

export default function ReviewDocumentDraftListPage() {
  // Catatan: hanya 100 draf terakhir diubah yang tampil; tambah halaman kalau ada yang lebih dari itu.
  const { data, isLoading } = useDocumentReviews({ status: 'DRAFT', limit: 100 })
  const hapus = useDeleteDocumentReview()
  return (
    <DraftListPage
      judul="Draf Review Dokumen"
      kembali="/dashboard/review-documents"
      izin="document_review.create.own"
      isLoading={isLoading}
      items={data?.items.map((dr) => ({ id: dr.id, judul: judulDraf(dr.document_name), diubah: dr.updated_at }))}
      lanjutkan={(id) => `/dashboard/review-documents/drafts/${id}`}
      onHapus={(id) => hapus.mutate(id)}
    />
  )
}
