package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type StudentRepository interface {
	FindByUserID(ctx context.Context, userID int) (model.Student, error)
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