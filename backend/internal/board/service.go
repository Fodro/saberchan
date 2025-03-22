package board

import (
	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/google/uuid"
)

type service struct {
	repo database.Repository
	conf *config.Config
}

func (s *service) CreateBoard(board *Board) error {
	boardDB := &database.Board{
		ID:          uuid.New(),
		Alias:       board.Alias,
		Name:        board.Name,
		Description: board.Description,
	}

	err := s.repo.AddBoard(boardDB)

	return err
}

func (s *service) CreatePost(threadID uuid.UUID, post *Post) error {
	ip := post.IP
	//TODO: add ip hashing
	hasAttachment := len(post.Attachments) > 0
	//TODO: add save attachments
	postDB := &database.Post{
		ID:                 uuid.New(),
		Text:               post.Text,
		ThreadID:           threadID,
		Sage:               post.Sage,
		OpMarker:           post.OpMarker,
		BrowserFingerprint: post.BrowserFingerprint,
		IP:                 ip,
		HasAttachment:      hasAttachment,
	}

	err := s.repo.AddPost(postDB)

	if err != nil {
		return err
	}

	if !post.Sage {
		return s.repo.BumpThread(threadID)
	}
	return nil
}

func (s *service) CreateThread(thread *Thread) error {
	threadDB := &database.Thread{
		ID:      uuid.New(),
		BoardID: thread.BoardID,
		Title:   thread.Title,
	}

	err := s.repo.AddThread(threadDB)
	if err != nil {
		return err
	}

	return s.CreatePost(threadDB.ID, thread.OriginalPost)
}

func (s *service) DeleteBoard(id uuid.UUID) error {
	panic("unimplemented")
}

func (s *service) DeletePost(id uuid.UUID) error {
	panic("unimplemented")
}

func (s *service) DeleteThread(id uuid.UUID) error {
	panic("unimplemented")
}

func (s *service) GetBoardWithThreads(alias string) (*Board, error) {
	boardDB, err := s.repo.GetBoardByAlias(alias)
	if err != nil {
		return nil, err
	}
	threadsDB, err := s.repo.GetThreads(boardDB.ID)
	if err != nil {
		return nil, err
	}
	threads := make([]*Thread, 0, len(threadsDB))
	for _, threadDB := range threadsDB {
		opPostDB, err := s.repo.GetOPPost(threadDB.ID)
		if err != nil {
			return nil, err
		}
		threads = append(threads, &Thread{
			ID:      threadDB.ID,
			BoardID: threadDB.BoardID,
			Title:   threadDB.Title,
			Locked:  threadDB.Locked,
			OriginalPost: &Post{
				ID:                 opPostDB.ID,
				Number:             opPostDB.Number,
				Text:               opPostDB.Text,
				ThreadID:           opPostDB.ThreadID,
				Sage:               opPostDB.Sage,
				OpMarker:           opPostDB.OpMarker,
				BrowserFingerprint: opPostDB.BrowserFingerprint,
				IP:                 opPostDB.IP,
				Attachments:        nil, //TODO: add attachments
			},
		})
	}

	return &Board{
		ID:          boardDB.ID,
		Alias:       boardDB.Alias,
		Name:        boardDB.Name,
		Locked:      boardDB.Locked,
		Description: boardDB.Description,
		Threads:     threads,
	}, nil

}

func (s *service) GetBoards() ([]*Board, error) {
	boardsDB, err := s.repo.GetBoards()
	if err != nil {
		return nil, err
	}
	boards := make([]*Board, 0, len(boardsDB))
	for _, boardDB := range boardsDB {
		boards = append(boards, &Board{
			ID:          boardDB.ID,
			Alias:       boardDB.Alias,
			Name:        boardDB.Name,
			Description: boardDB.Description,
			Locked:      boardDB.Locked,
		})
	}
	return boards, nil
}

func (s *service) GetThreadWithPosts(id uuid.UUID) (*Thread, error) {
	threadDB, err := s.repo.GetThread(id)
	if err != nil {
		return nil, err
	}
	postsDB, err := s.repo.GetPosts(id)
	if err != nil {
		return nil, err
	}
	var op *Post
	posts := make([]*Post, 0, len(postsDB))
	for i, postDB := range postsDB {
		if i == 0 {
			op = &Post{
				ID:                 postDB.ID,
				Number:             postDB.Number,
				Text:               postDB.Text,
				ThreadID:           postDB.ThreadID,
				Sage:               postDB.Sage,
				OpMarker:           postDB.OpMarker,
				BrowserFingerprint: postDB.BrowserFingerprint,
				IP:                 postDB.IP,
				Attachments:        nil, //TODO: add attachments
			}
			continue
		}
		posts = append(posts, &Post{
			ID:                 postDB.ID,
			Number:             postDB.Number,
			Text:               postDB.Text,
			ThreadID:           postDB.ThreadID,
			Sage:               postDB.Sage,
			OpMarker:           postDB.OpMarker,
			BrowserFingerprint: postDB.BrowserFingerprint,
			IP:                 postDB.IP,
			Attachments:        nil, //TODO: add attachments
		})
	}
	return &Thread{
		ID:           threadDB.ID,
		BoardID:      threadDB.BoardID,
		Title:        threadDB.Title,
		Locked:       threadDB.Locked,
		UpdatedAt:    threadDB.UpdatedAt.String(),
		OriginalPost: op,
		Posts:        posts,
	}, nil
}

func (s *service) UpdateBoard(board *Board) error {
	panic("unimplemented")
}

func NewService(repo database.Repository, conf *config.Config) Service {
	return &service{repo: repo, conf: conf}
}
