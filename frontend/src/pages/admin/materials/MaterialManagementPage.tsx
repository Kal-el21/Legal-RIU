import SharedMaterialListPage from '@/components/shared/MaterialListPage'
import ImportCard from '@/components/common/ImportCard'
import { useImportLegalMaterials } from '@/hooks/useMaterial'
import { materialService } from '@/services/material.service'

export default function MaterialManagementPage() {
  const importMutation = useImportLegalMaterials()

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <ImportCard
        title="Impor dari Excel"
        onImport={(file) => importMutation.mutateAsync(file)}
        onDownloadTemplate={() => materialService.downloadTemplate()}
      />
      <SharedMaterialListPage
        role="ADMIN"
        basePath="/admin/materials"
        title="Materi Legal"
        description="Kelola materi legal"
        showCreateButton={true}
        showEditButton={true}
      />
    </div>
  )
}
