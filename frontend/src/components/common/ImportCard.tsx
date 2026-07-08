import { useState } from 'react'
import { Upload, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface ImportCardProps {
  title: string
  onImport: (file: File) => Promise<{ imported: number; skipped: number; errors: { row: number; field: string; reason: string }[] }>
  onDownloadTemplate: () => Promise<{ blob: Blob; filename: string }>
}

export default function ImportCard({ title, onImport, onDownloadTemplate }: ImportCardProps) {
  const [importFile, setImportFile] = useState<File | null>(null)
  const [downloadingTemplate, setDownloadingTemplate] = useState(false)
  const [importing, setImporting] = useState(false)
  const [result, setResult] = useState<{ imported: number; skipped: number; errors: { row: number; field: string; reason: string }[] } | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleDownloadTemplate = async () => {
    try {
      setDownloadingTemplate(true)
      const { blob, filename } = await onDownloadTemplate()
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = filename
      anchor.click()
      URL.revokeObjectURL(url)
    } finally {
      setDownloadingTemplate(false)
    }
  }

  const handleImport = async () => {
    if (!importFile) return
    setImporting(true)
    setError(null)
    setResult(null)
    try {
      const res = await onImport(importFile)
      setResult(res)
      setImportFile(null)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Gagal mengimpor')
    } finally {
      setImporting(false)
    }
  }

  return (
    <div className="rounded-lg border border-gray-100 bg-gray-50 p-4">
      <p className="text-sm font-medium text-gray-700 mb-3">{title}</p>
      <div className="flex flex-wrap items-center gap-3">
        <input
          type="file"
          accept=".xlsx"
          className="block w-full text-xs text-gray-600 file:mr-3 file:rounded-lg file:border-0 file:bg-[#C8102E] file:px-3 file:py-1.5 file:text-white hover:file:opacity-90"
          onChange={(e) => setImportFile(e.target.files?.[0] ?? null)}
        />
        <Button type="button" variant="outline" onClick={handleDownloadTemplate} disabled={downloadingTemplate}>
          <Download className="h-4 w-4" />
          Unduh Template
        </Button>
        <Button
          type="button"
          onClick={handleImport}
          disabled={!importFile || importing}
          className="text-white"
          style={{ background: '#C8102E' }}
        >
          <Upload className="h-4 w-4" />
          {importing ? 'Mengimpor...' : 'Import'}
        </Button>
      </div>
      {error && (
        <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600">
          {error}
        </p>
      )}
      {result && (
        <p className="mt-2 rounded-lg bg-green-50 px-3 py-2 text-xs text-green-700">
          Impor selesai: {result.imported} berhasil, {result.skipped} dilewati.
          {result.errors.length > 0 && (
            <span className="block mt-1">
              Baris gagal: {result.errors.map((e) => `#${e.row} (${e.field}: ${e.reason})`).join(', ')}
            </span>
          )}
        </p>
      )}
    </div>
  )
}
