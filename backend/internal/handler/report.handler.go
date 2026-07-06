package handler

import (
	"errors"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/service"
	"legal-riu-portal/internal/utils"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	svc service.ReportService
}

func NewReportHandler(svc service.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) GetLegalCaseReport(c *gin.Context) {
	filter, err := parseReportFilter(c, "legal_case")
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	result, err := h.svc.GetLegalCaseReport(filter)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Success", result)
}

func (h *ReportHandler) GetLegalOpinionReport(c *gin.Context) {
	filter, err := parseReportFilter(c, "legal_opinion")
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	result, err := h.svc.GetLegalOpinionReport(filter)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Success", result)
}

func (h *ReportHandler) GetDocumentReviewReport(c *gin.Context) {
	filter, err := parseReportFilter(c, "document_review")
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	result, err := h.svc.GetDocumentReviewReport(filter)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.OK(c, "Success", result)
}

func parseReportFilter(c *gin.Context, feature string) (dto.ReportFilter, error) {
	groupBy := c.DefaultQuery("group_by", "")
	if groupBy == "" {
		return dto.ReportFilter{}, errors.New("group_by wajib diisi")
	}

	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	var dateFromPtr, dateToPtr *string
	if dateFrom != "" {
		if _, err := time.Parse("2006-01-02", dateFrom); err != nil {
			return dto.ReportFilter{}, errors.New("format date_from tidak valid (YYYY-MM-DD)")
		}
		dateFromPtr = &dateFrom
	}
	if dateTo != "" {
		if _, err := time.Parse("2006-01-02", dateTo); err != nil {
			return dto.ReportFilter{}, errors.New("format date_to tidak valid (YYYY-MM-DD)")
		}
		dateToPtr = &dateTo
	}

	filter := dto.ReportFilter{
		Feature:  feature,
		GroupBy:  groupBy,
		DateFrom: dateFromPtr,
		DateTo:   dateToPtr,
	}

	return filter, nil
}
