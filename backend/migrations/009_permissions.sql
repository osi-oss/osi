-- +goose Up
CREATE TABLE permissions (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE, -- например: members.invite
    description TEXT NOT NULL,
    group_name TEXT NOT NULL -- Members, Schedule, Admin
);

INSERT INTO
    permissions (code, description, group_name)
VALUES (
        'members.invite',
        'Приглашать сотрудников',
        'Members'
    ),
    (
        'members.view',
        'Просматривать сотрудников',
        'Members'
    ),
    (
        'positions.create',
        'Создавать должности',
        'Positions'
    ),
    (
        'departments.create',
        'Создавать отделы',
        'Departments'
    ),
    (
        'schedule.view',
        'Просмотр расписания',
        'Schedule'
    ),
    (
        'schedule.edit',
        'Редактирование расписания',
        'Schedule'
    ),
    (
        'permissions.manage',
        'Управление правами',
        'Admin'
    );

-- +goose Down
DROP TABLE permissions;