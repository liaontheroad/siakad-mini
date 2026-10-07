package repository

import "errors"

var ErrNotFound = errors.New("data tidak ditemukan")

var ErrDuplicate = errors.New("data sudah ada (duplikat)")