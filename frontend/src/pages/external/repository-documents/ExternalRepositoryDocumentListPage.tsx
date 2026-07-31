import RepositoryDocumentListPage from '@/components/shared/RepositoryDocumentListPage'

export default function ExternalRepositoryDocumentListPage() {
  return (
    <RepositoryDocumentListPage
      basePath="/external/repository-documents"
      title="Dokumen Repository"
      description="Daftar dokumen per feature"
    />
  )
}