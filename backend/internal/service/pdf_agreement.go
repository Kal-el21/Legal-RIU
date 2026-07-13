package service

import (
	"context"
	"errors"

	"legal-riu-portal/internal/entity"

	"github.com/jung-kurt/gofpdf/v2"
)

var errNilAgreement = errors.New("agreement document is nil")

func (s *pdfService) GenerateAgreementPreview(ctx context.Context, doc *entity.AgreementDocument) ([]byte, error) {
	if doc == nil {
		return nil, errNilAgreement
	}
	return s.renderAgreementPDF(doc, true, ctx)
}

func (s *pdfService) GenerateFinalAgreementPDF(ctx context.Context, doc *entity.AgreementDocument) ([]byte, error) {
	if doc == nil {
		return nil, errNilAgreement
	}
	return s.renderAgreementPDF(doc, false, ctx)
}

func (s *pdfService) addAgreementWatermark(pdf *gofpdf.Fpdf) {
	pdf.SetFont("Arial", "B", 48)
	pdf.SetTextColor(210, 210, 210)
	pages := pdf.PageNo()
	for i := 1; i <= pages; i++ {
		pdf.SetPage(i)
		pdf.TransformBegin()
		pdf.TransformRotate(45, 105, 148)
		pdf.SetXY(20, 130)
		pdf.Cell(170, 12, "DRAFT - FOR APPROVAL ONLY")
		pdf.TransformEnd()
	}
}
