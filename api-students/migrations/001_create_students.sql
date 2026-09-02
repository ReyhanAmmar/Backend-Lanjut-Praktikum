CREATE TABLE IF NOT EXISTS students (
    id         SERIAL PRIMARY KEY,
    nim        VARCHAR(20)   NOT NULL,
    name       VARCHAR(150)  NOT NULL,
    grade      NUMERIC(5,2)  NOT NULL DEFAULT 0 CHECK (grade >= 0 AND grade <= 100),
    is_active  BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key
    ON students (nim);

CREATE INDEX IF NOT EXISTS students_grade_idx
    ON students (grade);    