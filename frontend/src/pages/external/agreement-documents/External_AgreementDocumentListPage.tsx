import SharedAgreementDocumentListPage from '@/components/shared/AgreementDocumentListPage'

export default function External_AgreementDocumentListPage() {
  return (
    <SharedAgreementDocumentListPage
      basePath="/external/agreement-documents"
      title="Dokumen Perjanjian"
      description="Ajukan dan kelola dokumen perjanjian kerja sama"
      showCreateButton={true}
      createPath="/external/agreement-documents/new"
      viewPermission="agreement_document.view.own"
      createPermission="agreement_document.create.own"
    />
  )
}
