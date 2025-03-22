package database

import (
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	AddFile(f *File) error
	GetFilesForPost(postID uuid.UUID) ([]*File, error)
	DeleteFilesForPost(postID uuid.UUID) error

	Ping() error
}

type (
	File struct {
		ID uuid.UUID
		PostID uuid.UUID
		Key string
		CreatedAt time.Time
		DeletedAt *time.Time
	}
)