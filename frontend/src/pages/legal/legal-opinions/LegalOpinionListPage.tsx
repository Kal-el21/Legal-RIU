import SharedLegalOpinionListPage from '@/components/shared/LegalOpinionListPage'

export default function LegalLegalOpinionListPage() {
  return (
    <SharedLegalOpinionListPage
      basePath="/legal/legal-opinions"
      title="Legal Opinion"
      description="Review dan berikan opinion hukum"
      showColumnRequester={true}
      linkLabel="Review →"
    />
  )
}
