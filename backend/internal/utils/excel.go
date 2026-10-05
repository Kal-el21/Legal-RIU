package utils

import (
	"bytes"
	"encoding/csv"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

func NormalizeHeaders(header []string) []string {
	out := make([]string, len(header))
	for i, h := range header {
		out[i] = strings.ToLower(strings.TrimSpace(h))
	}
	return out
}

func IndexOfHeader(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	return -1
}

func IsEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func ReadSheet(file *multipart.FileHeader, sheetIndex int) ([][]string, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	if strings.EqualFold(filepath.Ext(file.Filename), ".csv") {
		return readCSV(src)
	}

	wb, err := excelize.OpenReader(src)
	if err != nil {
		return nil, err
	}
	defer wb.Close()

	sheetName := wb.GetSheetName(sheetIndex)
	rows, err := wb.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// readCSV membaca CSV hasil "Save As" Excel: pemisah bisa koma atau titik koma
// (regional Indonesia), dan bisa diawali BOM UTF-8.
// Catatan: isi file dianggap UTF-8; CSV non-UTF-8 (ANSI) dengan huruf beraksen
// akan terbaca salah, tambahkan konversi encoding kalau itu jadi kebutuhan.
func readCSV(r io.Reader) ([][]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	data = bytes.TrimPrefix(data, []byte("\xEF\xBB\xBF"))

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	header, _, _ := bytes.Cut(data, []byte("\n"))
	if bytes.Count(header, []byte(";")) > bytes.Count(header, []byte(",")) {
		reader.Comma = ';'
	}
	return reader.ReadAll()
}

func GenerateTemplate(wb *excelize.File, sheetName string, headers []string, examples [][]string) {
	wb.SetSheetName("Sheet1", sheetName)

	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		wb.SetCellValue(sheetName, cell, h)
	}

	for rowIdx, example := range examples {
		for colIdx, value := range example {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			wb.SetCellValue(sheetName, cell, value)
		}
	}
}

func CellValue(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}
