import { useEffect, useMemo, useState } from 'react'
import { CheckCircle2, Download, Eye, FileUp, Info, Upload } from 'lucide-react'
import {
  agreementTemplateService,
  type AgreementTemplate,
  type PlaceholderReference,
} from '@/services/agreement-template.service'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

const statusStyle: Record<AgreementTemplate['status'], string> = {
  ACTIVE: 'bg-green-50 text-green-700 border-green-200',
  DRAFT: 'bg-amber-50 text-amber-700 border-amber-200',
  ARCHIVED: 'bg-gray-100 text-gray-600 border-gray-200',
}

const statusLabel: Record<AgreementTemplate['status'], string> = {
  ACTIVE: 'Aktif',
  DRAFT: 'Draft',
  ARCHIVED: 'Diarsipkan',
}

function formatDate(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
}

interface Props {
  apiBase?: string
}

export default function AgreementTemplatePage({ apiBase = '/admin' }: Props) {
  const [items, setItems] = useState<AgreementTemplate[]>([])
  const [placeholders, setPlaceholders] = useState<PlaceholderReference[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isUploading, setIsUploading] = useState(false)
  const [activatingId, setActivatingId] = useState('')
  const [showReference, setShowReference] = useState(false)
  const [message, setMessage] = useState('')
  const [messageType, setMessageType] = useState<'success' | 'error' | ''>('')

  const [file, setFile] = useState<File | null>(null)
  const [name, setName] = useState('')
  const [effectiveDate, setEffectiveDate] = useState('')
  const [note, setNote] = useState('')

  const groupedPlaceholders = useMemo(() => {
    const groups = new Map<string, PlaceholderReference[]>()
    placeholders.forEach((item) => {
      const key = item.group || 'Lainnya'
      groups.set(key, [...(groups.get(key) || []), item])
    })
    return [...groups.entries()]
  }, [placeholders])

  const load = () =>
    agreementTemplateService.list(apiBase)
      .then((data) => setItems(data))
      .catch((error: Error) => { setMessage(error.message); setMessageType('error') })
      .finally(() => setIsLoading(false))

  useEffect(() => {
    void Promise.all([
      agreementTemplateService.list(apiBase).then((data) => setItems(data)),
      agreementTemplateService.placeholders(apiBase).then((data) => setPlaceholders(data)),
    ])
      .catch((error: Error) => { setMessage(error.message); setMessageType('error') })
      .finally(() => setIsLoading(false))
  }, [apiBase])

  const upload = async () => {
    if (!file) {
      setMessage('Pilih file template .docx terlebih dahulu')
      setMessageType('error')
      return
    }
    setIsUploading(true)
    setMessage('')
    setMessageType('')
    try {
      await agreementTemplateService.upload(apiBase, file, { name, note, effective_date: effectiveDate })
      setMessage('Template lolos uji coba dan tersimpan sebagai Draft. Tinjau pratinjaunya sebelum diaktifkan.')
      setMessageType('success')
      setFile(null)
      setName('')
      setEffectiveDate('')
      setNote('')
      await load()
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Gagal mengunggah template')
      setMessageType('error')
    } finally {
      setIsUploading(false)
    }
  }

  const activate = async (item: AgreementTemplate) => {
    if (!window.confirm(`Aktifkan "${item.name}" (versi ${item.version})? Versi aktif saat ini akan diarsipkan.`)) return
    setActivatingId(item.id)
    setMessage('')
    setMessageType('')
    try {
      await agreementTemplateService.activate(apiBase, item.id)
      setMessage(`Versi ${item.version} sekarang aktif untuk pengajuan baru.`)
      setMessageType('success')
      await load()
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Gagal mengaktifkan template')
      setMessageType('error')
    } finally {
      setActivatingId('')
    }
  }

  return <div className="p-6 max-w-7xl mx-auto">
    <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Template Perjanjian</h1>
    <p className="text-sm text-gray-500 mt-0.5 mb-6">
      Kelola versi template dokumen perjanjian. Versi lama tidak pernah dihapus agar dokumen yang sudah terbit tetap dapat ditelusuri.
    </p>

    {message && (
      <p className={`mb-4 text-sm ${messageType === 'error' ? 'text-red-600' : 'text-green-600'}`}>{message}</p>
    )}

    <div className="bg-white rounded-2xl border border-gray-100 p-6 mb-6">
      <h2 className="font-semibold flex items-center gap-2 mb-1" style={{ color: '#0B2545' }}>
        <FileUp className="w-4 h-4" /> Unggah Versi Baru
      </h2>
      <p className="text-sm text-gray-500 mb-4">
        File akan diuji otomatis: struktur .docx, penamaan placeholder, dan konversi ke PDF. Template yang gagal uji tidak akan tersimpan.
      </p>
      <div className="grid md:grid-cols-2 gap-4">
        <div>
          <Label className="mb-1.5 text-gray-700">File Template (.docx)</Label>
          <Input
            type="file"
            accept=".docx"
            onChange={(event) => setFile(event.target.files?.[0] || null)}
          />
        </div>
        <div>
          <Label className="mb-1.5 text-gray-700">Nama Versi</Label>
          <Input
            value={name}
            placeholder="mis. PKS - Revisi Klausul Denda"
            onChange={(event) => setName(event.target.value)}
          />
        </div>
        <div>
          <Label className="mb-1.5 text-gray-700">Tanggal Berlaku</Label>
          <Input type="date" value={effectiveDate} onChange={(event) => setEffectiveDate(event.target.value)} />
        </div>
        <div>
          <Label className="mb-1.5 text-gray-700">Catatan Perubahan</Label>
          <Textarea
            value={note}
            placeholder="Ringkas apa yang berubah pada versi ini"
            onChange={(event) => setNote(event.target.value)}
          />
        </div>
      </div>
      <Button
        onClick={() => void upload()}
        disabled={isUploading}
        className="mt-5 flex items-center gap-2 text-white transition hover:brightness-95"
        style={{ background: '#C8102E' }}
      >
        <Upload className="w-4 h-4" /> {isUploading ? 'Menguji template...' : 'Unggah & Uji'}
      </Button>
    </div>

    <div className="bg-white rounded-2xl border border-gray-100 p-6 mb-6">
      <button
        type="button"
        onClick={() => setShowReference((value) => !value)}
        className="font-semibold flex items-center gap-2 w-full text-left"
        style={{ color: '#0B2545' }}
      >
        <Info className="w-4 h-4" /> Referensi Placeholder
        <span className="ml-auto text-sm font-normal text-gray-400">
          {showReference ? 'Sembunyikan' : `Lihat ${placeholders.length} token`}
        </span>
      </button>
      <p className="text-sm text-gray-500 mt-1">
        Hanya token di bawah ini yang boleh dipakai. Ketik langsung di Word, huruf besar semua, diapit dua kurung kurawal.
      </p>
      {showReference && (
        <div className="mt-4 space-y-5">
          {groupedPlaceholders.map(([group, tokens]) => (
            <div key={group}>
              <h3 className="text-xs font-semibold uppercase tracking-wide text-gray-500 mb-2">{group}</h3>
              <div className="space-y-1.5">
                {tokens.map((token) => (
                  <div key={token.token} className="flex flex-wrap gap-x-3 gap-y-0.5 text-sm">
                    <code className="font-mono text-xs bg-gray-50 border border-gray-100 rounded px-1.5 py-0.5 text-gray-800">
                      {token.token}
                    </code>
                    <span className="text-gray-500">{token.description}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>

    <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
      <h2 className="font-semibold px-6 pt-6 pb-4" style={{ color: '#0B2545' }}>Riwayat Versi</h2>
      {isLoading ? (
        <div className="p-8 text-center text-gray-400">Memuat data...</div>
      ) : items.length === 0 ? (
        <div className="p-8 text-center text-gray-400">Belum ada template.</div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-xs uppercase tracking-wide text-gray-500 border-b border-gray-100">
                <th className="px-6 py-3 font-medium">Versi</th>
                <th className="px-6 py-3 font-medium">Nama</th>
                <th className="px-6 py-3 font-medium">Status</th>
                <th className="px-6 py-3 font-medium">Berlaku</th>
                <th className="px-6 py-3 font-medium">Diunggah</th>
                <th className="px-6 py-3 font-medium text-right">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.id} className="border-b border-gray-50 last:border-0">
                  <td className="px-6 py-4 font-medium text-gray-800">v{item.version}</td>
                  <td className="px-6 py-4">
                    <p className="text-gray-800">{item.name}</p>
                    <p className="text-xs text-gray-400 mt-0.5">{item.file_name}</p>
                    {item.note && <p className="text-xs text-gray-500 mt-1">{item.note}</p>}
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-block text-xs px-2 py-0.5 rounded-full border ${statusStyle[item.status]}`}>
                      {statusLabel[item.status]}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-gray-600">{formatDate(item.effective_date)}</td>
                  <td className="px-6 py-4 text-gray-600">
                    <p>{formatDate(item.created_at)}</p>
                    <p className="text-xs text-gray-400">{item.uploader?.full_name || 'Sistem'}</p>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center justify-end gap-2">
                      <a
                        href={agreementTemplateService.fileUrl(apiBase, item.id, 'preview')}
                        target="_blank"
                        rel="noreferrer"
                        title="Pratinjau PDF"
                        className="p-2 rounded-lg text-gray-500 hover:bg-gray-50 hover:text-gray-700"
                      >
                        <Eye className="w-4 h-4" />
                      </a>
                      <a
                        href={agreementTemplateService.fileUrl(apiBase, item.id, 'download')}
                        title="Unduh .docx"
                        className="p-2 rounded-lg text-gray-500 hover:bg-gray-50 hover:text-gray-700"
                      >
                        <Download className="w-4 h-4" />
                      </a>
                      {item.status !== 'ACTIVE' && (
                        <Button
                          onClick={() => void activate(item)}
                          disabled={activatingId === item.id}
                          className="flex items-center gap-1.5 h-8 px-3 text-xs text-white transition hover:brightness-95"
                          style={{ background: '#0B2545' }}
                        >
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          {activatingId === item.id ? 'Mengaktifkan...' : 'Aktifkan'}
                        </Button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  </div>
}
