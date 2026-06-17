package applications

// SaveApplicationRequest is the expected JSON body for POST /applications.
//
// For external jobs (e.g. from Remotive), provide source and external_id.
// For manual entries (e.g. from career pages, networking), leave source empty
// or set it to "manual". The server will auto-generate external_id.
type SaveApplicationRequest struct {
	Source                     string  `json:"source"`
	ExternalID                 string  `json:"external_id"`
	Title                      string  `json:"title"`
	CompanyName                string  `json:"company_name"`
	Category                   string  `json:"category,omitempty"`
	JobType                    string  `json:"job_type,omitempty"`
	CandidateRequiredLocation  string  `json:"candidate_required_location,omitempty"`
	SalaryText                 string  `json:"salary_text,omitempty"`
	ExternalURL                string  `json:"external_url,omitempty"`
	PublicationDate            *string `json:"publication_date,omitempty"`
	Description                string  `json:"description,omitempty"`
	Status                     string  `json:"status,omitempty"`
}

// ApplicationResponse is returned by POST /applications.
type ApplicationResponse struct {
	ID                        string  `json:"id"`
	UserID                    string  `json:"user_id"`
	Source                    string  `json:"source"`
	ExternalID                string  `json:"external_id"`
	Title                     string  `json:"title"`
	CompanyName               string  `json:"company_name"`
	Category                  string  `json:"category,omitempty"`
	JobType                   string  `json:"job_type,omitempty"`
	CandidateRequiredLocation string  `json:"candidate_required_location,omitempty"`
	SalaryText                string  `json:"salary_text,omitempty"`
	ExternalURL               string  `json:"external_url,omitempty"`
	PublicationDate           *string `json:"publication_date,omitempty"`
	Description               string  `json:"description,omitempty"`
	Status                    string  `json:"status"`
	CreatedAt                 string  `json:"created_at"`
	UpdatedAt                 string  `json:"updated_at"`
}