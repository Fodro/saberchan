package board

import (
	"github.com/google/uuid"
)

type Service interface {
	CreateBoard(board *Board) error
	GetBoardWithThreads(alias string) (*Board, error)
	GetBoards() ([]*Board, error)
	DeleteBoard(id uuid.UUID) error
	UpdateBoard(board *Board) error

	CreateThread(thread *Thread) error
	GetThreadWithPosts(id uuid.UUID) (*Thread, error)
	DeleteThread(id uuid.UUID) error

	CreatePost(threadID uuid.UUID, post *Post) error
	DeletePost(id uuid.UUID) error
}

type (
	Board struct {
		ID uuid.UUID `json:"id"`
		Alias string `json:"alias"`
		Name string `json:"name"`
		Description string `json:"description"`
		Locked bool `json:"locked"`
		Threads []*Thread `json:"threads"`
	}

	Thread struct {
		ID uuid.UUID `json:"id"`
		BoardID uuid.UUID `json:"board_id"`
		Title string `json:"title"`
		OriginalPost *Post `json:"original_post"`
		Locked bool `json:"locked"`
		UpdatedAt string `json:"updated_at"`
		Posts []*Post `json:"posts"`
		RepliesCount uint64 `json:"replies_count"`
	}

	Post struct {
		ID uuid.UUID `json:"id"`
		Number uint64 `json:"number"`
		Text string `json:"text"`
		ThreadID uuid.UUID `json:"thread_id"`
		Sage bool `json:"sage"`
		OpMarker bool `json:"op_marker"`
		BrowserFingerprint string `json:"browser_fingerprint"`
		IP string `json:"ip"`
		Attachments []Attachment `json:"attachments"`
	}

	Attachment struct {
		ID uuid.UUID `json:"id"`
		Link string `json:"link"`
		PostID uuid.UUID `json:"post_id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
)
