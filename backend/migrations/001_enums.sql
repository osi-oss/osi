-- +goose Up
CREATE TYPE org_status AS ENUM(
    'draft',
    'pending',
    'approved',
    'rejected'
);

CREATE TYPE member_status AS ENUM(
    'invited',
    'active',
    'blocked'
);

-- +goose Down
DROP TYPE IF EXISTS member_status;

DROP TYPE IF EXISTS org_status


CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    member_id BIGINT NOT NULL REFERENCES organization_members (id) ON DELETE CASCADE,
    position_id BIGINT NOT NULL REFERENCES positions (id),
    is_intern BOOLEAN DEFAULT FALSE,
    start_date DATE,
    end_date DATE
);