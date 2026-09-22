package service

import (
	"strings"
	"testing"
)

func TestSupportedPlaceholderTokens(t *testing.T) {
	tokens := map[string]bool{}
	for _, token := range SupportedPlaceholderTokens() {
		tokens[token] = true
	}
	for _, want := range []string{"{{NOMOR_PIHAK_PERTAMA}}", "{{NILAI_KONTRAK_TERBILANG}}", "{{DAFTAR_LAMPIRAN}}", scopeListPlaceholder} {
		if !tokens[want] {
			t.Fatalf("token %s harus termasuk yang didukung", want)
		}
	}
}

func TestValidateTemplatePlaceholders(t *testing.T) {
	if err := validateTemplatePlaceholders([]string{"{{NILAI_KONTRAK}}", scopeListPlaceholder}); err != nil {
		t.Fatalf("token yang dikenal tidak boleh ditolak: %v", err)
	}
	err := validateTemplatePlaceholders([]string{"{{NILAI_KONTRAK}}", "{{DENDA_KETERLAMBATAN}}"})
	if err == nil {
		t.Fatal("token tidak dikenal harus ditolak")
	}
	if !strings.Contains(err.Error(), "{{DENDA_KETERLAMBATAN}}") {
		t.Fatalf("pesan error harus menyebut token bermasalah, dapat: %v", err)
	}
}

func TestPlaceholderCatalogTerdokumentasi(t *testing.T) {
	for _, item := range PlaceholderCatalog() {
		if item.Group == "" || item.Description == "" {
			t.Fatalf("token %s belum punya kelompok/keterangan di katalog", item.Token)
		}
	}
}

// Template bawaan lama ditambal lewat pencocokan teks paragraf. Penambalan itu
// hanya boleh jalan ketika Legacy diaktifkan, agar template baru hasil unggahan
// tim legal tidak ikut diubah diam-diam.
func TestLegacyPatchingHanyaSaatLegacyAktif(t *testing.T) {
	registry, err := NewAgreementRegistry()
	if err != nil {
		t.Fatal(err)
	}
	definition, _ := registry.Get("PKS")

	patched, _, err := NewAgreementGenerator().Generate(definition.Template, sampleAgreementValues(), GenerateOptions{Legacy: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(generatedDocumentXML(t, patched), "PT Contoh Mitra Sejahtera") == false {
		t.Fatal("mode legacy seharusnya mengisi nilai ke template bawaan")
	}

	untouched, _, err := NewAgreementGenerator().Generate(definition.Template, sampleAgreementValues(), GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(generatedDocumentXML(t, untouched), "____________") {
		t.Fatal("tanpa mode legacy, template bawaan harus dibiarkan apa adanya")
	}
}

func TestInspectTemplateMenolakFileBukanDOCX(t *testing.T) {
	if _, _, err := InspectTemplate([]byte("bukan docx")); err == nil {
		t.Fatal("file rusak harus ditolak")
	}
}
