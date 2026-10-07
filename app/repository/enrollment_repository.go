package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

var (
	ErrAlreadyEnrolled = errors.New("mata kuliah sudah diambil")
	ErrQuotaFull       = errors.New("kuota mata kuliah penuh")
	ErrCourseNotFound  = errors.New("mata kuliah tidak ditemukan")
)

type SKSExceededError struct {
	Batas    int
	Terpakai int
	Sisa     int
	SKSBaru  int
	IPK      float64
}

func (e *SKSExceededError) Error() string {
	return "total SKS melebihi batas"
}

type CreateEnrollmentInput struct {
	StudentID     int
	CourseID      int
	TahunAkademik string
	MaxSKS        int
}

type EnrollmentRepository interface {
	CreateWithRules(ctx context.Context, in CreateEnrollmentInput) (model.Enrollment, error)
	FindByID(ctx context.Context, id int) (model.Enrollment, error)
	Delete(ctx context.Context, id int) error
}

type enrollmentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentPostgresRepository{pool: pool}
}

func (r *enrollmentPostgresRepository) CreateWithRules(ctx context.Context, in CreateEnrollmentInput) (model.Enrollment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Enrollment{}, err
	}
	defer tx.Rollback(ctx)

	var studentIPK float64
	var dummyDeleted *time.Time
	err = tx.QueryRow(ctx, `SELECT ipk_terakhir, deleted_at FROM students WHERE id = $1 FOR UPDATE`, in.StudentID).Scan(&studentIPK, &dummyDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, err
	}

	var courseSKS, courseKuota int
	err = tx.QueryRow(ctx, `SELECT sks, kuota FROM courses WHERE id = $1 FOR UPDATE`, in.CourseID).Scan(&courseSKS, &courseKuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrCourseNotFound
		}
		return model.Enrollment{}, err
	}

	var exists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM enrollments 
			WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3
		)
	`, in.StudentID, in.CourseID, in.TahunAkademik).Scan(&exists)
	if err != nil {
		return model.Enrollment{}, err
	}
	if exists {
		return model.Enrollment{}, ErrAlreadyEnrolled
	}

	var terisi int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM enrollments 
		WHERE course_id = $1 AND tahun_akademik = $2
	`, in.CourseID, in.TahunAkademik).Scan(&terisi)
	if err != nil {
		return model.Enrollment{}, err
	}
	if terisi >= courseKuota {
		return model.Enrollment{}, ErrQuotaFull
	}

	var totalTerpakai int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(c.sks), 0)
		FROM enrollments e
		JOIN courses c ON c.id = e.course_id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`, in.StudentID, in.TahunAkademik).Scan(&totalTerpakai)
	if err != nil {
		return model.Enrollment{}, err
	}

	sisa := in.MaxSKS - totalTerpakai
	if totalTerpakai+courseSKS > in.MaxSKS {
		return model.Enrollment{}, &SKSExceededError{
			Batas:    in.MaxSKS,
			Terpakai: totalTerpakai,
			Sisa:     sisa,
			SKSBaru:  courseSKS,
			IPK:      studentIPK,
		}
	}

	var en model.Enrollment
	insertQuery := `
		INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		VALUES ($1, $2, $3)
		RETURNING id, student_id, course_id, tahun_akademik, created_at
	`
	err = tx.QueryRow(ctx, insertQuery, in.StudentID, in.CourseID, in.TahunAkademik).Scan(
		&en.ID, &en.StudentID, &en.CourseID, &en.TahunAkademik, &en.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Enrollment{}, ErrAlreadyEnrolled
		}
		return model.Enrollment{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Enrollment{}, err
	}

	return en, nil
}

func (r *enrollmentPostgresRepository) FindByID(ctx context.Context, id int) (model.Enrollment, error) {
	query := `
		SELECT e.id, e.student_id, e.course_id, e.tahun_akademik, e.created_at
		FROM enrollments e
		WHERE e.id = $1
	`
	var en model.Enrollment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&en.ID, &en.StudentID, &en.CourseID, &en.TahunAkademik, &en.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, err
	}
	return en, nil
}

func (r *enrollmentPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM enrollments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}