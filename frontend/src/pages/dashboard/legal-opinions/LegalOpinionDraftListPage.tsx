import DraftListPage from '@/components/shared/DraftListPage'
import { useDeleteLegalOpinion, useLegalOpinions } from '@/hooks/useLegalOpinion'
import { judulDraf } from '@/lib/draf'

export default function LegalOpinionDraftListPage() {
  // Catatan: hanya 100 draf terakhir diubah yang tampil; tambah halaman kalau ada yang lebih dari itu.
  const { data, isLoading } = useLegalOpinions({ status: 'DRAFT', limit: 100 })
  const hapus = useDeleteLegalOpinion()
  return (
    <DraftListPage
      judul="Draf Legal Opinion"
      kembali="/dashboard/legal-opinions"
      izin="legal_opinion.create.own"
      isLoading={isLoading}
      items={data?.items.map((lo) => ({ id: lo.id, judul: judulDraf(lo.title), diubah: lo.updated_at }))}
      lanjutkan={(id) => `/dashboard/legal-opinions/drafts/${id}`}
      onHapus={(id) => hapus.mutate(id)}
    />
  )
}
