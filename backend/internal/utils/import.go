package utils

import "legal-riu-portal/internal/dto"

func AppendRowError(result *dto.ImportResult, row int, field string, reason string) {
	result.Skipped++
	result.Errors = append(result.Errors, dto.ImportRowError{
		Row:    row,
		Field:  field,
		Reason: reason,
	})
}
