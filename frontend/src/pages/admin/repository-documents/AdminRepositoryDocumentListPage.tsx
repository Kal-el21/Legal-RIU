import RepositoryDocumentListPage from '@/components/shared/RepositoryDocumentListPage'

export default function AdminRepositoryDocumentListPage() {
  return (
    <RepositoryDocumentListPage
      basePath="/admin/repository-documents"
      title="Repositori Dokumen"
      description="Arsip dokumen dari pengajuan dan perkara yang dapat Anda akses."
    />
  )
}