// Nama file dari header Content-Disposition, untuk unduhan lewat blob (PR 12).
// Backend menulis nama berhuruf non-Latin sebagai filename*=utf-8''<persen-encode>, nama
// berspasi atau bertanda petik sebagai filename="...", dan nama biasa tanpa tanda petik.
// filename* dibaca lebih dulu karena paling akurat; kalau tidak ada atau rusak, filename,
// lalu nama cadangan.
export function namaDariHeader(header: string | null | undefined, cadangan: string): string {
  if (!header) return cadangan
  const diperluas = /(?:^|;)\s*filename\*\s*=\s*"?[^']*'[^']*'([^";]+)"?/i.exec(header)
  if (diperluas) {
    try {
      const nama = decodeURIComponent(diperluas[1].trim())
      if (nama) return nama
    } catch {
      // Persen-encode rusak: lanjut ke filename biasa.
    }
  }
  const biasa = /(?:^|;)\s*filename\s*=\s*(?:"((?:[^"\\]|\\.)*)"|([^;]+))/i.exec(header)
  const nama = biasa ? (biasa[1] !== undefined ? biasa[1].replace(/\\(.)/g, '$1') : biasa[2].trim()) : ''
  return nama || cadangan
}
