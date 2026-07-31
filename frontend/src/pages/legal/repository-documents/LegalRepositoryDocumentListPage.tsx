import RepositoryDocumentListPage from '@/components/shared/RepositoryDocumentListPage'

export default function LegalRepositoryDocumentListPage() {
  return (
    <RepositoryDocumentListPage
      basePath="/legal/repository-documents"
      title="Dokumen Repository"
      description="Daftar dokumen per feature"
    />
  )
}