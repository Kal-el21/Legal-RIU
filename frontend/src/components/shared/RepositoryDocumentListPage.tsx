import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight, Download, Search, BadgeCheck, User } from 'lucide-react'
import { toast } from 'sonner'
import { useRepositoryDocuments, useDownloadRepositoryDocument } from '@/hooks/useRepositoryDocument'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
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

export default function RepositoryDocumentListPage({
  basePath,
  title,
  description,
}: RepositoryDocumentListPageProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const [search, setSearch] = useState(searchParams.get('search') ?? '')
  const urlSearch = searchParams.get('search') ?? ''
  const panel = `/${basePath.split('/')[1]}`
  const page = Math.max(1, Number(searchParams.get('page')) || 1)
  const featureCode = searchParams.get('feature_code') ?? ''
  const dateFrom = searchParams.get('date_from') ?? ''
  const dateTo = searchParams.get('date_to') ?? ''
  const type = searchParams.get('type') ?? ''
  const typeFilter = filterJenisDokumen(type)

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

      <div className="grid grid-cols-1 gap-3 mb-5 md:grid-cols-2 xl:grid-cols-5">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input
            aria-label="Cari dokumen"
            placeholder="Cari nama, nomor, atau judul asal..."
            className="pl-9"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <Select value={featureCode || 'all'} onValueChange={(value) => setFilter('feature_code', value === 'all' ? '' : value)}>
          <SelectTrigger aria-label="Saring modul">
            <SelectValue placeholder="Semua modul" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua modul</SelectItem>
            {MODUL_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={type || 'all'} onValueChange={(value) => setFilter('type', value === 'all' ? '' : value)}>
          <SelectTrigger aria-label="Saring jenis dokumen">
            <SelectValue placeholder="Semua jenis dokumen" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua jenis dokumen</SelectItem>
            {JENIS_DOKUMEN.map((option) => (
              <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Input
          aria-label="Tanggal unggah mulai"
          type="date"
          value={dateFrom}
          onChange={(event) => setFilter('date_from', event.target.value)}
        />
        <Input
          aria-label="Tanggal unggah sampai"
          type="date"
          value={dateTo}
          onChange={(event) => setFilter('date_to', event.target.value)}
        />
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : isError ? (
          <div role="alert" className="p-12 text-center text-red-600">Dokumen gagal dimuat. Silakan muat ulang halaman.</div>
        ) : !documents.length ? (
          <div className="p-16 text-center">
            <p className="font-medium text-gray-500">Tidak ada dokumen ditemukan</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="bg-slate-50">
                  <TableHead>Dokumen</TableHead>
                  <TableHead>Asal</TableHead>
                  <TableHead>Modul</TableHead>
                  <TableHead>Diunggah</TableHead>
                  <TableHead>Ukuran</TableHead>
                  <TableHead className="text-right">Unduh</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {documents.map((doc) => {
                  const href = tautanDokumenSumber(doc, panel)
                  return (
                    <TableRow key={doc.id}>
                      <TableCell className="min-w-52">
                        <div className="font-medium text-gray-900 break-all">{doc.file_name}</div>
                        <div className="mt-1 text-xs text-gray-500">{labelJenisDokumen(doc)}</div>
                      </TableCell>
                      <TableCell className="min-w-48">
                        {href ? (
                          <Link to={href} className="block font-medium text-blue-700 hover:underline">
                            {doc.source_number && <div>{doc.source_number}</div>}
                            {doc.source_title && <div className="mt-1 text-xs">{doc.source_title}</div>}
                            {!doc.source_number && !doc.source_title && 'Buka data'}
                          </Link>
                        ) : (
                          <div className="font-medium text-gray-700">
                            {doc.source_number && <div>{doc.source_number}</div>}
                            {doc.source_title && <div className="mt-1 text-xs text-gray-500">{doc.source_title}</div>}
                            {!doc.source_number && !doc.source_title && '-'}
                          </div>
                        )}
                      </TableCell>
                      <TableCell>{labelModul(doc.feature_code)}</TableCell>
                      <TableCell className="min-w-36">
                        <div>{formatTanggalUnggah(doc.uploaded_at ?? doc.created_at)}</div>
                        <div className="mt-1 inline-flex items-center gap-1 text-xs text-gray-500">
                          {doc.uploader_type === 'approver' ? <BadgeCheck className="h-3 w-3" /> : <User className="h-3 w-3" />}
                          {labelPengunggah(doc.uploader_type)}
                        </div>
                      </TableCell>
                      <TableCell className="whitespace-nowrap">{formatUkuranDokumen(doc.file_size)}</TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          aria-label={`Unduh ${doc.file_name}`}
                          disabled={downloadMutation.isPending}
                          onClick={() => handleDownload(doc.id)}
                        >
                          <Download className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
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
