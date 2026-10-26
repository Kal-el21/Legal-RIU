import RepositoryDocumentListPage from '@/components/shared/RepositoryDocumentListPage'

export default function UserRepositoryDocumentListPage() {
  return (
    <RepositoryDocumentListPage
      basePath="/dashboard/repository-documents"
      title="Repositori Dokumen"
      description="Arsip dokumen dari pengajuan dan perkara yang dapat Anda akses."
    />
  )
}