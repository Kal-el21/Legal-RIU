package service

import (
	"github.com/jung-kurt/gofpdf/v2"
)

// FieldPosition represents the absolute position (in mm) of a form field on the PDF template.
type FieldPosition struct {
	X     float64
	Y     float64
	Font  string
	Style string
	Size  float64
	Align string
	Page  int
}

// DefaultFieldPositions is a rough estimation for the PKS template.
// These coordinates MUST be calibrated against the actual .docx template.
// Origin is top-left corner of the page, unit is mm.
var DefaultFieldPositions = map[string]FieldPosition{
	"pihak_kedua_nama":      {X: 105, Y: 38, Font: "Arial", Style: "B", Size: 11, Align: "C", Page: 1},
	"pihak_kedua_bidang":    {X: 105, Y: 42, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"jenis_pekerjaan":       {X: 105, Y: 52, Font: "Arial", Style: "B", Size: 11, Align: "C", Page: 1},
	"nomor_pihak_pertama":   {X: 105, Y: 62, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"nomor_pihak_kedua":     {X: 105, Y: 68, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"tempat_ttd":            {X: 105, Y: 82, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"tanggal_ttd":           {X: 105, Y: 86, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"pihak_pertama_pejabat": {X: 105, Y: 185, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"pihak_pertama_jabatan": {X: 105, Y: 190, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"pihak_kedua_pejabat":   {X: 105, Y: 155, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"pihak_kedua_jabatan":   {X: 105, Y: 160, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"ruang_lingkup":         {X: 20, Y: 110, Font: "Arial", Style: "", Size: 11, Align: "L", Page: 1},
	"nilai_kontrak":         {X: 105, Y: 125, Font: "Arial", Style: "", Size: 11, Align: "R", Page: 1},
	"pihak_kedua_alamat":    {X: 105, Y: 46, Font: "Arial", Style: "", Size: 11, Align: "L", Page: 1},
	"pihak_kedua_telepon":   {X: 105, Y: 50, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"pihak_kedua_email":     {X: 105, Y: 54, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"pihak_kedua_pic":       {X: 105, Y: 58, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"surat_penawaran_nomor": {X: 105, Y: 70, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"surat_penawaran_perihal": {X: 20, Y: 74, Font: "Arial", Style: "", Size: 11, Align: "L", Page: 1},
	"surat_penawaran_tanggal": {X: 105, Y: 78, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"surat_penunjukan_nomor": {X: 105, Y: 90, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"surat_penunjukan_perihal": {X: 20, Y: 94, Font: "Arial", Style: "", Size: 11, Align: "L", Page: 1},
	"surat_penunjukan_tanggal": {X: 105, Y: 98, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"jangka_waktu_mulai":    {X: 105, Y: 130, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"jangka_waktu_selesai":  {X: 105, Y: 134, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"termin1_persen":        {X: 105, Y: 200, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"termin1_nilai":         {X: 105, Y: 205, Font: "Arial", Style: "", Size: 11, Align: "R", Page: 1},
	"termin2_persen":        {X: 105, Y: 210, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"termin2_nilai":         {X: 105, Y: 215, Font: "Arial", Style: "", Size: 11, Align: "R", Page: 1},
	"bank":                  {X: 105, Y: 220, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"nomor_rekening":        {X: 105, Y: 225, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"atas_nama":             {X: 105, Y: 230, Font: "Arial", Style: "", Size: 11, Align: "C", Page: 1},
	"lampiran":              {X: 20, Y: 240, Font: "Arial", Style: "", Size: 11, Align: "L", Page: 1},
}

func GetFieldPositions(version string) map[string]FieldPosition {
	return DefaultFieldPositions
}

func (s *pdfService) overlayField(pdf *gofpdf.Fpdf, text string, pos FieldPosition) {
	pdf.SetPage(pos.Page)
	pdf.SetFont(pos.Font, pos.Style, pos.Size)
	pdf.SetXY(pos.X, pos.Y)
	switch pos.Align {
	case "C":
		pageWidth := 170.0
		textWidth := pdf.GetStringWidth(text)
		if textWidth < pageWidth {
			pdf.SetX(20 + (pageWidth-textWidth)/2)
		}
		pdf.MultiCell(0, 5.5, text, "", "C", false)
	case "R":
		pdf.MultiCell(0, 5.5, text, "", "R", false)
	default:
		pdf.MultiCell(0, 5.5, text, "", "L", false)
	}
}
