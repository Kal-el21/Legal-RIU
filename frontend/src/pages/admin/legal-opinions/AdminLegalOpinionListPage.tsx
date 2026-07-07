import SharedLegalOpinionListPage from '@/components/shared/LegalOpinionListPage'

export default function AdminLegalOpinionListPage() {
  return (
    <SharedLegalOpinionListPage
      basePath="/admin/legal-opinions"
      title="Manage Legal Opinion"
      description="Kelola seluruh pengajuan Legal Opinion"
      showColumnRequester={true}
      linkLabel="Kelola →"
      viewPermission="legal_opinion.view.all"
    />
  )
}
