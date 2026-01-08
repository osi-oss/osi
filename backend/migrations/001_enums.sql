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

CREATE TYPE location_source_type AS ENUM(
    'registry',
    'manual'
);

-- +goose Down
DROP TYPE IF EXISTS location_source_type;

DROP TYPE IF EXISTS member_status;

DROP TYPE IF EXISTS org_status