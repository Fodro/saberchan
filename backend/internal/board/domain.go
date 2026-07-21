package board

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CreateBoard(ctx context.Context, board *BoardInput) error
	GetBoardWithThreads(ctx context.Context, alias string, limit, offset int, includeDeleted bool) (*Board, error)
	GetBoards(ctx context.Context, includeDeleted bool) ([]*Board, error)
	DeleteBoard(ctx context.Context, id uuid.UUID) error
	RestoreBoard(ctx context.Context, id uuid.UUID) error
	UpdateBoard(ctx context.Context, board *Board) error

	CreateThread(ctx context.Context, thread *Thread) (*Thread, error)
	GetThreadWithPosts(ctx context.Context, id uuid.UUID, includeDeleted bool) (*Thread, error)
	DeleteThread(ctx context.Context, id uuid.UUID) error
	RestoreThread(ctx context.Context, id uuid.UUID) error

	CreatePost(ctx context.Context, threadID uuid.UUID, post *Post) error
	DeletePost(ctx context.Context, id uuid.UUID) error
	RestorePost(ctx context.Context, id uuid.UUID) error
	GetBoardMetrics(ctx context.Context, from, to time.Time) ([]*BoardMetrics, error)
}

type (
	Board struct {
		ID           uuid.UUID  `json:"id"`
		Alias        string     `json:"alias"`
		Name         string     `json:"name"`
		Description  string     `json:"description"`
		Locked       bool       `json:"locked"`
		Threads      []*Thread  `json:"threads"`
		Author       string     `json:"author"`
		TotalThreads uint64     `json:"total_threads"`
		Limit        int        `json:"limit"`
		Offset       int        `json:"offset"`
		DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	}

	BoardInput struct {
		Alias       string `json:"alias"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Locked      bool   `json:"locked"`
		Author      string `json:"author"`
	}

	Thread struct {
		ID           uuid.UUID  `json:"id"`
		BoardID      uuid.UUID  `json:"board_id"`
		Title        string     `json:"title"`
		OriginalPost *Post      `json:"original_post"`
		Locked       bool       `json:"locked"`
		UpdatedAt    string     `json:"updated_at"`
		Posts        []*Post    `json:"posts"`
		RepliesCount uint64     `json:"replies_count"`
		IsAdmin      bool       `json:"is_admin,omitempty"`
		DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	}

	Post struct {
		ID                 uuid.UUID    `json:"id"`
		Number             uint64       `json:"number"`
		Text               string       `json:"text"`
		ThreadID           uuid.UUID    `json:"thread_id"`
		Sage               bool         `json:"sage"`
		OpMarker           bool         `json:"op_marker"`
		BrowserFingerprint string       `json:"browser_fingerprint"`
		IP                 string       `json:"ip"`
		CreatedAt          time.Time    `json:"created_at"`
		Attachments        []Attachment `json:"attachments"`
		DeletedAt          *time.Time   `json:"deleted_at,omitempty"`
	}

	Attachment struct {
		ID     uuid.UUID `json:"id"`
		Link   string    `json:"link"`
		PostID uuid.UUID `json:"post_id"`
		Name   string    `json:"name"`
		Type   string    `json:"type"`
		Body   string    `json:"body,omitempty"` // base64 (legacy JSON)
		Data   []byte    `json:"-"`              // raw bytes (multipart)
	}

	// BoardMetrics represents aggregated post metrics for a board.
	BoardMetrics struct {
		BoardID       uuid.UUID  `json:"board_id"`
		BoardAlias    string     `json:"board_alias"`
		PostCount     uint64     `json:"post_count"`
		DeletedCount  uint64     `json:"deleted_count"`
		SageCount     uint64     `json:"sage_count"`
		ThreadCount   uint64     `json:"thread_count"`
	}
)

var (
	ErrBoardLocked    = errors.New("board is locked")
	ErrNotImplemented = errors.New("not implemented")
	ErrNotFound       = errors.New("not found")
	ErrRestoreExpired = errors.New("restore window expired")
	ErrAlreadyDeleted = errors.New("already deleted")
)
