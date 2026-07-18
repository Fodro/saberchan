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
	GetBoardByAlias(ctx context.Context, alias string, includeDeleted bool) (*Board, error)
	GetBoardById(ctx context.Context, id uuid.UUID, includeDeleted bool) (*Board, error)
	GetBoards(ctx context.Context, includeDeleted bool) ([]Board, error)
	DeleteBoard(ctx context.Context, id uuid.UUID) error
	UpdateBoard(ctx context.Context, board *Board) error
	SoftDeleteBoard(ctx context.Context, id uuid.UUID) error
	RestoreBoard(ctx context.Context, id uuid.UUID) error
	ListBoardsDueForPurge(ctx context.Context, before time.Time) ([]Board, error)
	MarkBoardPurged(ctx context.Context, id uuid.UUID) error

	//thread
	AddThread(ctx context.Context, thread *Thread) error
	GetThread(ctx context.Context, id uuid.UUID, includeDeleted bool) (*Thread, error)
	GetThreads(ctx context.Context, boardID uuid.UUID) ([]Thread, error)
	GetBoardCatalog(ctx context.Context, boardID uuid.UUID, limit, offset int, includeDeleted bool) ([]CatalogThread, error)
	CountThreads(ctx context.Context, boardID uuid.UUID, includeDeleted bool) (uint64, error)
	DeleteThread(ctx context.Context, id uuid.UUID) error
	BumpThread(ctx context.Context, id uuid.UUID) error
		CheckIfThreadBelowBumpLimit(ctx context.Context, id uuid.UUID) (bool, error)
		GetFollowThreadInfos(ctx context.Context, ids []uuid.UUID) ([]FollowThreadInfo, error)
		SoftDeleteThread(ctx context.Context, id uuid.UUID) error
		RestoreThread(ctx context.Context, id uuid.UUID) error
		ListThreadsDueForPurge(ctx context.Context, before time.Time) ([]Thread, error)
		MarkThreadPurged(ctx context.Context, id uuid.UUID) error

	//post
	AddPost(ctx context.Context, post *Post) error
	GetPost(ctx context.Context, id uuid.UUID) (*Post, error)
	GetPosts(ctx context.Context, threadID uuid.UUID, includeDeleted bool) ([]Post, error)
	DeletePost(ctx context.Context, id uuid.UUID) error
	GetOPPost(ctx context.Context, threadID uuid.UUID) (*Post, error)
	GetRepliesForThread(ctx context.Context, threadID uuid.UUID, includeDeleted bool) (uint64, error)
	SoftDeletePost(ctx context.Context, id uuid.UUID) error
	RestorePost(ctx context.Context, id uuid.UUID) error
	ListPostsDueForPurge(ctx context.Context, before time.Time) ([]Post, error)
	MarkPostPurged(ctx context.Context, id uuid.UUID) error

	//attachment
	AddAttachment(ctx context.Context, attachment *Attachment) error
	GetAttachments(ctx context.Context, postID uuid.UUID) ([]Attachment, error)
	GetAttachmentsByPostIDs(ctx context.Context, postIDs []uuid.UUID) ([]Attachment, error)
	DeleteAttachmentsByPostID(ctx context.Context, postID uuid.UUID) error

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
		DeletedAt   *time.Time
		PurgedAt    *time.Time
	}

	Thread struct {
		ID        uuid.UUID
		BoardID   uuid.UUID
		Title     string
		Locked    bool
		UpdatedAt time.Time
		DeletedAt *time.Time
		PurgedAt  *time.Time
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
		DeletedAt          *time.Time
		PurgedAt           *time.Time
	}

	Attachment struct {
		ID     uuid.UUID
		PostID uuid.UUID
		Link   string
		Name   string
		Type   string
		Key    string
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

		// FollowThreadInfo is a batch row for follow status.
		FollowThreadInfo struct {
			ID             uuid.UUID
			Title          string
			BoardAlias     string
			RepliesCount   uint64 // posts - 1 (same as catalog)
			BelowBumpLimit bool
		}
	)
