import { Link } from 'react-router-dom'
import { Scale, ExternalLink } from 'lucide-react'

export default function Footer() {
  return (
    <footer style={{ background: '#0B2545' }}>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-10">
          <div>
            <div className="flex items-center gap-2.5 mb-4">
              <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: '#C8102E' }}>
                <Scale className="w-4 h-4 text-white" />
              </div>
              <div>
                <span className="font-bold text-sm text-white">Legal RIU</span>
                <p className="text-xs text-white/40 leading-none">Indonesia Re</p>
              </div>
            </div>
            <p className="text-sm text-white/50 leading-relaxed max-w-xs">
              Portal layanan hukum internal Indonesia Re. Membantu bisnis perusahaan sesuai dengan Good Corporate Governance.
            </p>
          </div>
          <div>
            <p className="text-xs font-semibold text-white/40 uppercase tracking-widest mb-4">Layanan</p>
            <div className="space-y-2.5">
              <Link to="/dashboard/legal-opinions/new" className="block text-sm text-white/60 hover:text-white transition-colors">Legal Opinion</Link>
              <Link to="/dashboard/review-documents/new" className="block text-sm text-white/60 hover:text-white transition-colors">Review Dokumen</Link>
              <span className="flex items-center gap-2 text-sm text-white/30">
                Konsultasi Legal
                <span className="text-xs px-1.5 py-0.5 rounded-full bg-white/10 text-white/40">Soon</span>
              </span>
            </div>
          </div>
          <div>
            <p className="text-xs font-semibold text-white/40 uppercase tracking-widest mb-4">Database Legal</p>
            <div className="space-y-2.5">
              {[
                { label: 'Undang-Undang', href: 'https://peraturan.bpk.go.id/Subjek' },
                { label: 'POJK', href: 'https://www.ojk.go.id/id/regulasi/otoritas-jasa-keuangan/peraturan-ojk/Default.aspx' },
                { label: 'Permen BUMN', href: 'https://jdih.bumn.go.id/peraturan' },
              ].map((item) => (
                <a key={item.label} href={item.href} target="_blank" rel="noopener noreferrer"
                  className="flex items-center gap-1.5 text-sm text-white/60 hover:text-white transition-colors">
                  {item.label} <ExternalLink className="w-3 h-3 opacity-50" />
                </a>
              ))}
            </div>
          </div>
        </div>
        <div className="mt-10 pt-6 border-t border-white/10 flex flex-col sm:flex-row items-center justify-between gap-3">
          <p className="text-xs text-white/30">© 2025 Legal RIU — Indonesia Re. All rights reserved.</p>
          <p className="text-xs text-white/20">Powered by Legal RIU</p>
        </div>
      </div>
    </footer>
  )
}