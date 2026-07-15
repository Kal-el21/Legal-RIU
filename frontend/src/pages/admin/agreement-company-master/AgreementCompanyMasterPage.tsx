import { useEffect, useState } from 'react'
import { Save } from 'lucide-react'
import { agreementService } from '@/services/agreement-document.service'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const fieldLabels: Record<string, string> = {
  name: 'Nama Perusahaan',
  address: 'Alamat',
  npwp: 'NPWP',
  phone: 'Telepon',
  email: 'Email',
  pic: 'PIC',
  default_signatory_name: 'Nama Penandatangan Default',
  default_signatory_position: 'Jabatan Penandatangan Default',
  default_signing_place: 'Tempat Penandatangan Default',
}

const fieldKeys = Object.keys(fieldLabels)

export default function AgreementCompanyMasterPage() {
  const [value, setValue] = useState<Record<string, string>>({})
  const [message, setMessage] = useState('')
  const [messageType, setMessageType] = useState<'success' | 'error' | ''>('')
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    void agreementService.master()
      .then((data) => setValue(data))
      .catch((error: Error) => { setMessage(error.message); setMessageType('error') })
      .finally(() => setIsLoading(false))
  }, [])

  const save = async () => {
    setIsSaving(true)
    setMessage('')
    setMessageType('')
    try {
      await agreementService.saveMaster(value)
      setMessage('Data berhasil disimpan')
      setMessageType('success')
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Gagal menyimpan data')
      setMessageType('error')
    } finally {
      setIsSaving(false)
    }
  }

  return <div className="p-6 max-w-7xl mx-auto">
    <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Master Pihak Pertama</h1>
    <p className="text-sm text-gray-500 mt-0.5 mb-6">Kelola data perusahaan pihak pertama</p>
    {message && <p className={`mb-4 text-sm ${messageType === 'error' ? 'text-red-600' : 'text-green-600'}`}>{message}</p>}
    <div className="bg-white rounded-2xl border border-gray-100 p-6">
      {isLoading ? (
        <div className="p-8 text-center text-gray-400">Memuat data...</div>
      ) : (
        <div className="grid md:grid-cols-2 gap-4">
          {fieldKeys.map((key) => (
            <div key={key}>
              <Label className="mb-1.5 text-gray-700">{fieldLabels[key]}</Label>
              <Input value={value[key] || ''} onChange={(event) => setValue({ ...value, [key]: event.target.value })} />
            </div>
          ))}
        </div>
      )}
      <Button onClick={() => void save()} disabled={isSaving} className="mt-5 flex items-center gap-2 text-white transition hover:brightness-95" style={{ background: '#C8102E' }}>
        <Save className="w-4 h-4" /> {isSaving ? 'Menyimpan...' : 'Simpan'}
      </Button>
    </div>
  </div>
}
