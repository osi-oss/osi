-- +goose Up
CREATE TABLE departments (
    id BIGSERIAL PRIMARY KEY,
    location_id BIGINT NOT NULL REFERENCES locations (id) ON DELETE CASCADE,
    parent_id BIGINT REFERENCES departments (id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    UNIQUE (location_id, name)
);

-- +goose Down
DROP TABLE departments;