package dto

type ImportRowError struct {
	Row    int    `json:"row"`
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type ImportResult struct {
	Imported int             `json:"imported"`
	Skipped  int             `json:"skipped"`
	Errors   []ImportRowError `json:"errors"`
}
