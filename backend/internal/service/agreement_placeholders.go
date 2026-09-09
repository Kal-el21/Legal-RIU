package service

import (
	"sort"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
)

// placeholderCatalog memberi label dan pengelompokan untuk setiap token.
// Daftar token yang benar-benar didukung tetap diturunkan dari placeholderValues
// agar tidak pernah melenceng dari yang sesungguhnya diisi generator.
var placeholderCatalog = map[string][2]string{
	"NOMOR_PIHAK_PERTAMA":       {"Identitas & Tanda Tangan", "Nomor perjanjian internal Indonesia Re"},
	"NOMOR_PIHAK_KEDUA":         {"Identitas & Tanda Tangan", "Nomor perjanjian dari pihak kedua"},
	"HARI_TTD":                  {"Identitas & Tanda Tangan", "Nama hari penandatanganan, mis. Senin"},
	"TANGGAL_TTD":               {"Identitas & Tanda Tangan", "Angka tanggal penandatanganan, mis. 17"},
	"BULAN_TTD":                 {"Identitas & Tanda Tangan", "Nama bulan penandatanganan, mis. Agustus"},
	"TAHUN_TTD":                 {"Identitas & Tanda Tangan", "Tahun penandatanganan, mis. 2026"},
	"TANGGAL_TTD_LENGKAP":       {"Identitas & Tanda Tangan", "Tanggal lengkap, mis. 17 Agustus 2026"},
	"TEMPAT_TTD":                {"Identitas & Tanda Tangan", "Kota penandatanganan"},
	"PIHAK_PERTAMA_NAMA":        {"Pihak Pertama", "Nama badan hukum, dari Master Pihak Pertama"},
	"PIHAK_PERTAMA_ALAMAT":      {"Pihak Pertama", "Alamat kantor"},
	"PIHAK_PERTAMA_TELEPON":     {"Pihak Pertama", "Nomor telepon"},
	"PIHAK_PERTAMA_EMAIL":       {"Pihak Pertama", "Alamat email"},
	"PIHAK_PERTAMA_PIC":         {"Pihak Pertama", "Nama PIC"},
	"PIHAK_PERTAMA_PEJABAT":     {"Pihak Pertama", "Nama pejabat penanda tangan"},
	"PIHAK_PERTAMA_JABATAN":     {"Pihak Pertama", "Jabatan pejabat penanda tangan"},
	"PIHAK_KEDUA_NAMA":          {"Pihak Kedua", "Nama badan hukum pihak kedua"},
	"PIHAK_KEDUA_BIDANG":        {"Pihak Kedua", "Bidang usaha pihak kedua"},
	"PIHAK_KEDUA_ALAMAT":        {"Pihak Kedua", "Alamat kantor"},
	"PIHAK_KEDUA_TELEPON":       {"Pihak Kedua", "Nomor telepon"},
	"PIHAK_KEDUA_EMAIL":         {"Pihak Kedua", "Alamat email"},
	"PIHAK_KEDUA_PIC":           {"Pihak Kedua", "Nama PIC"},
	"PIHAK_KEDUA_PEJABAT":       {"Pihak Kedua", "Nama pejabat penanda tangan"},
	"PIHAK_KEDUA_JABATAN":       {"Pihak Kedua", "Jabatan pejabat penanda tangan"},
	"JENIS_PEKERJAAN":           {"Ruang Lingkup & Surat Rujukan", "Jenis pekerjaan yang diperjanjikan"},
	"RUANG_LINGKUP":             {"Ruang Lingkup & Surat Rujukan", "Ruang lingkup sebagai satu blok teks"},
	scopeListToken:              {"Ruang Lingkup & Surat Rujukan", "Ruang lingkup sebagai daftar berhuruf a/b/c. Paragrafnya wajib diformat sebagai numbered list Word"},
	"SURAT_PENAWARAN_NOMOR":     {"Ruang Lingkup & Surat Rujukan", "Nomor surat penawaran"},
	"SURAT_PENAWARAN_PERIHAL":   {"Ruang Lingkup & Surat Rujukan", "Perihal surat penawaran"},
	"SURAT_PENAWARAN_TANGGAL":   {"Ruang Lingkup & Surat Rujukan", "Tanggal surat penawaran"},
	"SURAT_PENUNJUKAN_NOMOR":    {"Ruang Lingkup & Surat Rujukan", "Nomor surat penunjukan"},
	"SURAT_PENUNJUKAN_PERIHAL":  {"Ruang Lingkup & Surat Rujukan", "Perihal surat penunjukan"},
	"SURAT_PENUNJUKAN_TANGGAL":  {"Ruang Lingkup & Surat Rujukan", "Tanggal surat penunjukan"},
	"JANGKA_WAKTU_MULAI":        {"Ruang Lingkup & Surat Rujukan", "Tanggal mulai berlaku"},
	"JANGKA_WAKTU_SELESAI":      {"Ruang Lingkup & Surat Rujukan", "Tanggal berakhir"},
	"NILAI_KONTRAK":             {"Nilai Kontrak & Termin", "Nilai kontrak dalam format rupiah"},
	"NILAI_KONTRAK_TERBILANG":   {"Nilai Kontrak & Termin", "Nilai kontrak dalam huruf"},
	"TERMIN_1_PERSEN":           {"Nilai Kontrak & Termin", "Persentase termin pertama"},
	"TERMIN_1_PERSEN_TERBILANG": {"Nilai Kontrak & Termin", "Persentase termin pertama dalam huruf"},
	"TERMIN_1_NILAI":            {"Nilai Kontrak & Termin", "Nilai termin pertama"},
	"TERMIN_1_NILAI_TERBILANG":  {"Nilai Kontrak & Termin", "Nilai termin pertama dalam huruf"},
	"TERMIN_2_PERSEN":           {"Nilai Kontrak & Termin", "Persentase termin kedua"},
	"TERMIN_2_PERSEN_TERBILANG": {"Nilai Kontrak & Termin", "Persentase termin kedua dalam huruf"},
	"TERMIN_2_NILAI":            {"Nilai Kontrak & Termin", "Nilai termin kedua"},
	"TERMIN_2_NILAI_TERBILANG":  {"Nilai Kontrak & Termin", "Nilai termin kedua dalam huruf"},
	"BANK":                      {"Nilai Kontrak & Termin", "Nama bank pihak kedua"},
	"NOMOR_REKENING":            {"Nilai Kontrak & Termin", "Nomor rekening pihak kedua"},
	"ATAS_NAMA":                 {"Nilai Kontrak & Termin", "Nama pemilik rekening"},
	"DAFTAR_LAMPIRAN":           {"Lampiran", "Daftar bernomor otomatis dari file yang dilampirkan"},
}

