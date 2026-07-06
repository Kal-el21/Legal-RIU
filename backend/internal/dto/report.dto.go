package dto

type ReportChartSeries struct {
	Name   string    `json:"name"`
	Values []float64 `json:"data"`
}

type ReportChartResponse struct {
	Labels []string             `json:"labels"`
	Series []ReportChartSeries  `json:"series"`
}

type ReportFilter struct {
	Feature  string
	GroupBy  string
	DateFrom *string
	DateTo   *string
}
