import { useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { useMaterial, useCreateMaterial, useUpdateMaterial } from '@/hooks/useMaterial'

const schema = z.object({
  title: z.string().min(1, 'Judul wajib diisi'),
  excerpt: z.string().optional(),
  content: z.string().min(1, 'Konten wajib diisi'),
})

type FormData = z.infer<typeof schema>

export default function MaterialFormPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const isEdit = !!id

  const { data: material, isLoading } = useMaterial(id || '')
  const createMutation = useCreateMaterial()
  const updateMutation = useUpdateMaterial()

  const form = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { title: '', excerpt: '', content: '' },
  })

  useEffect(() => {
    if (material && isEdit) {
      form.reset({ title: material.title, excerpt: material.excerpt || '', content: material.content })
    }
  }, [material, isEdit, form])

  const onSubmit = async (data: FormData) => {
    if (isEdit && id) {
      await updateMutation.mutateAsync({ id, data })
    } else {
      await createMutation.mutateAsync(data)
    }
    navigate('/admin/materials')
  }

  if (isEdit && isLoading) {
    return <div className="p-6 text-center text-gray-400">Memuat data...</div>
  }

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <h1 className="text-2xl font-bold mb-6" style={{ color: '#0B2545' }}>{isEdit ? 'Edit Materi' : 'Tambah Materi'}</h1>

      <form onSubmit={form.handleSubmit(onSubmit)} className="bg-white rounded-2xl border border-gray-100 p-6 space-y-6">
        <div>
          <Label className="text-sm font-medium text-gray-700">Judul</Label>
          <Input {...form.register('title')} placeholder="Judul materi" className="mt-1.5" />
          {form.formState.errors.title && <p className="text-xs text-red-500 mt-1">{form.formState.errors.title.message}</p>}
        </div>

        <div>
          <Label className="text-sm font-medium text-gray-700">Excerpt</Label>
          <Input {...form.register('excerpt')} placeholder="Ringkasan singkat" className="mt-1.5" />
          {form.formState.errors.excerpt && <p className="text-xs text-red-500 mt-1">{form.formState.errors.excerpt.message}</p>}
        </div>

        <div>
          <Label className="text-sm font-medium text-gray-700">Konten</Label>
          <Textarea {...form.register('content')} rows={10} placeholder="Konten materi..." className="mt-1.5" />
          {form.formState.errors.content && <p className="text-xs text-red-500 mt-1">{form.formState.errors.content.message}</p>}
        </div>

        <div className="flex gap-2 pt-2">
          <Button type="button" variant="outline" onClick={() => navigate('/admin/materials')}>Batal</Button>
          <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending} className="text-white" style={{ background: '#C8102E' }}>
            {(createMutation.isPending || updateMutation.isPending) ? 'Menyimpan...' : 'Simpan'}
          </Button>
        </div>
      </form>
    </div>
  )
}
