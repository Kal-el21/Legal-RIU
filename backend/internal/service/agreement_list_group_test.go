package service

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"legal-riu-portal/internal/entity"
)

// bagianTermin: jenis uji dengan nilai kontrak, grup termin yang membagi nilai,
// dan daftar ruang lingkup.
func bagianTermin() []AgreementSectionDefinition {
	return []AgreementSectionDefinition{{Title: "Pembayaran", Fields: []AgreementFieldDefinition{
		{Name: "nilai_kontrak", Label: "Nilai Kontrak", Type: "money", Required: true},
		{Name: "termin", Label: "Termin", Type: "group", Required: true, SplitOf: "nilai_kontrak", Fields: []AgreementFieldDefinition{
			{Name: "persen", Label: "Persen", Type: "percent", Required: true},
			{Name: "nilai", Label: "Nilai", Type: "money", Required: true},
			{Name: "syarat", Label: "Syarat", Type: "text", Required: true},
		}},
		{Name: "ruang_lingkup", Label: "Ruang Lingkup", Type: "list", Required: true},
	}}}
}

// isianJSON membaca isian seperti saat dikirim browser (angka sebagai json.Number).
func isianJSON(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	data, err := decodeForm([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// teksParagraf: teks setiap paragraf isi dokumen hasil Generate.
func teksParagraf(t *testing.T, docx []byte) []string {
	t.Helper()
	var out []string
	for _, p := range wordParagraphRE.FindAllString(generatedDocumentXML(t, docx), -1) {
		out = append(out, paragraphText(p))
	}
	return out
}

const paragrafPoin = `<w:p w14:paraId="1A2B3C4D" w14:textId="77777777"><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t>{{RUANG_</w:t></w:r><w:r><w:t>LINGKUP}}</w:t></w:r></w:p>`

// Paragraf daftar diulang sekali per poin dengan format numbered list-nya, poin
// kosong dibuang saat validasi, dan daftar kosong tercetak "-".
func TestDaftarPoinDiulang(t *testing.T) {
	template := docxDenganIsi(t, `<w:p><w:r><w:t>Ruang lingkup:</w:t></w:r></w:p>`+paragrafPoin+`<w:p><w:r><w:t>Penutup.</w:t></w:r></w:p><w:sectPr/>`)
	bentuk := func(t *testing.T, f AgreementFieldDefinition, data map[string]interface{}) []byte {
		t.Helper()
		sections := []AgreementSectionDefinition{{Title: "Data", Fields: []AgreementFieldDefinition{f}}}
		if err := validateAgreementForm(sections, data); err != nil {
			t.Fatal(err)
		}
		out, _, err := NewAgreementGenerator().Generate(template, placeholderValues(&entity.AgreementDocument{}, sections, data, nil), GenerateOptions{})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	daftar := bagianTermin()[0].Fields[2]

	data := isianJSON(t, `{"ruang_lingkup": ["Analisis kebutuhan", "  ", "Implementasi & uji", "Pelatihan"]}`)
	out := bentuk(t, daftar, data)
	if got := fmt.Sprint(data["ruang_lingkup"]); got != "[Analisis kebutuhan Implementasi & uji Pelatihan]" {
		t.Errorf("isian tersimpan %s, poin kosong harus dibuang", got)
	}
	if got := strings.Join(teksParagraf(t, out), "|"); got != "Ruang lingkup:|Analisis kebutuhan|Implementasi & uji|Pelatihan|Penutup." {
		t.Errorf("paragraf %q", got)
	}
	doc := generatedDocumentXML(t, out)
	if n := strings.Count(doc, "<w:numPr>"); n != 3 {
		t.Errorf("%d paragraf bernomor, harus 3", n)
	}
	if strings.Contains(doc, "w14:paraId") {
		t.Error("salinan paragraf masih membawa id paragraf yang sama")
	}

	daftar.Required = false
	if got := strings.Join(teksParagraf(t, bentuk(t, daftar, isianJSON(t, `{"ruang_lingkup": []}`))), "|"); got != "Ruang lingkup:|-|Penutup." {
		t.Errorf("daftar kosong: %q", got)
	}
}

// Paragraf termin diulang sekali per baris dengan urutan, persen, dan nilai yang
// dihitung server; persen terakhir dan nilai dari browser diabaikan.
func TestTerminDiulang(t *testing.T) {
	template := docxDenganIsi(t, `<w:p><w:r><w:t>Nilai kontrak {{NILAI_KONTRAK}} dibagi menjadi {{TERMIN_JUMLAH}} termin:</w:t></w:r></w:p>`+
		`<w:p><w:r><w:t>Termin {{TERMIN_</w:t></w:r><w:r><w:t>URUTAN}}: {{TERMIN_PERSEN}}% ({{TERMIN_PERSEN_TERBILANG}} persen) sebesar {{TERMIN_NILAI}} ({{TERMIN_NILAI_TERBILANG}} rupiah), dibayarkan {{TERMIN_SYARAT}} (akhir termin {{TERMIN_URUTAN}}).</w:t></w:r></w:p>`+
		`<w:p><w:r><w:t>Penutup.</w:t></w:r></w:p><w:sectPr/>`)
	sections := bagianTermin()
	data := isianJSON(t, `{"nilai_kontrak": "Rp 750.000.000", "ruang_lingkup": ["Analisis"], "termin": [
		{"persen": 30, "nilai": 1, "syarat": "setelah penandatanganan"},
		{"persen": "30", "syarat": "setelah laporan"},
		{"persen": 99, "nilai": 5, "syarat": "setelah serah terima"}]}`)
	if err := validateAgreementForm(sections, data); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprint(data["termin"]); got != "[map[nilai:225000000 persen:30 syarat:setelah penandatanganan] map[nilai:225000000 persen:30 syarat:setelah laporan] map[nilai:300000000 persen:40 syarat:setelah serah terima]]" {
		t.Errorf("isian termin tersimpan %s", got)
	}
	out, _, err := NewAgreementGenerator().Generate(template, placeholderValues(&entity.AgreementDocument{}, sections, data, nil), GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	harap := []string{
		"Nilai kontrak Rp 750.000.000 dibagi menjadi 3 (tiga) termin:",
		"Termin pertama: 30% (tiga puluh persen) sebesar Rp 225.000.000 (dua ratus dua puluh lima juta rupiah), dibayarkan setelah penandatanganan (akhir termin pertama).",
		"Termin kedua: 30% (tiga puluh persen) sebesar Rp 225.000.000 (dua ratus dua puluh lima juta rupiah), dibayarkan setelah laporan (akhir termin kedua).",
		"Termin ketiga: 40% (empat puluh persen) sebesar Rp 300.000.000 (tiga ratus juta rupiah), dibayarkan setelah serah terima (akhir termin ketiga).",
		"Penutup.",
	}
	if got := teksParagraf(t, out); strings.Join(got, "\n") != strings.Join(harap, "\n") {
		t.Errorf("paragraf:\n%s\nharus:\n%s", strings.Join(got, "\n"), strings.Join(harap, "\n"))
	}
}

// Placeholder per baris satu grup yang tersebar di beberapa paragraf ditolak saat
// unggah, begitu juga paragraf daftar yang bukan numbered list.
func TestGrupTerpecahDitolak(t *testing.T) {
	sections := bagianTermin()
	for _, c := range []struct {
		nama, body, galat string
	}{
		{"SatuParagraf", `<w:p><w:r><w:t>Termin {{TERMIN_URUTAN}}: {{TERMIN_PERSEN}}% {{TERMIN_SYARAT}}</w:t></w:r></w:p>` + paragrafPoin, ""},
		{"GrupTerpecah", `<w:p><w:r><w:t>Termin {{TERMIN_URUTAN}}: {{TERMIN_PERSEN}}%</w:t></w:r></w:p><w:p><w:r><w:t>Syarat: {{TERMIN_SYARAT}}</w:t></w:r></w:p>` + paragrafPoin, "Termin tersebar di 2 paragraf"},
		{"JumlahBolehDiLuar", `<w:p><w:r><w:t>Dibagi {{TERMIN_JUMLAH}} termin.</w:t></w:r></w:p><w:p><w:r><w:t>Termin {{TERMIN_URUTAN}}: {{TERMIN_NILAI}}</w:t></w:r></w:p>` + paragrafPoin, ""},
		{"DaftarTanpaNomor", `<w:p><w:r><w:t>Termin {{TERMIN_URUTAN}}</w:t></w:r></w:p><w:p><w:r><w:t>{{RUANG_LINGKUP}}</w:t></w:r></w:p>`, "numbered list"},
	} {
		t.Run(c.nama, func(t *testing.T) {
			_, err := periksaTemplate(docxDenganIsi(t, c.body+`<w:sectPr/>`), sections)
			switch {
			case c.galat == "" && err != nil:
				t.Fatalf("ditolak: %v", err)
			case c.galat != "" && (err == nil || !strings.Contains(err.Error(), c.galat)):
				t.Fatalf("galat %v, harus menyebut %q", err, c.galat)
			}
		})
	}
}

// Isian yang memuat teks placeholder (diketik pengaju, atau nama file lampiran)
// tidak membuat pembentukan dokumen berputar tanpa henti. Sebelumnya isian
// "PT {{NAMA}} Jaya" untuk {{NAMA}} diganti terus dan memorinya terus bertambah.
func TestIsianBerisiPlaceholderTidakBerputar(t *testing.T) {
	template := docxDenganIsi(t, `<w:p><w:r><w:t>{{NA</w:t></w:r><w:r><w:t>MA}} dan {{NAMA}}</w:t></w:r></w:p>`+paragrafPoin+`<w:sectPr/>`)
	selesai := make(chan error, 1)
	go func() {
		isi := IsianDokumen{Nilai: map[string]string{"NAMA": "PT {{NAMA}} Jaya"}, Ulang: []ParagrafUlang{{Token: []string{"RUANG_LINGKUP"}, Baris: []map[string]string{{"RUANG_LINGKUP": "lihat {{RUANG_LINGKUP}}"}}}}}
		_, _, err := NewAgreementGenerator().Generate(template, isi, GenerateOptions{})
		selesai <- err
	}()
	select {
	case err := <-selesai:
		// Teks placeholder di dalam isian tersisa, jadi ditolak sebagai placeholder belum terisi.
		if err == nil || !strings.Contains(err.Error(), "placeholder belum terisi") {
			t.Errorf("galat %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("pembentukan dokumen tidak selesai dalam 5 detik")
	}
}

// Isian daftar dan grup yang tidak lengkap atau tidak berbentuk benar ditolak
// dengan pesan yang menyebut isiannya.
func TestValidasiDaftarDanGrup(t *testing.T) {
	const termin = `[{"persen": 30, "syarat": "a"}, {"syarat": "b"}]`
	for _, c := range []struct {
		nama, isian, galat string
	}{
		{"Lengkap", `{"nilai_kontrak": 1000, "termin": ` + termin + `, "ruang_lingkup": ["x"]}`, ""},
		{"DaftarKosong", `{"nilai_kontrak": 1000, "termin": ` + termin + `, "ruang_lingkup": ["", " "]}`, "Ruang Lingkup wajib diisi minimal satu poin"},
		{"DaftarBukanArray", `{"nilai_kontrak": 1000, "termin": ` + termin + `, "ruang_lingkup": "x\ny"}`, "Ruang Lingkup harus berupa daftar poin"},
		{"GrupKosong", `{"nilai_kontrak": 1000, "termin": [], "ruang_lingkup": ["x"]}`, "Termin wajib diisi minimal satu baris"},
		{"BarisBukanObjek", `{"nilai_kontrak": 1000, "termin": ["30"], "ruang_lingkup": ["x"]}`, "Termin baris 1 tidak valid"},
		{"SubFieldWajibKosong", `{"nilai_kontrak": 1000, "termin": [{"persen": 30, "syarat": "a"}, {"syarat": " "}], "ruang_lingkup": ["x"]}`, "Termin 2: Syarat wajib diisi"},
		{"PersenMelebihi100", `{"nilai_kontrak": 1000, "termin": [{"persen": 70, "syarat": "a"}, {"persen": 40, "syarat": "b"}, {"syarat": "c"}], "ruang_lingkup": ["x"]}`, "Termin: jumlah persen termin melebihi 100%"},
		{"PersenKosong", `{"nilai_kontrak": 1000, "termin": [{"syarat": "a"}, {"syarat": "b"}], "ruang_lingkup": ["x"]}`, "Termin: persen termin 1 harus lebih dari 0"},
	} {
		t.Run(c.nama, func(t *testing.T) {
			err := validateAgreementForm(bagianTermin(), isianJSON(t, c.isian))
			switch {
			case c.galat == "" && err != nil:
				t.Fatalf("ditolak: %v", err)
			case c.galat != "" && (err == nil || err.Error() != c.galat):
				t.Fatalf("galat %v, harus %q", err, c.galat)
			}
		})
	}
}

// Katalog jenis dengan grup dan daftar memuat placeholder per baris dan jumlah,
// definisinya lolos rapikanField, dan contoh termin dibagi dari contoh nilai kontrak.
func TestKatalogDaftarDanGrup(t *testing.T) {
	if _, err := rapikanField(bagianTermin()); err != nil {
		t.Fatal(err)
	}
	ada := map[string]bool{}
	for _, token := range SupportedPlaceholderTokens(bagianTermin()) {
		ada[token] = true
	}
	for _, token := range []string{"{{TERMIN_JUMLAH}}", "{{TERMIN_URUTAN}}", "{{TERMIN_PERSEN}}", "{{TERMIN_PERSEN_TERBILANG}}", "{{TERMIN_NILAI_TERBILANG}}", "{{TERMIN_SYARAT}}", "{{RUANG_LINGKUP}}"} {
		if !ada[token] {
			t.Errorf("katalog tanpa %s", token)
		}
	}
	for _, token := range []string{"{{TERMIN}}", scopeListPlaceholder} {
		if ada[token] {
			t.Errorf("katalog tidak boleh menawarkan %s", token)
		}
	}
	contoh := sampleAgreementValues(bagianTermin())
	for _, u := range contoh.Ulang {
		if u.Token[0] == "TERMIN_URUTAN" && (len(u.Baris) != 2 || u.Baris[1]["TERMIN_NILAI"] != "Rp 10.000.000") {
			t.Errorf("contoh termin %v, harus 2 baris masing-masing Rp 10.000.000", u.Baris)
		}
	}
}
