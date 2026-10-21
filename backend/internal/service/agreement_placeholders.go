package service

import (
	"fmt"
	"maps"
	"sort"
	"strconv"
	"strings"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
)

// Placeholder dokumen perjanjian ada dua macam:
//   - placeholder sistem, diisi dari data di luar isian pengaju (nomor perjanjian,
//     tanggal dan tempat tanda tangan, Pihak Pertama, lampiran). Daftarnya tetap di
//     sini dan tidak bisa dipakai sebagai kode field;
//   - placeholder field, dibentuk dari definisi field versi template: kodenya nama
//     field dalam huruf besar, turunannya mengikuti tipe (fieldPlaceholders).
var systemPlaceholders = map[string][2]string{
	"NOMOR_PIHAK_PERTAMA":   {"Identitas & Tanda Tangan", "Nomor perjanjian internal Indonesia Re"},
	"HARI_TTD":              {"Identitas & Tanda Tangan", "Nama hari penandatanganan, mis. Senin"},
	"TANGGAL_TTD":           {"Identitas & Tanda Tangan", "Angka tanggal penandatanganan, mis. 17"},
	"BULAN_TTD":             {"Identitas & Tanda Tangan", "Nama bulan penandatanganan, mis. Agustus"},
	"TAHUN_TTD":             {"Identitas & Tanda Tangan", "Tahun penandatanganan, mis. 2026"},
	"TANGGAL_TTD_LENGKAP":   {"Identitas & Tanda Tangan", "Tanggal lengkap, mis. 17 Agustus 2026"},
	"TEMPAT_TTD":            {"Identitas & Tanda Tangan", "Kota penandatanganan"},
	"PIHAK_PERTAMA_NAMA":    {"Pihak Pertama", "Nama badan hukum, dari Master Pihak Pertama"},
	"PIHAK_PERTAMA_ALAMAT":  {"Pihak Pertama", "Alamat kantor"},
	"PIHAK_PERTAMA_TELEPON": {"Pihak Pertama", "Nomor telepon"},
	"PIHAK_PERTAMA_EMAIL":   {"Pihak Pertama", "Alamat email"},
	"PIHAK_PERTAMA_PIC":     {"Pihak Pertama", "Nama PIC"},
	"PIHAK_PERTAMA_PEJABAT": {"Pihak Pertama", "Nama pejabat penanda tangan"},
	"PIHAK_PERTAMA_JABATAN": {"Pihak Pertama", "Jabatan pejabat penanda tangan"},
	"DAFTAR_LAMPIRAN":       {"Lampiran", "Daftar bernomor otomatis dari file yang dilampirkan"},
}

// scopeField: field ruang lingkup yang juga mengisi {{RUANG_LINGKUP_LIST}},
// satu paragraf per baris. Ditangani khusus sampai tipe daftar hadir di PR 10.
const scopeField = "ruang_lingkup"

// systemValues mengisi placeholder sistem.
func systemValues(d *entity.AgreementDocument, data, snap map[string]interface{}) map[string]string {
	date := valueString(data["tanggal_ttd"])
	dt, _ := parseDate(date)
	bagian := func(f func() string) string {
		if dt.IsZero() {
			return "-"
		}
		return f()
	}
	atts := make([]string, 0, len(d.Attachments))
	for i, a := range d.Attachments {
		atts = append(atts, fmt.Sprintf("%d. %s", i+1, a.FileName))
	}
	if len(atts) == 0 {
		atts = append(atts, "-")
	}
	return map[string]string{
		"NOMOR_PIHAK_PERTAMA":   strOrDash(d.AgreementNumber),
		"HARI_TTD":              bagian(func() string { return indoDays[dt.Weekday()] }),
		"TANGGAL_TTD":           bagian(func() string { return fmt.Sprint(dt.Day()) }),
		"BULAN_TTD":             bagian(func() string { return indoMonths[dt.Month()] }),
		"TAHUN_TTD":             bagian(func() string { return fmt.Sprint(dt.Year()) }),
		"TANGGAL_TTD_LENGKAP":   indoDate(date),
		"TEMPAT_TTD":            valueString(snap["signing_place"]),
		"PIHAK_PERTAMA_NAMA":    valueString(snap["name"]),
		"PIHAK_PERTAMA_ALAMAT":  valueString(snap["address"]),
		"PIHAK_PERTAMA_TELEPON": valueString(snap["phone"]),
		"PIHAK_PERTAMA_EMAIL":   valueString(snap["email"]),
		"PIHAK_PERTAMA_PIC":     valueString(snap["pic"]),
		"PIHAK_PERTAMA_PEJABAT": valueString(snap["signatory_name"]),
		"PIHAK_PERTAMA_JABATAN": valueString(snap["signatory_position"]),
		"DAFTAR_LAMPIRAN":       strings.Join(atts, "\n"),
	}
}

