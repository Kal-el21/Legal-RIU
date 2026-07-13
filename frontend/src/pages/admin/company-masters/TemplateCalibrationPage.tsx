import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Crosshair, Save, ArrowLeft, Loader2, Trash2, ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { companyMasterService } from '@/services/company-master.service'
import type { TemplateFieldPosition } from '@/types'

// A4 dimensions in millimetres — the coordinate space for field positions.
const PAGE_W_MM = 210
const PAGE_H_MM = 297

// The full set of fillable fields. `align` mirrors the backend defaults used by
// the PDF renderer. These keys MUST match the field names in pdf_fpdi.go.
const FIELD_DEFS: { key: string; label: string; align: 'L' | 'C' | 'R' }[] = [
  { key: 'pihak_kedua_nama', label: 'Nama Pihak Kedua', align: 'C' },
  { key: 'pihak_kedua_bidang', label: 'Bidang Pihak Kedua', align: 'C' },
  { key: 'jenis_pekerjaan', label: 'Jenis Pekerjaan', align: 'C' },
  { key: 'nomor_pihak_pertama', label: 'Nomor Pihak Pertama', align: 'C' },
  { key: 'nomor_pihak_kedua', label: 'Nomor Pihak Kedua', align: 'C' },
  { key: 'tempat_ttd', label: 'Tempat TTD', align: 'C' },
  { key: 'tanggal_ttd', label: 'Tanggal TTD', align: 'C' },
  { key: 'pihak_pertama_pejabat', label: 'Pejabat Pihak Pertama', align: 'C' },
  { key: 'pihak_pertama_jabatan', label: 'Jabatan Pihak Pertama', align: 'C' },
  { key: 'pihak_kedua_pejabat', label: 'Pejabat Pihak Kedua', align: 'C' },
  { key: 'pihak_kedua_jabatan', label: 'Jabatan Pihak Kedua', align: 'C' },
  { key: 'ruang_lingkup', label: 'Ruang Lingkup', align: 'L' },
  { key: 'nilai_kontrak', label: 'Nilai Kontrak', align: 'R' },
]

