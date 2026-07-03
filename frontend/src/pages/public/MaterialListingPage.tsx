import { useMaterials } from '@/hooks/useMaterial'
import { Link } from 'react-router-dom'
import { FileText } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'

export default function MaterialListingPage() {
  const { data, isLoading } = useMaterials()

  const items = data?.items ?? []

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold" style={{ color: '#0B2545' }}>Materi Legal</h1>
        <p className="text-sm text-gray-500 mt-1">Kumpulan materi dan artikel legal</p>
      </div>

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="bg-white rounded-2xl border border-gray-100 p-6">
              <Skeleton className="h-5 w-3/4 mb-3" />
              <Skeleton className="h-4 w-full mb-2" />
              <Skeleton className="h-4 w-2/3" />
            </div>
          ))}
        </div>
      ) : !items.length ? (
        <div className="text-center py-16">
          <div className="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mx-auto mb-4">
            <FileText className="w-7 h-7 text-gray-400" />
          </div>
          <p className="font-medium text-gray-500">Belum ada materi</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((item) => (
            <Link key={item.id} to={`/materi-legal/${item.id}`} className="bg-white rounded-2xl border border-gray-100 p-6 hover:shadow-md transition-shadow">
              <h3 className="text-base font-semibold text-gray-900 mb-2 line-clamp-2">{item.title}</h3>
              {item.excerpt && <p className="text-sm text-gray-500 line-clamp-3 mb-4">{item.excerpt}</p>}
              <p className="text-xs text-gray-400">{new Date(item.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })}</p>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
