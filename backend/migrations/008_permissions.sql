-- +goose Up
CREATE TABLE permissions (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE, -- например: members.invite
    description TEXT NOT NULL,
    group_name TEXT NOT NULL -- Members, Schedule, Admin
);

-- +goose Down
DROP TABLE permissions;