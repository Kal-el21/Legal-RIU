import SharedReviewDocumentListPage from '@/components/shared/ReviewDocumentListPage'

export default function ReviewDocumentListPage() {
  return (
    <SharedReviewDocumentListPage
      basePath="/dashboard/review-documents"
      title="Review Dokumen"
      description="Kelola pengajuan review dokumen Anda"
      showCreateButton={true}
      createPath="/dashboard/review-documents/new"
      linkLabel="Detail →"
    />
  )
}
