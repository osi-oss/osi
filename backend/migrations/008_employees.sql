-- +goose Up
CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    organization_id BIGINT NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    position_id BIGINT NOT NULL REFERENCES positions (id),
    status member_status DEFAULT 'invited',
    is_intern BOOLEAN DEFAULT FALSE,
    joined_at TIMESTAMPTZ,
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down

DROP TABLE employees;