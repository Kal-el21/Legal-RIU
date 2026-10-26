package dto

import "testing"

func permintaanDocumentReviewSah() CreateDocumentReviewRequest {
	return CreateDocumentReviewRequest{
		RequestorName:     "Nama",
		RequestorPosition: "Jabatan",
		RequestorDivision: "Divisi",
		RequestorEmail:    "nama@example.com",
		RequestorPhone:    "08123456789",
		DocumentName:      "Dokumen",
		SecondParty:       "Pihak kedua",
		DocumentType:      "Perjanjian",
	}
}

func TestCreateDocumentReviewValidate(t *testing.T) {
	tests := []struct {
		name  string
		field func(*CreateDocumentReviewRequest) *string
		label string
	}{
		{"nama", func(r *CreateDocumentReviewRequest) *string { return &r.RequestorName }, "Nama Lengkap"},
		{"jabatan", func(r *CreateDocumentReviewRequest) *string { return &r.RequestorPosition }, "Posisi Jabatan"},
		{"divisi", func(r *CreateDocumentReviewRequest) *string { return &r.RequestorDivision }, "Divisi pada RIU"},
		{"email", func(r *CreateDocumentReviewRequest) *string { return &r.RequestorEmail }, "Email Kantor"},
		{"telepon", func(r *CreateDocumentReviewRequest) *string { return &r.RequestorPhone }, "Nomor WhatsApp"},
		{"nama dokumen", func(r *CreateDocumentReviewRequest) *string { return &r.DocumentName }, "Nama Dokumen"},
		{"pihak kedua", func(r *CreateDocumentReviewRequest) *string { return &r.SecondParty }, "Pihak Kedua / Vendor / Broker / Ceding"},
		{"jenis dokumen", func(r *CreateDocumentReviewRequest) *string { return &r.DocumentType }, "Jenis Dokumen"},
	}

	for _, tt := range tests {
		for _, value := range []string{"", "   "} {
			t.Run(tt.name+"/"+value, func(t *testing.T) {
				req := permintaanDocumentReviewSah()
				*tt.field(&req) = value
				if err := req.Validate(); err == nil || err.Error() != tt.label+" wajib diisi" {
					t.Fatalf("Validate() error = %v, harap %q wajib diisi", err, tt.label)
				}
			})
		}
	}

	t.Run("email tidak valid", func(t *testing.T) {
		req := permintaanDocumentReviewSah()
		req.RequestorEmail = "bukan-surel"
		if err := req.Validate(); err == nil || err.Error() != "Email pemohon tidak valid" {
			t.Fatalf("Validate() error = %v", err)
		}
	})
	t.Run("lain-lain tanpa keterangan", func(t *testing.T) {
		req := permintaanDocumentReviewSah()
		req.DocumentType = "Lain-Lain"
		if err := req.Validate(); err == nil || err.Error() != "Jenis dokumen lainnya wajib diisi" {
			t.Fatalf("Validate() error = %v", err)
		}
		req.DocumentTypeOther = "Dokumen khusus"
		if err := req.Validate(); err != nil {
			t.Fatalf("Validate() dengan keterangan: %v", err)
		}
	})
	t.Run("opsional kosong dan semua teks dirapikan", func(t *testing.T) {
		req := permintaanDocumentReviewSah()
		req.ThirdParty = "   "
		req.AdditionalNote = "   "
		req.RequestorName = "  Nama  "
		req.Normalize()
		if err := req.Validate(); err != nil {
			t.Fatalf("Validate() untuk isian opsional kosong: %v", err)
		}
		if req.RequestorName != "Nama" || req.ThirdParty != "" || req.AdditionalNote != "" {
			t.Fatalf("isian setelah Normalize(): %+v", req)
		}
	})
}

func TestUpdateDocumentReviewValidatePakaiAturanSama(t *testing.T) {
	req := UpdateDocumentReviewRequest(permintaanDocumentReviewSah())
	req.DocumentType = "Lain-Lain"
	if err := req.Validate(); err == nil || err.Error() != "Jenis dokumen lainnya wajib diisi" {
		t.Fatalf("Validate() error = %v", err)
	}
}
