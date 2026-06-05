import { FileText, Award } from 'lucide-react'

const AKTA_ITEMS = [
  {
    icon: Award,
    title: 'Akta Perusahaan',
    desc: 'Akta pendirian dan perubahan anggaran dasar PT Reasuransi Indonesia Utama (Persero).',
    color: '#C8102E',
  },
  {
    icon: FileText,
    title: 'Akta Direksi',
    desc: 'Akta pengangkatan dan perubahan susunan Direksi PT Reasuransi Indonesia Utama (Persero).',
    color: '#0B2545',
  },
]

export default function AktaPerusahaanPage() {
  return (
    <div>
      {/* Hero */}
      <section className="relative overflow-hidden" style={{ background: 'linear-gradient(135deg, #0B2545 0%, #1A3A6B 100%)', minHeight: '280px' }}>
        <div className="absolute bottom-0 left-0 right-0 h-1" style={{ background: '#C8102E' }} />
        <div className="absolute inset-0 pointer-events-none overflow-hidden">
          <div className="absolute -top-20 -right-20 w-72 h-72 rounded-full opacity-10" style={{ background: '#C8102E' }} />
        </div>
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col justify-center" style={{ minHeight: '280px' }}>
          <p className="text-xs font-semibold uppercase tracking-widest mb-3" style={{ color: '#C8102E' }}>Dokumen Perusahaan</p>
          <h1 className="text-4xl lg:text-5xl font-bold text-white mb-4">Akta Perusahaan</h1>
          <p className="text-white/60 max-w-lg">
            Dokumen akta resmi PT Reasuransi Indonesia Utama (Persero) — Indonesia Re.
          </p>
        </div>
      </section>

      {/* Content */}
      <section className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-3xl mx-auto">
            {AKTA_ITEMS.map((item) => (
              <div key={item.title}
                className="group rounded-2xl p-10 border border-gray-100 hover:shadow-xl hover:-translate-y-1 transition-all duration-300 text-center cursor-pointer">
                <div className="w-20 h-20 rounded-2xl flex items-center justify-center mx-auto mb-6 transition-transform group-hover:scale-110"
                  style={{ background: item.color }}>
                  <item.icon className="w-9 h-9 text-white" />
                </div>
                <h3 className="text-xl font-bold mb-3" style={{ color: '#0B2545' }}>{item.title}</h3>
                <p className="text-sm text-gray-500 leading-relaxed">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}