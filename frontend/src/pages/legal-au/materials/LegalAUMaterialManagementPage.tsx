import SharedMaterialListPage from '@/components/shared/MaterialListPage'

export default function LegalAUMaterialManagementPage() {
  return (
    <SharedMaterialListPage
      role="LEGAL_AU"
      basePath="/legal-au/materials"
      title="Materi Legal"
      description="Kelola materi legal"
      showCreateButton={true}
      showEditButton={false}
    />
  )
}