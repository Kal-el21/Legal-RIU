import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props { children: ReactNode }
interface State { error: Error | null }

export default class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State { return { error } }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Unhandled application error', error, info.componentStack)
  }

  render() {
    if (!this.state.error) return this.props.children
    return <main className="min-h-screen bg-slate-50 flex items-center justify-center p-6"><section className="w-full max-w-lg rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm"><div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-50 text-xl text-red-600">!</div><h1 className="text-xl font-bold text-[#0B2545]">Terjadi kesalahan pada aplikasi</h1><p className="mt-2 text-sm text-slate-500">Halaman tidak dapat ditampilkan. Silakan muat ulang atau kembali ke halaman utama.</p><p className="mt-4 rounded-lg bg-slate-50 p-3 text-left text-xs text-slate-500 break-words">{this.state.error.message}</p><div className="mt-6 flex justify-center gap-3"><button className="rounded-lg border px-4 py-2 text-sm" onClick={() => window.location.assign('/')}>Halaman Utama</button><button className="rounded-lg bg-[#C8102E] px-4 py-2 text-sm text-white" onClick={() => window.location.reload()}>Muat Ulang</button></div></section></main>
  }
}
