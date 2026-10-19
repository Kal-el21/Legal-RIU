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

func TestInspectTemplateMenolakFileBukanDOCX(t *testing.T) {
	if _, _, err := InspectTemplate([]byte("bukan docx")); err == nil {
		t.Fatal("file rusak harus ditolak")
	}
}
