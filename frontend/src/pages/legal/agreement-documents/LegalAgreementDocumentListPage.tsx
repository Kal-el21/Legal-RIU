import SharedAgreementDocumentListPage from '@/components/shared/AgreementDocumentListPage'

export default function LegalAgreementDocumentListPage() {
  return (
    <SharedAgreementDocumentListPage
      basePath="/legal/agreement-documents"
      title="Dokumen Perjanjian"
      description="Review pengajuan dokumen perjanjian"
      linkLabel="Review →"
      viewPermission="agreement_document.view.all"
    />
  )
}
