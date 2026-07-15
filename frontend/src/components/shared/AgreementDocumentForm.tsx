import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { agreementService, type AgreementSchema } from '@/services/agreement-document.service'
import { Button } from '@/components/ui/button'

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
    <h1 className="text-2xl font-bold text-[#0B2545] mb-6">{id ? 'Revisi' : 'Pengajuan'} Perjanjian Kerja Sama</h1>
    {schema?.sections.map((section) => <section key={section.title} className="bg-white border rounded-xl p-5 mb-4">
      <h2 className="font-semibold mb-4">{section.title}</h2>
      <div className="grid md:grid-cols-2 gap-4">{section.fields.map((field) => <label key={field.name} className={field.type === 'textarea' ? 'md:col-span-2' : ''}>
        <span className="text-sm">{field.label}{field.required && ' *'}</span>
        {field.type === 'textarea'
          ? <textarea required={field.required} placeholder={field.name === 'ruang_lingkup' ? 'Tulis satu poin ruang lingkup per baris' : undefined} className="mt-1 w-full border rounded-lg p-2" value={form[field.name] || ''} onChange={(e) => setForm({ ...form, [field.name]: e.target.value })} />
          : <input required={field.required} type={field.type === 'money' || field.type === 'decimal' ? 'number' : field.type} className="mt-1 w-full border rounded-lg p-2" value={form[field.name] || ''} onChange={(e) => setForm({ ...form, [field.name]: e.target.value })} />}
      </label>)}</div>
    </section>)}
    <section className="bg-white border rounded-xl p-5 mb-6"><h2 className="font-semibold mb-3">Attachment tambahan</h2><input type="file" multiple onChange={(e) => setFiles(Array.from(e.target.files || []))} /></section>
    <Button onClick={submit}>{id ? 'Simpan dan Ajukan Ulang' : 'Kirim Pengajuan'}</Button>
  </div>
}
