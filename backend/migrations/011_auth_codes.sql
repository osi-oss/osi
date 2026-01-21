-- +goose Up
CREATE TABLE auth_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    code VARCHAR(4) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    attempts INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_auth_codes_user_id ON auth_codes (user_id);

CREATE INDEX idx_auth_codes_expires_at ON auth_codes (expires_at);

-- +goose Down
DROP TABLE IF EXISTS auth_codes;