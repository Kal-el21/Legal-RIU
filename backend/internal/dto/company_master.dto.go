package dto

type CreateCompanyMasterRequest struct {
	Name              string `json:"name" binding:"required"`
	Address           string `json:"address"`
	NPWP              string `json:"npwp"`
	Phone             string `json:"phone"`
	Email             string `json:"email"`
	DefaultPejabat    string `json:"default_pejabat"`
	DefaultJabatan    string `json:"default_jabatan"`
	DefaultTempatTtd  string `json:"default_tempat_ttd"`
	IsActive          bool   `json:"is_active"`
}
