// Pembantu draf pengajuan (PR Draf).

// formDraf: FormData untuk Simpan draf. Semua isian ikut, termasuk yang kosong, karena
// server mengganti seluruh isian draf dengan yang dikirim.
export function formDraf(isian: object, files: File[] = []): FormData {
  const form = new FormData()
  for (const [kunci, nilai] of Object.entries(isian)) {
    form.append(kunci, nilai == null ? '' : String(nilai))
  }
  files.forEach((f) => form.append('attachments', f))
  return form
}

// judulDraf: judul draf di daftar; draf yang belum diberi judul tetap bisa dikenali.
export function judulDraf(judul?: string | null): string {
  return judul?.trim() || '(Tanpa judul)'
}
