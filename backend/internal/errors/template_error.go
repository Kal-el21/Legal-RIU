package errors

import "errors"

var (
	ErrTemplateNotFound     = errors.New("template perjanjian belum diupload")
	ErrTemplateInvalid      = errors.New("file template harus berformat .docx")
	ErrTemplateTooLarge     = errors.New("ukuran file template melebihi 10MB")
	ErrConversionFailed     = errors.New("gagal mengkonversi template .docx ke PDF")
	ErrBasePDFNotFound      = errors.New("base PDF template tidak ditemukan")
	ErrPdftoppmFailed       = errors.New("gagal mengkonversi PDF ke image")
	ErrInvalidVersion       = errors.New("versi template tidak valid")
)
