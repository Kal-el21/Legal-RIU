import SharedReviewDocumentListPage from '@/components/shared/ReviewDocumentListPage'

export default function LegalReviewDocumentListPage() {
  return (
    <SharedReviewDocumentListPage
      basePath="/legal/review-documents"
      title="Review Dokumen"
      description="Review dan berikan masukan dokumen"
      linkLabel="Review →"
    />
  )
}
