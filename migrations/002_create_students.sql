CREATE TABLE IF NOT EXISTS students (
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER      NOT NULL UNIQUE REFERENCES users(id),
    nim          VARCHAR(12)  NOT NULL UNIQUE CHECK (nim ~ '^[0-9]{12}$'),
    nama         VARCHAR(100) NOT NULL,
    prodi        VARCHAR(100) NOT NULL,
    angkatan     INTEGER      NOT NULL CHECK (angkatan BETWEEN 1900 AND 2100),

    ipk_terakhir NUMERIC(3,2) NOT NULL DEFAULT 0.00
                 CHECK (ipk_terakhir BETWEEN 0.00 AND 4.00),
    deleted_at   TIMESTAMPTZ  NULL 
);

CREATE INDEX IF NOT EXISTS students_prodi_idx
    ON students (prodi) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS students_angkatan_idx
    ON students (angkatan) WHERE deleted_at IS NULL;