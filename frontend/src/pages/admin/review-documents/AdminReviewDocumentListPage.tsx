import SharedReviewDocumentListPage from '@/components/shared/ReviewDocumentListPage'

export default function AdminReviewDocumentListPage() {
  return (
    <SharedReviewDocumentListPage
      basePath="/admin/review-documents"
      title="Manage Review Dokumen"
      description="Kelola seluruh pengajuan Review Dokumen"
      linkLabel="Kelola →"
    />
  )
}
