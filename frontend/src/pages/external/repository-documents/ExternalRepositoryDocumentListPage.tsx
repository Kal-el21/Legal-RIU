import RepositoryDocumentListPage from '@/components/shared/RepositoryDocumentListPage'

export default function ExternalRepositoryDocumentListPage() {
  return (
    <RepositoryDocumentListPage
      basePath="/external/repository-documents"
      title="Repositori Dokumen"
      description="Arsip dokumen dari pengajuan dan perkara yang dapat Anda akses."
    />
  )
}