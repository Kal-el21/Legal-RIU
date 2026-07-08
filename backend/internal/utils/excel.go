package utils

import (
	"mime/multipart"
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
