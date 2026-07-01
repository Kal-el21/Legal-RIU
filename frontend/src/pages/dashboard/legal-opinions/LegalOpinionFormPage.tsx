import { useNavigate, useParams } from 'react-router-dom'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useState, useEffect } from 'react'
import { Upload, X, FileText, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useCreateLegalOpinion, useUpdateLegalOpinion, useLegalOpinion } from '@/hooks/useLegalOpinion'
import { useDivisions } from '@/hooks/useLegalCase'
import { useAuthStore } from '@/store/auth.store'
import { validateFile, formatFileSize } from '@/lib/utils'

const LEGAL_TYPES = [
  'Permasalahan Hukum',
  'Bisnis Teknik',
  'Bisnis Penunjang',
  'Perjanjian Reasuransi (Treaty/Fakultatif)',
  'Lain-Lain',
]

const schema = z.object({
  requestor_name: z.string().min(1, 'Wajib diisi'),
  requestor_position: z.string().min(1, 'Wajib diisi'),
  requestor_division: z.string().min(1, 'Wajib diisi'),
  requestor_email: z.string().email('Email tidak valid'),
  requestor_phone: z.string().min(1, 'Wajib diisi'),
  legal_type: z.string().min(1, 'Pilih jenis kajian'),
  legal_type_other: z.string().optional(),
  title: z.string().min(1, 'Wajib diisi'),
  chronology: z.string().min(10, 'Kronologis terlalu singkat'),
  question: z.string().min(5, 'Pertanyaan terlalu singkat'),
})

type FormData = z.infer<typeof schema>

