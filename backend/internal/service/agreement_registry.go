package service

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"legal-riu-portal/internal/assets"
	"legal-riu-portal/internal/entity"
)

type AgreementFieldDefinition struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	// Options: pilihan untuk tipe select; saran isian untuk tipe text.
	Options []string `json:"options,omitempty"`
	// Fields: sub-field tipe group, diisi sekali per baris.
	Fields []AgreementFieldDefinition `json:"fields,omitempty"`
	// SplitOf: kode field uang yang dibagi ke baris grup menurut persennya (termin).
	SplitOf string `json:"split_of,omitempty"`
	// PrimaryValue: field uang yang menjadi nilai kontrak untuk memilih penandatangan
	// Pihak Pertama (PR 11, keputusan P1). Paling banyak satu per jenis dokumen.
	PrimaryValue bool `json:"primary_value,omitempty"`
}
type AgreementSectionDefinition struct {
	Title  string                     `json:"title"`
	Fields []AgreementFieldDefinition `json:"fields"`
}
type AgreementTypeSchema struct {
	Code     string                       `json:"code"`
	Name     string                       `json:"name"`
	Sections []AgreementSectionDefinition `json:"sections"`
}
type AgreementTypeDefinition struct {
	Code     string
	Name     string
	Template []byte
	Schema   AgreementTypeSchema
}
type AgreementRegistry struct {
	types map[string]AgreementTypeDefinition
}

func NewAgreementRegistry() (*AgreementRegistry, error) {
	const base = "templates/pks-template/"
	template, err := assets.Files.ReadFile(base + "Draft-PKS-RIU.docx")
	if err != nil {
		return nil, fmt.Errorf("template PKS: %w", err)
	}
	if _, err = zip.NewReader(bytes.NewReader(template), int64(len(template))); err != nil {
		return nil, fmt.Errorf("template PKS yang di-embed bukan DOCX valid: %w", err)
	}
	raw, err := assets.Files.ReadFile(base + "form-schema.json")
	if err != nil {
		return nil, fmt.Errorf("schema PKS: %w", err)
	}
	var schema AgreementTypeSchema
	if err = json.Unmarshal(raw, &schema); err != nil {
		return nil, err
	}
	if schema.Code == "" || len(schema.Sections) == 0 {
		return nil, errors.New("schema PKS tidak valid")
	}
	d := AgreementTypeDefinition{Code: strings.ToUpper(schema.Code), Name: schema.Name, Template: template, Schema: schema}
	return &AgreementRegistry{types: map[string]AgreementTypeDefinition{d.Code: d}}, nil
}
func (r *AgreementRegistry) Get(code string) (AgreementTypeDefinition, bool) {
	// Kode jenis lama ("PERJANJIAN KERJA SAMA", "LAIN-LAIN", dst.) sudah
	// dipindah ke PKS oleh BackfillAgreementDocumentDocumentTypeCode.
	v, ok := r.types[strings.ToUpper(strings.TrimSpace(code))]
	return v, ok
}

// isianSistemPengaju selalu ada di form setiap jenis dokumen dan tidak dikelola
// Admin. Tempat penandatanganan diusulkan pengaju, boleh diganti approver, dan
// kalau kosong memakai tempat bawaan master Pihak Pertama (partyOneSnapshot).
var isianSistemPengaju = []AgreementFieldDefinition{
	{Name: "tempat_ttd", Label: "Usulan Tempat Penandatanganan", Type: "text"},
}

// templateSections membaca definisi field yang disimpan di versi template.
func templateSections(tpl *entity.AgreementTemplate) ([]AgreementSectionDefinition, error) {
	var sections []AgreementSectionDefinition
	if len(tpl.Fields) == 0 || json.Unmarshal(tpl.Fields, &sections) != nil || len(sections) == 0 {
		return nil, fmt.Errorf("definisi field template %s belum tersedia", tpl.Name)
	}
	return sections, nil
}

