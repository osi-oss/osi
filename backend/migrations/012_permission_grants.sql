-- Migration: Add scoped permissions support
-- This allows assigning permissions to specific locations or departments

-- Create scope type enum
CREATE TYPE scope_type AS ENUM ('organization', 'location', 'department');

-- Create permission grants table for scoped permissions
-- This table replaces the many-to-many relationship in member_permissions
-- for cases where you need scoped access
CREATE TABLE permission_grants (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    scope_type scope_type NOT NULL DEFAULT 'organization',
    scope_id BIGINT,  -- NULL = whole organization, otherwise location_id or department_id
    created_at TIMESTAMPTZ DEFAULT NOW(),

-- Ensure unique grants
UNIQUE(user_id, organization_id, permission_id, scope_type, scope_id)
);

-- Index for fast lookups
CREATE INDEX idx_permission_grants_user_org ON permission_grants (user_id, organization_id);

CREATE INDEX idx_permission_grants_scope ON permission_grants (scope_type, scope_id);

-- Add missing permissions that are used in middleware
INSERT INTO
    permissions (code, name, description)
VALUES (
        'locations.create',
        'Create locations',
        'Allows creating new locations in organization'
    ),
    (
        'locations.update',
        'Update locations',
        'Allows updating location information'
    ),
    (
        'locations.delete',
        'Delete locations',
        'Allows deleting locations'
    ),
    (
        'departments.update',
        'Update departments',
        'Allows updating department information'
    ),
    (
        'departments.delete',
        'Delete departments',
        'Allows deleting departments'
    ),
    (
        'positions.update',
        'Update positions',
        'Allows updating position information'
    ),
    (
        'positions.delete',
        'Delete positions',
        'Allows deleting positions'
    )
ON CONFLICT (code) DO NOTHING;

-- Comment explaining the scoped permissions model
COMMENT ON TABLE permission_grants IS 'Scoped permission grants for fine-grained access control. 
scope_type determines the level: organization (full access), location, or department.
scope_id references the specific location_id or department_id (NULL for organization scope).
Examples:
- departments.create with scope_type=location, scope_id=5 -> can create departments only in location 5
- members.invite with scope_type=organization, scope_id=NULL -> can invite to whole org';