import SharedLegalOpinionListPage from '@/components/shared/LegalOpinionListPage'

export default function LegalAU_OpinionListPage() {
  return (
    <SharedLegalOpinionListPage
      basePath="/legal-au/legal-opinions"
      title="Legal Opinion"
      description="Review dan berikan opinion hukum"
      showColumnRequester={true}
      linkLabel="Review →"
      viewPermission="legal_opinion.view.all"
      createPermission="legal_opinion.create.own"
    />
  )
}
