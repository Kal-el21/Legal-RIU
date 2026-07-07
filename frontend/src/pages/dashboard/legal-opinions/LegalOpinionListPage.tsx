import SharedLegalOpinionListPage from '@/components/shared/LegalOpinionListPage'

export default function LegalOpinionListPage() {
  return (
    <SharedLegalOpinionListPage
      basePath="/dashboard/legal-opinions"
      title="Legal Opinion"
      description="Kelola pengajuan legal opinion Anda"
      showCreateButton={true}
      createPath="/dashboard/legal-opinions/new"
      showColumnRequester={false}
      linkLabel="Detail →"
      viewPermission="legal_opinion.view.own"
      createPermission="legal_opinion.create.own"
    />
  )
}
