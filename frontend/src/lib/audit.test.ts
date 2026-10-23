import { describe, expect, it } from 'vitest'
import { AKSI, ENTITAS, jamWIB, kelompokPerHari, labelAksi, labelEntitas, perbandingan, pisahCatatan, tautanData } from './audit'

describe('label', () => {
  it('semua aksi berlabel bahasa Indonesia, kode yang tidak dikenal ditampilkan apa adanya', () => {
    expect(labelAksi('STATUS_CHANGE')).toBe('Status diubah')
    expect(labelAksi('AKSI_BARU')).toBe('AKSI_BARU')
    expect(Object.values(AKSI).every((a) => a.label && !a.label.includes('_'))).toBe(true)
  })

  it('jenis data berlabel, yang tidak dikenal apa adanya', () => {
    expect(labelEntitas('agreement_document')).toBe('Pengajuan perjanjian')
    expect(labelEntitas('baru')).toBe('baru')
    expect(Object.values(ENTITAS).some((l) => l.includes('_'))).toBe(false)
  })
})

describe('pisahCatatan', () => {
  it('memisahkan catatan approver dari keterangan', () => {
    expect(pisahCatatan('Status perjanjian PK-1 diubah dari UNDER_REVIEW ke NEED_REVISION. Catatan: lengkapi lampiran. Terima kasih'))
      .toEqual({ keterangan: 'Status perjanjian PK-1 diubah dari UNDER_REVIEW ke NEED_REVISION', catatan: 'lengkapi lampiran. Terima kasih' })
  })

  it('keterangan tanpa catatan, atau kosong', () => {
    expect(pisahCatatan('Pengajuan dibuat')).toEqual({ keterangan: 'Pengajuan dibuat' })
    expect(pisahCatatan(undefined)).toEqual({ keterangan: '' })
  })
})

describe('perbandingan', () => {
  it('status pengajuan dibaca labelnya', () => {
    expect(perbandingan({ old_value: 'UNDER_REVIEW', new_value: 'NEED_REVISION' })).toEqual({ lama: 'Sedang Direview', baru: 'Perlu Revisi' })
  })

  it('nilai lama yang tidak tercatat dibiarkan kosong; nilai bukan status apa adanya', () => {
    expect(perbandingan({ new_value: 'COMPLETED' })).toEqual({ lama: undefined, baru: 'Selesai' })
    expect(perbandingan({ old_value: 'Budi', new_value: 'Siti' })).toEqual({ lama: 'Budi', baru: 'Siti' })
  })

  it('tanpa nilai, atau nilai berupa ID, tidak ditampilkan', () => {
    expect(perbandingan({})).toBeNull()
    expect(perbandingan({ old_value: '8f14e45f-ceea-467a-9575-9a9d5a3c1b2e', new_value: '1679091c-5a88-4faf-b5e6-7b0f1c6a4b3d' })).toBeNull()
  })
})

describe('kelompokPerHari', () => {
  const sekarang = new Date('2026-10-23T03:00:00Z') // 23 Oktober 10.00 WIB

  it('mengelompokkan per tanggal WIB, bukan UTC, dengan label Hari ini dan Kemarin', () => {
    const baris = [
      { id: 'a', created_at: '2026-10-22T18:30:00Z' }, // 23 Okt 01.30 WIB
      { id: 'b', created_at: '2026-10-22T16:00:00Z' }, // 22 Okt 23.00 WIB
      { id: 'c', created_at: '2026-10-20T02:00:00Z' }, // 20 Okt
    ]
    const hasil = kelompokPerHari(baris, sekarang)
    expect(hasil.map((k) => [k.tanggal, k.label, k.items.map((i) => i.id)])).toEqual([
      ['2026-10-23', 'Hari ini', ['a']],
      ['2026-10-22', 'Kemarin', ['b']],
      ['2026-10-20', 'Selasa, 20 Oktober 2026', ['c']],
    ])
  })

  it('jam ditampilkan dalam WIB', () => {
    expect(jamWIB('2026-10-22T18:30:00Z')).toBe('01.30')
  })
})

describe('tautanData', () => {
  const id = '8f14e45f-ceea-467a-9575-9a9d5a3c1b2e'

  it('data berhalaman detail bertautan ke panel pengguna', () => {
    expect(tautanData({ action: 'STATUS_CHANGE', entity_type: 'legal_opinion', entity_id: id }, '/legal')).toBe(`/legal/legal-opinions/${id}`)
    expect(tautanData({ action: 'FILE_UPLOAD', entity_type: 'legal_case', entity_id: id }, '/admin')).toBe(`/admin/legal-cases/${id}`)
  })

  it('data master bertautan ke halaman kelolanya, hanya di panel admin', () => {
    expect(tautanData({ action: 'USER_UPDATE', entity_type: 'company', entity_id: id }, '/admin')).toBe('/admin/companies')
    expect(tautanData({ action: 'USER_UPDATE', entity_type: 'company', entity_id: id }, '/legal')).toBeNull()
  })

  it('baris impor (ID kosong) ke daftar datanya', () => {
    expect(tautanData({ action: 'FILE_UPLOAD', entity_type: 'legal_material', entity_id: '00000000-0000-0000-0000-000000000000' }, '/admin')).toBe('/admin/materials')
  })

  it('data yang dihapus, sesi login, dan panel lain tidak bertautan', () => {
    expect(tautanData({ action: 'DELETE', entity_type: 'legal_opinion', entity_id: id }, '/admin')).toBeNull()
    expect(tautanData({ action: 'LOGIN', entity_type: 'auth', entity_id: id }, '/admin')).toBeNull()
    expect(tautanData({ action: 'STATUS_CHANGE', entity_type: 'legal_opinion', entity_id: id }, '/legal-au')).toBeNull()
  })
})
