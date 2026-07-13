import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2, Upload, FileText, Crosshair } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { companyMasterService } from '@/services/company-master.service'
import type { CompanyMaster, CompanyMasterTemplate } from '@/types'

export default function CompanyMasterManagementPage() {
  const navigate = useNavigate()
  const [items, setItems] = useState<CompanyMaster[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState<Partial<CompanyMaster>>({})
  const [saving, setSaving] = useState(false)
  const [templates, setTemplates] = useState<CompanyMasterTemplate[]>([])
  const [uploading, setUploading] = useState(false)
  const [uploadVersion, setUploadVersion] = useState('1')
  const [uploadFile, setUploadFile] = useState<File | null>(null)

  const load = () => {
    setLoading(true)
    companyMasterService.getAll().then(setItems).finally(() => setLoading(false))
  }

  const loadTemplates = async () => {
    try {
      const active = await companyMasterService.getActiveTemplate()
      if (active) {
        setTemplates([active])
      } else {
        setTemplates([])
      }
    } catch {
      setTemplates([])
    }
  }

  useEffect(() => { load(); loadTemplates() }, [])

  const openCreate = () => {
    setEditingId(null)
    setForm({ is_active: true })
    setShowForm(true)
  }

  const openEdit = (item: CompanyMaster) => {
    setEditingId(item.id)
    setForm({ ...item })
    setShowForm(true)
  }

  const save = async () => {
    if (!form.name) {
      alert('Nama perusahaan wajib diisi')
      return
    }
    setSaving(true)
    try {
      if (editingId) {
        await companyMasterService.update(editingId, form as CompanyMaster)
      } else {
        await companyMasterService.create(form as CompanyMaster)
      }
      setShowForm(false)
      load()
    } catch (e: any) {
      alert(e?.response?.data?.message || 'Gagal menyimpan')
    } finally {
      setSaving(false)
    }
  }

  const remove = async (id: string) => {
    if (!confirm('Hapus data ini?')) return
    try {
      await companyMasterService.delete(id)
      load()
    } catch (e: any) {
      alert(e?.response?.data?.message || 'Gagal menghapus')
    }
  }

  const handleUploadTemplate = async () => {
    if (!uploadFile || !uploadVersion) {
      alert('Pilih file dan versi terlebih dahulu')
      return
    }
    setUploading(true)
    try {
      await companyMasterService.uploadTemplate(uploadVersion, uploadFile)
      setUploadFile(null)
      setUploadVersion('1')
      loadTemplates()
      alert('Template berhasil diupload')
    } catch (e: any) {
      alert(e?.response?.data?.message || 'Gagal upload template')
    } finally {
      setUploading(false)
    }
  }

  const handleDeleteTemplate = async (version: string) => {
    if (!confirm(`Hapus template v${version}?`)) return
    try {
      await companyMasterService.deleteTemplate(version)
      loadTemplates()
    } catch (e: any) {
      alert(e?.response?.data?.message || 'Gagal menghapus template')
    }
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Master Data Pihak Pertama</h1>
          <p className="text-sm text-gray-500 mt-0.5">Kelola data perusahaan Pihak Pertama dan template dokumen perjanjian</p>
        </div>
        <Button className="flex items-center gap-2 text-white" style={{ background: '#C8102E' }} onClick={openCreate}>
          <Plus className="w-4 h-4" /> Tambah Data
        </Button>
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 p-6 mb-6">
        <h2 className="text-lg font-semibold mb-4" style={{ color: '#0B2545' }}>
          Template Dokumen Perjanjian (.docx)
        </h2>
        <p className="text-sm text-gray-500 mb-4">
          Upload template .docx untuk digunakan dalam generate PDF perjanjian. Template akan dikonversi otomatis menjadi base PDF.
        </p>

        <div className="flex items-end gap-4 mb-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Versi</label>
            <input
              type="text"
              value={uploadVersion}
              onChange={(e) => setUploadVersion(e.target.value)}
              placeholder="1"
              className="w-24 px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm"
            />
          </div>
          <div className="flex-1">
            <label className="block text-sm font-medium text-gray-700 mb-1.5">File Template (.docx)</label>
            <input
              type="file"
              accept=".docx,.doc"
              onChange={(e) => setUploadFile(e.target.files?.[0] || null)}
              className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm"
            />
          </div>
          <Button onClick={handleUploadTemplate} disabled={uploading} className="text-white" style={{ background: '#C8102E' }}>
            <Upload className="w-4 h-4 mr-2" />
            {uploading ? 'Mengupload...' : 'Upload Template'}
          </Button>
        </div>

        {templates.length > 0 && (
          <div className="border border-gray-100 rounded-xl overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                  <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Versi</th>
                  <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Path Template</th>
                  <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Path Base PDF</th>
                  <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Uploaded At</th>
                  <th className="px-4 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {templates.map((tmpl) => (
                  <tr key={tmpl.version} className="hover:bg-gray-50/50 transition-colors">
                    <td className="px-4 py-3 text-sm text-gray-700 font-medium">{tmpl.version}</td>
                    <td className="px-4 py-3 text-sm text-gray-500">
                      <div className="flex items-center gap-2">
                        <FileText className="w-4 h-4 text-gray-400" />
                        {tmpl.template_path}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">
                      <div className="flex items-center gap-2">
                        <FileText className="w-4 h-4 text-gray-400" />
                        {tmpl.base_pdf_path}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">{new Date(tmpl.uploaded_at).toLocaleString('id-ID')}</td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => navigate(`/admin/company-masters/template/${tmpl.version}/calibrate`)}
                          className="p-1.5 rounded hover:bg-gray-100 text-blue-600"
                          title="Kalibrasi posisi field"
                        >
                          <Crosshair className="w-4 h-4" />
                        </button>
                        <button onClick={() => handleDeleteTemplate(tmpl.version)} className="p-1.5 rounded hover:bg-gray-100" title="Hapus">
                          <Trash2 className="w-4 h-4 text-red-500" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {templates.length === 0 && (
          <div className="p-8 text-center border border-dashed border-gray-200 rounded-xl">
            <p className="text-sm text-gray-400">Belum ada template yang diupload</p>
          </div>
        )}
      </div>

      {showForm && (
        <div className="bg-white rounded-2xl border border-gray-100 p-6 mb-6">
          <h2 className="text-lg font-semibold mb-4" style={{ color: '#0B2545' }}>
            {editingId ? 'Edit Data' : 'Tambah Data'}
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Nama Perusahaan *</label>
              <input value={form.name || ''} onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Alamat</label>
              <input value={form.address || ''} onChange={(e) => setForm({ ...form, address: e.target.value })}
                className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">NPWP</label>
              <input value={form.npwp || ''} onChange={(e) => setForm({ ...form, npwp: e.target.value })}
                className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Telepon</label>
              <input value={form.phone || ''} onChange={(e) => setForm({ ...form, phone: e.target.value })}
                className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Email</label>
              <input value={form.email || ''} onChange={(e) => setForm({ ...form, email: e.target.value })}
                className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Default Pejabat</label>
              <input value={form.default_pejabat || ''} onChange={(e) => setForm({ ...form, default_pejabat: e.target.value })}
                className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Default Jabatan</label>
              <input value={form.default_jabatan || ''} onChange={(e) => setForm({ ...form, default_jabatan: e.target.value })}
                className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Default Tempat TTD</label>
              <input value={form.default_tempat_ttd || ''} onChange={(e) => setForm({ ...form, default_tempat_ttd: e.target.value })}
                className="w-full px-3.5 py-2.5 rounded-xl border border-gray-200 focus:ring-2 focus:ring-red-100 outline-none text-sm" />
            </div>
            <div className="flex items-center gap-2">
              <input type="checkbox" id="is_active" checked={form.is_active || false}
                onChange={(e) => setForm({ ...form, is_active: e.target.checked })}
                className="w-4 h-4 rounded border-gray-300 text-red-600 focus:ring-red-500" />
              <label htmlFor="is_active" className="text-sm font-medium text-gray-700">Aktif</label>
            </div>
          </div>
          <div className="mt-6 flex gap-3">
            <Button onClick={save} disabled={saving} className="text-white" style={{ background: '#C8102E' }}>
              {saving ? 'Menyimpan...' : 'Simpan'}
            </Button>
            <Button onClick={() => setShowForm(false)} variant="outline">Batal</Button>
          </div>
        </div>
      )}

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {loading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !items.length ? (
          <div className="p-16 text-center">
            <p className="font-medium text-gray-500">Belum ada data</p>
          </div>
        ) : (
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase">Nama</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase">Alamat</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase">Telepon</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase">Email</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase">Default Pejabat</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase">Default Jabatan</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase">Tempat TTD</th>
                <th className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3.5" />
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {items.map((item) => (
                <tr key={item.id} className="hover:bg-gray-50/50 transition-colors">
                  <td className="px-6 py-4 text-sm text-gray-700">{item.name}</td>
                  <td className="px-6 py-4 text-sm text-gray-500 max-w-[200px] truncate">{item.address || '-'}</td>
                  <td className="px-6 py-4 text-sm text-gray-500">{item.phone || '-'}</td>
                  <td className="px-6 py-4 text-sm text-gray-500">{item.email || '-'}</td>
                  <td className="px-6 py-4 text-sm text-gray-500">{item.default_pejabat || '-'}</td>
                  <td className="px-6 py-4 text-sm text-gray-500">{item.default_jabatan || '-'}</td>
                  <td className="px-6 py-4 text-sm text-gray-500">{item.default_tempat_ttd || '-'}</td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 rounded text-xs font-medium ${item.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                      {item.is_active ? 'Aktif' : 'Nonaktif'}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex gap-2 justify-end">
                      <button onClick={() => openEdit(item)} className="p-1.5 rounded hover:bg-gray-100" title="Edit">
                        <Pencil className="w-4 h-4 text-gray-500" />
                      </button>
                      <button onClick={() => remove(item.id)} className="p-1.5 rounded hover:bg-gray-100" title="Hapus">
                        <Trash2 className="w-4 h-4 text-red-500" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
