-- Migration: Add scoped permissions support
-- Scoped permissions for positions (default for all employees on this position)
-- and individual employee permissions

-- +goose Up

-- Create scope type enum
CREATE TYPE scope_type AS ENUM ('organization', 'location', 'department', 'position');

-- Position permission grants (права должности с scope)
-- Все сотрудники на этой должности получают эти права
CREATE TABLE position_permission_grants (
    id BIGSERIAL PRIMARY KEY,
    position_id BIGINT NOT NULL REFERENCES positions (id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    scope_type scope_type NOT NULL DEFAULT 'organization',
    scope_id BIGINT, -- NULL = whole organization, otherwise location_id or department_id
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (
        position_id,
        permission_id,
        scope_type,
        scope_id
    )
);

-- Employee permission grants (индивидуальные права сотрудника)
-- Права конкретного работника, дополняют права должности
CREATE TABLE employee_permission_grants (
    id BIGSERIAL PRIMARY KEY,
    employee_id BIGINT NOT NULL REFERENCES employees (id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    scope_type scope_type NOT NULL DEFAULT 'organization',
    scope_id BIGINT, -- NULL = whole organization, otherwise location_id or department_id
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (
        employee_id,
        permission_id,
        scope_type,
        scope_id
    )
);

-- Indexes for fast lookups
CREATE INDEX idx_position_perm_grants_position ON position_permission_grants (position_id);

CREATE INDEX idx_position_perm_grants_scope ON position_permission_grants (scope_type, scope_id);

CREATE INDEX idx_employee_perm_grants_employee ON employee_permission_grants (employee_id);

CREATE INDEX idx_employee_perm_grants_scope ON employee_permission_grants (scope_type, scope_id);

-- Add missing permissions that are used in middleware
INSERT INTO
    permissions (code, description, group_name)
VALUES (
        'locations.create',
        'Создание локаций',
        'Locations'
    ),
    (
        'locations.update',
        'Редактирование локаций',
        'Locations'
    ),
    (
        'locations.delete',
        'Удаление локаций',
        'Locations'
    ),
    (
        'departments.update',
        'Редактирование отделов',
        'Departments'
    ),
    (
        'departments.delete',
        'Удаление отделов',
        'Departments'
    ),
    (
        'positions.update',
        'Редактирование должностей',
        'Positions'
    ),
    (
        'positions.delete',
        'Удаление должностей',
        'Positions'
    ),
    (
        'employees.view',
        'Просмотр сотрудников',
        'Employees'
    ),
    (
        'employees.manage',
        'Управление сотрудниками',
        'Employees'
    )
ON CONFLICT (code) DO NOTHING;

-- Comments
COMMENT ON TABLE position_permission_grants IS 'Scoped permission grants for positions. All employees on this position inherit these permissions.';

COMMENT ON TABLE employee_permission_grants IS 'Individual scoped permission grants for employees. Extends position permissions.';

-- +goose Down
DROP TABLE IF EXISTS employee_permission_grants;

DROP TABLE IF EXISTS position_permission_grants;

DROP TYPE IF EXISTS scope_type;