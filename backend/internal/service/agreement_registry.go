package service

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

func validateAgreementForm(sections []AgreementSectionDefinition, data map[string]interface{}) error {
	ada := map[string]bool{}
	for _, s := range sections {
		for _, f := range s.Fields {
			ada[f.Name] = true
			if f.Required && strings.TrimSpace(valueString(data[f.Name])) == "" {
				return errors.New(f.Label + " wajib diisi")
			}
		}
	}
	// Aturan di bawah terikat pada field tertentu, jadi hanya berlaku untuk jenis
	// yang punya field itu (PKS). Aturan termin diganti aturan grup di PR 10.
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
