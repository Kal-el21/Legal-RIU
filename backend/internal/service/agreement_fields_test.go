package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"legal-riu-portal/internal/entity"
)

// bagianPKS: definisi field PKS bawaan (form-schema.json), sama dengan yang
// dipasang migrasi ke versi template PKS.
func bagianPKS(t *testing.T) []AgreementSectionDefinition {
	t.Helper()
	r, err := NewAgreementRegistry()
	if err != nil {
		t.Fatal(err)
	}
	d, _ := r.Get("PKS")
	return d.Schema.Sections
}

// placeholderValuesSebelumPR9 adalah salinan apa adanya pemetaan manual sebelum
// PR 9 tahap 9.2, dipakai sebagai pembanding TestPKSSamaDenganSebelumnya.
func placeholderValuesSebelumPR9(d *entity.AgreementDocument, data, snap map[string]interface{}) map[string]string {
	get := func(k string) string {
		v := strings.TrimSpace(valueString(data[k]))
		if v == "" {
			return "-"
		}
		return v
	}
	date := valueString(data["tanggal_ttd"])
	dt, _ := parseDate(date)
	atts := make([]string, 0, len(d.Attachments))
	for i, a := range d.Attachments {
		atts = append(atts, fmt.Sprintf("%d. %s", i+1, a.FileName))
	}
	if len(atts) == 0 {
		atts = append(atts, "-")
	}
	money := func(k string) int64 { return valueInt64(data[k]) }
	pct := func(k string) int64 { return valueInt64(data[k]) }
	return map[string]string{"NOMOR_PIHAK_PERTAMA": strOrDash(d.AgreementNumber), "NOMOR_PIHAK_KEDUA": get("nomor_pihak_kedua"), "HARI_TTD": func() string {
		if dt.IsZero() {
			return "-"
		}
		return indoDays[dt.Weekday()]
	}(), "TANGGAL_TTD": func() string {
		if dt.IsZero() {
			return "-"
		}
		return fmt.Sprint(dt.Day())
	}(), "BULAN_TTD": func() string {
		if dt.IsZero() {
			return "-"
		}
		return indoMonths[dt.Month()]
	}(), "TAHUN_TTD": func() string {
		if dt.IsZero() {
			return "-"
		}
		return fmt.Sprint(dt.Year())
	}(), "TANGGAL_TTD_LENGKAP": indoDate(date), "TEMPAT_TTD": valueString(snap["signing_place"]), "PIHAK_PERTAMA_NAMA": valueString(snap["name"]), "PIHAK_PERTAMA_ALAMAT": valueString(snap["address"]), "PIHAK_PERTAMA_TELEPON": valueString(snap["phone"]), "PIHAK_PERTAMA_EMAIL": valueString(snap["email"]), "PIHAK_PERTAMA_PIC": valueString(snap["pic"]), "PIHAK_PERTAMA_PEJABAT": valueString(snap["signatory_name"]), "PIHAK_PERTAMA_JABATAN": valueString(snap["signatory_position"]), "PIHAK_KEDUA_NAMA": get("pihak_kedua_nama"), "PIHAK_KEDUA_BIDANG": get("pihak_kedua_bidang"), "PIHAK_KEDUA_ALAMAT": get("pihak_kedua_alamat"), "PIHAK_KEDUA_TELEPON": get("pihak_kedua_telepon"), "PIHAK_KEDUA_EMAIL": get("pihak_kedua_email"), "PIHAK_KEDUA_PIC": get("pihak_kedua_pic"), "PIHAK_KEDUA_PEJABAT": get("pihak_kedua_pejabat"), "PIHAK_KEDUA_JABATAN": get("pihak_kedua_jabatan"), "JENIS_PEKERJAAN": get("jenis_pekerjaan"), "RUANG_LINGKUP": get("ruang_lingkup"), "SURAT_PENAWARAN_NOMOR": get("surat_penawaran_nomor"), "SURAT_PENAWARAN_PERIHAL": get("surat_penawaran_perihal"), "SURAT_PENAWARAN_TANGGAL": indoDate(valueString(data["surat_penawaran_tanggal"])), "SURAT_PENUNJUKAN_NOMOR": get("surat_penunjukan_nomor"), "SURAT_PENUNJUKAN_PERIHAL": get("surat_penunjukan_perihal"), "SURAT_PENUNJUKAN_TANGGAL": indoDate(valueString(data["surat_penunjukan_tanggal"])), "JANGKA_WAKTU_MULAI": indoDate(valueString(data["jangka_waktu_mulai"])), "JANGKA_WAKTU_SELESAI": indoDate(valueString(data["jangka_waktu_selesai"])), "NILAI_KONTRAK": rupiah(money("nilai_kontrak")), "NILAI_KONTRAK_TERBILANG": terbilang(money("nilai_kontrak")), "TERMIN_1_PERSEN": fmt.Sprint(pct("termin_1_persen")), "TERMIN_1_PERSEN_TERBILANG": terbilang(pct("termin_1_persen")), "TERMIN_1_NILAI": rupiah(money("termin_1_nilai")), "TERMIN_1_NILAI_TERBILANG": terbilang(money("termin_1_nilai")), "TERMIN_2_PERSEN": fmt.Sprint(pct("termin_2_persen")), "TERMIN_2_PERSEN_TERBILANG": terbilang(pct("termin_2_persen")), "TERMIN_2_NILAI": rupiah(money("termin_2_nilai")), "TERMIN_2_NILAI_TERBILANG": terbilang(money("termin_2_nilai")), "BANK": get("bank"), "NOMOR_REKENING": get("nomor_rekening"), "ATAS_NAMA": get("atas_nama"), "DAFTAR_LAMPIRAN": strings.Join(atts, "\n")}
}

