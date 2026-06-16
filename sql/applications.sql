-- name: CreateApplication :one
INSERT INTO applications (
    user_id,
    source,
    external_id,
    title,
    company_name,
    category,
    job_type,
    candidate_required_location,
    salary_text,
    external_url,
    publication_date,
    description,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING id, user_id, source, external_id, title, company_name, category,
          job_type, candidate_required_location, salary_text, external_url,
          publication_date, description, status, created_at, updated_at;