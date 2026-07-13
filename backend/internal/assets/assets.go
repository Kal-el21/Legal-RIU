package assets

import (
	"context"
	"embed"
	"log"

	"legal-riu-portal/internal/storage"
)

//go:embed templates/pks-template/Draft-PKS-RIU.docx
var defaultTemplate embed.FS

// EnsureDefaultTemplate uploads the embedded default PKS template to MinIO
// if it does not already exist. It returns nil even when the template is
// already present or when MinIO is temporarily unavailable (the error is
// only logged) so that the application can still start up.
func EnsureDefaultTemplate(store *storage.MinIOClient) error {
	ctx := context.Background()
	templatePath := "templates/pks-template/v1.docx"

	if store.GetFileContentType(ctx, templatePath) != "" {
		log.Println("Default template already exists in MinIO")
		return nil
	}

	data, err := defaultTemplate.ReadFile("templates/pks-template/Draft-PKS-RIU.docx")
	if err != nil {
		return err
	}

	if _, err := store.UploadTemplate(ctx, "1", data); err != nil {
		return err
	}

	log.Println("Default template uploaded to MinIO")
	return nil
}
