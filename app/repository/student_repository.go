package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type DuplicateFieldError struct {
	Field string
}

func (e *DuplicateFieldError) Error() string {
	return fmt.Sprintf("field %s sudah terdaftar", e.Field)
}

type StudentRepository interface {
	FindByUserID(ctx context.Context, userID int) (model.Student, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindAll(ctx context.Context, q model.StudentListQuery) ([]model.Student, int, error)
	Create(ctx context.Context, user model.User, student model.Student) (model.Student, error)
	Update(ctx context.Context, student model.Student) (model.Student, error)
	SoftDelete(ctx context.Context, id int) error
	ListEnrollmentsForStudent(ctx context.Context, studentID int, tahunAkademik string) ([]model.EnrolledCourse, error)
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

func (r *studentPostgresRepository) FindByUserID(ctx context.Context, userID int) (model.Student, error) {
	query := `
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
		FROM students
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	var s model.Student
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, err
	}
	return s, nil
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	query := `
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
		FROM students
		WHERE id = $1 AND deleted_at IS NULL
	`
	var s model.Student
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, err
	}
	return s, nil
}

func (r *studentPostgresRepository) FindAll(ctx context.Context, q model.StudentListQuery) ([]model.Student, int, error) {
	whereClauses := []string{"deleted_at IS NULL"}
	var args []any
	argIdx := 1

	if q.Prodi != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(prodi) = LOWER($%d)", argIdx))
		args = append(args, q.Prodi)
		argIdx++
	}
	if q.Angkatan != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("angkatan = $%d", argIdx))
		args = append(args, *q.Angkatan)
		argIdx++
	}
	if q.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(nim ILIKE $%d OR nama ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+q.Search+"%")
		argIdx++
	}

	whereStr := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM students WHERE %s", whereStr)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	orderClause := "id ASC"
	switch q.Sort {
	case "nama":
		orderClause = "nama ASC"
	case "-ipk_terakhir":
		orderClause = "ipk_terakhir DESC"
	}

	offset := (q.Page - 1) * q.PerPage
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
		FROM students
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereStr, orderClause, argIdx, argIdx+1)

	dataArgs := append(args, q.PerPage, offset)
	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt); err != nil {
			return nil, 0, err
		}
		students = append(students, s)
	}

	if students == nil {
		students = []model.Student{}
	}

	return students, total, nil
}

func (r *studentPostgresRepository) Create(ctx context.Context, user model.User, student model.Student) (model.Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Student{}, err
	}
	defer tx.Rollback(ctx)

	var userID int
	userQuery := `
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	err = tx.QueryRow(ctx, userQuery, user.Email, user.Password, user.Role).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "users_email_lower_key") || strings.Contains(pgErr.ConstraintName, "users_email_key") {
				return model.Student{}, &DuplicateFieldError{Field: "email"}
			}
		}
		return model.Student{}, err
	}

	studentQuery := `
		INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
	`
	var s model.Student
	err = tx.QueryRow(ctx, studentQuery, userID, student.NIM, student.Nama, student.Prodi, student.Angkatan, student.IPKTerakhir).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "students_nim_key") {
				return model.Student{}, &DuplicateFieldError{Field: "nim"}
			}
		}
		return model.Student{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Student{}, err
	}

	return s, nil
}

func (r *studentPostgresRepository) Update(ctx context.Context, student model.Student) (model.Student, error) {
	query := `
		UPDATE students
		SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at
	`
	var s model.Student
	err := r.pool.QueryRow(ctx, query, student.Nama, student.Prodi, student.Angkatan, student.IPKTerakhir, student.ID).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, err
	}
	return s, nil
}

func (r *studentPostgresRepository) SoftDelete(ctx context.Context, id int) error {
	query := `
		UPDATE students
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *studentPostgresRepository) ListEnrollmentsForStudent(ctx context.Context, studentID int, tahunAkademik string) ([]model.EnrolledCourse, error) {
	query := `
		SELECT e.id, c.id, c.kode_mk, c.nama_mk, c.sks, e.tahun_akademik
		FROM enrollments e
		JOIN courses c ON c.id = e.course_id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`
	rows, err := r.pool.Query(ctx, query, studentID, tahunAkademik)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.EnrolledCourse
	for rows.Next() {
		var ec model.EnrolledCourse
		if err := rows.Scan(&ec.EnrollmentID, &ec.CourseID, &ec.KodeMK, &ec.NamaMK, &ec.SKS, &ec.TahunAkademik); err != nil {
			return nil, err
		}
		results = append(results, ec)
	}

	if results == nil {
		results = []model.EnrolledCourse{}
	}

	return results, nil
}