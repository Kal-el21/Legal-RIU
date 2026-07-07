import SharedMaterialListPage from '@/components/shared/MaterialListPage'

export default function DashboardMaterialManagementPage() {
  return (
    <SharedMaterialListPage
      role="USER"
      basePath="/dashboard/materials"
      title="Materi Legal"
      description="Kelola materi legal"
      showCreateButton={true}
      showEditButton={true}
    />
  )
}
