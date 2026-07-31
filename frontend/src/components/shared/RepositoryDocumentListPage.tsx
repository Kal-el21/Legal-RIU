import { useState } from 'react'
import { Download, Search, BadgeCheck, User } from 'lucide-react'
import { useRepositoryDocuments, useDownloadRepositoryDocument } from '@/hooks/useRepositoryDocument'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const FEATURE_OPTIONS = [
  { value: 'all', label: 'Semua' },
  { value: 'legal_opinion', label: 'Legal Opinion' },
  { value: 'document_review', label: 'Review Dokumen' },
  { value: 'case_management', label: 'Case Management' },
  { value: 'agreement_document', label: 'Dokumen Perjanjian' },
  { value: 'general', label: 'Umum' },
]

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

interface RepositoryDocumentListPageProps {
  basePath: string
  title: string
  description: string
}

export default function RepositoryDocumentListPage({
  basePath: _,
  title,
  description,
}: RepositoryDocumentListPageProps) {
  const [featureCode, setFeatureCode] = useState('all')
  const [search, setSearch] = useState('')

  const { data, isLoading } = useRepositoryDocuments({
    feature_code: featureCode === 'all' ? undefined : featureCode || undefined,
    search: search || undefined,
  })

  const downloadMutation = useDownloadRepositoryDocument()

  const handleDownload = (id: string) => {
    downloadMutation.mutate(id, {
      onSuccess: (result) => {
        const url = URL.createObjectURL(result.blob)
        const a = document.createElement('a')
        a.href = url
        a.download = result.filename
        a.click()
        URL.revokeObjectURL(url)
      },
    })
  }

  const filtered = data ?? []

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>{title}</h1>
          <p className="text-sm text-gray-500 mt-0.5">{description}</p>
        </div>
      </div>

      <div className="flex gap-4 mb-6">
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input
            placeholder="Cari judul atau file..."
            className="pl-9"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <Select value={featureCode} onValueChange={setFeatureCode}>
          <SelectTrigger className="w-56">
            <SelectValue placeholder="Filter feature" />
          </SelectTrigger>
          <SelectContent>
            {FEATURE_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-gray-400">Memuat data...</div>
        ) : !filtered.length ? (
          <div className="p-16 text-center">
            <p className="font-medium text-gray-500">Tidak ada dokumen ditemukan</p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <tr className="border-b border-gray-100" style={{ background: '#f8fafc' }}>
                <TableHead className="text-left px-6 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Judul</TableHead>
                <TableHead className="text-left px-4 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Feature</TableHead>
                <TableHead className="text-left px-4 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Uploader</TableHead>
                <TableHead className="text-left px-4 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">File</TableHead>
                <TableHead className="text-left px-4 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Ukuran</TableHead>
                <TableHead className="text-left px-4 py-3.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Tanggal</TableHead>
                <TableHead className="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 uppercase tracking-wide">Aksi</TableHead>
              </tr>
            </TableHeader>
            <TableBody className="divide-y divide-gray-50">
              {filtered.map((doc) => (
                <TableRow key={doc.id} className="hover:bg-gray-50/50 transition-colors">
                  <TableCell className="px-6 py-4 text-sm text-gray-900 font-medium max-w-xs truncate">{doc.title}</TableCell>
                  <TableCell className="px-4 py-4 text-sm">
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-700">
                      {doc.feature_code}
                    </span>
                  </TableCell>
                  <TableCell className="px-4 py-4 text-sm">
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-700">
                      {doc.uploader_type === 'approver' ? <BadgeCheck className="w-3 h-3" /> : <User className="w-3 h-3" />}
                      {doc.uploader_type}
                    </span>
                  </TableCell>
                  <TableCell className="px-4 py-4 text-sm text-gray-700 max-w-xs truncate">{doc.file_name}</TableCell>
                  <TableCell className="px-4 py-4 text-sm text-gray-500">{formatFileSize(doc.file_size)}</TableCell>
                  <TableCell className="px-4 py-4 text-sm text-gray-500">{formatDate(doc.created_at)}</TableCell>
                  <TableCell className="px-6 py-4 text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDownload(doc.id)}
                      className="h-7 px-2 text-xs"
                    >
                      <Download className="h-3 w-3" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>
    </div>
  )
}