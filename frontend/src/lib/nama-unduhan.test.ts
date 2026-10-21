import { describe, expect, it } from 'vitest'
import { namaDariHeader } from './nama-unduhan'

describe('namaDariHeader', () => {
  // Header di bawah persis keluaran utils.HeaderUnduhan di backend (lihat unduhan_test.go).
  it.each([
    ['attachment; filename=laporan.pdf', 'laporan.pdf'],
    ['attachment; filename="LO-202610-0003_Lampiran_Surat Penawaran Vendor.pdf"', 'LO-202610-0003_Lampiran_Surat Penawaran Vendor.pdf'],
    ['attachment; filename="Surat \\"Final\\".pdf"', 'Surat "Final".pdf'],
    ["attachment; filename*=utf-8''R%C3%A9sum%C3%A9%20Mitra.pdf", 'Résumé Mitra.pdf'],
    ["attachment; filename*=utf-8''%E5%A5%91%E7%B4%84%E6%9B%B8.docx", '契約書.docx'],
    ['inline; filename="C:_data_a.pdf"', 'C:_data_a.pdf'],
  ])('membaca %s', (header, harap) => {
    expect(namaDariHeader(header, 'cadangan')).toBe(harap)
  })

  it('mendahulukan filename* daripada filename', () => {
    expect(namaDariHeader("attachment; filename=\"Resume.pdf\"; filename*=UTF-8''R%C3%A9sum%C3%A9.pdf", 'x')).toBe('Résumé.pdf')
  })

  it('kembali ke filename kalau filename* rusak', () => {
    expect(namaDariHeader("attachment; filename*=utf-8''%E0%A4%A; filename=cadang.pdf", 'x')).toBe('cadang.pdf')
  })

  it.each([undefined, null, '', 'attachment', 'attachment; filename=""'])('memakai nama cadangan untuk %s', (header) => {
    expect(namaDariHeader(header, 'unduhan.pdf')).toBe('unduhan.pdf')
  })
})
