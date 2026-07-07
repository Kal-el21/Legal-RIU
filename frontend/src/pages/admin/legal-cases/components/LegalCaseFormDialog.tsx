import { useEffect, useMemo, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, Upload, Trash2, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import SearchableSelect from '@/components/common/SearchableSelect'
import { useQueryClient } from '@tanstack/react-query'
import { useCedants, useCompanies, useCreateCedant, useCreateLegalCase, useDivisions, useRegencies, useUpdateLegalCase, useCaseTypes, useCaseCategories } from '@/hooks/useLegalCase'
import { legalCaseService } from '@/services/legal-case.service'
import { useAuthStore } from '@/store/auth.store'
import type { LegalCase } from '@/types'

const DOC_EXTENSIONS = ['.pdf', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx', '.jpg', '.jpeg', '.png']
const PHOTO_EXTENSIONS = ['.jpg', '.jpeg', '.png', '.webp']
const MAX_FILE_SIZE = 100 * 1024 * 1024

function validateUploadFile(file: File, extensions: string[]): string | null {
  const ext = '.' + (file.name.split('.').pop()?.toLowerCase() ?? '')
  if (!extensions.includes(ext)) return 'Format file tidak didukung.'
  if (file.size > MAX_FILE_SIZE) return 'Ukuran file melebihi batas maksimal 100 MB.'
  return null
}

const schema = z.object({
  case_name: z.string().min(1, 'Wajib diisi'),
  case_summary: z.string().optional(),
  related_party_id: z.string().min(1, 'Pilih pihak terkait'),
  category_id: z.string().min(1, 'Wajib diisi'),
  specification: z.string().optional(),
  case_type_id: z.string().min(1, 'Pilih jenis kasus'),
  technical_reserve: z.string().optional(),
  case_value: z.number().min(0, 'Nilai tidak valid'),
  pic: z.string().min(1, 'Pilih PIC'),
  document_link: z.string().optional(),
  photo: z.string().optional(),
  current_status: z.string().optional(),
  case_date: z.string().min(1, 'Tanggal wajib diisi'),
  level: z.string().min(1, 'Wajib diisi'),
  additional_notes: z.string().optional(),
  location_regency_id: z.string().min(1, 'Pilih kabupaten/kota'),
  company_id: z.string().min(1, 'Wajib diisi'),
})

type FormData = z.infer<typeof schema>

interface LegalCaseFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  legalCase?: LegalCase | null
}

export default function LegalCaseFormDialog({ open, onOpenChange, legalCase }: LegalCaseFormDialogProps) {
  const isEdit = !!legalCase
  const queryClient = useQueryClient()
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const canManageDocument = hasPermission('case_management.manage_document')
  const { data: cedants = [] } = useCedants({ limit: 200 })
  const { data: regencies = [] } = useRegencies({ limit: 500 })
  const { data: divisions = [] } = useDivisions()
  const { data: companies = [] } = useCompanies()
  const { data: caseTypes = [] } = useCaseTypes()
  const { data: caseCategories = [] } = useCaseCategories()
  const createLegalCase = useCreateLegalCase()
  const updateLegalCase = useUpdateLegalCase()
  const createCedant = useCreateCedant()

  const [showCedantForm, setShowCedantForm] = useState(false)
  const [newCedantName, setNewCedantName] = useState('')
  const [newCedantDescription, setNewCedantDescription] = useState('')
  const [docFile, setDocFile] = useState<File | null>(null)
  const [photoFile, setPhotoFile] = useState<File | null>(null)
  const [uploadError, setUploadError] = useState('')

  const { control, register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: emptyDefaults(),
  })

  useEffect(() => {
    if (!open) return
    reset(legalCase ? valuesFromLegalCase(legalCase) : emptyDefaults())
    setShowCedantForm(false)
    setNewCedantName('')
    setNewCedantDescription('')
    setDocFile(null)
    setPhotoFile(null)
    setUploadError('')
  }, [legalCase, open, reset])

  const cedantOptions = useMemo(() => cedants.map((cedant) => ({
    value: cedant.id,
    label: cedant.name,
    description: cedant.description,
  })), [cedants])

  const regencyOptions = useMemo(() => regencies.map((regency) => ({
    value: regency.id,
    label: regency.label,
    description: regency.type,
  })), [regencies])

  const divisionOptions = useMemo(() => divisions.map((division) => ({
    value: division.id,
    label: division.name,
    description: division.description,
  })), [divisions])

  const handleCreateCedant = async () => {
    if (!newCedantName.trim()) return
    const cedant = await createCedant.mutateAsync({
      name: newCedantName.trim(),
      description: newCedantDescription.trim(),
    })
    setValue('related_party_id', cedant.id, { shouldValidate: true })
    setNewCedantName('')
    setNewCedantDescription('')
    setShowCedantForm(false)
  }

  const onSubmit = async (data: FormData) => {
    setUploadError('')
    try {
      if (isEdit && legalCase) {
        await updateLegalCase.mutateAsync({ id: legalCase.id, data })
      } else {
        const created = await createLegalCase.mutateAsync(data)
        const uploads: Promise<unknown>[] = []
        if (docFile) uploads.push(legalCaseService.uploadDocument(created.id, docFile))
        if (photoFile) uploads.push(legalCaseService.uploadPhoto(created.id, photoFile))
        if (uploads.length > 0) {
          await Promise.all(uploads)
        }
      }
      queryClient.invalidateQueries({ queryKey: ['legal-cases'] })
      onOpenChange(false)
    } catch (err) {
      setUploadError((err as Error)?.message ?? 'Terjadi kesalahan saat menyimpan')
    }
  }

  const isSaving = createLegalCase.isPending || updateLegalCase.isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-4xl">
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Kasus' : 'Tambah Kasus'}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <section className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label="Nama Kasus" error={errors.case_name?.message}>
              <Input {...register('case_name')} placeholder="Nama kasus" />
            </Field>
            <Field label="Tanggal" error={errors.case_date?.message}>
              <Input type="date" {...register('case_date')} />
            </Field>
            <Field label="Perusahaan" error={errors.company_id?.message}>
              <Controller
                name="company_id"
                control={control}
                render={({ field }) => (
                  <Select onValueChange={field.onChange} value={field.value}>
                    <SelectTrigger className="w-full"><SelectValue placeholder="Pilih perusahaan" /></SelectTrigger>
                    <SelectContent>
                      {companies.map((item) => <SelectItem key={item.id} value={item.id}>{item.name}</SelectItem>)}
                    </SelectContent>
                  </Select>
                )}
              />
            </Field>
            <Field label="Pihak Terkait" error={errors.related_party_id?.message}>
              <Controller
                name="related_party_id"
                control={control}
                render={({ field }) => (
                  <SearchableSelect
                    value={field.value}
                    options={cedantOptions}
                    placeholder="Pilih cedant"
                    emptyText="Cedant belum tersedia"
                    onChange={field.onChange}
                  />
                )}
              />
              <button
                type="button"
                onClick={() => setShowCedantForm((current) => !current)}
                className="mt-1 inline-flex items-center gap-1 text-xs font-medium text-[#C8102E]"
              >
                <Plus className="h-3 w-3" />
                Tambah cedant
              </button>
            </Field>
            <Field label="Lokasi Kabupaten/Kota" error={errors.location_regency_id?.message}>
              <Controller
                name="location_regency_id"
                control={control}
                render={({ field }) => (
                  <SearchableSelect
                    value={field.value}
                    options={regencyOptions}
                    placeholder="Pilih kabupaten/kota"
                    emptyText="Kabupaten/kota tidak ditemukan"
                    onChange={field.onChange}
                  />
                )}
              />
            </Field>

            {showCedantForm && (
              <div className="sm:col-span-2 rounded-lg border border-gray-100 bg-gray-50 p-4">
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_1fr_auto]">
                  <Input value={newCedantName} onChange={(event) => setNewCedantName(event.target.value)} placeholder="Nama cedant" />
                  <Input value={newCedantDescription} onChange={(event) => setNewCedantDescription(event.target.value)} placeholder="Deskripsi" />
                  <Button type="button" onClick={handleCreateCedant} disabled={!newCedantName.trim() || createCedant.isPending}>
                    Simpan
                  </Button>
                </div>
              </div>
            )}

            <Field label="Kategori" error={errors.category_id?.message}>
              <Controller
                name="category_id"
                control={control}
                render={({ field }) => (
                  <Select onValueChange={field.onChange} value={field.value}>
                    <SelectTrigger className="w-full"><SelectValue placeholder="Pilih kategori" /></SelectTrigger>
                    <SelectContent>
                      {caseCategories.map((item) => (
                        <SelectItem key={item.id} value={item.id}>{item.label}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
            </Field>
            <Field label="Jenis Kasus" error={errors.case_type_id?.message}>
              <Controller
                name="case_type_id"
                control={control}
                render={({ field }) => (
                  <Select onValueChange={field.onChange} value={field.value}>
                    <SelectTrigger className="w-full"><SelectValue placeholder="Pilih jenis kasus" /></SelectTrigger>
                    <SelectContent>
                      {caseTypes.map((item) => (
                        <SelectItem key={item.id} value={item.id}>{item.label}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
            </Field>
            <Field label="PIC" error={errors.pic?.message}>
              <Controller
                name="pic"
                control={control}
                render={({ field }) => (
                  <SearchableSelect
                    value={field.value}
                    options={divisionOptions}
                    placeholder="Pilih divisi PIC"
                    emptyText="Divisi belum tersedia"
                    onChange={field.onChange}
                  />
                )}
              />
            </Field>
            <Field label="Tingkat" error={errors.level?.message}>
              <Input {...register('level')} placeholder="Contoh: Pengadilan Negeri" />
            </Field>
            <Field label="Status Terkini" error={errors.current_status?.message}>
              <Input {...register('current_status')} placeholder="Status terkini" />
            </Field>
            <Field label="Nilai Kasus" error={errors.case_value?.message}>
              <Input type="number" min={0} step={1000} {...register('case_value', { valueAsNumber: true })} placeholder="0" />
            </Field>
            <Field label="Cadangan Teknis" error={errors.technical_reserve?.message}>
              <Input {...register('technical_reserve')} placeholder="Cadangan teknis" />
            </Field>

            <div className="sm:col-span-2">
              <DocumentAttachmentField
                label="Dokumen Pendukung"
                value={watch('document_link')}
                canManage={canManageDocument}
                legalCase={legalCase}
                bufferedFile={docFile}
                onBuffer={setDocFile}
                onBufferRemove={() => setDocFile(null)}
                queryClient={queryClient}
                setValue={setValue}
              />
            </div>

            <div className="sm:col-span-2">
              <PhotoAttachmentField
                label="Foto Kasus"
                value={watch('photo')}
                canManage={canManageDocument}
                legalCase={legalCase}
                bufferedFile={photoFile}
                onBuffer={setPhotoFile}
                onBufferRemove={() => setPhotoFile(null)}
                queryClient={queryClient}
                setValue={setValue}
              />
            </div>
          </section>

          <section className="space-y-4">
            <Field label="Ringkasan Kasus" error={errors.case_summary?.message}>
              <Textarea {...register('case_summary')} rows={4} placeholder="Ringkasan perkara" />
            </Field>
            <Field label="Spesifikasi Kasus" error={errors.specification?.message}>
              <Textarea {...register('specification')} rows={3} placeholder="Spesifikasi kasus" />
            </Field>
            <Field label="Catatan Tambahan" error={errors.additional_notes?.message}>
              <Textarea {...register('additional_notes')} rows={3} placeholder="Catatan tambahan" />
            </Field>
          </section>

          {(createLegalCase.isError || updateLegalCase.isError || createCedant.isError || uploadError) && (
            <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600">
              {uploadError || ((createLegalCase.error || updateLegalCase.error || createCedant.error) as Error)?.message || 'Terjadi kesalahan saat menyimpan'}
            </p>
          )}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Batal</Button>
            <Button type="submit" disabled={isSaving} className="text-white" style={{ background: '#C8102E' }}>
              {isSaving ? 'Menyimpan...' : 'Simpan'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

interface AttachmentFieldProps {
  label: string
  value?: string
  canManage?: boolean
  legalCase?: LegalCase | null
  bufferedFile: File | null
  onBuffer: (file: File | null) => void
  onBufferRemove: () => void
  queryClient: ReturnType<typeof useQueryClient>
  setValue: (name: 'document_link' | 'photo', value: string) => void
}

function DocumentAttachmentField({ label, value, canManage, legalCase, bufferedFile, onBuffer, onBufferRemove, queryClient, setValue }: AttachmentFieldProps) {
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    const err = validateUploadFile(file, DOC_EXTENSIONS)
    if (err) { setError(err); return }

    if (legalCase) {
      setUploading(true)
      setError('')
      try {
        const updated = await legalCaseService.uploadDocument(legalCase.id, file)
        setValue('document_link', updated.document_link ?? '')
        queryClient.invalidateQueries({ queryKey: ['legal-cases'] })
      } catch (err) {
        setError((err as Error)?.message ?? 'Gagal mengupload dokumen.')
      } finally {
        setUploading(false)
      }
    } else {
      setError('')
      onBuffer(file)
    }
  }

  const handleDelete = async () => {
    if (!legalCase) { onBufferRemove(); return }
    if (!window.confirm('Hapus dokumen pendukung?')) return
    setError('')
    try {
      await legalCaseService.deleteDocument(legalCase.id)
      setValue('document_link', '')
      queryClient.invalidateQueries({ queryKey: ['legal-cases'] })
    } catch (err) {
      setError((err as Error)?.message ?? 'Gagal menghapus dokumen.')
    }
  }

  const fileName = value?.split('/').pop() ?? bufferedFile?.name

  if (value || bufferedFile) {
    return (
      <div className="space-y-1.5">
        <Label className="text-sm font-medium text-gray-700">{label}</Label>
        <div className="flex items-center gap-2 rounded-lg border border-gray-200 bg-gray-50 px-3 py-2">
          <FileText className="h-4 w-4 shrink-0 text-gray-400" />
          <span className="min-w-0 flex-1 truncate text-xs text-gray-600">{fileName}</span>
          {canManage && (
            <Button type="button" variant="ghost" size="sm" onClick={handleDelete} className="h-7 px-2 text-xs text-red-600 hover:text-red-700">
              <Trash2 className="h-3 w-3" />
            </Button>
          )}
        </div>
        {error && <p className="text-xs text-red-500">{error}</p>}
      </div>
    )
  }

  if (!canManage) {
    return (
      <div className="space-y-1.5">
        <Label className="text-sm font-medium text-gray-700">{label}</Label>
        <p className="text-xs text-gray-400">Belum ada dokumen</p>
      </div>
    )
  }

  return (
    <div className="space-y-1.5">
      <Label className="text-sm font-medium text-gray-700">{label}</Label>
      <label className="flex cursor-pointer flex-col items-center gap-2 rounded-lg border-2 border-dashed border-gray-200 p-4 hover:bg-gray-50">
        <Upload className="h-5 w-5 text-gray-400" />
        <span className="text-center text-xs text-gray-500">
          {uploading ? 'Mengupload...' : 'Pilih dokumen untuk diupload'}
        </span>
        <input type="file" accept={DOC_EXTENSIONS.join(',')} className="hidden" onChange={handleFileChange} disabled={uploading} />
      </label>
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}

function PhotoAttachmentField({ label, value, canManage, legalCase, bufferedFile, onBuffer, onBufferRemove, queryClient, setValue }: AttachmentFieldProps) {
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')

  const previewUrl = bufferedFile
    ? URL.createObjectURL(bufferedFile)
    : value
      ? legalCaseService.getFileViewUrl(value)
      : null

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    const err = validateUploadFile(file, PHOTO_EXTENSIONS)
    if (err) { setError(err); return }

    if (legalCase) {
      setUploading(true)
      setError('')
      try {
        const updated = await legalCaseService.uploadPhoto(legalCase.id, file)
        setValue('photo', updated.photo ?? '')
        queryClient.invalidateQueries({ queryKey: ['legal-cases'] })
      } catch (err) {
        setError((err as Error)?.message ?? 'Gagal mengupload foto.')
      } finally {
        setUploading(false)
      }
    } else {
      setError('')
      onBuffer(file)
    }
  }

  const handleDelete = async () => {
    if (!legalCase) { onBufferRemove(); return }
    if (!window.confirm('Hapus foto kasus?')) return
    setError('')
    try {
      await legalCaseService.deletePhoto(legalCase.id)
      setValue('photo', '')
      queryClient.invalidateQueries({ queryKey: ['legal-cases'] })
    } catch (err) {
      setError((err as Error)?.message ?? 'Gagal menghapus foto.')
    }
  }

  if (value || bufferedFile) {
    return (
      <div className="space-y-1.5">
        <Label className="text-sm font-medium text-gray-700">{label}</Label>
        <div className="flex items-center gap-3">
          {previewUrl && <img src={previewUrl} alt="Foto kasus" className="h-20 w-20 rounded-lg object-cover border border-gray-200" />}
          <div className="flex flex-col gap-2">
            <span className="text-xs text-gray-600 truncate max-w-[200px]">{bufferedFile?.name ?? value?.split('/').pop()}</span>
            {canManage && (
              <Button type="button" variant="ghost" size="sm" onClick={handleDelete} className="h-7 w-fit px-2 text-xs text-red-600 hover:text-red-700">
                <Trash2 className="h-3 w-3 mr-1" /> Hapus
              </Button>
            )}
          </div>
        </div>
        {error && <p className="text-xs text-red-500">{error}</p>}
      </div>
    )
  }

  if (!canManage) {
    return (
      <div className="space-y-1.5">
        <Label className="text-sm font-medium text-gray-700">{label}</Label>
        <p className="text-xs text-gray-400">Belum ada foto</p>
      </div>
    )
  }

  return (
    <div className="space-y-1.5">
      <Label className="text-sm font-medium text-gray-700">{label}</Label>
      <label className="flex cursor-pointer flex-col items-center gap-2 rounded-lg border-2 border-dashed border-gray-200 p-4 hover:bg-gray-50">
        <Upload className="h-5 w-5 text-gray-400" />
        <span className="text-center text-xs text-gray-500">
          {uploading ? 'Mengupload...' : 'Pilih foto (jpg, png, webp)'}
        </span>
        <input type="file" accept={PHOTO_EXTENSIONS.join(',')} className="hidden" onChange={handleFileChange} disabled={uploading} />
      </label>
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}

function emptyDefaults(): FormData {
  return {
    case_name: '',
    case_summary: '',
    related_party_id: '',
    category_id: '',
    specification: '',
    case_type_id: '',
    technical_reserve: '',
    case_value: 0,
    pic: '',
    document_link: '',
    photo: '',
    current_status: '',
    case_date: '',
    level: '',
    additional_notes: '',
    location_regency_id: '',
    company_id: '',
  }
}

function valuesFromLegalCase(legalCase: LegalCase): FormData {
  return {
    case_name: legalCase.case_name,
    case_summary: legalCase.case_summary ?? '',
    related_party_id: legalCase.related_party_id,
    category_id: legalCase.category_id,
    specification: legalCase.specification ?? '',
    case_type_id: legalCase.case_type_id,
    technical_reserve: legalCase.technical_reserve ?? '',
    case_value: legalCase.case_value ?? 0,
    pic: legalCase.pic,
    document_link: legalCase.document_link ?? '',
    photo: legalCase.photo ?? '',
    current_status: legalCase.current_status ?? '',
    case_date: legalCase.case_date ? legalCase.case_date.slice(0, 10) : '',
    level: legalCase.level,
    additional_notes: legalCase.additional_notes ?? '',
    location_regency_id: legalCase.location_regency_id,
    company_id: legalCase.company_id,
  }
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label className="text-sm font-medium text-gray-700">{label}</Label>
      {children}
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}