// fieldPlaceholders membentuk placeholder satu field beserta turunannya:
//
//	money                       KODE (Rp 750.000.000), KODE_TERBILANG
//	number, percent             KODE (33,5), KODE_TERBILANG
//	date                        KODE (17 Agustus 2026), KODE_HARI, KODE_TERBILANG
//	text, textarea, email, tel  KODE
//	group                       KODE_JUMLAH (3 (tiga)); placeholder per baris di paragrafUlang
//	list                        tidak ada; placeholder per poin di paragrafUlang
//
// Isian kosong menjadi "-" di semua turunannya.
func fieldPlaceholders(f AgreementFieldDefinition, raw interface{}) map[string]string {
	key := strings.ToUpper(f.Name)
	switch f.Type {
	case "list":
		return map[string]string{}
	case "group":
		if baris, _ := barisIsian(raw); len(baris) > 0 {
			return map[string]string{key + "_JUMLAH": jumlahTerbilang(len(baris))}
		}
		return map[string]string{key + "_JUMLAH": "-"}
	}
	v := strings.TrimSpace(valueString(raw))
	var out map[string]string
	switch f.Type {
	case "money":
		n := valueInt64(raw)
		out = map[string]string{key: rupiah(n), key + "_TERBILANG": terbilang(n)}
	case "number", "percent":
		n := valueFloat(raw)
		out = map[string]string{key: angka(n), key + "_TERBILANG": terbilangAngka(n)}
	case "date":
		out = map[string]string{key: indoDate(v), key + "_HARI": "-", key + "_TERBILANG": "-"}
		if d, e := parseDate(v); e == nil {
			out[key+"_HARI"] = indoDays[d.Weekday()]
			out[key+"_TERBILANG"] = terbilang(int64(d.Day())) + " " + indoMonths[d.Month()] + " " + terbilang(int64(d.Year()))
		}
	default:
		out = map[string]string{key: v}
	}
	if v == "" {
		for k := range out {
			out[k] = "-"
		}
	}
	return out
}

// paragrafUlang membentuk paragraf berulang field daftar dan grup:
//
//	list   KODE, sekali per poin
//	group  KODE_URUTAN (pertama, kedua, ...) dan placeholder sub-field KODE_SUB, sekali per baris
func paragrafUlang(f AgreementFieldDefinition, raw interface{}) ParagrafUlang {
	key := strings.ToUpper(f.Name)
	baris, _ := barisIsian(raw)
	if f.Type == "list" {
		u := ParagrafUlang{Token: []string{key}}
		for _, b := range baris {
			if v := strings.TrimSpace(valueString(b)); v != "" {
				u.Baris = append(u.Baris, map[string]string{key: v})
			}
		}
		return u
	}
	u := ParagrafUlang{Token: []string{key + "_URUTAN"}}
	for _, sub := range f.Fields {
		for token := range fieldPlaceholders(subGrup(f, sub), nil) {
			u.Token = append(u.Token, token)
		}
	}
	for i, b := range baris {
		isi, _ := b.(map[string]interface{})
		nilai := map[string]string{key + "_URUTAN": urutanTerbilang(i + 1)}
		for _, sub := range f.Fields {
			maps.Copy(nilai, fieldPlaceholders(subGrup(f, sub), isi[sub.Name]))
		}
		u.Baris = append(u.Baris, nilai)
	}
	return u
}

// subGrup: sub-field dengan kode lengkapnya, mis. persen di grup termin menjadi termin_persen.
func subGrup(f, sub AgreementFieldDefinition) AgreementFieldDefinition {
	sub.Name = f.Name + "_" + sub.Name
	return sub
}

func placeholderValues(d *entity.AgreementDocument, sections []AgreementSectionDefinition, data, snap map[string]interface{}) IsianDokumen {
	isi := IsianDokumen{Nilai: systemValues(d, data, snap)}
	for _, s := range sections {
		for _, f := range s.Fields {
			maps.Copy(isi.Nilai, fieldPlaceholders(f, data[f.Name]))
			if f.Type == "list" || f.Type == "group" {
				isi.Ulang = append(isi.Ulang, paragrafUlang(f, data[f.Name]))
			}
		}
	}
	return isi
}

// angka menulis bilangan dengan koma desimal, tanpa nol di belakang: 33,5.
func angka(n float64) string {
	return strings.Replace(strconv.FormatFloat(n, 'f', -1, 64), ".", ",", 1)
}

// terbilangAngka: 33,5 menjadi "tiga puluh tiga koma lima"; angka di belakang
// koma dibaca satu per satu.
func terbilangAngka(n float64) string {
	bulat, pecahan, _ := strings.Cut(strconv.FormatFloat(n, 'f', -1, 64), ".")
	i, _ := strconv.ParseInt(bulat, 10, 64)
	out := terbilang(i)
	if i == 0 {
		out = "nol"
	}
	if pecahan != "" {
		digit := []string{"nol", "satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan"}
		out += " koma"
		for _, c := range pecahan {
			out += " " + digit[c-'0']
		}
	}
	return out
}

