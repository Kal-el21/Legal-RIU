import SharedAgreementDocumentDetail from '@/components/shared/AgreementDocumentDetail'

export default function AgreementDocumentDetailPage() {
  return (
    <SharedAgreementDocumentDetail
      basePath="/dashboard/agreement-documents"
      canEdit={true}
    />
  )
}
