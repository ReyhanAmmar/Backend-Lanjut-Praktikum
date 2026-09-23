CREATE TABLE IF NOT EXISTS roles (
    name        VARCHAR(20) PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
    ('admin', 'Akses penuh terhadap seluruh data student dan pengaturan'),
    ('staff', 'Boleh melihat data seluruh student, tetapi tidak boleh mengubah'),
    ('user',  'Hanya boleh mengelola data student miliknya sendiri')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(50) PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

INSERT INTO permissions (name, description) VALUES
    ('student:list',       'Melihat daftar seluruh student'),
    ('student:read:any',   'Melihat data student mana pun'),
    ('student:create',     'Membuat student baru'),
    ('student:update:any', 'Mengubah data student mana pun'),
    ('student:delete',     'Mengapus student'),
    ('role:assign',        'Mengubah role milik student lain')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20) NOT NULL REFERENCES roles(name)       ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('admin', 'role:assign'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;

UPDATE students SET role = 'user' WHERE role NOT IN (SELECT name FROM roles);

ALTER TABLE students DROP CONSTRAINT IF EXISTS students_role_fkey;

ALTER TABLE students
    ADD CONSTRAINT students_role_fkey
    FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE;

CREATE INDEX IF NOT EXISTS students_role_idx ON students (role);