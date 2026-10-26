package dto

import "testing"

func permintaanLegalOpinionSah() CreateLegalOpinionRequest {
	return CreateLegalOpinionRequest{
		RequestorName:     "Nama",
		RequestorPosition: "Jabatan",
		RequestorDivision: "Divisi",
		RequestorEmail:    "nama@example.com",
		RequestorPhone:    "08123456789",
		LegalType:         "Kontrak",
		Title:             "Judul",
		Chronology:        "Kronologi",
		Question:          "Pertanyaan",
	}
}

func TestCreateLegalOpinionValidate(t *testing.T) {
	tests := []struct {
		name  string
		field func(*CreateLegalOpinionRequest) *string
		label string
	}{
		{"nama", func(r *CreateLegalOpinionRequest) *string { return &r.RequestorName }, "Nama Lengkap"},
		{"jabatan", func(r *CreateLegalOpinionRequest) *string { return &r.RequestorPosition }, "Posisi Jabatan"},
		{"divisi", func(r *CreateLegalOpinionRequest) *string { return &r.RequestorDivision }, "Divisi pada RIU"},
		{"email", func(r *CreateLegalOpinionRequest) *string { return &r.RequestorEmail }, "Email Kantor"},
		{"telepon", func(r *CreateLegalOpinionRequest) *string { return &r.RequestorPhone }, "Nomor WhatsApp"},
		{"jenis kajian", func(r *CreateLegalOpinionRequest) *string { return &r.LegalType }, "Jenis Kajian Hukum"},
		{"judul", func(r *CreateLegalOpinionRequest) *string { return &r.Title }, "Judul Permasalahan Hukum"},
		{"kronologi", func(r *CreateLegalOpinionRequest) *string { return &r.Chronology }, "Kronologis Permasalahan"},
		{"pertanyaan", func(r *CreateLegalOpinionRequest) *string { return &r.Question }, "Pertanyaan Permasalahan Hukum"},
	}

	for _, tt := range tests {
		for _, value := range []string{"", "   "} {
			t.Run(tt.name+"/"+value, func(t *testing.T) {
				req := permintaanLegalOpinionSah()
				*tt.field(&req) = value
				if err := req.Validate(); err == nil || err.Error() != tt.label+" wajib diisi" {
					t.Fatalf("Validate() error = %v, harap %q wajib diisi", err, tt.label)
				}
			})
		}
	}

	t.Run("email tidak valid", func(t *testing.T) {
		req := permintaanLegalOpinionSah()
		req.RequestorEmail = "bukan-surel"
		if err := req.Validate(); err == nil || err.Error() != "Email pemohon tidak valid" {
			t.Fatalf("Validate() error = %v", err)
		}
	})
	t.Run("lain-lain tanpa keterangan", func(t *testing.T) {
		req := permintaanLegalOpinionSah()
		req.LegalType = "Lain-Lain"
		if err := req.Validate(); err == nil || err.Error() != "Jenis kajian lainnya wajib diisi" {
			t.Fatalf("Validate() error = %v", err)
		}
		req.LegalTypeOther = "Kajian khusus"
		if err := req.Validate(); err != nil {
			t.Fatalf("Validate() dengan keterangan: %v", err)
		}
	})
	t.Run("opsional kosong dan isian dirapikan", func(t *testing.T) {
		req := permintaanLegalOpinionSah()
		req.LegalTypeOther = "   "
		req.Normalize()
		if err := req.Validate(); err != nil {
			t.Fatalf("Validate() untuk isian opsional kosong: %v", err)
		}
		req.Title = "  Judul  "
		req.Normalize()
		if req.Title != "Judul" {
			t.Errorf("Title = %q, harap tanpa spasi tepi", req.Title)
		}
	})
}

func TestUpdateLegalOpinionValidatePakaiAturanSama(t *testing.T) {
	req := UpdateLegalOpinionRequest(permintaanLegalOpinionSah())
	req.LegalType = "Lain-Lain"
	if err := req.Validate(); err == nil || err.Error() != "Jenis kajian lainnya wajib diisi" {
		t.Fatalf("Validate() error = %v", err)
	}
}
