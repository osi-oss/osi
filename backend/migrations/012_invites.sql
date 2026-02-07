-- +goose Up
CREATE TABLE invites (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    position_id BIGINT NOT NULL REFERENCES positions (id) ON DELETE CASCADE,
    invited_user_id BIGINT REFERENCES users (id) ON DELETE CASCADE, -- NULLABLE if user not registered
    invited_by_user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    invited_email VARCHAR(255) NOT NULL,
    invited_at TIMESTAMPTZ DEFAULT now(),
    accepted_at TIMESTAMPTZ,
    declined_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (
        organization_id,
        position_id,
        invited_email
    )
);

CREATE INDEX idx_invites_invited_user_id ON invites (invited_user_id);

CREATE INDEX idx_invites_invited_email ON invites (invited_email);

CREATE INDEX idx_invites_invited_by_user_id ON invites (invited_by_user_id);

CREATE INDEX idx_invites_status ON invites (status);

CREATE INDEX idx_invites_organization_id ON invites (organization_id);

-- +goose Down
DROP TABLE IF EXISTS invites CASCADE;