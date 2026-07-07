import SharedLegalOpinionListPage from '@/components/shared/LegalOpinionListPage'

export default function External_OpinionListPage() {
  return (
    <SharedLegalOpinionListPage
      basePath="/external/legal-opinions"
      title="Legal Opinion"
      description="Review dan berikan opinion hukum"
      showColumnRequester={true}
      linkLabel="Review →"
      viewPermission="legal_opinion.view.own"
      createPermission="legal_opinion.create.own"
    />
  )
}
