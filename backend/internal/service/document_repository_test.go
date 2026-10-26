package service

import (
	"testing"

	"legal-riu-portal/internal/entity"
)

func TestNamaUnduhanRepositori(t *testing.T) {
	nomorLO := "LO-202610-0001"
	nomorRD := "RD-202610-0001"
	nomorPKS := "PKS-202610-0001"
	tests := []struct {
		name string
		doc  entity.RepositoryDocument
		want string
	}{
		{"lampiran LO", entity.RepositoryDocument{SourceType: "legal_opinion_attachments", SourceNumber: &nomorLO, FileName: "Surat Asli.pdf"}, "LO-202610-0001_Lampiran_Surat Asli.pdf"},
		{"hasil LO", entity.RepositoryDocument{SourceType: "legal_opinion_results", SourceNumber: &nomorLO, FileName: "Hasil Kajian.pdf"}, "LO-202610-0001_Hasil_Hasil Kajian.pdf"},
		{"lampiran RD", entity.RepositoryDocument{SourceType: "document_review_attachments", SourceNumber: &nomorRD, FileName: "Dokumen Asli.pdf"}, "RD-202610-0001_Lampiran_Dokumen Asli.pdf"},
		{"hasil RD", entity.RepositoryDocument{SourceType: "document_review_results", SourceNumber: &nomorRD, FileName: "Hasil Review.pdf"}, "RD-202610-0001_Hasil_Hasil Review.pdf"},
		{"lampiran perjanjian", entity.RepositoryDocument{SourceType: "agreement_attachments", SourceNumber: &nomorPKS, FileName: "Lampiran PKS.pdf"}, "PKS-202610-0001_Lampiran_Lampiran PKS.pdf"},
		{"final perjanjian", entity.RepositoryDocument{SourceType: "agreement_documents", FileName: "PKS-202610-0001_PKS_Mitra.pdf"}, "PKS-202610-0001_PKS_Mitra.pdf"},
	}
	svc := &repositoryDocumentService{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := svc.namaUnduhan(&test.doc); got != test.want {
				t.Errorf("nama unduhan: %q, harap %q", got, test.want)
			}
		})
	}
}
