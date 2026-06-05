import { Link } from 'react-router-dom'
import { ArrowRight, FileText, FileSearch, MessageSquare, ExternalLink, BookOpen, Building2, Scale, ChevronRight } from 'lucide-react'
import { useAuthStore } from '@/store/auth.store'

const SERVICES = [
  {
    icon: FileText,
    title: 'Legal Opinion',
    desc: 'Pengajuan LO kepada Konsultan Hukum yang terdaftar dalam database Legal.',
    href: '/dashboard/legal-opinions/new',
    available: true,
    accent: '#C8102E',
  },
  {
    icon: FileSearch,
    title: 'Review Dokumen',
    desc: 'Melakukan review dokumen hukum perusahaan secara profesional.',
    href: '/dashboard/review-documents/new',
    available: true,
    accent: '#0B2545',
  },
  {
    icon: MessageSquare,
    title: 'Konsultasi Legal',
    desc: 'Konsultasi secara langsung dan tidak langsung kepada team Legal RIU.',
    href: '#',
    available: false,
    accent: '#64748b',
  },
]

const DATABASE_LEGAL = [
  { label: 'Undang-Undang', icon: BookOpen, href: 'https://peraturan.bpk.go.id/Subjek', color: '#1e3a5f' },
  { label: 'POJK', icon: Scale, href: 'https://www.ojk.go.id/id/regulasi/otoritas-jasa-keuangan/peraturan-ojk/Default.aspx', color: '#C8102E' },
  { label: 'Permen BUMN', icon: Building2, href: 'https://jdih.bumn.go.id/peraturan', color: '#1e3a5f' },
  { label: 'Dokumen Perusahaan', icon: FileText, href: '#', color: '#94a3b8', soon: true },
]

