package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=mocks/mock_repository.go -package=mocks github.com/Fodro/saberchan/internal/database Repository

type Repository interface {
	// InTx runs fn inside a single DB transaction. fn receives a Repository
	// bound to that transaction (nested InTx uses a savepoint when supported).
	InTx(ctx context.Context, fn func(tx Repository) error) error

	//confg
	AddConfig(ctx context.Context, config *Config) error
	ChangeCurrConfig(ctx context.Context, configId uuid.UUID) error
	GetCurrentConfig(ctx context.Context) (*Config, error)
	GetConfigs(ctx context.Context) ([]Config, error)

	//board
	AddBoard(ctx context.Context, board *Board) error
	GetBoardByAlias(ctx context.Context, alias string) (*Board, error)
	GetBoardById(ctx context.Context, id uuid.UUID) (*Board, error)
	GetBoards(ctx context.Context) ([]Board, error)
	DeleteBoard(ctx context.Context, id uuid.UUID) error
	UpdateBoard(ctx context.Context, board *Board) error

	//thread
	AddThread(ctx context.Context, thread *Thread) error
	GetThread(ctx context.Context, id uuid.UUID) (*Thread, error)
	GetThreads(ctx context.Context, boardID uuid.UUID) ([]Thread, error)
	GetBoardCatalog(ctx context.Context, boardID uuid.UUID, limit, offset int) ([]CatalogThread, error)
	CountThreads(ctx context.Context, boardID uuid.UUID) (uint64, error)
	DeleteThread(ctx context.Context, id uuid.UUID) error
	BumpThread(ctx context.Context, id uuid.UUID) error
	CheckIfThreadBelowBumpLimit(ctx context.Context, id uuid.UUID) (bool, error)

	//post
	AddPost(ctx context.Context, post *Post) error
	GetPost(ctx context.Context, id uuid.UUID) (*Post, error)
	GetPosts(ctx context.Context, threadID uuid.UUID) ([]Post, error)
	DeletePost(ctx context.Context, id uuid.UUID) error
	GetOPPost(ctx context.Context, threadID uuid.UUID) (*Post, error)
	GetRepliesForThread(ctx context.Context, threadID uuid.UUID) (uint64, error)

	//attachment
	AddAttachment(ctx context.Context, attachment *Attachment) error
	GetAttachments(ctx context.Context, postID uuid.UUID) ([]Attachment, error)
	GetAttachmentsByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Attachment, error)

	Ping(ctx context.Context) error
}

type (
	Board struct {
		ID          uuid.UUID
		Alias       string
		Name        string
		Description string
		Locked      bool
		Author      string
	}

	Thread struct {
		ID        uuid.UUID
		BoardID   uuid.UUID
		Title     string
		Locked    bool
		UpdatedAt time.Time
	}

	Post struct {
		ID                 uuid.UUID
		Number             uint64
		Text               string
		ThreadID           uuid.UUID
		Sage               bool
		OpMarker           bool
		BrowserFingerprint string
		IP                 string
		HasAttachment      bool
		CreatedAt          time.Time
	}

	Attachment struct {
		ID     uuid.UUID
		PostID uuid.UUID
		Link   string
		Name   string
		Type   string
	}

	Config struct {
		Nickname  string
		BumpLimit uint
		Current   bool
		SiteName  string
		CreatedAt time.Time
	}

	// CatalogThread is a board-catalog row: thread + OP post + reply count (no N+1).
	CatalogThread struct {
		Thread
		OP           Post
		RepliesCount uint64
	}
)
