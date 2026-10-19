package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"regexp"
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
	out, _, e := NewAgreementGenerator().Generate(d.Template, values, GenerateOptions{})
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

// Isian masuk ke tempatnya di template PKS bawaan: nomor, surat rujukan, termin,
// rekening dan alamat dengan tab rata, serta poin ruang lingkup.
func TestPKSFieldPlacement(t *testing.T) {
	registry, err := NewAgreementRegistry()
	if err != nil {
		t.Fatal(err)
	}
	definition, _ := registry.Get("PKS")
	values := map[string]string{}
	for _, key := range []string{"NOMOR_PIHAK_PERTAMA", "NOMOR_PIHAK_KEDUA", "HARI_TTD", "TANGGAL_TTD", "BULAN_TTD", "TAHUN_TTD", "TANGGAL_TTD_LENGKAP", "TEMPAT_TTD", "PIHAK_PERTAMA_PEJABAT", "PIHAK_PERTAMA_JABATAN", "PIHAK_KEDUA_NAMA", "PIHAK_KEDUA_ALAMAT", "PIHAK_KEDUA_PEJABAT", "PIHAK_KEDUA_JABATAN", "PIHAK_KEDUA_BIDANG", "JENIS_PEKERJAAN", "SURAT_PENAWARAN_NOMOR", "SURAT_PENAWARAN_PERIHAL", "SURAT_PENAWARAN_TANGGAL", "SURAT_PENUNJUKAN_NOMOR", "SURAT_PENUNJUKAN_PERIHAL", "SURAT_PENUNJUKAN_TANGGAL", "RUANG_LINGKUP", "JANGKA_WAKTU_MULAI", "JANGKA_WAKTU_SELESAI", "NILAI_KONTRAK", "NILAI_KONTRAK_TERBILANG", "TERMIN_1_PERSEN", "TERMIN_1_PERSEN_TERBILANG", "TERMIN_1_NILAI", "TERMIN_1_NILAI_TERBILANG", "TERMIN_2_PERSEN", "TERMIN_2_PERSEN_TERBILANG", "TERMIN_2_NILAI", "TERMIN_2_NILAI_TERBILANG", "BANK", "NOMOR_REKENING", "ATAS_NAMA", "DAFTAR_LAMPIRAN", "PIHAK_PERTAMA_ALAMAT", "PIHAK_PERTAMA_TELEPON", "PIHAK_PERTAMA_EMAIL", "PIHAK_PERTAMA_PIC", "PIHAK_KEDUA_TELEPON", "PIHAK_KEDUA_EMAIL", "PIHAK_KEDUA_PIC"} {
		values[key] = "TEST-" + key
	}
	values["NOMOR_PIHAK_KEDUA"] = "SECOND-32868382846921"
	values["PIHAK_KEDUA_NAMA"] = "PT PIHAK KEDUA"
	values["SURAT_PENAWARAN_NOMOR"] = "OFFER-001"
	values["SURAT_PENAWARAN_PERIHAL"] = "PENAWARAN IKAN"
	values["SURAT_PENAWARAN_TANGGAL"] = "15 Juli 2026"
	values["RUANG_LINGKUP"] = "a. Analisis kebutuhan\n- Implementasi sistem\n3. Pelatihan pengguna"
	values["TERMIN_1_PERSEN"] = "50"
	values["TERMIN_1_PERSEN_TERBILANG"] = "lima puluh"
	values["TERMIN_1_NILAI"] = "Rp 10.000.000"
	values["TERMIN_1_NILAI_TERBILANG"] = "sepuluh juta"
	values["TERMIN_2_PERSEN"] = "50"
	values["TERMIN_2_PERSEN_TERBILANG"] = "lima puluh"
	values["TERMIN_2_NILAI"] = "Rp 10.000.000"
	values["TERMIN_2_NILAI_TERBILANG"] = "sepuluh juta"

	out, _, err := NewAgreementGenerator().Generate(definition.Template, values, GenerateOptions{Draft: true})
	if err != nil {
		t.Fatal(err)
	}
	xmlText := generatedDocumentXML(t, out)
	paragraphs := wordParagraphRE.FindAllString(xmlText, -1)
	texts := make([]string, len(paragraphs))
	for i, paragraph := range paragraphs {
		texts[i] = strings.TrimSpace(paragraphText(paragraph))
	}

	assertNextParagraph(t, texts, "NOMOR PIHAK KEDUA", "SECOND-32868382846921")
	assertContains(t, texts, "Surat Penawaran No. OFFER-001 perihal PENAWARAN IKAN pada tanggal 15 Juli 2026")
	assertContains(t, texts, "Termin pertama: Pembayaran sebesar 50% (lima puluh persen) dari Nilai Kontrak atau sebesar Rp 10.000.000 (sepuluh juta rupiah)")
	assertContains(t, texts, "Termin kedua: Pembayaran sebesar 50% (lima puluh persen) dari Nilai Kontrak atau sebesar Rp 10.000.000 (sepuluh juta rupiah)")
	assertKeyValueTabs(t, paragraphs, texts, "Bank:TEST-BANK", 2880, 3060)
	assertKeyValueTabs(t, paragraphs, texts, "Nomor Rekening:TEST-NOMOR_REKENING", 2880, 3060)
	assertKeyValueTabs(t, paragraphs, texts, "Atas Nama:TEST-ATAS_NAMA", 2880, 3060)
	assertKeyValueTabs(t, paragraphs, texts, "Alamat:TEST-PIHAK_PERTAMA_ALAMAT", 1800, 1980)
	assertKeyValueTabs(t, paragraphs, texts, "Telepon:TEST-PIHAK_KEDUA_TELEPON", 1800, 1980)
	for _, item := range []string{"Analisis kebutuhan", "Implementasi sistem", "Pelatihan pengguna"} {
		assertContains(t, texts, item)
	}
	assertContains(t, texts, "RUANG LINGKUP PEKERJAAN")
}

