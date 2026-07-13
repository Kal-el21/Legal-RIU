package service

import (
	"fmt"
	"strings"
	"time"

	"legal-riu-portal/internal/entity"
)

// agreementTokens are placeholders injected into the verbatim PKS template and
// replaced with data from the AgreementDocument before rendering.
const (
	tokPihakKeduaNama    = "{{PIHAK_KEDUA_NAMA}}"
	tokJenisPekerjaan    = "{{JENIS_PEKERJAAN}}"
	tokNomorPP           = "{{NOMOR_PP}}"
	tokNomorPK           = "{{NOMOR_PK}}"
	tokHari              = "{{HARI}}"
	tokTanggal           = "{{TANGGAL}}"
	tokBulan             = "{{BULAN}}"
	tokTahun             = "{{TAHUN}}"
	tokTempat            = "{{TEMPAT}}"
	tokPPPejabat         = "{{PP_PEJABAT}}"
	tokPPJabatan         = "{{PP_JABATAN}}"
	tokPKBidang          = "{{PK_BIDANG}}"
	tokPenawaranNo       = "{{PENAWARAN_NO}}"
	tokPenawaranPerihal  = "{{PENAWARAN_PERIHAL}}"
	tokPenawaranTgl      = "{{PENAWARAN_TGL}}"
	tokPenunjukanNo      = "{{PENUNJUKAN_NO}}"
	tokPenunjukanPerihal = "{{PENUNJUKAN_PERIHAL}}"
	tokPenunjukanTgl     = "{{PENUNJUKAN_TGL}}"
	tokRuangLingkup      = "{{RUANG_LINGKUP}}"
	tokJangkaMulai       = "{{JANGKA_MULAI}}"
	tokJangkaSelesai     = "{{JANGKA_SELESAI}}"
	tokNilaiKontrak      = "{{NILAI_KONTRAK}}"
	tokNilaiTerbilang    = "{{NILAI_TERBILANG}}"
	tokTermin1Persen     = "{{TERMIN1_PERSEN}}"
	tokTermin1Nilai      = "{{TERMIN1_NILAI}}"
	tokTermin1Terbilang  = "{{TERMIN1_TERBILANG}}"
	tokTermin2Persen     = "{{TERMIN2_PERSEN}}"
	tokTermin2Nilai      = "{{TERMIN2_NILAI}}"
	tokTermin2Terbilang  = "{{TERMIN2_TERBILANG}}"
	tokBank              = "{{BANK}}"
	tokNoRekening        = "{{NO_REKENING}}"
	tokAtasNama          = "{{ATAS_NAMA}}"
	tokPKAlamat          = "{{PK_ALAMAT}}"
	tokPKTelepon         = "{{PK_TELEPON}}"
	tokPKEmail           = "{{PK_EMAIL}}"
	tokPKPic             = "{{PK_PIC}}"
	tokPKPejabat         = "{{PK_PEJABAT}}"
	tokPKJabatan         = "{{PK_JABATAN}}"
)

