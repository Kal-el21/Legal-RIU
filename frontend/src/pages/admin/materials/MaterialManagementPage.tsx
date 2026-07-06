import SharedMaterialListPage from '@/components/shared/MaterialListPage'

export default function MaterialManagementPage() {
  return (
    <SharedMaterialListPage
      role="ADMIN"
      basePath="/admin/materials"
      title="Materi Legal"
      description="Kelola materi legal"
      showCreateButton={true}
      showEditButton={true}
    />
  )
}
