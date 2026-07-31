import RepositoryDocumentListPage from '@/components/shared/RepositoryDocumentListPage'

export default function AdminRepositoryDocumentListPage() {
  return (
    <RepositoryDocumentListPage
      basePath="/admin/repository-documents"
      title="Dokumen Repository"
      description="Daftar dokumen per feature"
    />
  )
}