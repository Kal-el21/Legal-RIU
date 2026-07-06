import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/store/auth.store'
import LegalCaseFormDialog from '@/pages/admin/legal-cases/components/LegalCaseFormDialog'

export default function LegalAUCaseFormPage() {
  const navigate = useNavigate()
  const hasPermission = useAuthStore((state) => state.hasPermission)

  useEffect(() => {
    if (!hasPermission('case_management.create')) {
      navigate('/legal-au/cases', { replace: true })
    }
  }, [hasPermission, navigate])

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="flex items-center gap-3 mb-6">
        <Button variant="ghost" onClick={() => navigate('/legal-au/cases')} className="-ml-2">
          <ArrowLeft className="w-4 h-4 mr-2" /> Kembali
        </Button>
        <h1 className="text-2xl font-bold" style={{ color: '#0B2545' }}>Tambah Kasus</h1>
      </div>
      <LegalCaseFormDialog
        open={true}
        onOpenChange={(open) => {
          if (!open) navigate('/legal-au/cases')
        }}
        legalCase={null}
      />
    </div>
  )
}
