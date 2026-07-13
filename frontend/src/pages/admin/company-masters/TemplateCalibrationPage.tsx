import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Crosshair, Save, ArrowLeft, Loader2, Trash2, ChevronLeft, ChevronRight, Plus, CircleDot } from 'lucide-react'
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
  { key: 'pihak_kedua_alamat', label: 'Alamat Pihak Kedua', align: 'L' },
  { key: 'pihak_kedua_telepon', label: 'Telepon Pihak Kedua', align: 'C' },
  { key: 'pihak_kedua_email', label: 'Email Pihak Kedua', align: 'C' },
  { key: 'pihak_kedua_pic', label: 'PIC Pihak Kedua', align: 'C' },
  { key: 'jenis_pekerjaan', label: 'Jenis Pekerjaan', align: 'C' },
  { key: 'nomor_pihak_pertama', label: 'Nomor Pihak Pertama', align: 'C' },
  { key: 'nomor_pihak_kedua', label: 'Nomor Pihak Kedua', align: 'C' },
  { key: 'surat_penawaran_nomor', label: 'No. Surat Penawaran', align: 'C' },
  { key: 'surat_penawaran_perihal', label: 'Perihal Surat Penawaran', align: 'L' },
  { key: 'surat_penawaran_tanggal', label: 'Tanggal Surat Penawaran', align: 'C' },
  { key: 'surat_penunjukan_nomor', label: 'No. Surat Penunjukan', align: 'C' },
  { key: 'surat_penunjukan_perihal', label: 'Perihal Surat Penunjukan', align: 'L' },
  { key: 'surat_penunjukan_tanggal', label: 'Tanggal Surat Penunjukan', align: 'C' },
  { key: 'jangka_waktu_mulai', label: 'Jangka Waktu Mulai', align: 'C' },
  { key: 'jangka_waktu_selesai', label: 'Jangka Waktu Selesai', align: 'C' },
  { key: 'tempat_ttd', label: 'Tempat TTD', align: 'C' },
  { key: 'tanggal_ttd', label: 'Tanggal TTD', align: 'C' },
  { key: 'pihak_pertama_pejabat', label: 'Pejabat Pihak Pertama', align: 'C' },
  { key: 'pihak_pertama_jabatan', label: 'Jabatan Pihak Pertama', align: 'C' },
  { key: 'pihak_kedua_pejabat', label: 'Pejabat Pihak Kedua', align: 'C' },
  { key: 'pihak_kedua_jabatan', label: 'Jabatan Pihak Kedua', align: 'C' },
  { key: 'ruang_lingkup', label: 'Ruang Lingkup', align: 'L' },
  { key: 'nilai_kontrak', label: 'Nilai Kontrak', align: 'R' },
  { key: 'termin1_persen', label: 'Termin 1 Persen', align: 'C' },
  { key: 'termin1_nilai', label: 'Termin 1 Nilai', align: 'R' },
  { key: 'termin2_persen', label: 'Termin 2 Persen', align: 'C' },
  { key: 'termin2_nilai', label: 'Termin 2 Nilai', align: 'R' },
  { key: 'bank', label: 'Bank', align: 'C' },
  { key: 'nomor_rekening', label: 'No. Rekening', align: 'C' },
  { key: 'atas_nama', label: 'Atas Nama', align: 'C' },
  { key: 'lampiran', label: 'Lampiran', align: 'L' },
]

const fieldDefOf = (key: string) => FIELD_DEFS.find((f) => f.key === key)
const fieldLabel = (key: string) => fieldDefOf(key)?.label ?? key