// validateAgreementForm memeriksa isian terhadap definisi field. Isian daftar dan
// grup sekaligus dirapikan di data: poin kosong dibuang, dan grup yang membagi
// nilai dihitung ulang (bagiNilaiGrup), jadi persen terakhir dan nilai rupiah dari
// browser tidak dipercaya begitu saja.
func validateAgreementForm(sections []AgreementSectionDefinition, data map[string]interface{}) error {
	ada := map[string]bool{}
	for _, s := range sections {
		for _, f := range s.Fields {
			ada[f.Name] = true
			var e error
			switch f.Type {
			case "list":
				e = periksaDaftar(f, data)
			case "group":
				e = periksaGrup(f, data)
			default:
				e = periksaIsian(f, f.Label, data[f.Name])
			}
			if e != nil {
				return e
			}
		}
	}
	// Aturan di bawah terikat pada field tertentu, jadi hanya berlaku untuk jenis
	// yang punya field itu: versi PKS sebelum PR 10, dengan termin 1 dan 2 (D1).
	if ada["jangka_waktu_mulai"] && ada["jangka_waktu_selesai"] {
		start, end := valueString(data["jangka_waktu_mulai"]), valueString(data["jangka_waktu_selesai"])
		if start != "" && end != "" && end < start {
			return errors.New("tanggal selesai tidak boleh sebelum tanggal mulai")
		}
	}
	if ada["termin_1_persen"] && ada["termin_2_persen"] {
		p1, p2 := valueFloat(data["termin_1_persen"]), valueFloat(data["termin_2_persen"])
		if p1 < 0 || p2 < 0 || p1 > 100 || p2 > 100 || p1+p2 != 100 {
			return errors.New("jumlah persentase termin harus 100")
		}
	}
	if ada["nilai_kontrak"] && ada["termin_1_nilai"] && ada["termin_2_nilai"] {
		c, t1, t2 := valueInt64(data["nilai_kontrak"]), valueInt64(data["termin_1_nilai"]), valueInt64(data["termin_2_nilai"])
		if c < 0 || t1 < 0 || t2 < 0 || t1+t2 != c {
			return errors.New("jumlah nilai termin harus sama dengan nilai kontrak")
		}
	}
	return nil
}

func periksaIsian(f AgreementFieldDefinition, label string, raw interface{}) error {
	v := strings.TrimSpace(valueString(raw))
	if f.Required && v == "" {
		return errors.New(label + " wajib diisi")
	}
	if f.Type == "select" && v != "" && !slices.Contains(f.Options, v) {
		return errors.New(label + " harus salah satu pilihan yang tersedia")
	}
	return nil
}

// barisIsian membaca isian daftar atau grup (array JSON); isian kosong menjadi
// daftar kosong.
func barisIsian(raw interface{}) ([]interface{}, bool) {
	if raw == nil {
		return nil, true
	}
	v, ok := raw.([]interface{})
	return v, ok
}

func periksaDaftar(f AgreementFieldDefinition, data map[string]interface{}) error {
	baris, ok := barisIsian(data[f.Name])
	if !ok {
		return errors.New(f.Label + " harus berupa daftar poin")
	}
	poin := []interface{}{}
	for _, b := range baris {
		if v := strings.TrimSpace(valueString(b)); v != "" {
			poin = append(poin, v)
		}
	}
	if f.Required && len(poin) == 0 {
		return errors.New(f.Label + " wajib diisi minimal satu poin")
	}
	if _, ada := data[f.Name]; ada {
		data[f.Name] = poin
	}
	return nil
}

func periksaGrup(f AgreementFieldDefinition, data map[string]interface{}) error {
	baris, ok := barisIsian(data[f.Name])
	if !ok {
		return errors.New(f.Label + " tidak valid")
	}
	for i, b := range baris {
		if _, ok := b.(map[string]interface{}); !ok {
			return fmt.Errorf("%s baris %d tidak valid", f.Label, i+1)
		}
	}
	if f.Required && len(baris) == 0 {
		return errors.New(f.Label + " wajib diisi minimal satu baris")
	}
	if e := bagiNilaiGrup(f, data); e != nil {
		return e
	}
	for i, b := range baris {
		for _, sub := range f.Fields {
			if e := periksaIsian(sub, fmt.Sprintf("%s %d: %s", f.Label, i+1, sub.Label), b.(map[string]interface{})[sub.Name]); e != nil {
				return e
			}
		}
	}
	return nil
}

// subBagi: kode sub-field persen dan sub-field uang pada grup yang membagi nilai.
func subBagi(f AgreementFieldDefinition) (persen, nilai string) {
	for _, sub := range f.Fields {
		switch {
		case sub.Type == "percent" && persen == "":
			persen = sub.Name
		case sub.Type == "money" && nilai == "":
			nilai = sub.Name
		}
	}
	return persen, nilai
}

