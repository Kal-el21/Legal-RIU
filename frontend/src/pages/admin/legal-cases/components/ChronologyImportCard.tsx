import { useState } from 'react'
import { Upload, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useImportCaseChronology } from '@/hooks/useLegalCase'
import { legalCaseService } from '@/services/legal-case.service'

interface ChronologyImportCardProps {
  caseId: string
}

export default function ChronologyImportCard({ caseId }: ChronologyImportCardProps) {
  const importChronology = useImportCaseChronology(caseId)
  const [importFile, setImportFile] = useState<File | null>(null)
  const [downloadingTemplate, setDownloadingTemplate] = useState(false)

  const handleDownloadTemplate = async () => {
    try {
      setDownloadingTemplate(true)
      const { blob, filename } = await legalCaseService.downloadChronologyTemplate()
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
    try {
      await importChronology.mutateAsync(importFile)
      setImportFile(null)
    } catch {
      // Error rendered from mutation state.
    }
  }

  return (
    <div className="rounded-lg border border-gray-100 bg-gray-50 p-4">
      <p className="text-sm font-medium text-gray-700 mb-3">Impor dari Excel</p>
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
          disabled={!importFile || importChronology.isPending}
          className="text-white"
          style={{ background: '#C8102E' }}
        >
          <Upload className="h-4 w-4" />
          {importChronology.isPending ? 'Mengimpor...' : 'Import'}
        </Button>
      </div>
      {importChronology.isError && (
        <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600">
          {(importChronology.error as Error)?.message ?? 'Gagal mengimpor'}
        </p>
      )}
      {importChronology.isSuccess && importChronology.data && (
        <p className="mt-2 rounded-lg bg-green-50 px-3 py-2 text-xs text-green-700">
          Impor selesai: {importChronology.data.imported} berhasil, {importChronology.data.skipped} dilewati.
          {importChronology.data.errors.length > 0 && (
            <span className="block mt-1">
              Baris gagal: {importChronology.data.errors.map((e) => `#${e.row}`).join(', ')}
            </span>
          )}
        </p>
      )}
    </div>
  )
}
