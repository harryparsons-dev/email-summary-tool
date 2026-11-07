-- AI summaries. A row with no email_id summarises the whole project; a row
-- with an email_id summarises that one email.
--
-- Summaries are generated in the background: rows start as 'pending' and move
-- to 'completed' (summary filled in) or 'failed' (error filled in).
CREATE TABLE summaries (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    email_id UUID REFERENCES emails (id) ON DELETE CASCADE,

    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    summary TEXT NOT NULL DEFAULT '',
    error TEXT NOT NULL DEFAULT '',

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX summaries_project_id_email_id_created_at_idx
    ON summaries (project_id, email_id, created_at DESC);