func TestIndonesianNumberParsing(t *testing.T) {
	for input, expected := range map[string]int64{"10000000": 10000000, "10.000.000": 10000000, "Rp 10.000.000": 10000000} {
		if actual := valueInt64(input); actual != expected {
			t.Fatalf("valueInt64(%q) = %d, ingin %d", input, actual, expected)
		}
	}
	if actual := valueFloat("50,5"); actual != 50.5 {
		t.Fatalf("valueFloat(50,5) = %v", actual)
	}
}

func generatedDocumentXML(t *testing.T, docx []byte) string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range r.File {
		if file.Name != "word/document.xml" {
			continue
		}
		stream, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(stream)
		stream.Close()
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	t.Fatal("word/document.xml tidak ditemukan")
	return ""
}

func assertNextParagraph(t *testing.T, paragraphs []string, label, expected string) {
	t.Helper()
	for i, paragraph := range paragraphs {
		if paragraph == label && i+1 < len(paragraphs) {
			if paragraphs[i+1] != expected {
				t.Fatalf("paragraf setelah %q = %q, ingin %q", label, paragraphs[i+1], expected)
			}
			return
		}
	}
	t.Fatalf("label %q tidak ditemukan", label)
}

func assertContains(t *testing.T, paragraphs []string, expected string) {
	t.Helper()
	for _, paragraph := range paragraphs {
		if strings.Contains(paragraph, expected) {
			return
		}
	}
	t.Fatalf("teks %q tidak ditemukan", expected)
}

func assertKeyValueTabs(t *testing.T, paragraphs, texts []string, expected string, colonPosition, valuePosition int) {
	t.Helper()
	for i, text := range texts {
		if text != expected {
			continue
		}
		colonTab := fmt.Sprintf(`w:pos="%d"`, colonPosition)
		valueTab := fmt.Sprintf(`w:pos="%d"`, valuePosition)
		if !strings.Contains(paragraphs[i], colonTab) || !strings.Contains(paragraphs[i], valueTab) || strings.Count(paragraphs[i], "<w:tab/>") != 2 {
			t.Fatalf("tab stop untuk %q tidak sesuai", expected)
		}
		return
	}
	t.Fatalf("rincian %q tidak ditemukan", expected)
}

var (
	wordTableRE       = regexp.MustCompile(`(?s)<w:tbl>.*?</w:tbl>`)
	pageBreakBeforeRE = regexp.MustCompile(`<w:pageBreakBefore(?:\s+w:val="(?:1|true|on)")?\s*/>`)
	pageBreakRunRE    = regexp.MustCompile(`<w:br\b[^>]*w:type="page"`)
)

