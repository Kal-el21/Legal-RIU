import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Download, Edit, FileText, FileDown, Plus, Trash2, Upload } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { useCreateCaseChronology, useDeleteCaseChronology, useDeleteLegalCase, useLegalCase, useAdminDownloadPDF } from '@/hooks/useLegalCase'
import { getLegalCaseRouteBase, legalCaseService } from '@/services/legal-case.service'
import { formatCurrency, formatDate, formatDateTime, formatFileSize, validateFile } from '@/lib/utils'
import { useAuthStore } from '@/store/auth.store'
import type { CaseChronology } from '@/types'
import LegalCaseFormDialog from './components/LegalCaseFormDialog'

export default function AdminLegalCaseDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { data: legalCase, isLoading } = useLegalCase(id!)
  const deleteCase = useDeleteLegalCase()
  const createChronology = useCreateCaseChronology(id!)
  const deleteChronology = useDeleteCaseChronology(id!)
  const downloadPDF = useAdminDownloadPDF()
  const caseRouteBase = getLegalCaseRouteBase()
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const canDelete = hasPermission('case_management.delete')
  const canManageDocument = hasPermission('case_management.manage_document')
  const canManageChronology = hasPermission('case_management.manage_chronology')

  const [editOpen, setEditOpen] = useState(false)
  const [agendaDate, setAgendaDate] = useState('')
  const [agenda, setAgenda] = useState('')
  const [description, setDescription] = useState('')
  const [files, setFiles] = useState<File[]>([])
  const [fileError, setFileError] = useState('')

  const handleDeleteCase = async () => {
    if (!id || !window.confirm('Hapus kasus hukum ini?')) return
    await deleteCase.mutateAsync(id)
    navigate(caseRouteBase)
  }

  const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const selected = Array.from(event.target.files ?? [])
    const validFiles: File[] = []

    for (const file of selected) {
      const error = await validateFile(file)
      if (error) {
        setFileError(`${file.name}: ${error}`)
        event.target.value = ''
        return
      }
      validFiles.push(file)
    }

    setFileError('')
    setFiles((current) => [...current, ...validFiles])
    event.target.value = ''
  }

  const handleSubmitChronology = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!agendaDate || !agenda.trim()) return

    try {
      await createChronology.mutateAsync({
        agenda_date: agendaDate,
        agenda,
        description,
        files,
      })

      setAgendaDate('')
      setAgenda('')
      setDescription('')
      setFiles([])
      setFileError('')
    } catch {
      // Error detail is rendered from the mutation state.
    }
  }

  const handleDownload = async (path: string) => {
    const { blob, filename } = await legalCaseService.downloadFile(path)
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = filename
    anchor.click()
    URL.revokeObjectURL(url)
  }

  const handleDownloadPDF = async () => {
    if (!id) return
    const { blob, filename } = await downloadPDF.mutateAsync(id)
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = filename
    anchor.click()
    URL.revokeObjectURL(url)
  }

  const handleDeleteChronology = async (chronologyID: string) => {
    if (!window.confirm('Hapus kronologi ini?')) return
    await deleteChronology.mutateAsync(chronologyID)
  }

  if (isLoading) return <div className="p-12 text-center text-gray-400">Memuat data...</div>
  if (!legalCase) return <div className="p-12 text-center text-gray-500">Kasus hukum tidak ditemukan</div>

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex items-start gap-3">
          <button onClick={() => navigate(caseRouteBase)} className="mt-0.5 rounded-lg p-2 hover:bg-gray-100" title="Kembali">
            <ArrowLeft className="h-5 w-5 text-gray-600" />
          </button>
          <div>
            <p className="text-xs font-medium text-gray-400">Kasus Hukum</p>
            <h1 className="mt-1 text-2xl font-bold" style={{ color: '#0B2545' }}>{legalCase.case_name}</h1>
            <p className="mt-0.5 text-sm text-gray-500">{legalCase.location_regency?.label ?? '-'} - {formatDate(legalCase.case_date)}</p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setEditOpen(true)}>
            <Edit className="h-4 w-4" />
            Edit Kasus
          </Button>
          <Button variant="outline" onClick={handleDownloadPDF} disabled={downloadPDF.isPending}>
            <FileDown className="h-4 w-4" />
            Download PDF
          </Button>
          {canDelete && (
            <Button variant="destructive" onClick={handleDeleteCase} disabled={deleteCase.isPending}>
              <Trash2 className="h-4 w-4" />
              Hapus
            </Button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <Section title="Informasi Umum">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Info label="Nama Kasus" value={legalCase.case_name} />
              <Info label="Jenis Kasus" value={legalCase.case_type?.label ?? legalCase.case_type_id} />
              <Info label="Pihak Terkait" value={legalCase.related_party?.name ?? '-'} />
              <Info label="Kategori Kasus" value={legalCase.category?.label ?? legalCase.category_id} />
              <Info label="Lokasi" value={legalCase.location_regency?.label ?? '-'} />
              <Info label="Penanggung Jawab" value={legalCase.pic_division?.name ?? legalCase.pic ?? '-'} />
              <Info label="Cadangan Teknis" value={legalCase.technical_reserve || '-'} />
              <Info label="Nilai Kasus" value={formatCurrency(legalCase.case_value)} />
              <Info label="Status Terkini" value={legalCase.current_status || '-'} />
              <Info label="Tanggal" value={formatDate(legalCase.case_date)} />
              <Info label="Tingkat" value={legalCase.level} />
              <DocumentUpload
                label="Dokumen Pendukung"
                documentLink={legalCase.document_link}
                canManage={canManageDocument}
                onUpload={async (file) => {
                  await legalCaseService.uploadDocument(legalCase.id, file)
                  await queryClient.invalidateQueries({ queryKey: ['legal-cases'] })
                }}
                onDelete={async () => {
                  if (!window.confirm('Hapus dokumen pendukung?')) return
                  await legalCaseService.deleteDocument(legalCase.id)
                  await queryClient.invalidateQueries({ queryKey: ['legal-cases'] })
                }}
              />
            </div>
            <TextInfo label="Ringkasan Kasus / Perkara" value={legalCase.case_summary || '-'} />
            <TextInfo label="Spesifikasi Kasus" value={legalCase.specification || '-'} />
            <TextInfo label="Catatan Tambahan" value={legalCase.additional_notes || '-'} />
          </Section>
        </div>

        <div className="space-y-6">
          <Section title="Posisi Kasus">
            <div className="flex items-start gap-3">
              <div className="mt-1 h-3 w-3 rounded-full bg-[#C8102E]" />
              <div>
                <p className="text-sm font-semibold text-gray-900">{legalCase.current_status || 'Belum ada status'}</p>
                <p className="mt-1 text-xs text-gray-400">Update terakhir {formatDateTime(legalCase.status_updated_at || legalCase.updated_at)}</p>
              </div>
            </div>
            {(legalCase.chronologies?.length ?? 0) > 0 && (
              <div className="mt-5 space-y-3">
                {legalCase.chronologies?.slice(0, 3).map((chronology) => (
                  <div key={chronology.id} className="border-l-2 border-gray-100 pl-3">
                    <p className="text-xs text-gray-400">{formatDate(chronology.agenda_date)}</p>
                    <p className="mt-0.5 text-sm font-medium text-gray-700">{chronology.agenda}</p>
                  </div>
                ))}
              </div>
            )}
          </Section>

          <Section title="Update Kronologi Kasus">
            <form onSubmit={handleSubmitChronology} className="space-y-4">
              <Field label="Tanggal">
                <Input type="date" value={agendaDate} onChange={(event) => setAgendaDate(event.target.value)} required />
              </Field>
              <Field label="Status Terkini">
                <Textarea value={agenda} onChange={(event) => setAgenda(event.target.value)} placeholder="Status terkini" required />
              </Field>
              <Field label="Tindak Lanjut">
                <Textarea value={description} onChange={(event) => setDescription(event.target.value)} rows={4} placeholder="Tindak lanjut" />
              </Field>
              <Field label="Dokumen">
                <label className="flex cursor-pointer flex-col items-center gap-2 rounded-lg border-2 border-dashed border-gray-200 p-5 hover:bg-gray-50">
                  <Upload className="h-5 w-5 text-gray-400" />
                  <span className="text-center text-xs text-gray-500">Pilih dokumen</span>
                  <input type="file" multiple accept=".pdf,.doc,.docx" className="hidden" onChange={handleFileChange} />
                </label>
                {fileError && <p className="text-xs text-red-500">{fileError}</p>}
                {files.length > 0 && (
                  <div className="mt-2 space-y-2">
                    {files.map((file, index) => (
                      <div key={`${file.name}-${index}`} className="flex items-center gap-2 rounded-lg bg-gray-50 px-3 py-2">
                        <FileText className="h-4 w-4 shrink-0 text-gray-400" />
                        <span className="min-w-0 flex-1 truncate text-xs text-gray-600">{file.name}</span>
                        <span className="text-xs text-gray-400">{formatFileSize(file.size)}</span>
                        <button type="button" onClick={() => setFiles((current) => current.filter((_, i) => i !== index))} className="text-gray-400 hover:text-red-600">
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </Field>

              {createChronology.isError && (
                <p className="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600">{(createChronology.error as Error)?.message}</p>
              )}
              {createChronology.isSuccess && (
                <p className="rounded-lg bg-green-50 px-3 py-2 text-xs text-green-700">Kronologi berhasil ditambahkan.</p>
              )}

              <Button type="submit" disabled={!agendaDate || !agenda.trim() || createChronology.isPending} className="w-full text-white" style={{ background: '#C8102E' }}>
                <Plus className="h-4 w-4" />
                {createChronology.isPending ? 'Menyimpan...' : 'Tambah Kronologi'}
              </Button>
            </form>
          </Section>
        </div>
      </div>

      <div className="mt-6">
        <Section title="Kronologi Sidang">
          {(legalCase.chronologies?.length ?? 0) === 0 ? (
            <div className="rounded-lg bg-gray-50 p-8 text-center text-sm text-gray-400">Belum ada kronologi sidang</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-100 bg-gray-50">
                    <th className="whitespace-nowrap px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500">Tanggal</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500">Status Terkini</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500">Tindak Lanjut</th>
                    <th className="px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500">Dokumen</th>
                    <th className="px-4 py-3 text-right text-xs font-semibold uppercase text-gray-500">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-50">
                  {legalCase.chronologies?.map((chronology) => (
                    <ChronologyRow
                      key={chronology.id}
                      chronology={chronology}
                      canDelete={canManageChronology}
                      onDownload={handleDownload}
                      onDelete={handleDeleteChronology}
                    />
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Section>
      </div>

      <LegalCaseFormDialog open={editOpen} onOpenChange={setEditOpen} legalCase={legalCase} />
    </div>
  )
}

function ChronologyRow({ chronology, canDelete, onDownload, onDelete }: {
  chronology: CaseChronology
  canDelete?: boolean
  onDownload: (path: string) => void
  onDelete: (id: string) => void
}) {
  return (
    <tr>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-600">{formatDate(chronology.agenda_date)}</td>
      <td className="px-4 py-3 text-sm font-medium text-gray-800">{chronology.agenda}</td>
      <td className="px-4 py-3 text-sm text-gray-500">{chronology.description || '-'}</td>
      <td className="px-4 py-3">
        {chronology.documents.length === 0 ? (
          <span className="text-sm text-gray-400">-</span>
        ) : (
          <div className="flex flex-wrap gap-1">
            {chronology.documents.map((path) => (
              <button key={path} onClick={() => onDownload(path)} className="inline-flex items-center gap-1 rounded-lg bg-gray-100 px-2 py-1 text-xs text-gray-600 hover:bg-gray-200">
                <Download className="h-3 w-3" />
                {path.split('/').pop()}
              </button>
            ))}
          </div>
        )}
      </td>
      <td className="px-4 py-3 text-right">
        {canDelete && (
          <button onClick={() => onDelete(chronology.id)} className="rounded-lg p-2 text-gray-400 hover:bg-red-50 hover:text-red-600" title="Hapus kronologi">
            <Trash2 className="h-4 w-4" />
          </button>
        )}
      </td>
    </tr>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="rounded-2xl border border-gray-100 bg-white p-5">
      <h2 className="mb-4 text-sm font-semibold" style={{ color: '#0B2545' }}>{title}</h2>
      <div className="space-y-4">{children}</div>
    </section>
  )
}

function Info({ label, value, link }: { label: string; value: string; link?: string }) {
  return (
    <div>
      <p className="text-xs text-gray-400">{label}</p>
      {link ? (
        <a href={link} target="_blank" rel="noreferrer" className="mt-0.5 block truncate text-sm font-medium text-[#C8102E] hover:underline">{value}</a>
      ) : (
        <p className="mt-0.5 text-sm font-medium text-gray-800">{value}</p>
      )}
    </div>
  )
}

function TextInfo({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs text-gray-400">{label}</p>
      <p className="mt-1 whitespace-pre-wrap text-sm leading-relaxed text-gray-700">{value}</p>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label className="text-sm font-medium text-gray-700">{label}</Label>
      {children}
    </div>
  )
}

function DocumentUpload({ label, documentLink, canManage, onUpload, onDelete }: {
  label: string
  documentLink?: string
  canManage?: boolean
  onUpload: (file: File) => void
  onDelete: () => void
}) {
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const isExternalDocument = !!documentLink && /^https?:\/\//i.test(documentLink)

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const allowedExtensions = ['.pdf', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx', '.jpg', '.jpeg', '.png']
    const ext = '.' + file.name.split('.').pop()?.toLowerCase()
    if (!allowedExtensions.includes(ext)) {
      setError('Format file tidak didukung.')
      e.target.value = ''
      return
    }
    if (file.size > 100 * 1024 * 1024) {
      setError('Ukuran file melebihi batas maksimal 100 MB.')
      e.target.value = ''
      return
    }

    setError('')
    setUploading(true)
    try {
      await onUpload(file)
    } catch (err) {
      setError((err as Error)?.message ?? 'Gagal mengupload dokumen.')
    } finally {
      setUploading(false)
      e.target.value = ''
    }
  }

  const fileName = documentLink?.split('/').pop()

  return (
    <div className="space-y-1.5">
      <Label className="text-sm font-medium text-gray-700">{label}</Label>
      {documentLink ? (
        <div className="flex items-center gap-2 rounded-lg border border-gray-200 bg-gray-50 px-3 py-2">
          <FileText className="h-4 w-4 shrink-0 text-gray-400" />
          <span className="min-w-0 flex-1 truncate text-xs text-gray-600">{fileName}</span>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => {
              if (isExternalDocument) {
                window.open(documentLink, '_blank', 'noopener,noreferrer')
                return
              }
              legalCaseService.downloadFile(documentLink).then(({ blob, filename }) => {
                const url = URL.createObjectURL(blob)
                const a = document.createElement('a')
                a.href = url
                a.download = filename
                a.click()
                URL.revokeObjectURL(url)
              })
            }}
            className="h-7 px-2 text-xs"
          >
            <Download className="h-3 w-3" />
          </Button>
          {canManage && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={onDelete}
              className="h-7 px-2 text-xs text-red-600 hover:text-red-700"
            >
              <Trash2 className="h-3 w-3" />
            </Button>
          )}
        </div>
      ) : (
        canManage && (
          <label className="flex cursor-pointer flex-col items-center gap-2 rounded-lg border-2 border-dashed border-gray-200 p-4 hover:bg-gray-50">
            <Upload className="h-5 w-5 text-gray-400" />
            <span className="text-center text-xs text-gray-500">
              {uploading ? 'Mengupload...' : 'Pilih dokumen untuk diupload'}
            </span>
            <input
              type="file"
              accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.jpg,.jpeg,.png"
              className="hidden"
              onChange={handleFileChange}
              disabled={uploading}
            />
          </label>
        )
      )}
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}
