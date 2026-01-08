-- +goose Up
CREATE TABLE position_permissions (
    position_id BIGINT NOT NULL REFERENCES positions (id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    PRIMARY KEY (position_id, permission_id)
);

-- +goose Down
DROP TABLE position_permissions;