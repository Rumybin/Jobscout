CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE applications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source     TEXT NOT NULL,
    external_id TEXT NOT NULL DEFAULT '',
    title      TEXT NOT NULL,
    company_name TEXT NOT NULL,
    category   TEXT NOT NULL DEFAULT '',
    job_type   TEXT NOT NULL DEFAULT '',
    candidate_required_location TEXT NOT NULL DEFAULT '',
    salary_text TEXT NOT NULL DEFAULT '',
    external_url TEXT NOT NULL DEFAULT '',
    publication_date TIMESTAMPTZ,
    description TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'Wishlist',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT unique_user_source_external UNIQUE (user_id, source, external_id)
);

CREATE INDEX idx_applications_user_id ON applications(user_id);
CREATE INDEX idx_applications_status ON applications(status);