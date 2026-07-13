import SharedAgreementDocumentListPage from '@/components/shared/AgreementDocumentListPage'

export default function AdminAgreementDocumentListPage() {
  return (
    <SharedAgreementDocumentListPage
      basePath="/admin/agreement-documents"
      title="Dokumen Perjanjian"
      description="Kelola seluruh pengajuan dokumen perjanjian"
      linkLabel="Review →"
      viewPermission="agreement_document.view.all"
    />
  )
}