// Jaring pengaman refactor tahap 9.2: untuk data PKS yang sama, pemetaan dari
// definisi field menghasilkan nilai yang persis sama dengan pemetaan manual lama
// di setiap placeholder lama. Persen termin memakai bilangan bulat, karena persen
// desimal salah cetak di pemetaan lama (33.5 menjadi 335).
func TestPKSSamaDenganSebelumnya(t *testing.T) {
	nomor := "001/PKS/RM.01.01/2026"
	lengkap := &entity.AgreementDocument{AgreementNumber: &nomor, Attachments: []entity.AgreementAttachment{{FileName: "penawaran.pdf"}, {FileName: "spk.pdf"}}}
	snap := map[string]interface{}{"name": "PT Reasuransi Indonesia Utama (Persero)", "address": "Jl. Salemba Raya No. 30", "phone": "021 3920101", "email": "corsec@indonesiare.co.id", "pic": "Divisi Legal", "signatory_name": "Budi", "signatory_position": "Direktur", "signing_place": "Jakarta"}
	isi := func() map[string]interface{} {
		return map[string]interface{}{
			"nomor_pihak_kedua": "002/MITRA/2026", "tanggal_ttd": "2026-08-17", "tempat_ttd": "Bandung",
			"pihak_kedua_nama": "  PT Mitra Sejahtera ", "pihak_kedua_bidang": "konsultan", "pihak_kedua_alamat": "Jl. Mitra No. 1\nJakarta",
			"pihak_kedua_telepon": "021 555", "pihak_kedua_email": "mitra@contoh.co.id", "pihak_kedua_pic": "Rina",
			"pihak_kedua_pejabat": "Andi", "pihak_kedua_jabatan": "Direktur Utama",
			"jenis_pekerjaan": "jasa konsultasi", "ruang_lingkup": "Analisis\nImplementasi\nPelatihan",
			"surat_penawaran_nomor": "PNW-01", "surat_penawaran_perihal": "Penawaran", "surat_penawaran_tanggal": "2026-07-15",
			"surat_penunjukan_nomor": "PNJ-01", "surat_penunjukan_perihal": "Penunjukan", "surat_penunjukan_tanggal": "2026-07-30",
			"jangka_waktu_mulai": "2026-09-01", "jangka_waktu_selesai": "2027-08-31",
			"nilai_kontrak": json.Number("1500000000"), "termin_1_persen": json.Number("40"), "termin_1_nilai": "Rp 600.000.000",
			"termin_2_persen": "60", "termin_2_nilai": float64(900000000),
			"bank": "Bank Mitra", "nomor_rekening": "1234567890", "atas_nama": "PT Mitra Sejahtera",
		}
	}
	opsionalKosong := isi()
	for _, k := range []string{"nomor_pihak_kedua", "tanggal_ttd", "pihak_kedua_telepon", "pihak_kedua_email", "pihak_kedua_pic"} {
		delete(opsionalKosong, k)
	}

	for _, c := range []struct {
		nama string
		doc  *entity.AgreementDocument
		data map[string]interface{}
	}{
		{"Lengkap", lengkap, isi()},
		{"OpsionalKosong", &entity.AgreementDocument{}, opsionalKosong},
	} {
		t.Run(c.nama, func(t *testing.T) {
			lama := placeholderValuesSebelumPR9(c.doc, c.data, snap)
			baru := placeholderValues(c.doc, bagianPKS(t), c.data, snap)
			for k, v := range lama {
				if baru.Nilai[k] != v {
					t.Errorf("%s: %q, sebelumnya %q", k, baru.Nilai[k], v)
				}
			}
		})
	}
}

