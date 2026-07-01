package service

import (
	"bytes"
	"errors"
	"fmt"

	"legal-riu-portal/internal/entity"

	"github.com/jung-kurt/gofpdf/v2"
)

type PDFService interface {
	GenerateLegalOpinionPDF(lo *entity.LegalOpinion) ([]byte, error)
}

type pdfService struct{}

func NewPDFService() PDFService {
	return &pdfService{}
}

func (s *pdfService) GenerateLegalOpinionPDF(lo *entity.LegalOpinion) ([]byte, error) {
	if lo == nil {
		return nil, errors.New("legal opinion is nil")
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(11, 37, 69) // #0B2545
	pdf.Cell(0, 12, "Legal Opinion Report")
	pdf.Ln(10)

	// Ticket Number
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(50, 8, "Ticket Number:", "0", 0, "", false, 0, "")
	pdf.Cell(0, 8, lo.TicketNumber)
	pdf.Ln(8)

	// Status
	pdf.CellFormat(50, 8, "Status:", "0", 0, "", false, 0, "")
	pdf.Cell(0, 8, string(lo.Status))
	pdf.Ln(12)

	// Section: Requestor Information
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(11, 37, 69)
	pdf.Cell(0, 10, "Informasi Pemohon")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)

	pdf.CellFormat(40, 7, "Nama:", "0", 0, "", false, 0, "")
	pdf.Cell(0, 7, lo.RequestorName)
	pdf.Ln(7)

	pdf.CellFormat(40, 7, "Jabatan:", "0", 0, "", false, 0, "")
	pdf.Cell(0, 7, lo.RequestorPosition)
	pdf.Ln(7)

	pdf.CellFormat(40, 7, "Divisi:", "0", 0, "", false, 0, "")
	pdf.Cell(0, 7, lo.RequestorDivision)
	pdf.Ln(7)

	pdf.CellFormat(40, 7, "Email:", "0", 0, "", false, 0, "")
	pdf.Cell(0, 7, lo.RequestorEmail)
	pdf.Ln(7)

	pdf.CellFormat(40, 7, "WhatsApp:", "0", 0, "", false, 0, "")
	pdf.Cell(0, 7, lo.RequestorPhone)
	pdf.Ln(12)

	// Section: Request Details
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(11, 37, 69)
	pdf.Cell(0, 10, "Detail Permohonan")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)

	pdf.CellFormat(40, 7, "Jenis Kajian:", "0", 0, "", false, 0, "")
	legalType := lo.LegalType
	if legalType == "Lain-Lain" && lo.LegalTypeOther != "" {
		legalType = lo.LegalTypeOther
	}
	pdf.Cell(0, 7, legalType)
	pdf.Ln(7)

	pdf.CellFormat(40, 7, "Judul:", "0", 0, "", false, 0, "")
	pdf.MultiCell(0, 7, lo.Title, "", "", false)
	pdf.Ln(4)

	// Chronology
	pdf.CellFormat(40, 7, "Kronologis:", "0", 0, "", false, 0, "")
	pdf.Ln(7)
	pdf.MultiCell(0, 6, lo.Chronology, "", "", false)
	pdf.Ln(4)

	// Question
	pdf.CellFormat(40, 7, "Pertanyaan:", "0", 0, "", false, 0, "")
	pdf.Ln(7)
	pdf.MultiCell(0, 6, lo.Question, "", "", false)
	pdf.Ln(4)

	// Admin Note (if exists)
	if lo.AdminNote != "" {
		pdf.Ln(4)
		pdf.SetFont("Arial", "B", 12)
		pdf.SetTextColor(11, 37, 69)
		pdf.Cell(0, 10, "Catatan Admin")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(0, 0, 0)
		pdf.MultiCell(0, 6, lo.AdminNote, "", "", false)
		pdf.Ln(4)
	}

	// Attachments (if any)
	if len(lo.Attachments) > 0 {
		pdf.Ln(4)
		pdf.SetFont("Arial", "B", 12)
		pdf.SetTextColor(11, 37, 69)
		pdf.Cell(0, 10, "Dokumen Lampiran")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(0, 0, 0)
		for i, att := range lo.Attachments {
			pdf.CellFormat(10, 6, fmt.Sprintf("%d.", i+1), "0", 0, "", false, 0, "")
			pdf.Cell(0, 6, att.FileName)
			pdf.Ln(6)
		}
	}

	// Results (if any)
	if len(lo.Results) > 0 {
		pdf.Ln(4)
		pdf.SetFont("Arial", "B", 12)
		pdf.SetTextColor(11, 37, 69)
		pdf.Cell(0, 10, "Hasil Kajian")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(0, 0, 0)
		for i, res := range lo.Results {
			pdf.CellFormat(10, 6, fmt.Sprintf("%d.", i+1), "0", 0, "", false, 0, "")
			pdf.Cell(0, 6, res.FileName)
			if res.Notes != "" {
				pdf.Ln(5)
				pdf.MultiCell(0, 5, "Catatan: "+res.Notes, "", "", false)
			}
			pdf.Ln(6)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
