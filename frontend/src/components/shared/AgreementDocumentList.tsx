import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { agreementService, type AgreementDocument } from '@/services/agreement-document.service'
import { Button } from '@/components/ui/button'

export default function AgreementDocumentList({ basePath, apiBase = '', requester = false }: { basePath: string; apiBase?: string; requester?: boolean }) {
  const [items, setItems] = useState<AgreementDocument[]>([])
  useEffect(() => { void agreementService.list(apiBase).then(setItems) }, [apiBase])
  return <div className="p-6 max-w-6xl mx-auto">
    <div className="flex justify-between mb-6"><div><h1 className="text-2xl font-bold text-[#0B2545]">Dokumen Perjanjian</h1><p className="text-sm text-gray-500">Pengajuan dan generation dokumen perjanjian</p></div>{requester && <Link to={`${basePath}/new`}><Button>Pengajuan Baru</Button></Link>}</div>
    <div className="bg-white rounded-xl border overflow-hidden"><table className="w-full text-sm"><thead className="bg-gray-50"><tr><th className="p-3 text-left">Ticket</th><th>Tipe</th><th>Nomor</th><th>Status</th><th>Aksi</th></tr></thead><tbody>{items.map((item) => <tr key={item.id} className="border-t"><td className="p-3">{item.ticket_number}</td><td>{item.document_type_code}</td><td>{item.agreement_number}</td><td>{item.status}</td><td className="space-x-3"><Link className="text-red-600" to={`${basePath}/${item.id}`}>Detail</Link>{requester && (item.status === 'SUBMITTED' || item.status === 'NEED_REVISION') && <Link className="text-blue-600" to={`${basePath}/${item.id}/edit`}>{item.status === 'NEED_REVISION' ? 'Revisi' : 'Edit'}</Link>}</td></tr>)}</tbody></table>{!items.length && <p className="p-8 text-center text-gray-400">Belum ada pengajuan.</p>}</div>
  </div>
}
