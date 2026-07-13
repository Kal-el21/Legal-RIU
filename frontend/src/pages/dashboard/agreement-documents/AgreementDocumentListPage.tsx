import SharedAgreementDocumentListPage from '@/components/shared/AgreementDocumentListPage'

export default function AgreementDocumentListPage() {
  return (
    <SharedAgreementDocumentListPage
      basePath="/dashboard/agreement-documents"
      title="Dokumen Perjanjian"
      description="Ajukan dan kelola dokumen perjanjian kerja sama"
      showCreateButton={true}
      createPath="/dashboard/agreement-documents/new"
      viewPermission="agreement_document.view.own"
      createPermission="agreement_document.create.own"
    />
  )
}
