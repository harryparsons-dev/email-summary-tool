-- Each project's inbound address must be unique so incoming mail maps to
-- exactly one project.
CREATE UNIQUE INDEX projects_email_address_key ON projects (email_address);

CREATE TABLE emails (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,

    -- SMTP envelope: who the sending server says it is from and which of our
    -- addresses it was delivered to.
    envelope_from TEXT NOT NULL DEFAULT '',
    envelope_to TEXT NOT NULL,

    -- Parsed message headers and bodies.
    message_id TEXT NOT NULL DEFAULT '',
    from_address TEXT NOT NULL DEFAULT '',
    from_name TEXT NOT NULL DEFAULT '',
    to_addresses TEXT NOT NULL DEFAULT '',
    subject TEXT NOT NULL DEFAULT '',
    sent_at TIMESTAMPTZ,
    text_body TEXT NOT NULL DEFAULT '',
    html_body TEXT NOT NULL DEFAULT '',
    attachments JSONB NOT NULL DEFAULT '[]',

    -- The original message, kept so it can be re-parsed later.
    raw_message BYTEA NOT NULL,

    received_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX emails_project_id_received_at_idx ON emails (project_id, received_at DESC);
