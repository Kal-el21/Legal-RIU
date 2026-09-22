package dto

type UploadAgreementTemplateRequest struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Note          string `json:"note"`
	EffectiveDate string `json:"effective_date"`
}

// PlaceholderReference adalah daftar token yang boleh dipakai tim legal saat
// menyusun template. Ditampilkan sebagai panel referensi di halaman admin.
type PlaceholderReference struct {
	Token       string `json:"token"`
	Group       string `json:"group"`
	Description string `json:"description"`
}

type AgreementTemplateValidation struct {
	Placeholders        []string `json:"placeholders"`
	UnknownPlaceholders []string `json:"unknown_placeholders"`
	UnusedPlaceholders  []string `json:"unused_placeholders"`
	Warnings            []string `json:"warnings"`
}
