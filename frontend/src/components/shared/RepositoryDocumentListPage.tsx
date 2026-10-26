import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { ChevronDown, ChevronLeft, ChevronRight, Download, Search, BadgeCheck, User, X } from 'lucide-react'
import { toast } from 'sonner'
import { useRepositoryDocuments, useDownloadRepositoryDocument } from '@/hooks/useRepositoryDocument'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type { RepositoryDocument } from '@/types'
import {
  filterJenisDokumen,
  formatTanggalUnggah,
  formatUkuranDokumen,
  JENIS_DOKUMEN,
  labelJenisDokumen,
  labelModul,
  labelPengunggah,
  tautanDokumenSumber,
} from '@/lib/repository-document'

const MODUL_OPTIONS = [
  { value: 'legal_opinion', label: 'Legal Opinion' },
  { value: 'document_review', label: 'Review Dokumen' },
  { value: 'agreement_document', label: 'Perjanjian' },
  { value: 'case_management', label: 'Perkara' },
]

const LIMIT = 20

interface RepositoryDocumentListPageProps {
  basePath: string
  title: string
  description: string
}

function NamaDokumen({ doc }: { doc: RepositoryDocument }) {
  return (
    <div className="min-w-0">
      <div className="line-clamp-2 break-words font-medium leading-5 text-gray-900" title={doc.file_name}>
        {doc.file_name}
      </div>
      <div className="mt-1 text-xs text-gray-500">{labelJenisDokumen(doc)}</div>
    </div>
  )
}

function AsalDokumen({ doc, panel }: { doc: RepositoryDocument; panel: string }) {
  const href = tautanDokumenSumber(doc, panel)
  const asal = (
    <>
      {doc.source_number && (
        <span className="inline-flex max-w-full truncate rounded-md bg-gray-100 px-2 py-1 font-mono text-xs font-medium text-gray-700">
          {doc.source_number}
        </span>
      )}
      {doc.source_title && (
        <span className="line-clamp-2 text-sm leading-5 text-gray-600">
          {doc.source_title}
        </span>
      )}
      {!doc.source_number && !doc.source_title && <span className="text-sm text-gray-500">-</span>}
    </>
  )

  return href ? (
    <Link
      to={href}
      className="flex min-w-0 flex-col items-start gap-1.5 text-[#0B2545] hover:text-[#C8102E] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#0B2545] focus-visible:ring-offset-2"
    >
      {asal}
    </Link>
  ) : (
    <div className="flex min-w-0 flex-col items-start gap-1.5">{asal}</div>
  )
}

function PengunggahDokumen({ type }: { type: string }) {
  return (
    <span className="inline-flex items-center gap-1 text-xs text-gray-500">
      {type === 'approver' ? <BadgeCheck className="h-3.5 w-3.5" /> : <User className="h-3.5 w-3.5" />}
      {labelPengunggah(type)}
    </span>
  )
}

