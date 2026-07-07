import SharedMaterialListPage from '@/components/shared/MaterialListPage'

export default function ExternalMaterialManagementPage() {
  return (
    <SharedMaterialListPage
      role="EXTERNAL"
      basePath="/external/materials"
      title="Materi Legal"
      description="Kelola materi legal"
      showCreateButton={true}
      showEditButton={true}
    />
  )
}
