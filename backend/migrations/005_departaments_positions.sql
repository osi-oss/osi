-- +goose Up
CREATE TABLE departments (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    parent_id BIGINT REFERENCES departments (id),
    name VARCHAR(255) NOT NULL,
    description TEXT
);

CREATE TABLE positions (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    department_id BIGINT REFERENCES departments (id),
    name VARCHAR(255) NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    description TEXT
);

-- +goose Down
DROP TABLE positions;
DROP TABLE departments;

