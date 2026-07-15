import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Send, ArrowLeft } from 'lucide-react'
import { agreementService, type AgreementSchema } from '@/services/agreement-document.service'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
export default function AgreementDocumentForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [schema, setSchema] = useState<AgreementSchema>()
  const [form, setForm] = useState<Record<string, string>>({})
  const [files, setFiles] = useState<File[]>([])
  const [currentStatus, setCurrentStatus] = useState('')

  useEffect(() => {
    void agreementService.schema().then(setSchema)
    if (id) {
      void agreementService.get(id).then((doc) => {
        setCurrentStatus(doc.status)
        const values: Record<string, string> = {}
        Object.entries(doc.form_data).forEach(([key, value]) => { values[key] = String(value ?? '') })
        setForm(values)
      })
    }
  }, [id])

  const submit = async () => {
    if (id) {
      await agreementService.update(id, form)
      if (currentStatus === 'NEED_REVISION') await agreementService.resubmit(id, files)
    } else {
      await agreementService.create({ document_type_code: 'PKS', form_data: form }, files)
    }
    navigate('/dashboard/agreement-documents')
  }

   return <div className="p-6 max-w-4xl mx-auto">
     <button onClick={() => navigate(-1)} className="p-2 rounded-lg hover:bg-gray-100 mt-0.5 mb-2" title="Kembali">
       <ArrowLeft className="w-5 h-5 text-gray-600" />
     </button>
     <h1 className="text-2xl font-bold mb-6" style={{ color: '#0B2545' }}>{id ? 'Revisi' : 'Pengajuan'} Perjanjian Kerja Sama</h1>
    {schema?.sections.map((section) => <section key={section.title} className="bg-white rounded-2xl border border-gray-100 p-6 mb-5">
      <h2 className="text-lg font-semibold mb-4" style={{ color: '#0B2545' }}>{section.title}</h2>
      <div className="grid md:grid-cols-2 gap-4">
        {section.fields.map((field) => (
          <div key={field.name} className={field.type === 'textarea' ? 'md:col-span-2' : ''}>
            <Label className="mb-1.5 text-gray-700">{field.label}{field.required && ' *'}</Label>
            {field.type === 'textarea'
              ? <Textarea required={field.required} placeholder={field.name === 'ruang_lingkup' ? 'Tulis satu poin ruang lingkup per baris' : undefined} value={form[field.name] || ''} onChange={(e) => setForm({ ...form, [field.name]: e.target.value })} />
              : <Input required={field.required} type={field.type === 'money' || field.type === 'decimal' ? 'number' : field.type} value={form[field.name] || ''} onChange={(e) => setForm({ ...form, [field.name]: e.target.value })} />}
          </div>
        ))}
      </div>
    </section>)}
    <section className="bg-white rounded-2xl border border-gray-100 p-6 mb-6">
      <h2 className="text-lg font-semibold mb-3" style={{ color: '#0B2545' }}>Lampiran Tambahan</h2>
      <input type="file" multiple onChange={(e) => setFiles(Array.from(e.target.files || []))} className="block w-full text-sm text-gray-600 file:mr-3 file:rounded-lg file:border-0 file:bg-gray-100 file:px-3 file:py-1.5 file:text-sm file:font-medium file:text-gray-700 hover:file:bg-gray-200" />
    </section>
    <Button onClick={submit} className="flex items-center gap-2 text-white transition hover:brightness-95" style={{ background: '#C8102E' }}>
      <Send className="w-4 h-4" /> {id ? 'Simpan dan Ajukan Ulang' : 'Kirim Pengajuan'}
    </Button>
  </div>
}
