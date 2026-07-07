import SharedReviewDocumentListPage from '@/components/shared/ReviewDocumentListPage'

export default function LegalAU_ReviewDocumentListPage() {
  return (
    <SharedReviewDocumentListPage
      basePath="/legal-au/review-documents"
      title="Review Dokumen"
      description="Review dan berikan masukan dokumen"
      linkLabel="Review →"
      viewPermission="document_review.view.all"
      createPermission="document_review.create.own"
    />
  )
}