export default function TemplateCalibrationPage() {
  const { version = '1' } = useParams()
  const navigate = useNavigate()
  const imageRef = useRef<HTMLImageElement>(null)

  const [previewUrl, setPreviewUrl] = useState<string>('')
  const [positions, setPositions] = useState<Record<string, TemplateFieldPosition>>({})
  const [selected, setSelected] = useState<string>(FIELD_DEFS[0].key)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)

  useEffect(() => {
    let url = previewUrl
    setLoading(true)
    Promise.all([
      companyMasterService.getTemplatePreview(version, currentPage),
      companyMasterService.getFieldPositions(version).catch(() => [] as TemplateFieldPosition[]),
    ])
      .then(([u, saved]) => {
        url = u
        setPreviewUrl(u)
        const map: Record<string, TemplateFieldPosition> = {}
        for (const f of saved) {
          map[f.field_name] = f
        }
        setPositions(map)
      })
      .catch((err) => {
        const message =
          err?.response?.data?.message || err?.message || 'Unknown error'
        alert(
          `Gagal memuat template v${version}.\n\nAlasan: ${message}\n\nPastikan:\n1. Template sudah diupload\n2. Base PDF sudah di-generate\n3. Server dan LibreOffice berjalan dengan baik`,
        )
      })
      .finally(() => setLoading(false))
    return () => {
      if (url) URL.revokeObjectURL(url)
    }
  }, [version, currentPage])

  const setField = useCallback(
    (key: string, patch: Partial<TemplateFieldPosition>) => {
      setPositions((prev) => {
        const existing = prev[key] || {
          template_version: version,
          field_name: key,
          x: PAGE_W_MM / 2,
          y: PAGE_H_MM / 2,
          font: 'Arial',
          style: '',
          size: 11,
          align: FIELD_DEFS.find((f) => f.key === key)?.align || 'C',
          page_number: 1,
        }
        return { ...prev, [key]: { ...existing, ...patch, field_name: key } }
      })
    },
    [version],
  )

  const handleImageClick = (e: React.MouseEvent<HTMLImageElement>) => {
    const img = imageRef.current
    if (!img) return
    const rect = img.getBoundingClientRect()
    const px = e.clientX - rect.left
    const py = e.clientY - rect.top
    const xMm = (px / rect.width) * PAGE_W_MM
    const yMm = (py / rect.height) * PAGE_H_MM
    setField(selected, { x: Math.round(xMm * 10) / 10, y: Math.round(yMm * 10) / 10, page_number: currentPage })
  }

  const handleReset = () => {
    if (!confirm('Reset semua posisi ke default?')) return
    setPositions({})
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload: TemplateFieldPosition[] = FIELD_DEFS.map((def) => {
        const p = positions[def.key]
        return {
          template_version: version,
          field_name: def.key,
          x: p?.x ?? PAGE_W_MM / 2,
          y: p?.y ?? PAGE_H_MM / 2,
          font: 'Arial',
          style: '',
          size: 11,
          align: p?.align ?? def.align,
          page_number: p?.page_number ?? 1,
        }
      })
      await companyMasterService.saveFieldPositions(version, payload)
      alert('Posisi field berhasil disimpan')
    } catch (e: any) {
      alert(e?.response?.data?.message || 'Gagal menyimpan posisi')
    } finally {
      setSaving(false)
    }
  }

  const selectedPos = positions[selected]

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <Button variant="outline" size="sm" onClick={() => navigate('/admin/company-masters')}>
            <ArrowLeft className="w-4 h-4 mr-1" /> Kembali
          </Button>
          <div>
            <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Kalibrasi Template v{version}</h1>
            <p className="text-sm text-gray-500 mt-0.5">
              Klik pada pratinjau PDF untuk menempatkan field yang dipilih.
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleReset} disabled={saving}>
            <Trash2 className="w-4 h-4 mr-1" /> Reset
          </Button>
          <div className="flex items-center gap-1 border rounded-lg px-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              disabled={currentPage <= 1}
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <span className="text-sm font-medium w-16 text-center">
              Halaman {currentPage}
            </span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setCurrentPage((p) => Math.min(24, p + 1))}
              disabled={currentPage >= 24}
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
          <Button onClick={handleSave} disabled={saving} className="text-white" style={{ background: '#C8102E' }}>
            {saving ? <Loader2 className="w-4 h-4 mr-1 animate-spin" /> : <Save className="w-4 h-4 mr-1" />}
            Simpan Posisi
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-[280px_1fr] gap-6">
        {/* Sidebar */}
        <div className="bg-white rounded-2xl border border-gray-100 p-4 h-fit">
          <h2 className="text-sm font-semibold mb-3" style={{ color: '#0B2545' }}>Daftar Field</h2>
          <div className="space-y-1">
            {FIELD_DEFS.map((def) => {
              const pos = positions[def.key]
              const isSel = def.key === selected
              return (
                <button
                  key={def.key}
                  onClick={() => setSelected(def.key)}
                  className={`w-full text-left px-3 py-2 rounded-lg text-sm flex items-center justify-between transition-colors ${
                    isSel ? 'bg-red-50 text-[#C8102E] font-medium' : 'hover:bg-gray-50 text-gray-700'
                  }`}
                >
                  <span className="flex items-center gap-2">
                    <Crosshair className="w-3.5 h-3.5" />
                    {def.label}
                  </span>
                  {pos ? (
                    <span className="text-xs text-gray-500">
                      hal {pos.page_number ?? 1}
                    </span>
                  ) : (
                    <span className="text-xs text-gray-300">—</span>
                  )}
                </button>
              )
            })}
          </div>

          {selectedPos && (
            <div className="mt-4 pt-4 border-t border-gray-100 space-y-3">
              <div className="text-xs text-gray-500">Posisi field terpilih (mm)</div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">X</label>
                  <input
                    type="number" step="0.1" min="0" max={PAGE_W_MM}
                    value={selectedPos.x}
                    onChange={(e) => setField(selected, { x: parseFloat(e.target.value) || 0 })}
                    className="w-full px-2 py-1.5 rounded-lg border border-gray-200 text-sm outline-none focus:ring-2 focus:ring-red-100"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-600 mb-1">Y</label>
                  <input
                    type="number" step="0.1" min="0" max={PAGE_H_MM}
                    value={selectedPos.y}
                    onChange={(e) => setField(selected, { y: parseFloat(e.target.value) || 0 })}
                    className="w-full px-2 py-1.5 rounded-lg border border-gray-200 text-sm outline-none focus:ring-2 focus:ring-red-100"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Halaman</label>
                <input
                  type="number"
                  step="1"
                  min="1"
                  max={24}
                  value={selectedPos.page_number ?? 1}
                  onChange={(e) => setField(selected, { page_number: parseInt(e.target.value) || 1 })}
                  className="w-full px-2 py-1.5 rounded-lg border border-gray-200 text-sm outline-none focus:ring-2 focus:ring-red-100"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 mb-1">Align</label>
                <select
                  value={selectedPos.align ?? 'C'}
                  onChange={(e) => setField(selected, { align: e.target.value as 'L' | 'C' | 'R' })}
                  className="w-full px-2 py-1.5 rounded-lg border border-gray-200 text-sm outline-none focus:ring-2 focus:ring-red-100"
                >
                  <option value="L">Left</option>
                  <option value="C">Center</option>
                  <option value="R">Right</option>
                </select>
              </div>
            </div>
          )}
        </div>

        {/* Preview */}
        <div className="bg-white rounded-2xl border border-gray-100 p-4">
          {loading ? (
            <div className="flex items-center justify-center h-[70vh] text-gray-400">
              <Loader2 className="w-6 h-6 mr-2 animate-spin" /> Memuat pratinjau...
            </div>
          ) : (
            <div className="relative inline-block w-full">
              {previewUrl ? (
                <img
                  ref={imageRef}
                  src={previewUrl}
                  alt={`Template v${version}`}
                  onClick={handleImageClick}
                  className="w-full h-auto select-none cursor-crosshair rounded-lg border border-gray-200"
                  draggable={false}
                />
              ) : (
                <div className="flex items-center justify-center h-[60vh] text-gray-400 border border-dashed border-gray-200 rounded-lg">
                  Pratinjau tidak tersedia
                </div>
              )}

              {/* Markers */}
              {previewUrl &&
                FIELD_DEFS.map((def) => {
                  const pos = positions[def.key]
                  if (!pos) return null
                  if ((pos.page_number ?? 1) !== currentPage) return null
                  const leftPct = (pos.x / PAGE_W_MM) * 100
                  const topPct = (pos.y / PAGE_H_MM) * 100
                  const isSel = def.key === selected
                  return (
                    <div
                      key={def.key}
                      title={`${def.label} (halaman ${pos.page_number ?? 1})`}
                      onClick={() => setSelected(def.key)}
                      className={`absolute -translate-x-1/2 -translate-y-1/2 w-5 h-5 rounded-full border-2 border-white shadow cursor-pointer flex items-center justify-center text-[9px] font-bold ${
                        isSel ? 'bg-[#C8102E] z-10 scale-125' : 'bg-blue-500'
                      }`}
                      style={{ left: `${leftPct}%`, top: `${topPct}%`, color: 'white' }}
                    >
                      {FIELD_DEFS.indexOf(def) + 1}
                    </div>
                  )
                })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
