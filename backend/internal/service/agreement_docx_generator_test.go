package service

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestPKSTemplateGeneration(t *testing.T) {
	r, e := NewAgreementRegistry()
	if e != nil {
		t.Fatal(e)
	}
	d, _ := r.Get("PKS")
	values := map[string]string{}
	for _, k := range []string{"NOMOR_PIHAK_PERTAMA", "NOMOR_PIHAK_KEDUA", "HARI_TTD", "TANGGAL_TTD", "BULAN_TTD", "TAHUN_TTD", "TANGGAL_TTD_LENGKAP", "TEMPAT_TTD", "PIHAK_PERTAMA_PEJABAT", "PIHAK_PERTAMA_JABATAN", "PIHAK_KEDUA_NAMA", "PIHAK_KEDUA_ALAMAT", "PIHAK_KEDUA_PEJABAT", "PIHAK_KEDUA_JABATAN", "PIHAK_KEDUA_BIDANG", "JENIS_PEKERJAAN", "SURAT_PENAWARAN_NOMOR", "SURAT_PENAWARAN_PERIHAL", "SURAT_PENAWARAN_TANGGAL", "SURAT_PENUNJUKAN_NOMOR", "SURAT_PENUNJUKAN_PERIHAL", "SURAT_PENUNJUKAN_TANGGAL", "RUANG_LINGKUP", "JANGKA_WAKTU_MULAI", "JANGKA_WAKTU_SELESAI", "NILAI_KONTRAK", "NILAI_KONTRAK_TERBILANG", "TERMIN_1_PERSEN", "TERMIN_1_PERSEN_TERBILANG", "TERMIN_1_NILAI", "TERMIN_1_NILAI_TERBILANG", "TERMIN_2_PERSEN", "TERMIN_2_PERSEN_TERBILANG", "TERMIN_2_NILAI", "TERMIN_2_NILAI_TERBILANG", "BANK", "NOMOR_REKENING", "ATAS_NAMA", "DAFTAR_LAMPIRAN"} {
		values[k] = "TEST"
	}
	for _, k := range []string{"PIHAK_PERTAMA_ALAMAT", "PIHAK_PERTAMA_TELEPON", "PIHAK_PERTAMA_EMAIL", "PIHAK_PERTAMA_PIC", "PIHAK_KEDUA_TELEPON", "PIHAK_KEDUA_EMAIL", "PIHAK_KEDUA_PIC"} {
		values[k] = "TEST"
	}
	out, _, e := NewAgreementGenerator().Generate(d.Template, values, false)
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, f := range z.File {
		if f.Name == "word/document.xml" {
			s, _ := f.Open()
			b, _ := io.ReadAll(s)
			s.Close()
			found = strings.Contains(string(b), "TEST")
		}
	}
	if !found {
		t.Fatal("generated values not found")
	}
}

func TestPKSLegacyDocumentTypeAliases(t *testing.T) {
	r, err := NewAgreementRegistry()
	if err != nil {
		t.Fatal(err)
	}

	for _, code := range []string{"PKS", " pks ", "", "PERJANJIAN KERJA SAMA", "PERJANJIAN_KERJA_SAMA", "lain-lain", "LAIN LAIN", "OTHER"} {
		definition, ok := r.Get(code)
		if !ok || definition.Code != "PKS" || len(definition.Template) == 0 {
			t.Fatalf("kode lama %q tidak menghasilkan template PKS", code)
		}
	}
}
