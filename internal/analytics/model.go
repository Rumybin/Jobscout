package analytics

// SummaryResponse is returned by GET /analytics/summary.
type SummaryResponse struct {
	TotalSavedJobs int64            `json:"total_saved_jobs"`
	ByStatus       map[string]int64 `json:"by_status"`
	BySource       map[string]int64 `json:"by_source"`
	ByCategory     map[string]int64 `json:"by_category"`
	SavedPerMonth  []MonthCount     `json:"saved_per_month"`
}

// MonthCount represents how many jobs were saved in one month.
type MonthCount struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}
