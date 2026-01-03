-- +goose Up
CREATE TABLE member_permissions (
    member_id BIGINT NOT NULL REFERENCES organization_members (id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    PRIMARY KEY (member_id, permission_id)
);

-- +goose Down
DROP TABLE member_permissions;