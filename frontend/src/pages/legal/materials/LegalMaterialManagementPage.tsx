import SharedMaterialListPage from '@/components/shared/MaterialListPage'

export default function LegalMaterialManagementPage() {
  return (
    <SharedMaterialListPage
      basePath="/legal/materials"
      title="Materi Legal"
      description="Kelola materi legal"
      showCreateButton={true}
      showEditButton={false}
    />
  )
}
