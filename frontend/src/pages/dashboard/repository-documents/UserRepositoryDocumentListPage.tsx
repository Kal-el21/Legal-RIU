import RepositoryDocumentListPage from '@/components/shared/RepositoryDocumentListPage'

export default function UserRepositoryDocumentListPage() {
  return (
    <RepositoryDocumentListPage
      basePath="/dashboard/repository-documents"
      title="Dokumen Repository"
      description="Daftar dokumen per feature"
    />
  )
}