export default function LegalOpinionFormPage() {
  const { id } = useParams()
  const isEdit = !!id
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)

  const [files, setFiles] = useState<File[]>([])
  const [fileErrors, setFileErrors] = useState<string[]>([])

  const { data: existing } = useLegalOpinion(id ?? '')
  const { data: divisions = [] } = useDivisions()
  const createMutation = useCreateLegalOpinion()
  const updateMutation = useUpdateLegalOpinion()
  const divisionOptions = divisions.map((division) => division.name)

  const { register, handleSubmit, watch, setValue, reset, control, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      requestor_name: user?.full_name ?? '',
      requestor_position: user?.position ?? '',
      requestor_division: user?.division ?? '',
      requestor_email: user?.email ?? '',
    },
  })

  const legalType = watch('legal_type')

  useEffect(() => {
    if (existing && isEdit) {
      reset({
        requestor_name: existing.requestor_name,
        requestor_position: existing.requestor_position,
        requestor_division: existing.requestor_division,
        requestor_email: existing.requestor_email,
        requestor_phone: existing.requestor_phone,
        legal_type: existing.legal_type,
        legal_type_other: existing.legal_type_other,
        title: existing.title,
        chronology: existing.chronology,
        question: existing.question,
      })
    }
  }, [existing, isEdit, reset])

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = Array.from(e.target.files ?? [])
    const errs: string[] = []
    const validFiles: File[] = []
    
    for (const f of selected) {
      const err = await validateFile(f)
      if (err) errs.push(`${f.name}: ${err}`)
      else validFiles.push(f)
    }
    
    setFileErrors(errs)
    setFiles((prev) => [...prev, ...validFiles])
    e.target.value = ''
  }

  const removeFile = (index: number) => setFiles((prev) => prev.filter((_, i) => i !== index))

  const onSubmit = async (data: FormData) => {
    if (isEdit) {
      await updateMutation.mutateAsync({ id: id!, data })
    } else {
      await createMutation.mutateAsync({ ...data, attachments: files })
    }
    navigate('/dashboard/legal-opinions')
  }

  const isLoading = createMutation.isPending || updateMutation.isPending

  return (
    <div className="p-6 max-w-3xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-3 mb-8">
        <button onClick={() => navigate(-1)} className="p-2 rounded-lg hover:bg-gray-100 transition-colors">
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </button>
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>
            {isEdit ? 'Edit Legal Opinion' : 'Buat Legal Opinion'}
          </h1>
          <p className="text-sm text-gray-500">Isi formulir pengajuan dengan lengkap dan benar</p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-8">
        {/* Informasi Pemohon */}
        <section className="bg-white rounded-2xl border border-gray-100 p-6">
          <h2 className="text-base font-semibold mb-5" style={{ color: '#0B2545' }}>Informasi Pemohon</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">
            <Field label="Nama Lengkap" error={errors.requestor_name?.message}>
              <Input {...register('requestor_name')} placeholder="Nama lengkap" />
            </Field>
            <Field label="Posisi Jabatan" error={errors.requestor_position?.message}>
              <Input {...register('requestor_position')} placeholder="Jabatan Anda" />
            </Field>
            <Field label="Divisi pada RIU" error={errors.requestor_division?.message}>
              <Controller
                name="requestor_division"
                control={control}
                render={({ field }) => (
                  <Select onValueChange={field.onChange} value={field.value}>
                    <SelectTrigger><SelectValue placeholder="Pilih divisi" /></SelectTrigger>
                    <SelectContent>
                      {field.value && !divisionOptions.includes(field.value) && (
                        <SelectItem value={field.value}>{field.value}</SelectItem>
                      )}
                      {divisionOptions.map((d) => <SelectItem key={d} value={d}>{d}</SelectItem>)}
                    </SelectContent>
                  </Select>
                )}
              />
            </Field>
            <Field label="Email Kantor" error={errors.requestor_email?.message}>
              <Input {...register('requestor_email')} type="email" placeholder="email@indonesiare.co.id" />
            </Field>
            <Field label="Nomor WhatsApp" error={errors.requestor_phone?.message} className="sm:col-span-2">
              <Input {...register('requestor_phone')} placeholder="08xxxxxxxxxx" />
            </Field>
          </div>
        </section>

        {/* Detail Permasalahan */}
        <section className="bg-white rounded-2xl border border-gray-100 p-6">
          <h2 className="text-base font-semibold mb-5" style={{ color: '#0B2545' }}>Detail Permasalahan</h2>
          <div className="space-y-5">
            <Field label="Jenis Kajian Hukum" error={errors.legal_type?.message}>
              <Select onValueChange={(v) => setValue('legal_type', v)} defaultValue={existing?.legal_type}>
                <SelectTrigger><SelectValue placeholder="Pilih jenis kajian" /></SelectTrigger>
                <SelectContent>
                  {LEGAL_TYPES.map((t) => <SelectItem key={t} value={t}>{t}</SelectItem>)}
                </SelectContent>
              </Select>
            </Field>

            {legalType === 'Lain-Lain' && (
              <Field label="Jenis Kajian Lainnya" error={errors.legal_type_other?.message}>
                <Input {...register('legal_type_other')} placeholder="Sebutkan jenis kajian" />
              </Field>
            )}

            <Field label="Judul Permasalahan Hukum" error={errors.title?.message}>
              <Input {...register('title')} placeholder="Judul singkat permasalahan" />
            </Field>

            <Field label="Kronologis Permasalahan" error={errors.chronology?.message}>
              <Textarea {...register('chronology')} rows={5} placeholder="Jelaskan kronologis permasalahan secara rinci..." />
            </Field>

            <Field label="Pertanyaan Permasalahan Hukum" error={errors.question?.message}>
              <Textarea {...register('question')} rows={3} placeholder="Tuliskan pertanyaan hukum yang ingin dijawab..." />
            </Field>
          </div>
        </section>

        {/* Upload Dokumen — only on create */}
        {!isEdit && (
          <section className="bg-white rounded-2xl border border-gray-100 p-6">
            <h2 className="text-base font-semibold mb-2" style={{ color: '#0B2545' }}>Dokumen Pendukung</h2>
            <p className="text-xs text-gray-400 mb-5">Format: PDF, DOC, DOCX — Maks. 100 MB per file (opsional)</p>

            <label className="flex flex-col items-center justify-center gap-3 p-8 border-2 border-dashed border-gray-200 rounded-xl cursor-pointer hover:border-gray-300 hover:bg-gray-50 transition-colors">
              <Upload className="w-8 h-8 text-gray-400" />
              <div className="text-center">
                <p className="text-sm font-medium text-gray-600">Klik untuk upload file</p>
                <p className="text-xs text-gray-400 mt-0.5">PDF, DOC, atau DOCX</p>
              </div>
              <input type="file" multiple accept=".pdf,.doc,.docx" className="hidden" onChange={handleFileChange} />
            </label>

            {fileErrors.length > 0 && (
              <div className="mt-3 space-y-1">
                {fileErrors.map((e, i) => <p key={i} className="text-xs text-red-500">{e}</p>)}
              </div>
            )}

            {files.length > 0 && (
              <div className="mt-4 space-y-2">
                {files.map((f, i) => (
                  <div key={i} className="flex items-center gap-3 p-3 rounded-lg bg-gray-50 border border-gray-100">
                    <FileText className="w-4 h-4 text-gray-400 flex-shrink-0" />
                    <span className="text-sm text-gray-700 flex-1 truncate">{f.name}</span>
                    <span className="text-xs text-gray-400">{formatFileSize(f.size)}</span>
                    <button type="button" onClick={() => removeFile(i)} className="text-gray-400 hover:text-red-500 transition-colors">
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </section>
        )}

        {/* Error */}
        {(createMutation.isError || updateMutation.isError) && (
          <div className="p-4 rounded-xl bg-red-50 border border-red-200">
            <p className="text-sm text-red-600">
              {((createMutation.error || updateMutation.error) as Error)?.message ?? 'Terjadi kesalahan'}
            </p>
          </div>
        )}

        {/* Submit */}
        <div className="flex gap-3 justify-end">
          <Button type="button" variant="outline" onClick={() => navigate(-1)}>Batal</Button>
          <Button type="submit" disabled={isLoading} className="text-white" style={{ background: '#C8102E' }}>
            {isLoading ? 'Menyimpan...' : isEdit ? 'Simpan Perubahan' : 'Ajukan'}
          </Button>
        </div>
      </form>
    </div>
  )
}

function Field({ label, error, children, className }: {
  label: string; error?: string; children: React.ReactNode; className?: string
}) {
  return (
    <div className={`space-y-1.5 ${className ?? ''}`}>
      <Label className="text-sm font-medium text-gray-700">{label}</Label>
      {children}
      {error && <p className="text-xs text-red-500">{error}</p>}
    </div>
  )
}
