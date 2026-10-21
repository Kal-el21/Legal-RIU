package utils

import (
	"mime"
	"strings"
)

// NamaUnduhan menyusun nama file unduhan dari bagian-bagiannya, dipisah "_"; bagian kosong
// dilewati. Contoh: NamaUnduhan("LO-202610-0003", "Lampiran", "Surat.pdf") menjadi
// "LO-202610-0003_Lampiran_Surat.pdf".
func NamaUnduhan(bagian ...string) string {
	isi := make([]string, 0, len(bagian))
	for _, b := range bagian {
		if b = strings.TrimSpace(b); b != "" {
			isi = append(isi, b)
		}
	}
	return strings.Join(isi, "_")
}

// HeaderUnduhan membentuk isi header Content-Disposition yang aman untuk nama file apa pun.
// jenis: "attachment" (diunduh) atau "inline" (dibuka di browser).
// Spasi dan tanda petik dikutip, huruf non-Latin ditulis dalam bentuk filename* (RFC 5987),
// pemisah folder diganti "_" dan karakter kontrol dibuang, jadi nama tidak bisa membawa path
// atau menyisipkan header lain.
func HeaderUnduhan(jenis, nama string) string {
	nama = strings.TrimSpace(strings.Map(func(r rune) rune {
		switch {
		case r == '/' || r == '\\':
			return '_'
		case r < ' ' || r == 0x7f:
			return -1
		}
		return r
	}, nama))
	if strings.Trim(nama, "._ ") == "" {
		nama = "unduhan"
	}
	return mime.FormatMediaType(jenis, map[string]string{"filename": nama})
}
