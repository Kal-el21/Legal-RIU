package service

import (
	"errors"
	"sort"
	"time"

	"legal-riu-portal/internal/dto"

	"gorm.io/gorm"
)

type ReportService interface {
	GetLegalCaseReport(filter dto.ReportFilter) (*dto.ReportChartResponse, error)
	GetLegalOpinionReport(filter dto.ReportFilter) (*dto.ReportChartResponse, error)
	GetDocumentReviewReport(filter dto.ReportFilter) (*dto.ReportChartResponse, error)
}

type reportService struct {
	db *gorm.DB
}

func NewReportService(db *gorm.DB) ReportService {
	return &reportService{db: db}
}

type reportAggRow struct {
	Label      string
	SeriesName string
	Count      float64
}

func (s *reportService) GetLegalCaseReport(filter dto.ReportFilter) (*dto.ReportChartResponse, error) {
	allowed := map[string]bool{
		"company": true, "case_type": true, "category": true,
		"status": true, "level": true, "location": true, "pic": true,
	}
	if !allowed[filter.GroupBy] {
		return nil, errors.New("group_by tidak valid")
	}

	query := s.db.Table("legal_cases")

	if filter.DateFrom != nil {
		query = query.Where("case_date >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		to := *filter.DateTo
		parsed, _ := time.Parse("2006-01-02", to)
		query = query.Where("case_date < ?", parsed.AddDate(0, 0, 1))
	}

	switch filter.GroupBy {
	case "company":
		query = query.Joins("LEFT JOIN companies ON companies.id = legal_cases.company_id")
		query = query.Select("TO_CHAR(legal_cases.case_date, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(companies.name, 'Tidak Diketahui') as series_name")
	case "case_type":
		query = query.Joins("LEFT JOIN case_types ON case_types.id = legal_cases.case_type_id")
		query = query.Select("TO_CHAR(legal_cases.case_date, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(case_types.label, 'Tidak Diketahui') as series_name")
	case "category":
		query = query.Joins("LEFT JOIN case_categories ON case_categories.id = legal_cases.category_id")
		query = query.Select("TO_CHAR(legal_cases.case_date, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(case_categories.label, 'Tidak Diketahui') as series_name")
	case "status":
		query = query.Select("TO_CHAR(legal_cases.case_date, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(legal_cases.current_status, 'Tidak Diketahui') as series_name")
	case "level":
		query = query.Select("TO_CHAR(legal_cases.case_date, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(legal_cases.level, 'Tidak Diketahui') as series_name")
	case "location":
		query = query.Joins("LEFT JOIN regencies ON regencies.id = legal_cases.location_regency_id")
		query = query.Select("TO_CHAR(legal_cases.case_date, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(regencies.name, 'Tidak Diketahui') as series_name")
	case "pic":
		query = query.Joins("LEFT JOIN divisions ON divisions.id = legal_cases.pic")
		query = query.Select("TO_CHAR(legal_cases.case_date, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(divisions.name, 'Tidak Diketahui') as series_name")
	}

	query = query.Group("label, series_name").Order("label ASC")

	var rows []reportAggRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	return pivotReport(rows)
}

func (s *reportService) GetLegalOpinionReport(filter dto.ReportFilter) (*dto.ReportChartResponse, error) {
	allowed := map[string]bool{
		"company": true, "legal_type": true, "status": true, "division": true,
	}
	if !allowed[filter.GroupBy] {
		return nil, errors.New("group_by tidak valid")
	}

	query := s.db.Table("legal_opinions")

	if filter.DateFrom != nil {
		query = query.Where("legal_opinions.created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		to := *filter.DateTo
		parsed, _ := time.Parse("2006-01-02", to)
		query = query.Where("legal_opinions.created_at < ?", parsed.AddDate(0, 0, 1))
	}

	switch filter.GroupBy {
	case "company":
		query = query.Joins("LEFT JOIN users ON users.id = legal_opinions.user_id")
		query = query.Joins("LEFT JOIN companies ON companies.id = users.company_id")
		query = query.Select("TO_CHAR(legal_opinions.created_at, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(companies.name, 'Tidak Diketahui') as series_name")
	case "legal_type":
		query = query.Select("TO_CHAR(legal_opinions.created_at, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(legal_opinions.legal_type, 'Tidak Diketahui') as series_name")
	case "status":
		query = query.Select("TO_CHAR(legal_opinions.created_at, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(legal_opinions.status, 'Tidak Diketahui') as series_name")
	case "division":
		query = query.Joins("LEFT JOIN users ON users.id = legal_opinions.user_id")
		query = query.Joins("LEFT JOIN divisions ON divisions.id = users.division_id")
		query = query.Select("TO_CHAR(legal_opinions.created_at, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(divisions.name, 'Tidak Diketahui') as series_name")
	}

	query = query.Group("label, series_name").Order("label ASC")

	var rows []reportAggRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	return pivotReport(rows)
}

func (s *reportService) GetDocumentReviewReport(filter dto.ReportFilter) (*dto.ReportChartResponse, error) {
	allowed := map[string]bool{
		"company": true, "document_type": true, "status": true, "division": true,
	}
	if !allowed[filter.GroupBy] {
		return nil, errors.New("group_by tidak valid")
	}

	query := s.db.Table("document_reviews")

	if filter.DateFrom != nil {
		query = query.Where("document_reviews.created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		to := *filter.DateTo
		parsed, _ := time.Parse("2006-01-02", to)
		query = query.Where("document_reviews.created_at < ?", parsed.AddDate(0, 0, 1))
	}

	switch filter.GroupBy {
	case "company":
		query = query.Joins("LEFT JOIN users ON users.id = document_reviews.user_id")
		query = query.Joins("LEFT JOIN companies ON companies.id = users.company_id")
		query = query.Select("TO_CHAR(document_reviews.created_at, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(companies.name, 'Tidak Diketahui') as series_name")
	case "document_type":
		query = query.Select("TO_CHAR(document_reviews.created_at, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(document_reviews.document_type, 'Tidak Diketahui') as series_name")
	case "status":
		query = query.Select("TO_CHAR(document_reviews.created_at, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(document_reviews.status, 'Tidak Diketahui') as series_name")
	case "division":
		query = query.Joins("LEFT JOIN users ON users.id = document_reviews.user_id")
		query = query.Joins("LEFT JOIN divisions ON divisions.id = users.division_id")
		query = query.Select("TO_CHAR(document_reviews.created_at, 'YYYY-MM') as label, COUNT(*) as count, COALESCE(divisions.name, 'Tidak Diketahui') as series_name")
	}

	query = query.Group("label, series_name").Order("label ASC")

	var rows []reportAggRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	return pivotReport(rows)
}

func pivotReport(rows []reportAggRow) (*dto.ReportChartResponse, error) {
	if len(rows) == 0 {
		return &dto.ReportChartResponse{Labels: []string{}, Series: []dto.ReportChartSeries{}}, nil
	}

	labelSet := make(map[string]bool)
	seriesSet := make(map[string]bool)
	for _, row := range rows {
		labelSet[row.Label] = true
		seriesSet[row.SeriesName] = true
	}

	labels := make([]string, 0, len(labelSet))
	for label := range labelSet {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	seriesNames := make([]string, 0, len(seriesSet))
	for name := range seriesSet {
		seriesNames = append(seriesNames, name)
	}
	sort.Strings(seriesNames)

	counts := make(map[string]map[string]float64)
	for _, row := range rows {
		if counts[row.SeriesName] == nil {
			counts[row.SeriesName] = make(map[string]float64)
		}
		counts[row.SeriesName][row.Label] = row.Count
	}

	series := make([]dto.ReportChartSeries, 0, len(seriesNames))
	for _, name := range seriesNames {
		data := make([]float64, len(labels))
		for i, label := range labels {
			if c, ok := counts[name][label]; ok {
				data[i] = c
			}
		}
		series = append(series, dto.ReportChartSeries{
			Name:   name,
			Values: data,
		})
	}

	return &dto.ReportChartResponse{
		Labels: labels,
		Series: series,
	}, nil
}
