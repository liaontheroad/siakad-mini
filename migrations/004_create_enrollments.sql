CREATE TABLE IF NOT EXISTS enrollments (
    id             SERIAL PRIMARY KEY,
    student_id     INTEGER     NOT NULL REFERENCES students(id),
    course_id      INTEGER     NOT NULL REFERENCES courses(id),
    tahun_akademik VARCHAR(20) NOT NULL
                   CHECK (tahun_akademik ~ '^[0-9]{4}/[0-9]{4}-(Ganjil|Genap)$'),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT enrollments_unique UNIQUE (student_id, course_id, tahun_akademik)
);

CREATE INDEX IF NOT EXISTS enrollments_course_year_idx
    ON enrollments (course_id, tahun_akademik);