const scopeListToken = "RUANG_LINGKUP_LIST"

// sampleAgreementValues membentuk nilai contoh untuk seluruh token melalui jalur
// yang sama dengan dokumen sungguhan, sehingga dry-run saat unggah menguji
// generator apa adanya.
func sampleAgreementValues() map[string]string {
	number := "001/CONTOH/RM.01.01/2026"
	doc := &entity.AgreementDocument{
		TicketNumber:    "PK-CONTOH-0001",
		AgreementNumber: &number,
		Attachments: []entity.AgreementAttachment{
			{FileName: "lampiran-contoh-1.pdf"},
			{FileName: "lampiran-contoh-2.pdf"},
		},
	}
	data := map[string]interface{}{
		"nomor_pihak_kedua":        "002/CONTOH/2026",
		"tanggal_ttd":              "2026-08-17",
		"pihak_kedua_nama":         "PT Contoh Mitra Sejahtera",
		"pihak_kedua_bidang":       "jasa konsultasi teknologi informasi",
		"pihak_kedua_alamat":       "Jl. Contoh No. 1, Jakarta Selatan",
		"pihak_kedua_telepon":      "021 1234567",
		"pihak_kedua_email":        "kontak@contoh.co.id",
		"pihak_kedua_pic":          "Rina Contoh",
		"pihak_kedua_pejabat":      "Andi Contoh",
		"pihak_kedua_jabatan":      "Direktur Utama",
		"jenis_pekerjaan":          "pengadaan jasa konsultasi",
		"ruang_lingkup":            "Analisis kebutuhan sistem\nImplementasi dan konfigurasi\nPelatihan pengguna",
		"surat_penawaran_nomor":    "PNW-001/2026",
		"surat_penawaran_perihal":  "Penawaran Jasa Konsultasi",
		"surat_penawaran_tanggal":  "2026-07-15",
		"surat_penunjukan_nomor":   "PNJ-001/2026",
		"surat_penunjukan_perihal": "Penunjukan Penyedia Jasa",
		"surat_penunjukan_tanggal": "2026-07-30",
		"jangka_waktu_mulai":       "2026-09-01",
		"jangka_waktu_selesai":     "2027-08-31",
		"nilai_kontrak":            "20000000",
		"termin_1_persen":          "50",
		"termin_1_nilai":           "10000000",
		"termin_2_persen":          "50",
		"termin_2_nilai":           "10000000",
		"bank":                     "Bank Contoh",
		"nomor_rekening":           "1234567890",
		"atas_nama":                "PT Contoh Mitra Sejahtera",
	}
	snap := map[string]interface{}{
		"name":               "PT Reasuransi Indonesia Utama (Persero)",
		"address":            "Jl. Salemba Raya No. 30, Jakarta Pusat",
		"phone":              "021 3920101",
		"email":              "corsec@indonesiare.co.id",
		"pic":                "Divisi Legal",
		"signatory_name":     "Nama Pejabat Contoh",
		"signatory_position": "Direktur Contoh",
		"signing_place":      "Jakarta",
	}
	return placeholderValues(doc, data, snap)
}

// SupportedPlaceholderTokens mengembalikan seluruh token yang dikenal generator,
// dalam bentuk lengkap dengan kurung kurawal.
func SupportedPlaceholderTokens() []string {
	values := sampleAgreementValues()
	out := make([]string, 0, len(values)+1)
	for key := range values {
		out = append(out, "{{"+key+"}}")
	}
	out = append(out, scopeListPlaceholder)
	sort.Strings(out)
	return out
}

func PlaceholderCatalog() []dto.PlaceholderReference {
	tokens := SupportedPlaceholderTokens()
	out := make([]dto.PlaceholderReference, 0, len(tokens))
	for _, token := range tokens {
		key := token[2 : len(token)-2]
		meta := placeholderCatalog[key]
		out = append(out, dto.PlaceholderReference{Token: token, Group: meta[0], Description: meta[1]})
	}
	return out
}