export default function TemplateCalibrationPage() {
  const { version = '1' } = useParams()
  const navigate = useNavigate()
  const imageRef = useRef<HTMLImageElement>(null)

  // occurrences: field name -> list of calibrated positions (one per occurrence)
  const [occurrences, setOccurrences] = useState<Record<string, TemplateFieldPosition[]>>({})
  const [selectedField, setSelectedField] = useState<string>(FIELD_DEFS[0].key)
  const [selectedOcc, setSelectedOcc] = useState<number | null>(0)
  const [previewUrl, setPreviewUrl] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)

  // Preview image — re-fetch only when page or version changes. This must NOT
  // depend on version alone, otherwise changing page wouldn't update the image.
  useEffect(() => {
    let url = ''
    setLoading(true)
    companyMasterService
      .getTemplatePreview(version, currentPage)
      .then((u) => {
        url = u
        setPreviewUrl(u)
      })
      .catch(() => {
        alert('Gagal memuat preview halaman ' + currentPage)
      })
      .finally(() => setLoading(false))
    return () => {
      if (url) URL.revokeObjectURL(url)
    }
  }, [version, currentPage])

  // Field positions — fetch ONLY when version changes. Keeping currentPage out
  // of this dependency prevents page navigation from overwriting unsaved local
  // occurrences (the root cause of markers disappearing on page change).
  useEffect(() => {
    let mounted = true
    companyMasterService
      .getFieldPositions(version)
      .then((saved) => {
        if (!mounted) return
        const grouped: Record<string, TemplateFieldPosition[]> = {}
        for (const f of saved) {
          if (!grouped[f.field_name]) grouped[f.field_name] = []
          grouped[f.field_name].push(f)
        }
        setOccurrences(grouped)
        const firstField = FIELD_DEFS[0].key
        setSelectedField(firstField)
        setSelectedOcc(grouped[firstField]?.length ? 0 : null)
      })
      .catch(() => {
        if (!mounted) return
        alert('Gagal memuat posisi field')
      })
    return () => {
      mounted = false
    }
  }, [version])

  const setOccurrence = useCallback(
    (field: string, index: number, patch: Partial<TemplateFieldPosition>) => {
      setOccurrences((prev) => {
        const list = [...(prev[field] || [])]
        const base: TemplateFieldPosition =
          list[index] ||
          {
            template_version: version,
            field_name: field,
            x: PAGE_W_MM / 2,
            y: PAGE_H_MM / 2,
            font: 'Arial',
            style: '',
            size: 11,
            align: fieldDefOf(field)?.align || 'C',
            page_number: 1,
            occurrence_index: list.length + 1,
          }
        list[index] = { ...base, ...patch, field_name: field, occurrence_index: index + 1 }
        return { ...prev, [field]: list }
      })
    },
    [version],
  )

  const deleteOccurrence = useCallback((field: string, index: number) => {
    setOccurrences((prev) => {
      const list = (prev[field] || []).filter((_, i) => i !== index)
      // occurrence_index kept as-is; handleSave assigns from array index
      return { ...prev, [field]: list }
    })
    setSelectedOcc((cur) => (cur === index ? null : cur && cur > index ? cur - 1 : cur))
  }, [])

  const addOccurrence = useCallback(
    (field: string) => {
      let newIndex = 0
      setOccurrences((prev) => {
        const list = [...(prev[field] || [])]
        newIndex = list.length
        list.push({
          template_version: version,
          field_name: field,
          x: PAGE_W_MM / 2,
          y: PAGE_H_MM / 2,
          font: 'Arial',
          style: '',
          size: 11,
          align: fieldDefOf(field)?.align || 'C',
          page_number: currentPage,
          occurrence_index: newIndex + 1,
        })
        return { ...prev, [field]: list }
      })
      setTimeout(() => setSelectedOcc(newIndex), 0)
    },
    [version, currentPage],
  )

  const handleImageClick = (e: React.MouseEvent<HTMLImageElement>) => {
    const img = imageRef.current
    if (!img) return
    const rect = img.getBoundingClientRect()
    const px = e.clientX - rect.left
    const py = e.clientY - rect.top
    const xMm = (px / rect.width) * PAGE_W_MM
    const yMm = (py / rect.height) * PAGE_H_MM

    const occs = occurrences[selectedField] || []
    if (selectedOcc !== null && occs[selectedOcc]) {
      setOccurrence(selectedField, selectedOcc, {
        x: Math.round(xMm * 10) / 10,
        y: Math.round(yMm * 10) / 10,
        page_number: currentPage,
      })
    } else {
      // No occurrence selected for this field yet — create a new one here.
      const newIndex = occs.length
      setOccurrence(selectedField, newIndex, {
        x: Math.round(xMm * 10) / 10,
        y: Math.round(yMm * 10) / 10,
        page_number: currentPage,
      })
      setSelectedOcc(newIndex)
    }
  }

  const handleReset = () => {
    if (!confirm('Reset semua posisi ke default?')) return
    setOccurrences({})
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload: TemplateFieldPosition[] = []
      for (const def of FIELD_DEFS) {
        const list = occurrences[def.key] || []
        list.forEach((p, i) => {
          payload.push({
            template_version: version,
            field_name: def.key,
            occurrence_index: i + 1,
            x: p?.x ?? PAGE_W_MM / 2,
            y: p?.y ?? PAGE_H_MM / 2,
            font: 'Arial',
            style: '',
            size: 11,
            align: p?.align ?? def.align,
            page_number: p?.page_number ?? 1,
          })
        })
      }
      await companyMasterService.saveFieldPositions(version, payload)
      alert('Posisi field berhasil disimpan')
    } catch (e: any) {
      alert(e?.response?.data?.message || 'Gagal menyimpan posisi')
    } finally {
      setSaving(false)
    }
  }

  const selectedOccs = occurrences[selectedField] || []
  const selectedPos = selectedOcc !== null ? selectedOccs[selectedOcc] : undefined
  const fieldIndex = FIELD_DEFS.findIndex((f) => f.key === selectedField)

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

      <div className="grid grid-cols-1 lg:grid-cols-[300px_1fr] gap-6">
        {/* Sidebar */}
        <div className="bg-white rounded-2xl border border-gray-100 p-4 h-fit">
          <h2 className="text-sm font-semibold mb-3" style={{ color: '#0B2545' }}>Daftar Field</h2>
          <div className="space-y-1 max-h-[60vh] overflow-y-auto pr-1">
            {FIELD_DEFS.map((def) => {
              const occs = occurrences[def.key] || []
              const isSel = def.key === selectedField
              return (
                <button
                  key={def.key}
                  onClick={() => {
                    setSelectedField(def.key)
                    setSelectedOcc(occs.length ? 0 : null)
                  }}
                  className={`w-full text-left px-3 py-2 rounded-lg text-sm flex items-center justify-between transition-colors ${
                    isSel ? 'bg-red-50 text-[#C8102E] font-medium' : 'hover:bg-gray-50 text-gray-700'
                  }`}
                >
                  <span className="flex items-center gap-2">
                    <Crosshair className="w-3.5 h-3.5" />
                    {def.label}
                  </span>
                  {occs.length > 0 ? (
                    <span className="text-xs text-gray-500">
                      {occs.length} kemunculan
                    </span>
                  ) : (
                    <span className="text-xs text-gray-300">—</span>
                  )}
                </button>
              )
            })}
          </div>

          {/* Occurrences of the selected field */}
          {selectedField && (
            <div className="mt-4 pt-4 border-t border-gray-100">
              <div className="flex items-center justify-between mb-2">
                <h3 className="text-xs font-semibold text-gray-600">Kemunculan: {fieldLabel(selectedField)}</h3>
                <Button variant="outline" size="sm" onClick={() => addOccurrence(selectedField)}>
                  <Plus className="w-3.5 h-3.5 mr-1" /> Tambah
                </Button>
              </div>

              {selectedOccs.length === 0 ? (
                <p className="text-xs text-gray-400">Belum ada. Klik gambar untuk menempatkan.</p>
              ) : (
                <div className="space-y-1">
                  {selectedOccs.map((pos, idx) => {
                    const active = selectedOcc === idx
                    return (
                      <div
                        key={idx}
                        className={`flex items-center justify-between px-2 py-1.5 rounded-lg text-xs ${
                          active ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-50 text-gray-600'
                        }`}
                      >
                        <button
                          className="flex items-center gap-1.5"
                          onClick={() => setSelectedOcc(idx)}
                        >
                          <CircleDot className="w-3.5 h-3.5" />
                          Kemunculan {idx + 1} — hal {pos.page_number ?? 1}
                        </button>
                        <button
                          className="text-red-500 hover:text-red-700"
                          onClick={() => deleteOccurrence(selectedField, idx)}
                          title="Hapus kemunculan"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    )
                  })}
                </div>
              )}

              {/* Editor for the selected occurrence */}
              {selectedPos && (
                <div className="mt-3 space-y-3">
                  <div className="text-xs text-gray-500">
                    Posisi kemunculan {selectedOcc! + 1} (mm) — hal {selectedPos.page_number ?? 1}
                  </div>
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label className="block text-xs font-medium text-gray-600 mb-1">X</label>
                      <input
                        type="number" step="0.1" min="0" max={PAGE_W_MM}
                        value={selectedPos.x}
                        onChange={(e) => setOccurrence(selectedField, selectedOcc!, { x: parseFloat(e.target.value) || 0 })}
                        className="w-full px-2 py-1.5 rounded-lg border border-gray-200 text-sm outline-none focus:ring-2 focus:ring-red-100"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-600 mb-1">Y</label>
                      <input
                        type="number" step="0.1" min="0" max={PAGE_H_MM}
                        value={selectedPos.y}
                        onChange={(e) => setOccurrence(selectedField, selectedOcc!, { y: parseFloat(e.target.value) || 0 })}
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
                      onChange={(e) => setOccurrence(selectedField, selectedOcc!, { page_number: parseInt(e.target.value) || 1 })}
                      className="w-full px-2 py-1.5 rounded-lg border border-gray-200 text-sm outline-none focus:ring-2 focus:ring-red-100"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-gray-600 mb-1">Align</label>
                    <select
                      value={selectedPos.align ?? 'C'}
                      onChange={(e) => setOccurrence(selectedField, selectedOcc!, { align: e.target.value as 'L' | 'C' | 'R' })}
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

              {/* Markers — one per occurrence, only those on the current page */}
              {previewUrl &&
                FIELD_DEFS.flatMap((def) =>
                  (occurrences[def.key] || []).map((pos, idx) => {
                    if ((pos.page_number ?? 1) !== currentPage) return null
                    const isSel = def.key === selectedField && idx === selectedOcc
                    return (
                      <div
                        key={`${def.key}-${idx}`}
                        title={`${def.label} (kemunculan ${idx + 1}, halaman ${pos.page_number ?? 1})`}
                        onClick={() => {
                          setSelectedField(def.key)
                          setSelectedOcc(idx)
                        }}
                        className={`absolute -translate-x-1/2 -translate-y-1/2 w-5 h-5 rounded-full border-2 border-white shadow cursor-pointer flex items-center justify-center text-[9px] font-bold ${
                          isSel ? 'bg-[#C8102E] z-10 scale-125' : 'bg-blue-500'
                        }`}
                        style={{ left: `${(pos.x / PAGE_W_MM) * 100}%`, top: `${(pos.y / PAGE_H_MM) * 100}%`, color: 'white' }}
                      >
                        {fieldIndex + 1}
                      </div>
                    )
                  }),
                )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
