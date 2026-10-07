package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type CourseRepository interface {
	FindAll(ctx context.Context, q model.CourseListQuery) ([]model.CourseView, error)
}

type coursePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &coursePostgresRepository{pool: pool}
}

func (r *coursePostgresRepository) FindAll(ctx context.Context, q model.CourseListQuery) ([]model.CourseView, error) {
	whereClauses := []string{"1=1"}
	var args []any
	argIdx := 1

	if q.Semester != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("c.semester = $%d", argIdx))
		args = append(args, *q.Semester)
		argIdx++
	}

	if q.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+q.Search+"%")
		argIdx++
	}

	whereStr := strings.Join(whereClauses, " AND ")

	query := fmt.Sprintf(`
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       COALESCE(en.terisi, 0) AS terisi
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*) AS terisi
			FROM enrollments
			WHERE tahun_akademik = $%d
			GROUP BY course_id
		) en ON en.course_id = c.id
		WHERE %s
	`, argIdx, whereStr)

	args = append(args, q.TahunAkademik)
	argIdx++

	if q.AvailableOnly {
		query += " AND c.kuota > COALESCE(en.terisi, 0)"
	}

	query += " ORDER BY c.kode_mk ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.CourseView
	for rows.Next() {
		var cv model.CourseView
		if err := rows.Scan(&cv.ID, &cv.KodeMK, &cv.NamaMK, &cv.SKS, &cv.Semester, &cv.Kuota, &cv.Terisi); err != nil {
			return nil, err
		}
		cv.SisaKuota = cv.Kuota - cv.Terisi
		if cv.SisaKuota < 0 {
			cv.SisaKuota = 0
		}
		result = append(result, cv)
	}

	if result == nil {
		result = []model.CourseView{}
	}

	return result, nil
}