// bagiNilaiGrup mengisi ulang persen dan nilai rupiah setiap baris grup yang
// membagi nilai, dengan bagiTermin (keputusan K5 dan D5). Baris grup harus sudah
// diperiksa berbentuk objek.
func bagiNilaiGrup(f AgreementFieldDefinition, data map[string]interface{}) error {
	if f.SplitOf == "" {
		return nil
	}
	baris, _ := barisIsian(data[f.Name])
	persen, nilai := subBagi(f)
	p := make([]float64, len(baris))
	for i, b := range baris {
		p[i] = valueFloat(b.(map[string]interface{})[persen])
	}
	hasil, e := bagiTermin(valueInt64(data[f.SplitOf]), p)
	if e != nil {
		return fmt.Errorf("%s: %w", f.Label, e)
	}
	for i, b := range baris {
		b.(map[string]interface{})[persen] = hasil[i].Persen
		b.(map[string]interface{})[nilai] = hasil[i].Nilai
	}
	return nil
}

// ── Rincian perubahan untuk audit log (PR Audit Log tahap L.3b, keputusan L8) ──

// rincianIsian membandingkan isian pengajuan lama dan baru menurut definisi template:
// "Label: lama → baru" per isian yang berubah. Isian daftar dan grup (ruang lingkup, termin)
// cukup "Label diubah", supaya keterangan tetap terbaca.
func rincianIsian(sections []AgreementSectionDefinition, lama, baru map[string]interface{}) []string {
	var hasil []string
	for _, s := range sections {
		for _, f := range s.Fields {
			label := f.Label
			if label == "" {
				label = f.Name
			}
			if f.Type == "list" || f.Type == "group" {
				if bentukDaftar(lama[f.Name]) != bentukDaftar(baru[f.Name]) {
					hasil = append(hasil, label+" diubah")
				}
				continue
			}
			if r, ok := rincianNilai(label, f.Type, lama[f.Name], baru[f.Name]); ok {
				hasil = append(hasil, r)
			}
		}
	}
	return hasil
}

// rincianNilai: "Label: lama → baru" kalau tampilan nilainya berbeda.
func rincianNilai(label, tipe string, lama, baru interface{}) (string, bool) {
	l, b := tampilIsian(tipe, lama), tampilIsian(tipe, baru)
	return label + ": " + l + " → " + b, l != b
}

// tampilIsian: nilai isian seperti yang dibaca orang, dipotong 100 karakter.
func tampilIsian(tipe string, v interface{}) string {
	s := strings.TrimSpace(valueString(v))
	if s == "" {
		return "(kosong)"
	}
	switch tipe {
	case "money":
		s = rupiah(valueInt64(v))
	case "date":
		s = indoDate(s)
	}
	if r := []rune(s); len(r) > 100 {
		s = string(r[:100]) + "…"
	}
	return s
}

// bentukDaftar: bentuk pembanding isian daftar dan grup; kosong, null, dan [] dianggap sama.
func bentukDaftar(v interface{}) string {
	b, _ := json.Marshal(v)
	if s := string(b); s != "null" && s != "[]" {
		return s
	}
	return ""
}

func valueString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return fmt.Sprint(x)
	}
}
func valueFloat(v interface{}) float64 {
	raw := strings.TrimSpace(valueString(v))
	raw = strings.NewReplacer("Rp", "", "rp", "", "%", "", " ", "").Replace(raw)
	if strings.Contains(raw, ",") {
		raw = strings.ReplaceAll(raw, ".", "")
		raw = strings.Replace(raw, ",", ".", 1)
	} else if strings.Count(raw, ".") > 1 {
		raw = strings.ReplaceAll(raw, ".", "")
	}
	n, _ := strconv.ParseFloat(raw, 64)
	return n
}

func valueInt64(v interface{}) int64 {
	raw := strings.TrimSpace(valueString(v))
	raw = strings.NewReplacer("Rp", "", "rp", "", " ", "", ".", "").Replace(raw)
	if comma := strings.Index(raw, ","); comma >= 0 {
		raw = raw[:comma]
	}
	n, _ := strconv.ParseInt(raw, 10, 64)
	if n != 0 || raw == "0" {
		return n
	}
	return int64(valueFloat(v))
}