// Placeholder turunan per tipe field, termasuk isian kosong.
func TestNilaiPerTipe(t *testing.T) {
	for _, c := range []struct {
		nama  string
		tipe  string
		isian interface{}
		harap map[string]string
	}{
		{"Uang", "money", "750000000", map[string]string{"KODE": "Rp 750.000.000", "KODE_TERBILANG": "tujuh ratus lima puluh juta"}},
		{"UangBerformat", "money", "Rp 1.000.000", map[string]string{"KODE": "Rp 1.000.000", "KODE_TERBILANG": "satu juta"}},
		{"PersenBulat", "percent", json.Number("40"), map[string]string{"KODE": "40", "KODE_TERBILANG": "empat puluh"}},
		{"PersenDesimal", "percent", "33.5", map[string]string{"KODE": "33,5", "KODE_TERBILANG": "tiga puluh tiga koma lima"}},
		{"AngkaNol", "number", "0", map[string]string{"KODE": "0", "KODE_TERBILANG": "nol"}},
		{"Tanggal", "date", "2026-08-17", map[string]string{"KODE": "17 Agustus 2026", "KODE_HARI": "Senin", "KODE_TERBILANG": "tujuh belas Agustus dua ribu dua puluh enam"}},
		{"TanggalTidakTerbaca", "date", "17/08/2026", map[string]string{"KODE": "17/08/2026", "KODE_HARI": "-", "KODE_TERBILANG": "-"}},
		{"Teks", "text", "  PT Contoh  ", map[string]string{"KODE": "PT Contoh"}},
		{"Email", "email", "a@contoh.co.id", map[string]string{"KODE": "a@contoh.co.id"}},
		{"UangKosong", "money", "", map[string]string{"KODE": "-", "KODE_TERBILANG": "-"}},
		{"PersenKosong", "percent", nil, map[string]string{"KODE": "-", "KODE_TERBILANG": "-"}},
		{"TanggalKosong", "date", " ", map[string]string{"KODE": "-", "KODE_HARI": "-", "KODE_TERBILANG": "-"}},
		{"TeksKosong", "textarea", "", map[string]string{"KODE": "-"}},
	} {
		t.Run(c.nama, func(t *testing.T) {
			hasil := fieldPlaceholders(AgreementFieldDefinition{Name: "kode", Type: c.tipe}, c.isian)
			if len(hasil) != len(c.harap) {
				t.Errorf("placeholder %v, harus %v", hasil, c.harap)
			}
			for k, v := range c.harap {
				if hasil[k] != v {
					t.Errorf("%s = %q, harus %q", k, hasil[k], v)
				}
			}
		})
	}
}

// Isian bertipe pilihan hanya menerima salah satu pilihannya; kosong boleh kalau
// tidak wajib.
func TestPilihanHarusTersedia(t *testing.T) {
	bagian := []AgreementSectionDefinition{{Title: "Data", Fields: []AgreementFieldDefinition{{Name: "jenis_mitra", Label: "Jenis Mitra", Type: "select", Options: []string{"Badan usaha", "Perorangan"}}}}}
	for isian, boleh := range map[string]bool{"Perorangan": true, "": true, "Koperasi": false} {
		e := validateAgreementForm(bagian, map[string]interface{}{"jenis_mitra": isian})
		if (e == nil) != boleh {
			t.Errorf("isian %q: galat %v, harus diterima %v", isian, e, boleh)
		}
	}
}

// Katalog placeholder dibentuk dari field jenis dokumennya: jenis tanpa field
// ruang lingkup tidak menawarkan {{RUANG_LINGKUP_LIST}}, dan placeholder sistem
// selalu ada.
func TestKatalogDariField(t *testing.T) {
	ada := func(sections []AgreementSectionDefinition) map[string]bool {
		out := map[string]bool{}
		for _, token := range SupportedPlaceholderTokens(sections) {
			out[token] = true
		}
		return out
	}
	pks := ada(bagianPKS(t))
	for _, token := range []string{"{{NOMOR_PIHAK_PERTAMA}}", "{{TEMPAT_TTD}}", "{{NILAI_KONTRAK_TERBILANG}}", "{{SURAT_PENAWARAN_TANGGAL_HARI}}", scopeListPlaceholder} {
		if !pks[token] {
			t.Errorf("katalog PKS tanpa %s", token)
		}
	}
	nda := ada([]AgreementSectionDefinition{{Title: "Pihak Kedua", Fields: []AgreementFieldDefinition{{Name: "pihak_kedua_nama", Label: "Nama", Type: "text"}}}})
	if !nda["{{PIHAK_KEDUA_NAMA}}"] || !nda["{{DAFTAR_LAMPIRAN}}"] || nda["{{NILAI_KONTRAK}}"] || nda[scopeListPlaceholder] {
		t.Errorf("katalog jenis tanpa ruang lingkup: %v", nda)
	}
}