export default function HomePage() {
  const { isAuthenticated } = useAuthStore()

  return (
    <div>
      {/* Hero */}
      <section className="relative overflow-hidden" style={{ background: 'linear-gradient(135deg, #0B2545 0%, #1A3A6B 55%, #0B2545 100%)', minHeight: '560px' }}>
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute -top-32 -right-32 w-96 h-96 rounded-full opacity-10" style={{ background: '#C8102E' }} />
          <div className="absolute top-20 -left-20 w-72 h-72 rounded-full border border-white/10" />
          <div className="absolute bottom-0 right-1/3 w-48 h-48 rounded-full border border-white/5" />
          <div className="absolute bottom-0 left-0 right-0 h-1" style={{ background: '#C8102E' }} />
        </div>
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 lg:py-32">
          <div className="max-w-2xl">
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/10 text-white/70 text-xs font-medium mb-6 border border-white/10">
              <span className="w-1.5 h-1.5 rounded-full bg-green-400" />
              Portal Layanan Hukum Internal
            </div>
            <h1 className="text-5xl lg:text-6xl font-bold text-white leading-tight mb-6">
              Legal <span style={{ color: '#C8102E' }}>RIU</span>
              <br />
              <span className="text-white/80 text-4xl lg:text-5xl">Indonesia Re</span>
            </h1>
            <p className="text-lg text-white/60 leading-relaxed mb-10 max-w-xl">
              Membantu bisnis perusahaan sesuai dengan Good Corporate Governance melalui layanan hukum yang profesional dan efisien.
            </p>
            <div className="flex flex-wrap gap-3">
              <Link to={isAuthenticated ? '/dashboard/legal-opinions/new' : '/login'}
                className="inline-flex items-center gap-2 px-6 py-3 rounded-xl text-sm font-semibold text-white transition-all hover:opacity-90"
                style={{ background: '#C8102E' }}>
                Legal Opinion <ArrowRight className="w-4 h-4" />
              </Link>
              <Link to={isAuthenticated ? '/dashboard/review-documents/new' : '/login'}
                className="inline-flex items-center gap-2 px-6 py-3 rounded-xl text-sm font-semibold text-white border border-white/20 hover:bg-white/10 transition-all">
                Review Dokumen <ArrowRight className="w-4 h-4" />
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Services */}
      <section className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-14">
            <p className="text-xs font-semibold uppercase tracking-widest mb-3" style={{ color: '#C8102E' }}>Layanan Kami</p>
            <h2 className="text-3xl font-bold" style={{ color: '#0B2545' }}>Pelayanan Legal RIU</h2>
            <div className="w-12 h-1 rounded-full mx-auto mt-4" style={{ background: '#C8102E' }} />
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {SERVICES.map((svc) => (
              <div key={svc.title} className={`group relative rounded-2xl p-8 border transition-all duration-300 ${svc.available ? 'hover:shadow-xl hover:-translate-y-1 border-gray-100' : 'border-dashed border-gray-200 opacity-60'}`}>
                {!svc.available && <span className="absolute top-4 right-4 text-xs px-2 py-1 rounded-full bg-gray-100 text-gray-400 font-medium">Coming Soon</span>}
                <div className="w-14 h-14 rounded-2xl flex items-center justify-center mb-6 transition-transform group-hover:scale-110" style={{ background: svc.available ? svc.accent : '#f1f5f9' }}>
                  <svc.icon className="w-6 h-6" style={{ color: svc.available ? 'white' : '#94a3b8' }} />
                </div>
                <h3 className="text-lg font-bold mb-3" style={{ color: '#0B2545' }}>{svc.title}</h3>
                <p className="text-sm text-gray-500 leading-relaxed mb-6">{svc.desc}</p>
                {svc.available && (
                  <Link to={isAuthenticated ? svc.href : '/login'}
                    className="inline-flex items-center gap-1.5 text-sm font-semibold transition-colors" style={{ color: svc.accent }}>
                    Ajukan sekarang <ChevronRight className="w-4 h-4" />
                  </Link>
                )}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Mission */}
      <section className="py-20" style={{ background: '#f8fafc' }}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="rounded-3xl overflow-hidden grid lg:grid-cols-2" style={{ background: '#0B2545' }}>
            <div className="p-12 lg:p-16 flex flex-col justify-center">
              <p className="text-xs font-semibold uppercase tracking-widest mb-4" style={{ color: '#C8102E' }}>Misi Kami</p>
              <h2 className="text-3xl font-bold text-white leading-tight mb-6">
                Tujuan Membantu Bisnis Perusahaan Sesuai dengan Good Corporate Governance
              </h2>
              <p className="text-white/60 text-sm leading-relaxed">
                Tim Legal Compliance & Risk Management Indonesia Re siap memberikan layanan hukum yang profesional, cepat, dan dapat diandalkan untuk seluruh pegawai Indonesia Re.
              </p>
            </div>
            <div className="relative hidden lg:flex items-center justify-center p-12" style={{ background: 'linear-gradient(135deg, #1A3A6B, #C8102E)' }}>
              <div className="grid grid-cols-2 gap-4 w-full">
                {['Legal Opinion', 'Review Dokumen', 'Compliance', 'Risk Management'].map((item) => (
                  <div key={item} className="bg-white/10 backdrop-blur-sm rounded-xl p-4 border border-white/20">
                    <p className="text-white text-sm font-medium">{item}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Database Legal */}
      <section className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-14">
            <p className="text-xs font-semibold uppercase tracking-widest mb-3" style={{ color: '#C8102E' }}>Referensi Hukum</p>
            <h2 className="text-3xl font-bold" style={{ color: '#0B2545' }}>Database Legal</h2>
            <div className="w-12 h-1 rounded-full mx-auto mt-4" style={{ background: '#C8102E' }} />
          </div>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            {DATABASE_LEGAL.map((item) =>
              item.soon ? (
                <div key={item.label} className="rounded-2xl p-6 border-2 border-dashed border-gray-200 text-center opacity-50">
                  <div className="w-12 h-12 rounded-xl flex items-center justify-center mx-auto mb-4 bg-gray-100">
                    <item.icon className="w-6 h-6 text-gray-400" />
                  </div>
                  <p className="text-sm font-semibold text-gray-400">{item.label}</p>
                  <span className="text-xs text-gray-400 mt-1 block">Coming Soon</span>
                </div>
              ) : (
                <a key={item.label} href={item.href} target="_blank" rel="noopener noreferrer"
                  className="group rounded-2xl p-6 border border-gray-100 text-center hover:shadow-xl hover:-translate-y-1 transition-all duration-300 block">
                  <div className="w-12 h-12 rounded-xl flex items-center justify-center mx-auto mb-4 transition-transform group-hover:scale-110" style={{ background: item.color }}>
                    <item.icon className="w-6 h-6 text-white" />
                  </div>
                  <p className="text-sm font-semibold" style={{ color: '#0B2545' }}>{item.label}</p>
                  <ExternalLink className="w-3.5 h-3.5 text-gray-300 mx-auto mt-2 group-hover:text-gray-500 transition-colors" />
                </a>
              )
            )}
          </div>
        </div>
      </section>
    </div>
  )
}