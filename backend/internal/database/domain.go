package database

import (
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	//confg
	AddConfig(config *Config) error
	ChangeCurrConfig(configId uuid.UUID) error
	GetCurrentConfig() (*Config, error)
	GetConfigs() ([]Config, error)

	//board
	AddBoard(board *Board) error
	GetBoardByAlias(alias string) (*Board, error)
	GetBoards() ([]Board, error)
	DeleteBoard(id uuid.UUID) error
	UpdateBoard(board *Board) error

	//thread
	AddThread(thread *Thread) error
	GetThread(id uuid.UUID) (*Thread, error)
	GetThreads(boardID uuid.UUID) ([]Thread, error)
	DeleteThread(id uuid.UUID) error
	BumpThread(id uuid.UUID) error

	//post
	AddPost(post *Post) error
	GetPost(id uuid.UUID) (*Post, error)
	GetPosts(threadID uuid.UUID) ([]Post, error)
	DeletePost(id uuid.UUID) error
	GetOPPost(threadID uuid.UUID) (*Post, error)
	GetRepliesForThread(threadID uuid.UUID) (uint64, error)

	//attachment
	AddAttachment(attachment *Attachment) error
	GetAttachments(postID uuid.UUID) ([]Attachment, error)

	Ping() error
}

type (
	Board struct {
		ID          uuid.UUID
		Alias       string
		Name        string
		Description string
		Locked      bool
		Author string
	}

	Thread struct {
		ID                 uuid.UUID
		BoardID            uuid.UUID
		Title              string
		Locked             bool
		UpdatedAt          time.Time
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
		HasAttachment 	bool
		CreatedAt time.Time
	}

	Attachment struct {
		ID       uuid.UUID
		PostID   uuid.UUID
		Link    string
		Name   string
		Type  string
	}

	Config struct {
		Nickname  string
		BumpLimit uint
		Current   bool
		SiteName string
		CreatedAt time.Time
	}
)
