package service

import (
	"maps"
	"testing"

	"legal-riu-portal/internal/assets"
	"legal-riu-portal/internal/entity"
)

func bacaAset(t *testing.T, path string) []byte {
	t.Helper()
	data, err := assets.Files.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Template PKS bawaan PR 10 dengan 2 termin menghasilkan dokumen yang sama persis
// dengan template PR 8 untuk isian yang sama, kalimat demi kalimat. Termin kedua
// dan persennya kini dihitung server, dan ruang lingkup diisi sebagai daftar.
func TestTemplateBaruSamaDenganPR9(t *testing.T) {
	const (
		syarat1 = "setelah Perjanjian ini ditandatangani oleh PARA PIHAK"
		syarat2 = "setelah penyampaian Hasil Laporan Pekerjaan dan Berita Acara Serah Terima Pekerjaan disetujui oleh PIHAK PERTAMA dan ditandatangani oleh PARA PIHAK"
	)
	nomor := "001/PKS/RM.01.01/2026"
	doc := &entity.AgreementDocument{AgreementNumber: &nomor, Attachments: []entity.AgreementAttachment{{FileName: "penawaran.pdf"}}}
	snap := map[string]interface{}{"name": "PT Reasuransi Indonesia Utama (Persero)", "signatory_name": "Budi", "signatory_position": "Direktur", "signing_place": "Jakarta"}
	bentuk := func(t *testing.T, template string, sections []AgreementSectionDefinition, data map[string]interface{}) []string {
		t.Helper()
		if err := validateAgreementForm(sections, data); err != nil {
			t.Fatal(err)
		}
		out, _, err := NewAgreementGenerator().Generate(bacaAset(t, template), placeholderValues(doc, sections, data, snap), GenerateOptions{})
		if err != nil {
			t.Fatal(err)
		}
		return teksParagraf(t, out)
	}

	lama := map[string]interface{}{"tanggal_ttd": "2026-08-17"}
	for _, s := range bagianPKSPR9(t) {
		for _, f := range s.Fields {
			lama[f.Name] = contohNilai(f)
		}
	}
	maps.Copy(lama, map[string]interface{}{
		"ruang_lingkup": "Analisis kebutuhan\nImplementasi sistem\nPelatihan pengguna",
		"nilai_kontrak": "750000000", "termin_1_persen": "40", "termin_1_nilai": "300000000", "termin_2_persen": "60", "termin_2_nilai": "450000000",
	})
	baru := maps.Clone(lama)
	for _, k := range []string{"termin_1_persen", "termin_1_nilai", "termin_2_persen", "termin_2_nilai"} {
		delete(baru, k)
	}
	baru["ruang_lingkup"] = []interface{}{"Analisis kebutuhan", "Implementasi sistem", "Pelatihan pengguna"}
	baru["termin"] = []interface{}{map[string]interface{}{"persen": "40", "syarat": syarat1}, map[string]interface{}{"syarat": syarat2}}

	sebelum := bentuk(t, "templates/pks-template/pr9/Draft-PKS-RIU.docx", bagianPKSPR9(t), lama)
	sesudah := bentuk(t, "templates/pks-template/Draft-PKS-RIU.docx", bagianPKS(t), baru)
	if len(sebelum) != len(sesudah) {
		t.Errorf("%d paragraf, sebelumnya %d", len(sesudah), len(sebelum))
	}
	for i := range min(len(sebelum), len(sesudah)) {
		if sebelum[i] != sesudah[i] {
			t.Fatalf("paragraf %d berbeda:\n%q\nsebelumnya:\n%q", i, sesudah[i], sebelum[i])
		}
	}
}

// Definisi field dan template PKS bawaan dipasang seed tanpa melewati Jenis
// Perjanjian dan unggah template, jadi keduanya harus lolos pemeriksaan yang sama:
// field sah, placeholder dikenal, ruang lingkup bernomor, termin dalam satu
// paragraf, dan semua field terpakai. Saran termin memuat kedua syarat lama.
func TestTemplatePKSBawaanLolosPemeriksaan(t *testing.T) {
	sections, err := rapikanField(bagianPKS(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		template string
		sections []AgreementSectionDefinition
	}{
		{"templates/pks-template/Draft-PKS-RIU.docx", sections},
		{"templates/pks-template/pr9/Draft-PKS-RIU.docx", bagianPKSPR9(t)},
	} {
		hasil, err := periksaTemplate(bacaAset(t, c.template), c.sections)
		if err != nil {
			t.Fatalf("%s: %v", c.template, err)
		}
		if len(hasil.UnusedPlaceholders) > 0 {
			t.Errorf("%s: field tidak terpakai %v", c.template, hasil.UnusedPlaceholders)
		}
	}
	for _, s := range sections {
		for _, f := range s.Fields {
			if f.Name == "termin" && (f.SplitOf != "nilai_kontrak" || len(f.Fields[2].Options) != 2) {
				t.Errorf("grup termin: %+v", f)
			}
		}
	}
}
