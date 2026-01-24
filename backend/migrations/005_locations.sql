-- +goose Up

CREATE TABLE locations (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    address TEXT,
    source location_source_type NOT NULL DEFAULT 'manual',
    -- 'registry' | 'manual'
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (organization_id, name)
);

-- +goose Down
DROP TABLE IF EXISTS locations CASCADE;

-- | Сценарий            | source   | is_verified |
-- | ------------------- | -------- | ----------- |
-- | Филиал из ФНС       | registry | true        |
-- | Ручной ввод         | manual   | false       |
-- | Подтверждён админом | manual   | true        |