// contohNilai: isian contoh per tipe untuk uji coba unggah dan pratinjau template.
func contohNilai(f AgreementFieldDefinition) interface{} {
	switch f.Type {
	case "list":
		return []interface{}{"[" + f.Label + " poin 1]", "[" + f.Label + " poin 2]"}
	case "group":
		baris := make([]interface{}, 2)
		for i := range baris {
			isi := map[string]interface{}{}
			for _, sub := range f.Fields {
				isi[sub.Name] = contohNilai(sub)
			}
			baris[i] = isi
		}
		return baris
	case "money":
		return "20000000"
	case "number":
		return "3"
	case "percent":
		return "50"
	case "date":
		return "2026-08-17"
	case "email":
		return "kontak@contoh.co.id"
	case "tel":
		return "021 1234567"
	case "textarea":
		return "[" + f.Label + " baris 1]\n[" + f.Label + " baris 2]\n[" + f.Label + " baris 3]"
	default:
		return "[" + f.Label + "]"
	}
}

// sampleAgreementValues membentuk nilai contoh untuk seluruh token melalui jalur
// yang sama dengan dokumen sungguhan, sehingga dry-run saat unggah menguji
// generator apa adanya.
func sampleAgreementValues(sections []AgreementSectionDefinition) IsianDokumen {
	number := "001/CONTOH/RM.01.01/2026"
	doc := &entity.AgreementDocument{
		TicketNumber:    "PK-CONTOH-0001",
		AgreementNumber: &number,
		Attachments: []entity.AgreementAttachment{
			{FileName: "lampiran-contoh-1.pdf"},
			{FileName: "lampiran-contoh-2.pdf"},
		},
	}
	data := map[string]interface{}{"tanggal_ttd": "2026-08-17"}
	for _, s := range sections {
		for _, f := range s.Fields {
			data[f.Name] = contohNilai(f)
		}
	}
	// Contoh termin dibagi dari contoh nilai kontrak, jadi angkanya masuk akal.
	for _, s := range sections {
		for _, f := range s.Fields {
			_ = bagiNilaiGrup(f, data)
		}
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
	return placeholderValues(doc, sections, data, snap)
}

// PlaceholderCatalog mengembalikan seluruh token yang dikenal untuk definisi field
// itu, beserta kelompok dan keterangannya. Token field dikelompokkan per bagian.
func PlaceholderCatalog(sections []AgreementSectionDefinition) []dto.PlaceholderReference {
	out := make([]dto.PlaceholderReference, 0, len(systemPlaceholders))
	for key, meta := range systemPlaceholders {
		out = append(out, dto.PlaceholderReference{Token: "{{" + key + "}}", Group: meta[0], Description: meta[1]})
	}
	for _, s := range sections {
		for _, f := range s.Fields {
			for token, desc := range tokenField(f) {
				out = append(out, dto.PlaceholderReference{Token: "{{" + token + "}}", Group: s.Title, Description: desc})
			}
			if f.Name == scopeField && f.Type != "list" {
				out = append(out, dto.PlaceholderReference{Token: scopeListPlaceholder, Group: s.Title, Description: f.Label + " sebagai daftar berhuruf a/b/c. Paragrafnya wajib diformat sebagai numbered list Word"})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Token < out[j].Token })
	return out
}

// tokenField: semua placeholder satu field, termasuk turunan dan placeholder per
// baris, beserta keterangannya untuk katalog.
func tokenField(f AgreementFieldDefinition) map[string]string {
	out := map[string]string{}
	turunan := func(g AgreementFieldDefinition, label string) {
		key := strings.ToUpper(g.Name)
		for token := range fieldPlaceholders(g, contohNilai(g)) {
			desc := label
			switch strings.TrimPrefix(token, key) {
			case "_TERBILANG":
				desc += " dalam huruf"
			case "_HARI":
				desc += " (nama hari)"
			}
			out[token] = desc
		}
	}
	key := strings.ToUpper(f.Name)
	switch f.Type {
	case "list":
		out[key] = f.Label + ", satu paragraf per poin. Paragrafnya wajib diformat sebagai numbered list Word"
	case "group":
		out[key+"_JUMLAH"] = "Jumlah baris " + f.Label + ", mis. 3 (tiga)"
		out[key+"_URUTAN"] = "Urutan baris " + f.Label + ": pertama, kedua, dst. Semua placeholder per baris grup ini ditulis di satu paragraf, yang diulang sekali per baris"
		for _, sub := range f.Fields {
			turunan(subGrup(f, sub), f.Label+": "+sub.Label)
		}
	default:
		turunan(f, f.Label)
	}
	return out
}

// SupportedPlaceholderTokens mengembalikan token yang dikenal untuk definisi field
// itu, lengkap dengan kurung kurawal.
func SupportedPlaceholderTokens(sections []AgreementSectionDefinition) []string {
	catalog := PlaceholderCatalog(sections)
	out := make([]string, len(catalog))
	for i, item := range catalog {
		out[i] = item.Token
	}
	return out
}
