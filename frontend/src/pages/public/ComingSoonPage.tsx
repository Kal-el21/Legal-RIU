import { Clock } from 'lucide-react'
import { Link } from 'react-router-dom'

interface ComingSoonPageProps {
  title: string
  description?: string
}

export default function ComingSoonPage({ title, description }: ComingSoonPageProps) {
  return (
    <div>
      {/* Hero */}
      <section className="relative overflow-hidden" style={{ background: 'linear-gradient(135deg, #0B2545 0%, #1A3A6B 100%)', minHeight: '280px' }}>
        <div className="absolute bottom-0 left-0 right-0 h-1" style={{ background: '#C8102E' }} />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col justify-center" style={{ minHeight: '280px' }}>
          <p className="text-xs font-semibold uppercase tracking-widest mb-3" style={{ color: '#C8102E' }}>Legal RIU</p>
          <h1 className="text-4xl lg:text-5xl font-bold text-white mb-4">{title}</h1>
          {description && <p className="text-white/60 max-w-lg">{description}</p>}
        </div>
      </section>

      {/* Coming soon body */}
      <section className="py-32 bg-white">
        <div className="max-w-md mx-auto px-4 text-center">
          <div className="w-20 h-20 rounded-2xl flex items-center justify-center mx-auto mb-8" style={{ background: '#f1f5f9' }}>
            <Clock className="w-9 h-9 text-gray-400" />
          </div>
          <h2 className="text-2xl font-bold mb-3" style={{ color: '#0B2545' }}>Segera Hadir</h2>
          <p className="text-gray-500 text-sm leading-relaxed mb-8">
            Halaman ini sedang dalam pengembangan dan akan segera tersedia. Pantau terus pembaruan dari Legal RIU.
          </p>
          <Link to="/"
            className="inline-flex items-center gap-2 px-6 py-3 rounded-xl text-sm font-semibold text-white transition-all hover:opacity-90"
            style={{ background: '#C8102E' }}>
            Kembali ke Beranda
          </Link>
        </div>
      </section>
    </div>
  )
}