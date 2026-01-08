-- +goose Up
CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    member_id BIGINT NOT NULL REFERENCES organization_members (id) ON DELETE CASCADE,
    position_id BIGINT NOT NULL REFERENCES positions (id),
    is_intern BOOLEAN DEFAULT FALSE,
    start_date DATE,
    end_date DATE
);

-- +goose Down

DROP TABLE employees;