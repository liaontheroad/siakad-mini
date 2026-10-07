package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"siakad-mini/config"
	"siakad-mini/database"
)

const bcryptCost = 12

type studentSeed struct {
	nama     string
	prodi    string
	angkatan int
	ipk      float64
}

type courseSeed struct {
	kode     string
	nama     string
	sks      int
	semester int
	kuota    int
}

var students = []studentSeed{
	{"Rina Putri", "Sistem Informasi", 2022, 3.45},
	{"Budi Santoso", "Teknik Informatika", 2022, 3.80},
	{"Citra Dewi", "Sains Data", 2023, 3.12},
	{"Dimas Pratama", "Teknik Informatika", 2021, 3.00},
	{"Eka Wulandari", "Sistem Informasi", 2023, 3.67},
	{"Fajar Nugroho", "Sains Data", 2022, 3.25},
	{"Gita Lestari", "Teknik Informatika", 2024, 3.90},
	{"Hendra Wijaya", "Sistem Informasi", 2021, 3.05},
	{"Intan Permata", "Teknik Informatika", 2023, 2.99},
	{"Joko Susilo", "Sains Data", 2022, 2.75},
	{"Kartika Sari", "Sistem Informasi", 2024, 2.60},
	{"Lutfi Hakim", "Teknik Informatika", 2021, 2.50},
	{"Maya Anggraini", "Sains Data", 2023, 2.88},
	{"Nanda Firmansyah", "Sistem Informasi", 2022, 2.55},
	{"Oki Ramadhan", "Teknik Informatika", 2024, 2.49},
	{"Putri Amelia", "Sains Data", 2021, 2.20},
	{"Qori Rahman", "Sistem Informasi", 2023, 1.95},
	{"Rizky Maulana", "Teknik Informatika", 2022, 2.10},
	{"Sinta Bella", "Sains Data", 2024, 1.80},
	{"Tegar Saputra", "Sistem Informasi", 2021, 0.00},
}

var courses = []courseSeed{
	{"TI101", "Algoritma dan Pemrograman", 3, 1, 30},
	{"TI102", "Matematika Diskrit", 3, 1, 30},
	{"TI103", "Basis Data", 3, 3, 25},
	{"TI104", "Struktur Data", 3, 2, 25},
	{"TI105", "Jaringan Komputer", 3, 3, 25},
	{"TI106", "Pemrograman Web", 4, 4, 20},
	{"TI107", "Rekayasa Perangkat Lunak", 3, 4, 20},
	{"TI108", "Sistem Operasi", 3, 3, 20},
	{"TI109", "Kecerdasan Buatan", 4, 5, 3},
	{"TI110", "Etika Profesi", 2, 6, 2},
}

func main() {
	config.LoadEnv()
	ctx := context.Background()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("memulai transaction: %v", err)
	}
	defer tx.Rollback(ctx) 

	if err := seedAdmin(ctx, tx); err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	if err := seedStudents(ctx, tx); err != nil {
		log.Fatalf("seed mahasiswa: %v", err)
	}
	if err := seedCourses(ctx, tx); err != nil {
		log.Fatalf("seed mata kuliah: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit: %v", err)
	}
	log.Println("seeder selesai")
}

func ensureUser(ctx context.Context, tx pgx.Tx, email, password, role string) (int, bool, error) {
	var id int
	err := tx.QueryRow(ctx,
		`SELECT id FROM users WHERE LOWER(email) = LOWER($1)`, email,
	).Scan(&id)
	if err == nil {
		return id, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, false, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return 0, false, fmt.Errorf("hash password: %w", err)
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id`,
		email, string(hash), role,
	).Scan(&id)
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func seedAdmin(ctx context.Context, tx pgx.Tx) error {
	email := config.GetEnv("SEED_ADMIN_EMAIL", "admin@siakad.test")
	password := config.GetEnv("SEED_ADMIN_PASSWORD", "Admin12345!")

	_, created, err := ensureUser(ctx, tx, email, password, "admin")
	if err != nil {
		return err
	}
	if created {
		log.Printf("admin dibuat: %s", email)
	} else {
		log.Printf("admin sudah ada: %s (dilewati)", email)
	}
	return nil
}

func seedStudents(ctx context.Context, tx pgx.Tx) error {
	inserted := 0
	for i, s := range students {
		nim := fmt.Sprintf("1872%02d%06d", s.angkatan%100, i+1)
		email := nim + "@student.siakad.test"

		userID, _, err := ensureUser(ctx, tx, email, nim, "mahasiswa")
		if err != nil {
			return fmt.Errorf("akun %s: %w", nim, err)
		}

		tag, err := tx.Exec(ctx,
			`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT DO NOTHING`,
			userID, nim, s.nama, s.prodi, s.angkatan, s.ipk,
		)
		if err != nil {
			return fmt.Errorf("mahasiswa %s: %w", nim, err)
		}
		inserted += int(tag.RowsAffected())
	}
	log.Printf("mahasiswa: %d baru dari %d", inserted, len(students))
	return nil
}

func seedCourses(ctx context.Context, tx pgx.Tx) error {
	inserted := 0
	for _, c := range courses {
		tag, err := tx.Exec(ctx,
			`INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (kode_mk) DO NOTHING`,
			c.kode, c.nama, c.sks, c.semester, c.kuota,
		)
		if err != nil {
			return fmt.Errorf("mata kuliah %s: %w", c.kode, err)
		}
		inserted += int(tag.RowsAffected())
	}
	log.Printf("mata kuliah: %d baru dari %d", inserted, len(courses))
	return nil
}