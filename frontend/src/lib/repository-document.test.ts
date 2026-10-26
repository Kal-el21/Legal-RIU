import { describe, expect, it } from 'vitest'
import {
  filterJenisDokumen,
  formatTanggalUnggah,
  formatUkuranDokumen,
  labelJenisDokumen,
  labelModul,
  labelPengunggah,
  tautanDokumenSumber,
} from './repository-document'

describe('tampilan repositori dokumen', () => {
  it('menerjemahkan label modul, jenis, dan pengunggah', () => {
    expect(labelModul('legal_opinion')).toBe('Legal Opinion')
    expect(labelModul('case_management')).toBe('Perkara')
    expect(labelJenisDokumen({ source_type: 'legal_opinion_results' })).toBe('Hasil Legal Opinion')
    expect(labelJenisDokumen({ source_type: 'agreement_documents', mime_type: 'application/pdf' })).toBe('Perjanjian Final (PDF)')
    expect(labelJenisDokumen({
      source_type: 'agreement_documents',
      mime_type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    })).toBe('Perjanjian Final (DOCX)')
    expect(labelPengunggah('requester')).toBe('Pengaju')
    expect(labelPengunggah('approver')).toBe('Approver')
  })

  it('membentuk filter jenis dokumen termasuk MIME final perjanjian', () => {
    expect(filterJenisDokumen('')).toEqual({})
    expect(filterJenisDokumen('legal_cases_photo')).toEqual({ source_type: 'legal_cases_photo' })
    expect(filterJenisDokumen('agreement_documents:application/pdf')).toEqual({
      source_type: 'agreement_documents',
      mime_type: 'application/pdf',
    })
  })

  it('menampilkan ukuran dan tanggal dengan aman', () => {
    expect(formatUkuranDokumen(0)).toBe('-')
    expect(formatUkuranDokumen(1024)).toBe('1.0 KB')
    expect(formatTanggalUnggah('2026-10-25T18:00:00Z')).toBe('26 Okt 2026')
    expect(formatTanggalUnggah(undefined)).toBe('-')
    expect(formatTanggalUnggah('tanggal-salah')).toBe('-')
  })

  it('membentuk tautan hanya untuk panel yang memiliki halaman sumber', () => {
    const doc = { source_type: 'legal_case', source_id: '123' }
    expect(tautanDokumenSumber({ ...doc, source_type: 'legal_cases' }, '/legal-au')).toBe('/legal-au/cases/123')
    expect(tautanDokumenSumber({ ...doc, source_type: 'legal_opinion_attachments' }, '/external')).toBe('/external/legal-opinions/123')
    expect(tautanDokumenSumber({ ...doc, source_type: 'agreement_documents' }, '/external')).toBeNull()
    expect(tautanDokumenSumber({ ...doc, source_type: 'case_chronologies' }, '/legal')).toBeNull()
  })
})
