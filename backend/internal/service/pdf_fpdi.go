package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"legal-riu-portal/internal/entity"

	"github.com/jung-kurt/gofpdf/v2"
)

func (s *pdfService) renderAgreementPDF(doc *entity.AgreementDocument, watermark bool, ctx context.Context) ([]byte, error) {
	if doc == nil {
		return nil, errors.New("agreement document is nil")
	}

	version := doc.TemplateVersion
	if version == "" {
		version = "1"
	}

	basePDF, err := s.templateSvc.EnsureBasePDFExists(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get base PDF template: %w", err)
	}

	tmpDir := fmt.Sprintf("%s/templates", os.TempDir())
	os.MkdirAll(tmpDir, 0755)

	pdfPath := filepath.Join(tmpDir, fmt.Sprintf("base-v%s-%s.pdf", version, doc.ID.String()[:8]))
	if err := os.WriteFile(pdfPath, basePDF, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp PDF: %w", err)
	}
	defer os.Remove(pdfPath)

	cmd := exec.Command("pdftoppm", "-png", "-r", "200", pdfPath, filepath.Join(tmpDir, "page"))
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm failed: %s: %s", err.Error(), string(output))
	}

	pattern := filepath.Join(tmpDir, "page-*.png")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("no PDF pages generated")
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i] < matches[j]
	})

	pdf := gofpdf.New("P", "mm", "A4", "")

	for _, imgPath := range matches {
		pdf.AddPage()
		pdf.ImageOptions(imgPath, 0, 0, 210, 297, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		defer os.Remove(imgPath)
	}

	positions := DefaultFieldPositions
	if s.fieldPositionRepo != nil {
		dbVersion := version
		if dbPositions, err := s.fieldPositionRepo.GetByVersion(dbVersion); err == nil && len(dbPositions) > 0 {
			positions = make(map[string]FieldPosition)
			for _, pos := range dbPositions {
				positions[pos.FieldName] = FieldPosition{
					X:     pos.X,
					Y:     pos.Y,
					Font:  pos.Font,
					Style: pos.Style,
					Size:  pos.Size,
					Align: pos.Align,
					Page:  pos.PageNumber,
				}
			}
			log.Printf("Loaded %d field positions from DB for version %s", len(positions), dbVersion)
		} else if err != nil {
			log.Printf("Warning: failed to load field positions for version %s, using defaults: %v", dbVersion, err)
		}
	}

	fields := map[string]string{
		"pihak_kedua_nama":      fdString(doc, "pihak_kedua_nama"),
		"pihak_kedua_bidang":    fdString(doc, "pihak_kedua_bidang"),
		"jenis_pekerjaan":       fdString(doc, "jenis_pekerjaan"),
		"nomor_pihak_pertama":   fdString(doc, "nomor_pihak_pertama"),
		"nomor_pihak_kedua":     fdString(doc, "nomor_pihak_kedua"),
		"tempat_ttd":            fdString(doc, "tempat_ttd"),
		"tanggal_ttd":           formatTanggalID(fdString(doc, "tanggal_ttd")),
		"pihak_pertama_pejabat": doc.PihakPertamaPejabat,
		"pihak_pertama_jabatan": doc.PihakPertamaJabatan,
		"pihak_kedua_pejabat":   fdString(doc, "pihak_kedua_pejabat"),
		"pihak_kedua_jabatan":   fdString(doc, "pihak_kedua_jabatan"),
		"ruang_lingkup":         fdString(doc, "ruang_lingkup"),
		"nilai_kontrak":         formatRupiah(parseFloatField(fdString(doc, "nilai_kontrak"))),
	}

	fieldsByPage := make(map[int]map[string]string)
	for fieldName, value := range fields {
		if value == "" || value == ".............................." {
			continue
		}
		if pos, ok := positions[fieldName]; ok {
			page := pos.Page
			if page == 0 {
				page = 1
			}
			if fieldsByPage[page] == nil {
				fieldsByPage[page] = make(map[string]string)
			}
			fieldsByPage[page][fieldName] = value
		}
	}

	for _, pageFields := range fieldsByPage {
		for fieldName, value := range pageFields {
			if pos, ok := positions[fieldName]; ok {
				s.overlayField(pdf, value, pos)
			}
		}
	}

	if watermark {
		s.addAgreementWatermark(pdf)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func parseFloatField(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
