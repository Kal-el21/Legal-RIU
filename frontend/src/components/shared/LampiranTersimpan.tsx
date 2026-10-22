import { FileText, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { formatFileSize } from '@/lib/utils'

interface Lampiran {
  id: string
  file_name: string
  file_size?: number
}

// Lampiran yang sudah tersimpan di draf (PR Draf). Menghapusnya permanen, jadi diminta
// konfirmasi dulu.
export default function LampiranTersimpan({ items, onHapus, disabled }: { items?: Lampiran[]; onHapus: (id: string) => void; disabled?: boolean }) {
  if (!items?.length) return null
  return (
    <div className="mb-4 space-y-2">
      <p className="text-xs font-medium text-gray-500">Sudah tersimpan di draf</p>
      {items.map((a) => (
        <div key={a.id} className="flex items-center gap-3 p-3 rounded-lg bg-gray-50 border border-gray-100">
          <FileText className="w-4 h-4 text-gray-400 flex-shrink-0" />
          <span className="text-sm text-gray-700 flex-1 truncate">{a.file_name}</span>
          {a.file_size != null && <span className="text-xs text-gray-400">{formatFileSize(a.file_size)}</span>}
          <Button variant="outline" size="icon-xs" type="button" disabled={disabled}
            onClick={() => { if (confirm(`Hapus lampiran "${a.file_name}" dari draf?`)) onHapus(a.id) }}
            aria-label="Hapus lampiran" title="Hapus lampiran" className="text-red-600 hover:bg-red-50 hover:text-red-700">
            <X className="w-4 h-4" />
          </Button>
        </div>
      ))}
    </div>
  )
}
