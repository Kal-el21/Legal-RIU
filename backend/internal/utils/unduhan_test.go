package utils

import (
	"mime"
	"strings"
	"testing"
)

func TestNamaUnduhan(t *testing.T) {
	kasus := []struct {
		bagian []string
		harap  string
	}{
		{[]string{"LO-202610-0003", "Lampiran", "Surat Penawaran.pdf"}, "LO-202610-0003_Lampiran_Surat Penawaran.pdf"},
		{[]string{"PK-202610-0001", "PKS", " "}, "PK-202610-0001_PKS"},
		{[]string{"", "Perkara"}, "Perkara"},
		{nil, ""},
	}
	for _, k := range kasus {
		if got := NamaUnduhan(k.bagian...); got != k.harap {
			t.Errorf("NamaUnduhan(%q) = %q, mau %q", k.bagian, got, k.harap)
		}
	}
}

func TestHeaderUnduhan(t *testing.T) {
	kasus := []struct {
		nama, header, hasil string
	}{
		{"laporan.pdf", "attachment; filename=laporan.pdf", "laporan.pdf"},
		{"Surat Penawaran Vendor.pdf", `attachment; filename="Surat Penawaran Vendor.pdf"`, "Surat Penawaran Vendor.pdf"},
		{`Surat "Final".pdf`, `attachment; filename="Surat \"Final\".pdf"`, `Surat "Final".pdf`},
		{"Résumé Mitra.pdf", "attachment; filename*=utf-8''R%C3%A9sum%C3%A9%20Mitra.pdf", "Résumé Mitra.pdf"},
		{"契約書.docx", "attachment; filename*=utf-8''%E5%A5%91%E7%B4%84%E6%9B%B8.docx", "契約書.docx"},
		{"../../etc/passwd", "attachment; filename=.._.._etc_passwd", ".._.._etc_passwd"},
		{`C:\data\a.pdf`, `attachment; filename="C:_data_a.pdf"`, "C:_data_a.pdf"},
		{"a.pdf\r\nSet-Cookie: x=1", `attachment; filename="a.pdfSet-Cookie: x=1"`, "a.pdfSet-Cookie: x=1"},
		{"", "attachment; filename=unduhan", "unduhan"},
		{"  ../ ", "attachment; filename=unduhan", "unduhan"},
	}
	for _, k := range kasus {
		h := HeaderUnduhan("attachment", k.nama)
		if h != k.header {
			t.Errorf("HeaderUnduhan(%q) = %q, mau %q", k.nama, h, k.header)
		}
		if strings.ContainsAny(h, "\r\n") {
			t.Errorf("HeaderUnduhan(%q) berisi baris baru: %q", k.nama, h)
		}
		// Browser membaca header dengan aturan yang sama dengan mime.ParseMediaType.
		jenis, p, err := mime.ParseMediaType(h)
		if err != nil || jenis != "attachment" || p["filename"] != k.hasil {
			t.Errorf("baca ulang %q = %q %q (%v), mau nama %q", h, jenis, p["filename"], err, k.hasil)
		}
	}
	if h := HeaderUnduhan("inline", "pratinjau.pdf"); h != "inline; filename=pratinjau.pdf" {
		t.Errorf("inline = %q", h)
	}
}