func fdString(doc *entity.AgreementDocument, key string) string {
	m, err := unmarshalFormData(doc.FormData)
	if err != nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return formatNumber(val)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (s *pdfService) buildAgreementText(doc *entity.AgreementDocument) string {
	pp := doc.PihakPertama
	ppPejabat := doc.PihakPertamaPejabat
	ppJabatan := doc.PihakPertamaJabatan
	if ppPejabat == "" {
		ppPejabat = pp.DefaultPejabat
	}
	if ppJabatan == "" {
		ppJabatan = pp.DefaultJabatan
	}

	formData, err := unmarshalFormData(doc.FormData)
	if err != nil {
		formData = map[string]interface{}{}
	}

	nilai := parseFloat(formData["nilai_kontrak"])
	t1 := parseFloat(formData["termin1_nilai"])
	t2 := parseFloat(formData["termin2_nilai"])
	t1p := parseFloat(formData["termin1_persen"])
	t2p := parseFloat(formData["termin2_persen"])

	if v, ok := formData["tanggal_ttd"].(string); ok && v != "" {
		formData["tanggal_ttd"] = formatTanggalID(v)
	}

	text := agreementTemplate
	repl := map[string]string{
		tokPihakKeduaNama:    fdString(doc, "pihak_kedua_nama"),
		tokJenisPekerjaan:    fdString(doc, "jenis_pekerjaan"),
		tokNomorPP:           fdString(doc, "nomor_pihak_pertama"),
		tokNomorPK:           fdString(doc, "nomor_pihak_kedua"),
		tokHari:              hariFromDate(fdString(doc, "tanggal_ttd")),
		tokTanggal:           fdString(doc, "tanggal_ttd"),
		tokBulan:             bulanFromDate(fdString(doc, "tanggal_ttd")),
		tokTahun:             tahunFromDate(fdString(doc, "tanggal_ttd")),
		tokTempat:            fdString(doc, "tempat_ttd"),
		tokPPPejabat:         ppPejabat,
		tokPPJabatan:         ppJabatan,
		tokPKBidang:          fdString(doc, "pihak_kedua_bidang"),
		tokPenawaranNo:       fdString(doc, "surat_penawaran_nomor"),
		tokPenawaranPerihal:  fdString(doc, "surat_penawaran_perihal"),
		tokPenawaranTgl:      fdString(doc, "surat_penawaran_tanggal"),
		tokPenunjukanNo:      fdString(doc, "surat_penunjukan_nomor"),
		tokPenunjukanPerihal: fdString(doc, "surat_penunjukan_perihal"),
		tokPenunjukanTgl:     fdString(doc, "surat_penunjukan_tanggal"),
		tokRuangLingkup:      fdString(doc, "ruang_lingkup"),
		tokJangkaMulai:       fdString(doc, "jangka_waktu_mulai"),
		tokJangkaSelesai:     fdString(doc, "jangka_waktu_selesai"),
		tokNilaiKontrak:      formatRupiah(nilai),
		tokNilaiTerbilang:    terbilang(nilai) + " rupiah",
		tokTermin1Persen:     formatNumber(t1p),
		tokTermin1Nilai:      formatRupiah(t1),
		tokTermin1Terbilang:  terbilang(t1) + " rupiah",
		tokTermin2Persen:     formatNumber(t2p),
		tokTermin2Nilai:      formatRupiah(t2),
		tokTermin2Terbilang:  terbilang(t2) + " rupiah",
		tokBank:              fdString(doc, "bank"),
		tokNoRekening:        fdString(doc, "nomor_rekening"),
		tokAtasNama:          fdString(doc, "atas_nama"),
		tokPKAlamat:          fdString(doc, "pihak_kedua_alamat"),
		tokPKTelepon:         fdString(doc, "pihak_kedua_telepon"),
		tokPKEmail:           fdString(doc, "pihak_kedua_email"),
		tokPKPic:             fdString(doc, "pihak_kedua_pic"),
		tokPKPejabat:         fdString(doc, "pihak_kedua_pejabat"),
		tokPKJabatan:         fdString(doc, "pihak_kedua_jabatan"),
	}

	for token, value := range repl {
		if value == "" {
			value = ".............................."
		}
		text = strings.ReplaceAll(text, token, value)
	}
	return text
}

func parseFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

func hariFromDate(s string) string {
	t, err := parseFlexDate(s)
	if err != nil {
		return "............"
	}
	days := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	return days[int(t.Weekday())]
}

func bulanFromDate(s string) string {
	t, err := parseFlexDate(s)
	if err != nil {
		return "............"
	}
	months := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	return months[int(t.Month())-1]
}

func tahunFromDate(s string) string {
	t, err := parseFlexDate(s)
	if err != nil {
		return "......"
	}
	return fmt.Sprintf("%d", t.Year())
}

func parseFlexDate(s string) (time.Time, error) {
	layouts := []string{"2006-01-02", "02/01/2006", "02-01-2006", "2006/01/02"}
	var lastErr error
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, lastErr
}

func formatTanggalID(s string) string {
	t, err := parseFlexDate(s)
	if err != nil {
		return s
	}
	return fmt.Sprintf("%02d/%02d/%d", t.Day(), t.Month(), t.Year())
}
