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

	// positionsByField groups every calibrated occurrence of a field. Each field
	// can have multiple occurrences (e.g. a signature block repeated across
	// pages); every occurrence renders with the same value at its own position.
	positionsByField := make(map[string][]FieldPosition)

	if s.fieldPositionRepo != nil {
		dbVersion := version
		if dbPositions, err := s.fieldPositionRepo.GetByVersion(dbVersion); err == nil && len(dbPositions) > 0 {
			for _, pos := range dbPositions {
				positionsByField[pos.FieldName] = append(positionsByField[pos.FieldName], FieldPosition{
					X:     pos.X,
					Y:     pos.Y,
					Font:  pos.Font,
					Style: pos.Style,
					Size:  pos.Size,
					Align: pos.Align,
					Page:  pos.PageNumber,
				})
			}
			log.Printf("Loaded %d field positions from DB for version %s", len(dbPositions), dbVersion)
		} else if err != nil {
			log.Printf("Warning: failed to load field positions for version %s, using defaults: %v", dbVersion, err)
		}
	}

	// Fall back to hardcoded defaults for any field without a calibrated entry.
	for name, pos := range DefaultFieldPositions {
		if _, ok := positionsByField[name]; !ok {
			positionsByField[name] = []FieldPosition{pos}
		}
	}

	type fieldEntry struct {
		Name  string
		Value string
		Pos   FieldPosition
	}

	var entries []fieldEntry
	addField := func(name, value string) {
		if value == "" || value == ".............................." {
			return
		}
		occs, ok := positionsByField[name]
		if !ok || len(occs) == 0 {
			return
		}
		for _, pos := range occs {
			entries = append(entries, fieldEntry{Name: name, Value: value, Pos: pos})
		}
	}

	addField("pihak_kedua_nama", fdString(doc, "pihak_kedua_nama"))
	addField("pihak_kedua_bidang", fdString(doc, "pihak_kedua_bidang"))
	addField("pihak_kedua_alamat", fdString(doc, "pihak_kedua_alamat"))
	addField("pihak_kedua_telepon", fdString(doc, "pihak_kedua_telepon"))
	addField("pihak_kedua_email", fdString(doc, "pihak_kedua_email"))
	addField("pihak_kedua_pic", fdString(doc, "pihak_kedua_pic"))
	addField("jenis_pekerjaan", fdString(doc, "jenis_pekerjaan"))
	addField("nomor_pihak_pertama", fdString(doc, "nomor_pihak_pertama"))
	addField("nomor_pihak_kedua", fdString(doc, "nomor_pihak_kedua"))
	addField("surat_penawaran_nomor", fdString(doc, "surat_penawaran_nomor"))
	addField("surat_penawaran_perihal", fdString(doc, "surat_penawaran_perihal"))
	addField("surat_penawaran_tanggal", fdString(doc, "surat_penawaran_tanggal"))
	addField("surat_penunjukan_nomor", fdString(doc, "surat_penunjukan_nomor"))
	addField("surat_penunjukan_perihal", fdString(doc, "surat_penunjukan_perihal"))
	addField("surat_penunjukan_tanggal", fdString(doc, "surat_penunjukan_tanggal"))
	addField("jangka_waktu_mulai", fdString(doc, "jangka_waktu_mulai"))
	addField("jangka_waktu_selesai", fdString(doc, "jangka_waktu_selesai"))
	addField("tempat_ttd", fdString(doc, "tempat_ttd"))
	addField("tanggal_ttd", formatTanggalID(fdString(doc, "tanggal_ttd")))
	addField("pihak_pertama_pejabat", doc.PihakPertamaPejabat)
	addField("pihak_pertama_jabatan", doc.PihakPertamaJabatan)
	addField("pihak_kedua_pejabat", fdString(doc, "pihak_kedua_pejabat"))
	addField("pihak_kedua_jabatan", fdString(doc, "pihak_kedua_jabatan"))
	addField("ruang_lingkup", fdString(doc, "ruang_lingkup"))
	addField("nilai_kontrak", formatRupiah(parseFloatField(fdString(doc, "nilai_kontrak"))))
	addField("termin1_persen", fdString(doc, "termin1_persen"))
	addField("termin1_nilai", formatRupiah(parseFloatField(fdString(doc, "termin1_nilai"))))
	addField("termin2_persen", fdString(doc, "termin2_persen"))
	addField("termin2_nilai", formatRupiah(parseFloatField(fdString(doc, "termin2_nilai"))))
	addField("bank", fdString(doc, "bank"))
	addField("nomor_rekening", fdString(doc, "nomor_rekening"))
	addField("atas_nama", fdString(doc, "atas_nama"))
	addField("lampiran", fdString(doc, "lampiran"))

	fieldsByPage := make(map[int][]fieldEntry)
	for _, entry := range entries {
		page := entry.Pos.Page
		if page == 0 {
			page = 1
		}
		fieldsByPage[page] = append(fieldsByPage[page], entry)
	}

	for _, pageEntries := range fieldsByPage {
		for _, entry := range pageEntries {
			s.overlayField(pdf, entry.Value, entry.Pos)
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
