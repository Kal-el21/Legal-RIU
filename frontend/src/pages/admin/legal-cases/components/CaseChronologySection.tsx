import { useState, useEffect } from 'react'
import { Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useCreateCaseChronology, useDeleteCaseChronology, useLegalCase } from '@/hooks/useLegalCase'
import { useAuthStore } from '@/store/auth.store'
import { formatDate } from '@/lib/utils'

interface CaseChronologySectionProps {
  caseId: string
}

export default function CaseChronologySection({ caseId }: CaseChronologySectionProps) {
  const { data: legalCase } = useLegalCase(caseId)
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const canManageChronology = hasPermission('case_management.manage_chronology')
  const createChronology = useCreateCaseChronology(caseId)
  const deleteChronology = useDeleteCaseChronology(caseId)
  const [agendaDate, setAgendaDate] = useState('')
  const [agenda, setAgenda] = useState('')

  const chronologies = legalCase?.chronologies ?? []

  useEffect(() => {
    if (legalCase) {
      const latest = chronologies[0]
      if (latest) {
        setAgendaDate(latest.agenda_date)
        setAgenda(latest.agenda)
      }
    }
  }, [legalCase])

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!agendaDate || !agenda.trim()) return
    await createChronology.mutateAsync({
      agenda_date: agendaDate,
      agenda: agenda.trim(),
    })
    setAgenda('')
  }

  const handleDelete = async (chronologyID: string) => {
    if (!window.confirm('Hapus kronologi ini?')) return
    await deleteChronology.mutateAsync(chronologyID)
  }

  return (
    <div className="bg-white rounded-2xl border border-gray-100 p-6">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Kronologi Kasus</h2>

      {canManageChronology && (
        <form onSubmit={handleSubmit} className="space-y-4 mb-6">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <Label className="text-sm font-medium text-gray-700">Tanggal Agenda</Label>
              <Input
                type="date"
                value={agendaDate}
                onChange={(e) => setAgendaDate(e.target.value)}
                className="mt-1.5"
              />
            </div>
            <div>
              <Label className="text-sm font-medium text-gray-700">Agenda</Label>
              <Input
                value={agenda}
                onChange={(e) => setAgenda(e.target.value)}
                placeholder="Judul agenda"
                className="mt-1.5"
              />
            </div>
          </div>
          <Button
            type="submit"
            disabled={!agendaDate || !agenda.trim() || createChronology.isPending}
            className="w-full text-white"
            style={{ background: '#C8102E' }}
          >
            {createChronology.isPending ? 'Menyimpan...' : 'Tambah Kronologi'}
          </Button>
        </form>
      )}

      <div className="space-y-3">
        {chronologies.length === 0 && (
          <p className="text-sm text-gray-400 text-center py-4">Belum ada kronologi</p>
        )}
        {chronologies.map((chronology) => (
          <div key={chronology.id} className="flex items-start justify-between rounded-lg border border-gray-100 p-4">
            <div>
              <p className="text-xs text-gray-400">{formatDate(chronology.agenda_date)}</p>
              <p className="text-sm font-medium text-gray-800 mt-0.5">{chronology.agenda}</p>
              {chronology.description && (
                <p className="text-sm text-gray-500 mt-1">{chronology.description}</p>
              )}
            </div>
            {canManageChronology && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => handleDelete(chronology.id)}
                className="text-red-600 hover:text-red-700"
              >
                <Trash2 className="w-4 h-4" />
              </Button>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
