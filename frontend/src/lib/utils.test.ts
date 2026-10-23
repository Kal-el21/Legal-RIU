import { describe, expect, it } from 'vitest'
import { catatanWajib, pilahLampiran } from './utils'

describe('catatanWajib', () => {
  it('hanya Perlu Revisi dan Ditolak yang wajib bercatatan', () => {
    expect(['NEED_REVISION', 'REJECTED'].every(catatanWajib)).toBe(true)
    expect(['UNDER_REVIEW', 'COMPLETED', ''].some(catatanWajib)).toBe(false)
  })
})

const pdf = (nama: string) => new File(['%PDF-1.7 isi'], nama, { type: 'application/pdf' })

describe('pilahLampiran', () => {
  it('meloloskan file yang sah, urutannya tetap', async () => {
    const a = pdf('a.pdf')
    const b = pdf('b.pdf')
    expect(await pilahLampiran([a, b])).toEqual({ sah: [a, b], ditolak: [] })
  })

  it('menolak ekstensi yang tidak didukung dengan menyebut nama filenya', async () => {
    const sah = pdf('surat.pdf')
    const { sah: lolos, ditolak } = await pilahLampiran([new File(['MZ'], 'program.exe'), sah])
    expect(lolos).toEqual([sah])
    expect(ditolak).toEqual(['program.exe: Format file tidak didukung. Gunakan: PDF, DOC, DOCX'])
  })

  it('menolak file berekstensi .pdf yang isinya bukan PDF', async () => {
    const palsu = new File(['bukan pdf'], 'palsu.pdf', { type: 'application/pdf' })
    const { sah, ditolak } = await pilahLampiran([palsu])
    expect(sah).toEqual([])
    expect(ditolak).toEqual(['palsu.pdf: Format file tidak valid - konten tidak sesuai ekstensi'])
  })
})