// Template PKS bawaan harus diisi lewat placeholder, dan tiap BAGIAN mulai di
// halaman baru lewat pemisah halaman, bukan deretan Enter kosong. Deretan Enter
// ikut terdorong saat isian di atasnya panjang, sehingga muncul ruang atau
// halaman kosong.
func TestTemplatePKSBawaan(t *testing.T) {
	registry, err := NewAgreementRegistry()
	if err != nil {
		t.Fatal(err)
	}
	definition, _ := registry.Get("PKS")

	placeholders, _, err := InspectTemplate(definition.Template)
	if err != nil {
		t.Fatal(err)
	}
	if len(placeholders) == 0 {
		t.Error("template bawaan belum berisi placeholder; masih template lama yang diisi lewat pencocokan kalimat")
	}
	known := map[string]bool{}
	for _, token := range SupportedPlaceholderTokens() {
		known[token] = true
	}
	for _, p := range placeholders {
		if !known[p] {
			t.Errorf("placeholder %s tidak dikenal generator", p)
		}
	}

	// Watermark DRAFT di header tercetak di setiap halaman, termasuk dokumen
	// final. Penanda draft ditambahkan sistem hanya untuk pratinjau.
	z, err := zip.NewReader(bytes.NewReader(definition.Template), int64(len(definition.Template)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range z.File {
		if !strings.HasPrefix(f.Name, "word/header") {
			continue
		}
		rc, _ := f.Open()
		data, _ := io.ReadAll(rc)
		rc.Close()
		if strings.Contains(string(data), "WaterMark") {
			t.Errorf("%s berisi watermark; template bawaan tidak boleh membawa tanda DRAFT", f.Name)
		}
	}

	xmlText := generatedDocumentXML(t, definition.Template)

	// Ruang kosong di dalam tabel (mis. tempat tanda tangan basah) disengaja,
	// jadi hanya paragraf di luar tabel yang diperiksa.
	run, longest, before := 0, 0, ""
	for _, paragraph := range wordParagraphRE.FindAllString(wordTableRE.ReplaceAllString(xmlText, ""), -1) {
		text := strings.TrimSpace(paragraphText(paragraph))
		if text == "" {
			run++
			continue
		}
		if run > longest {
			longest, before = run, text
		}
		run = 0
	}
	if longest >= 4 {
		t.Errorf("ada %d paragraf kosong berturut-turut sebelum %q; pakai pemisah halaman, bukan Enter berulang", longest, before)
	}

	paragraphs := wordParagraphRE.FindAllString(xmlText, -1)
	for _, heading := range []string{"BAGIAN I", "BAGIAN II", "BAGIAN III"} {
		i := -1
		for k, paragraph := range paragraphs {
			if strings.TrimSpace(paragraphText(paragraph)) == heading {
				i = k
				break
			}
		}
		if i < 1 {
			t.Errorf("judul %q tidak ditemukan", heading)
			continue
		}
		breaks := len(pageBreakBeforeRE.FindAllString(paragraphs[i], -1)) +
			len(pageBreakRunRE.FindAllString(paragraphs[i], -1)) +
			len(pageBreakRunRE.FindAllString(paragraphs[i-1], -1))
		if breaks != 1 {
			t.Errorf("judul %q didahului %d pemisah halaman, harus tepat 1", heading, breaks)
		}
	}
}

// Isian yang panjang harus tampil utuh di dokumen, tanpa placeholder yang tersisa.
func TestIsianPanjangTerisiSemua(t *testing.T) {
	registry, err := NewAgreementRegistry()
	if err != nil {
		t.Fatal(err)
	}
	definition, _ := registry.Get("PKS")

	values := sampleAgreementValues()
	alamat := strings.TrimSpace(strings.Repeat("Jl. Contoh Panjang Nomor 123 Blok A, Kawasan Industri, ", 11))
	values["PIHAK_KEDUA_ALAMAT"] = alamat
	poin := make([]string, 20)
	for i := range poin {
		poin[i] = fmt.Sprintf("Poin ruang lingkup nomor %d", i+1)
	}
	values["RUANG_LINGKUP"] = strings.Join(poin, "\n")

	out, _, err := NewAgreementGenerator().Generate(definition.Template, values, GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	xmlText := generatedDocumentXML(t, out)
	texts := map[string]bool{}
	var all []string
	for _, paragraph := range wordParagraphRE.FindAllString(xmlText, -1) {
		text := strings.TrimSpace(paragraphText(paragraph))
		texts[text] = true
		all = append(all, text)
	}

	if len(alamat) < 600 {
		t.Fatalf("alamat uji hanya %d karakter", len(alamat))
	}
	if !strings.Contains(strings.Join(all, "\n"), alamat) {
		t.Error("alamat pihak kedua yang panjang tidak tampil utuh")
	}
	for _, item := range poin {
		if !texts[item] {
			t.Errorf("poin ruang lingkup %q tidak tampil sebagai paragraf sendiri", item)
		}
	}
	if left := placeholderRE.FindAllString(xmlText, -1); len(left) > 0 {
		t.Errorf("placeholder tersisa: %v", left)
	}
}
