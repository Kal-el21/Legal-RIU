import { useParams, useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { formatDateTime } from '@/lib/utils'
import { useMaterial } from '@/hooks/useMaterial'
import RichTextViewer from '@/components/common/RichTextViewer'

export default function LegalMaterialDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const { data: material, isLoading } = useMaterial(id!)

  if (isLoading) return <div className="p-12 text-center text-gray-400">Memuat data...</div>
  if (!material) return <div className="p-12 text-center text-gray-500">Materi tidak ditemukan</div>

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="flex items-start gap-3 mb-6">
        <button onClick={() => navigate(-1)} className="p-2 rounded-lg hover:bg-gray-100 mt-0.5">
          <ArrowLeft className="w-5 h-5 text-gray-600" />
        </button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>{material.title}</h1>
          {material.excerpt && (
            <p className="text-sm text-gray-500 mt-0.5">{material.excerpt}</p>
          )}
        </div>
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 p-6 space-y-5">
        <div>
          <h3 className="text-sm font-semibold mb-3" style={{ color: '#0B2545' }}>Konten</h3>
          <RichTextViewer content={material.content} />
        </div>

        <div className="border-t border-gray-100 pt-4">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-xs text-gray-400">Dibuat</p>
              <p className="text-gray-700">{formatDateTime(material.created_at)}</p>
            </div>
            <div>
              <p className="text-xs text-gray-400">Diperbarui</p>
              <p className="text-gray-700">{formatDateTime(material.updated_at)}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}