export default function RepositoryDocumentListPage({
  basePath,
  title,
  description,
}: RepositoryDocumentListPageProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const [search, setSearch] = useState(searchParams.get('search') ?? '')
  const [filterOpen, setFilterOpen] = useState(false)
  const urlSearch = searchParams.get('search') ?? ''
  const panel = `/${basePath.split('/')[1]}`
  const page = Math.max(1, Number(searchParams.get('page')) || 1)
  const featureCode = searchParams.get('feature_code') ?? ''
  const dateFrom = searchParams.get('date_from') ?? ''
  const dateTo = searchParams.get('date_to') ?? ''
  const type = searchParams.get('type') ?? ''
  const typeFilter = filterJenisDokumen(type)
  const activeTypeLabel = JENIS_DOKUMEN.find((option) => option.value === type)?.label

  useEffect(() => {
    const timer = window.setTimeout(() => setSearch(urlSearch), 0)
    return () => window.clearTimeout(timer)
  }, [urlSearch])

  useEffect(() => {
    const timer = window.setTimeout(() => {
      if (search.trim() === urlSearch) return
      setSearchParams((current) => {
        const next = new URLSearchParams(current)
        if (search.trim()) next.set('search', search.trim())
        else next.delete('search')
        next.delete('page')
        return next
      })
    }, 300)
    return () => window.clearTimeout(timer)
  }, [search, setSearchParams, urlSearch])

  const setFilter = (key: string, value: string) => {
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      if (value) next.set(key, value)
      else next.delete(key)
      if (key !== 'page') next.delete('page')
      return next
    })
  }

  const clearFilters = () => {
    setSearch('')
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      for (const key of ['search', 'feature_code', 'type', 'date_from', 'date_to', 'page']) next.delete(key)
      return next
    })
  }

  const { data, isLoading, isError } = useRepositoryDocuments({
    page,
    limit: LIMIT,
    feature_code: featureCode || undefined,
    ...typeFilter,
    date_from: dateFrom || undefined,
    date_to: dateTo || undefined,
    search: urlSearch || undefined,
  })

  const downloadMutation = useDownloadRepositoryDocument()
  const handleDownload = (id: string) => {
    downloadMutation.mutate(id, {
      onSuccess: (result) => {
        const url = URL.createObjectURL(result.blob)
        const anchor = document.createElement('a')
        anchor.href = url
        anchor.download = result.filename
        anchor.click()
        window.setTimeout(() => URL.revokeObjectURL(url), 1000)
      },
      onError: () => toast.error('Gagal mengunduh dokumen. Silakan coba lagi.'),
    })
  }

  const documents = data?.items ?? []
  const totalPages = data?.total_pages ?? 0
  const firstItem = data?.total ? (page - 1) * LIMIT + 1 : 0
  const lastItem = data?.total ? Math.min(page * LIMIT, data.total) : 0

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>{title}</h1>
        <p className="text-sm text-gray-500 mt-0.5">{description}</p>
      </div>

      <section aria-label="Pencarian dan filter dokumen" className="mb-5 rounded-2xl border border-gray-100 bg-white p-4">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-[minmax(0,1fr)_220px_auto]">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <Input
              aria-label="Cari dokumen"
              placeholder="Cari nama file, nomor, atau judul asal..."
              className="pl-9"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>
          <Select value={featureCode || 'all'} onValueChange={(value) => setFilter('feature_code', value === 'all' ? '' : value)}>
            <SelectTrigger aria-label="Saring modul" className="w-full">
              <SelectValue placeholder="Semua modul" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Semua modul</SelectItem>
              {MODUL_OPTIONS.map((option) => (
                <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            type="button"
            variant="outline"
            aria-expanded={filterOpen}
            aria-controls="repository-document-extra-filters"
            onClick={() => setFilterOpen((open) => !open)}
            className="justify-between gap-2"
          >
            Filter lainnya
            <ChevronDown className={`h-4 w-4 transition-transform ${filterOpen ? 'rotate-180' : ''}`} />
          </Button>
        </div>

        <div
          id="repository-document-extra-filters"
          hidden={!filterOpen}
          className="mt-4 grid grid-cols-1 gap-3 border-t border-gray-100 pt-4 sm:grid-cols-2"
        >
          <div className="space-y-1.5">
            <Label htmlFor="repository-document-type" className="text-xs font-medium text-gray-600">Jenis dokumen</Label>
            <Select value={type || 'all'} onValueChange={(value) => setFilter('type', value === 'all' ? '' : value)}>
              <SelectTrigger id="repository-document-type" aria-label="Saring jenis dokumen" className="w-full">
                <SelectValue placeholder="Semua jenis dokumen" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">Semua jenis dokumen</SelectItem>
                {JENIS_DOKUMEN.map((option) => (
                  <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <fieldset className="space-y-1.5">
            <legend className="text-xs font-medium text-gray-600">Tanggal unggah</legend>
            <div className="grid grid-cols-2 gap-2">
              <div className="space-y-1">
                <Label htmlFor="repository-document-date-from" className="text-xs text-gray-500">Dari</Label>
                <Input
                  id="repository-document-date-from"
                  aria-label="Tanggal unggah dari"
                  type="date"
                  max={dateTo || undefined}
                  value={dateFrom}
                  onChange={(event) => setFilter('date_from', event.target.value)}
                />
              </div>
              <div className="space-y-1">
                <Label htmlFor="repository-document-date-to" className="text-xs text-gray-500">Sampai</Label>
                <Input
                  id="repository-document-date-to"
                  aria-label="Tanggal unggah sampai"
                  type="date"
                  min={dateFrom || undefined}
                  value={dateTo}
                  onChange={(event) => setFilter('date_to', event.target.value)}
                />
              </div>
            </div>
          </fieldset>
        </div>

        {(urlSearch || featureCode || type || dateFrom || dateTo) && (
          <div className="mt-4 flex flex-wrap items-center gap-2 border-t border-gray-100 pt-3">
            <span className="text-xs font-medium text-gray-500">Filter aktif:</span>
            {urlSearch && (
              <Button type="button" variant="outline" size="sm" title={`Pencarian: ${urlSearch}`} onClick={() => { setSearch(''); setFilter('search', '') }} className="h-7 max-w-full gap-1.5 px-2 text-xs">
                <span className="truncate">Pencarian: {urlSearch}</span>
                <X className="h-3 w-3 shrink-0" aria-hidden="true" />
                <span className="sr-only">Hapus filter pencarian</span>
              </Button>
            )}
            {featureCode && (
              <Button type="button" variant="outline" size="sm" onClick={() => setFilter('feature_code', '')} className="h-7 gap-1.5 px-2 text-xs">
                {MODUL_OPTIONS.find((option) => option.value === featureCode)?.label ?? featureCode}
                <X className="h-3 w-3" aria-hidden="true" />
                <span className="sr-only">Hapus filter modul</span>
              </Button>
            )}
            {type && (
              <Button type="button" variant="outline" size="sm" onClick={() => setFilter('type', '')} className="h-7 gap-1.5 px-2 text-xs">
                {activeTypeLabel ?? type}
                <X className="h-3 w-3" aria-hidden="true" />
                <span className="sr-only">Hapus filter jenis dokumen</span>
              </Button>
            )}
            {dateFrom && (
              <Button type="button" variant="outline" size="sm" onClick={() => setFilter('date_from', '')} className="h-7 gap-1.5 px-2 text-xs">
                Dari {formatTanggalUnggah(dateFrom)}
                <X className="h-3 w-3" aria-hidden="true" />
                <span className="sr-only">Hapus tanggal mulai</span>
              </Button>
            )}
            {dateTo && (
              <Button type="button" variant="outline" size="sm" onClick={() => setFilter('date_to', '')} className="h-7 gap-1.5 px-2 text-xs">
                Sampai {formatTanggalUnggah(dateTo)}
                <X className="h-3 w-3" aria-hidden="true" />
                <span className="sr-only">Hapus tanggal akhir</span>
              </Button>
            )}
            <Button type="button" variant="ghost" size="sm" onClick={clearFilters} className="h-7 px-2 text-xs text-[#0B2545] hover:text-[#C8102E]">
              Bersihkan filter
            </Button>
          </div>
        )}
      </section>

      <div className="overflow-hidden rounded-2xl border border-gray-100 bg-white">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : isError ? (
          <div role="alert" className="p-12 text-center text-red-600">Dokumen gagal dimuat. Silakan muat ulang halaman.</div>
        ) : !documents.length ? (
          <div className="p-16 text-center">
            <p className="font-medium text-gray-500">Tidak ada dokumen ditemukan</p>
          </div>
        ) : (
          <>
            <div className="hidden xl:block">
              <Table>
                <TableHeader>
                  <TableRow className="border-b border-gray-100 bg-slate-50">
                    <TableHead className="px-6 py-3.5 text-xs font-semibold uppercase tracking-wide text-gray-500">Dokumen</TableHead>
                    <TableHead className="px-4 py-3.5 text-xs font-semibold uppercase tracking-wide text-gray-500">Asal</TableHead>
                    <TableHead className="px-4 py-3.5 text-xs font-semibold uppercase tracking-wide text-gray-500">Modul</TableHead>
                    <TableHead className="px-4 py-3.5 text-xs font-semibold uppercase tracking-wide text-gray-500">Diunggah</TableHead>
                    <TableHead className="px-4 py-3.5 text-right text-xs font-semibold uppercase tracking-wide text-gray-500">Ukuran</TableHead>
                    <TableHead className="px-6 py-3.5 text-right text-xs font-semibold uppercase tracking-wide text-gray-500">Unduh</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody className="divide-y divide-gray-50">
                  {documents.map((doc) => (
                    <TableRow key={doc.id}>
                      <TableCell className="max-w-[22rem] whitespace-normal px-6 py-4">
                        <NamaDokumen doc={doc} />
                      </TableCell>
                      <TableCell className="max-w-[18rem] whitespace-normal px-4 py-4">
                        <AsalDokumen doc={doc} panel={panel} />
                      </TableCell>
                      <TableCell className="whitespace-nowrap px-4 py-4 text-sm text-gray-700">
                        {labelModul(doc.feature_code)}
                      </TableCell>
                      <TableCell className="whitespace-nowrap px-4 py-4">
                        <div className="text-sm text-gray-700">{formatTanggalUnggah(doc.uploaded_at ?? doc.created_at)}</div>
                        <div className="mt-1"><PengunggahDokumen type={doc.uploader_type} /></div>
                      </TableCell>
                      <TableCell className="whitespace-nowrap px-4 py-4 text-right text-sm tabular-nums text-gray-600">
                        {formatUkuranDokumen(doc.file_size)}
                      </TableCell>
                      <TableCell className="px-6 py-4 text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          aria-label={`Unduh ${doc.file_name}`}
                          disabled={downloadMutation.isPending}
                          onClick={() => handleDownload(doc.id)}
                          className="text-gray-500 hover:text-[#C8102E] focus-visible:ring-[#0B2545]"
                        >
                          <Download className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            <ul className="divide-y divide-gray-100 xl:hidden">
              {documents.map((doc) => (
                <li key={doc.id} className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <NamaDokumen doc={doc} />
                    </div>
                    <Button
                      variant="outline"
                      size="icon"
                      aria-label={`Unduh ${doc.file_name}`}
                      disabled={downloadMutation.isPending}
                      onClick={() => handleDownload(doc.id)}
                      className="shrink-0 text-gray-600 hover:border-[#C8102E] hover:text-[#C8102E] focus-visible:ring-[#0B2545]"
                    >
                      <Download className="h-4 w-4" />
                    </Button>
                  </div>
                  <div className="mt-3">
                    <AsalDokumen doc={doc} panel={panel} />
                  </div>
                  <dl className="mt-3 grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
                    <div>
                      <dt className="text-xs text-gray-500">Modul</dt>
                      <dd className="mt-0.5 text-gray-700">{labelModul(doc.feature_code)}</dd>
                    </div>
                    <div>
                      <dt className="text-xs text-gray-500">Diunggah</dt>
                      <dd className="mt-0.5 text-gray-700">{formatTanggalUnggah(doc.uploaded_at ?? doc.created_at)}</dd>
                    </div>
                    <div>
                      <dt className="text-xs text-gray-500">Pengunggah</dt>
                      <dd className="mt-0.5"><PengunggahDokumen type={doc.uploader_type} /></dd>
                    </div>
                    <div>
                      <dt className="text-xs text-gray-500">Ukuran</dt>
                      <dd className="mt-0.5 tabular-nums text-gray-700">{formatUkuranDokumen(doc.file_size)}</dd>
                    </div>
                  </dl>
                </li>
              ))}
            </ul>
          </>
        )}
      </div>

      <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-sm text-gray-600">
        <span>Menampilkan {firstItem}–{lastItem} dari {data?.total ?? 0} dokumen</span>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            aria-label="Halaman sebelumnya"
            disabled={page <= 1 || isLoading}
            onClick={() => setFilter('page', String(page - 1))}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span>Halaman {page} dari {totalPages}</span>
          <Button
            variant="outline"
            size="sm"
            aria-label="Halaman berikutnya"
            disabled={page >= totalPages || isLoading}
            onClick={() => setFilter('page', String(page + 1))}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
