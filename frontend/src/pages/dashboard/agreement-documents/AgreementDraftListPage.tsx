import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import DraftListPage from '@/components/shared/DraftListPage'
import { agreementService, type AgreementDocument } from '@/services/agreement-document.service'
import { judulDraf } from '@/lib/draf'

export default function AgreementDraftListPage() {
  const [items, setItems] = useState<AgreementDocument[]>()
  const [muatUlang, setMuatUlang] = useState(0)

  // Catatan: hanya 100 draf terakhir diubah yang tampil; tambah halaman kalau ada yang lebih dari itu.
  useEffect(() => {
    agreementService.list('', { status: 'DRAFT', limit: 100 })
      .then(setItems)
      .catch((e: Error) => {
        setItems([])
        toast.error(e.message)
      })
  }, [muatUlang])

  const hapus = async (id: string) => {
    try {
      await agreementService.remove(id)
      toast.success('Draf dihapus')
      setMuatUlang((n) => n + 1)
    } catch (e) {
      toast.error((e as Error).message || 'Gagal menghapus draf')
    }
  }

  return (
    <DraftListPage
      judul="Draf Perjanjian"
      kembali="/dashboard/agreement-documents"
      izin="agreement_document.create.own"
      isLoading={!items}
      items={items?.map((d) => ({
        id: d.id,
        judul: `${d.document_type_code} · ${judulDraf(String(d.form_data?.pihak_kedua_nama ?? ''))}`,
        diubah: d.updated_at ?? d.created_at ?? '',
      }))}
      lanjutkan={(id) => `/dashboard/agreement-documents/drafts/${id}`}
      onHapus={(id) => void hapus(id)}
    />
  )
}
