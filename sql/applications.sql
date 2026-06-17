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

-- name: GetApplicationByID :one
SELECT id, user_id, source, external_id, title, company_name, category,
       job_type, candidate_required_location, salary_text, external_url,
       publication_date, description, status, created_at, updated_at
FROM applications
WHERE id = $1 AND user_id = $2;

-- name: GetApplicationsByUser :many
SELECT id, user_id, source, external_id, title, company_name, category,
       job_type, candidate_required_location, salary_text, external_url,
       publication_date, description, status, created_at, updated_at
FROM applications
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateApplication :one
UPDATE applications
SET
    title = CASE WHEN sqlc.arg(update_title)::boolean THEN sqlc.arg(title)::text ELSE title END,
    company_name = CASE WHEN sqlc.arg(update_company_name)::boolean THEN sqlc.arg(company_name)::text ELSE company_name END,
    category = CASE WHEN sqlc.arg(update_category)::boolean THEN sqlc.arg(category)::text ELSE category END,
    job_type = CASE WHEN sqlc.arg(update_job_type)::boolean THEN sqlc.arg(job_type)::text ELSE job_type END,
    candidate_required_location = CASE WHEN sqlc.arg(update_candidate_required_location)::boolean THEN sqlc.arg(candidate_required_location)::text ELSE candidate_required_location END,
    salary_text = CASE WHEN sqlc.arg(update_salary_text)::boolean THEN sqlc.arg(salary_text)::text ELSE salary_text END,
    external_url = CASE WHEN sqlc.arg(update_external_url)::boolean THEN sqlc.arg(external_url)::text ELSE external_url END,
    publication_date = CASE WHEN sqlc.arg(update_publication_date)::boolean THEN sqlc.arg(publication_date)::timestamptz ELSE publication_date END,
    description = CASE WHEN sqlc.arg(update_description)::boolean THEN sqlc.arg(description)::text ELSE description END,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, source, external_id, title, company_name, category,
          job_type, candidate_required_location, salary_text, external_url,
          publication_date, description, status, created_at, updated_at;

-- name: DeleteApplication :one
DELETE FROM applications
WHERE id = $1 AND user_id = $2
RETURNING id;

-- name: UpdateApplicationStatus :one
UPDATE applications
SET status = $3, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, source, external_id, title, company_name, category,
          job_type, candidate_required_location, salary_text, external_url,
          publication_date, description, status, created_at, updated_at;
