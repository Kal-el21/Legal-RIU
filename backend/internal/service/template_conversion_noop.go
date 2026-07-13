package service

import (
	"context"
	"fmt"
)

type noOpTemplateConversionService struct{}

func (n *noOpTemplateConversionService) ConvertDocxToPDF(ctx context.Context, docxData []byte, version string) ([]byte, error) {
	return nil, fmt.Errorf("template conversion not supported")
}

func (n *noOpTemplateConversionService) EnsureBasePDFExists(ctx context.Context, version string) ([]byte, error) {
	return nil, fmt.Errorf("template conversion not supported")
}

func NewNoOpTemplateConversionService() TemplateConversionService {
	return &noOpTemplateConversionService{}
}
