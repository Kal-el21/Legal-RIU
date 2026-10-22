import { describe, expect, it } from 'vitest'
import { formDraf, judulDraf } from './draf'

describe('formDraf', () => {
  it('mengirim semua isian, termasuk yang kosong, supaya isian yang dihapus ikut terhapus di server', () => {
    const form = formDraf({ title: 'Judul', chronology: '', legal_type_other: undefined, question: null })
    expect(form.get('title')).toBe('Judul')
    expect(form.get('chronology')).toBe('')
    expect(form.get('legal_type_other')).toBe('')
    expect(form.get('question')).toBe('')
  })

  it('menyertakan lampiran di field attachments', () => {
    const a = new File(['a'], 'a.pdf')
    const b = new File(['b'], 'b.pdf')
    const form = formDraf({ title: 'x' }, [a, b])
    expect(form.getAll('attachments').map((f) => (f as File).name)).toEqual(['a.pdf', 'b.pdf'])
  })
})

describe('judulDraf', () => {
  it.each([
    ['Kajian kontrak', 'Kajian kontrak'],
    ['  Kajian  ', 'Kajian'],
    ['', '(Tanpa judul)'],
    ['   ', '(Tanpa judul)'],
    [undefined, '(Tanpa judul)'],
    [null, '(Tanpa judul)'],
  ])('%j menjadi %j', (judul, harap) => {
    expect(judulDraf(judul)).toBe(harap)
  })
})
