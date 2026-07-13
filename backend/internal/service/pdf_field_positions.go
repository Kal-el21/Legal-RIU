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
