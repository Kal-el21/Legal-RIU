package utils

import (
	"mime"
	"strings"
)

// HeaderUnduhan membentuk isi header Content-Disposition yang aman untuk nama file apa pun.
// jenis: "attachment" (diunduh) atau "inline" (dibuka di browser).
// Spasi dan tanda petik dikutip, huruf non-Latin ditulis sebagai filename*=utf-8''...,
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
