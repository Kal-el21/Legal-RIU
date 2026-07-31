package handler

import (
	"net/http"

	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type RepositoryDocumentHandler struct {
	svc service.RepositoryDocumentService
}

func NewRepositoryDocumentHandler(svc service.RepositoryDocumentService) *RepositoryDocumentHandler {
	return &RepositoryDocumentHandler{svc: svc}
}

func (h *RepositoryDocumentHandler) GetAll(c *gin.Context) {
	featureCode := c.Query("feature_code")
	search := c.Query("search")
	items, err := h.svc.GetAllDocuments(featureCode, search)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}
	utils.OK(c, "Success", items)
}

func (h *RepositoryDocumentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.svc.GetDocumentByID(id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.OK(c, "Success", item)
}

func (h *RepositoryDocumentHandler) Download(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.BadRequest(c, "ID tidak boleh kosong", nil)
		return
	}
	obj, fileName, err := h.svc.DownloadDocument(id)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	defer obj.Close()
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", obj, map[string]string{
		"Content-Disposition": `attachment; filename="` + fileName + `"`,
	})
}