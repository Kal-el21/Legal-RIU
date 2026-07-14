import { useEffect, useState } from 'react'
import { agreementService } from '@/services/agreement-document.service'
import { Button } from '@/components/ui/button'

const fields = ['name', 'address', 'npwp', 'phone', 'email', 'pic', 'default_signatory_name', 'default_signatory_position', 'default_signing_place']

export default function AgreementCompanyMasterPage() {
  const [value, setValue] = useState<Record<string, string>>({})
  const [message, setMessage] = useState('')
  useEffect(() => { void agreementService.master().then(setValue).catch((error: Error) => setMessage(error.message)) }, [])
  const save = async () => { try { await agreementService.saveMaster(value); setMessage('Data berhasil disimpan') } catch (error) { setMessage(error instanceof Error ? error.message : 'Gagal menyimpan data') } }
  return <div className="p-6 max-w-3xl mx-auto"><h1 className="text-2xl font-bold text-[#0B2545] mb-5">Master Pihak Pertama</h1>{message && <p className="mb-3 text-sm">{message}</p>}<div className="bg-white border rounded-xl p-5 grid md:grid-cols-2 gap-4">{fields.map((field) => <label key={field}><span className="text-sm">{field.split('_').join(' ')}</span><input className="w-full border rounded p-2" value={value[field] || ''} onChange={(event) => setValue({ ...value, [field]: event.target.value })} /></label>)}<Button onClick={() => void save()}>Simpan</Button></div></div>
}
