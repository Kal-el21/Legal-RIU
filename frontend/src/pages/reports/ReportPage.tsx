import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { BarChart3, Calendar, Filter } from 'lucide-react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { useReport } from '@/hooks/useReport'
import { useAuthStore } from '@/store/auth.store'
import { REPORT_GROUP_BY_OPTIONS } from '@/types'
import type { ReportFeature, ReportGroupBy, ReportChartResponse } from '@/types'

const TABS: { key: ReportFeature; label: string }[] = [
  { key: 'legal-cases', label: 'Manajemen Kasus' },
  { key: 'legal-opinions', label: 'Legal Opinion' },
  { key: 'document-reviews', label: 'Review Dokumen' },
]

const PERMISSIONS: Record<ReportFeature, string> = {
  'legal-cases': 'report.legal_case.view',
  'legal-opinions': 'report.legal_opinion.view',
  'document-reviews': 'report.document_review.view',
}

export default function ReportPage() {
  const navigate = useNavigate()
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const [activeTab, setActiveTab] = useState<ReportFeature>('legal-cases')
  const [groupBy, setGroupBy] = useState<ReportGroupBy>('company')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')

  useEffect(() => {
    if (!hasPermission(PERMISSIONS[activeTab])) {
      navigate(`/${activeTab === 'legal-cases' ? 'admin' : 'legal'}`, { replace: true })
    }
  }, [hasPermission, activeTab, navigate])

  const { data, isLoading, isFetching } = useReport(activeTab, groupBy, dateFrom || undefined, dateTo || undefined)

  const chartData = transformChartData(data)

  const handleTabChange = (tab: ReportFeature) => {
    setActiveTab(tab)
    const options = REPORT_GROUP_BY_OPTIONS[tab]
    setGroupBy(options[0]?.value ?? 'company')
  }

  return (
    <div className="p-6 max-w-7xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Laporan</h1>
        <p className="text-sm text-gray-500 mt-0.5">Analisis data legal case, legal opinion, dan review dokumen</p>
      </div>

      <div className="bg-white rounded-2xl border border-gray-100">
        <div className="flex border-b border-gray-100">
          {TABS.map((tab) => (
            <button
              key={tab.key}
              onClick={() => handleTabChange(tab.key)}
              className={cn(
                'flex-1 px-4 py-3 text-sm font-medium transition-all',
                activeTab === tab.key
                  ? 'text-white border-b-2'
                  : 'text-gray-500 hover:text-gray-700'
              )}
              style={activeTab === tab.key ? { background: '#0B2545', borderColor: '#C8102E' } : undefined}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <div className="p-4 flex flex-wrap items-center gap-3">
          <div className="flex items-center gap-2">
            <Calendar className="w-4 h-4 text-gray-400" />
            <input
              type="date"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
              className="px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#0B2545]/20"
              placeholder="Dari"
            />
            <span className="text-gray-400 text-sm">s/d</span>
            <input
              type="date"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
              className="px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#0B2545]/20"
              placeholder="Sampai"
            />
          </div>

          <div className="flex items-center gap-2 ml-auto">
            <Filter className="w-4 h-4 text-gray-400" />
            <select
              value={groupBy}
              onChange={(e) => setGroupBy(e.target.value as ReportGroupBy)}
              className="px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-[#0B2545]/20"
            >
              {REPORT_GROUP_BY_OPTIONS[activeTab].map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>
        </div>

        <div className="p-4">
          {isLoading || isFetching ? (
            <div className="flex items-center justify-center h-64">
              <div className="w-8 h-8 border-2 border-[#0B2545] border-t-transparent rounded-full animate-spin" />
            </div>
          ) : !chartData.length ? (
            <div className="flex flex-col items-center justify-center h-64 text-gray-400">
              <BarChart3 className="w-12 h-12 mb-2 opacity-40" />
              <p className="text-sm">Tidak ada data untuk filter ini</p>
            </div>
          ) : (
            <div className="h-96">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={chartData} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                  <XAxis dataKey="name" tick={{ fontSize: 12 }} stroke="#94a3b8" />
                  <YAxis tick={{ fontSize: 12 }} stroke="#94a3b8" />
                  <Tooltip
                    contentStyle={{ borderRadius: '8px', border: '1px solid #e2e8f0', fontSize: '12px' }}
                    cursor={{ fill: 'rgba(11, 37, 69, 0.05)' }}
                  />
                  <Legend />
                  {data?.series.map((series, index) => (
                    <Bar
                      key={series.name}
                      dataKey={series.name}
                      fill={CHART_COLORS[index % CHART_COLORS.length]}
                      radius={[2, 2, 0, 0]}
                    />
                  ))}
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

function transformChartData(response?: ReportChartResponse) {
  if (!response?.labels?.length || !response?.series?.length) return []
  return response.labels.map((label) => {
    const point: Record<string, string | number> = { name: label }
    response.series.forEach((s) => {
      point[s.name] = s.data[response.labels.indexOf(label)] ?? 0
    })
    return point
  })
}

const CHART_COLORS = [
  '#C8102E', '#0B2545', '#0891B2', '#D97706', '#059669',
  '#7C3AED', '#DB2777', '#EA580C', '#2563EB', '#4F46E5',
]

function cn(...classes: (string | boolean | undefined)[]) {
  return classes.filter(Boolean).join(' ')
}
