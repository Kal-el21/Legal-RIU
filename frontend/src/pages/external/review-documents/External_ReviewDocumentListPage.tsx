import SharedReviewDocumentListPage from '@/components/shared/ReviewDocumentListPage'

export default function External_ReviewDocumentListPage() {
  return (
    <SharedReviewDocumentListPage
      basePath="/external/review-documents"
      title="Review Dokumen"
      description="Review dan berikan masukan dokumen"
      linkLabel="Review →"
      viewPermission="document_review.view.own"
      createPermission="document_review.create.own"
    />
  )
}
