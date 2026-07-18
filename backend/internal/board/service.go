package board

import (
	"context"
	"encoding/base64"
	"errors"
	"log"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/Fodro/saberchan/internal/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type service struct {
	repo database.Repository
	file file.Service
	conf *config.Config
}

func (s *service) CreateBoard(ctx context.Context, board *BoardInput) error {
	boardDB := &database.Board{
		ID:          uuid.New(),
		Alias:       board.Alias,
		Name:        board.Name,
		Description: board.Description,
		Author:      board.Author,
		Locked:      board.Locked,
	}

	return s.repo.AddBoard(ctx, boardDB)
}

func (s *service) cleanupUploads(ctx context.Context, keys []string) {
	for _, key := range keys {
		if err := s.file.DeleteFile(ctx, key); err != nil {
			log.Printf("cleanup: failed to delete orphaned upload %s: %v", key, err)
		}
	}
}

// persistPostInTx inserts the post (and attachments) inside an open transaction.
// S3 uploads run here too so a failed DB step rolls back the post; uploadedKeys
// is used by the caller to delete objects if the surrounding InTx fails (including commit).
func (s *service) persistPostInTx(
	ctx context.Context,
	tx database.Repository,
	threadID, postID uuid.UUID,
	post *Post,
	uploadedKeys *[]string,
) error {
	hasAttachment := len(post.Attachments) > 0
	if err := tx.AddPost(ctx, &database.Post{
		ID:                 postID,
		Text:               post.Text,
		ThreadID:           threadID,
		Sage:               post.Sage,
		OpMarker:           post.OpMarker,
		BrowserFingerprint: post.BrowserFingerprint,
		IP:                 post.IP,
		HasAttachment:      hasAttachment,
	}); err != nil {
		return err
	}

	for _, attachment := range post.Attachments {
		data, err := attachmentBytes(attachment)
		if err != nil {
			log.Printf("error decoding attachment: %v", err)
			return err
		}
		fileResp, err := s.file.UploadFile(ctx, &file.FileReq{
			PostID: postID,
			Name:   attachment.Name,
			Type:   attachment.Type,
			Data:   data,
		})
		if err != nil {
			log.Printf("error uploading file: %v", err)
			return err
		}
		*uploadedKeys = append(*uploadedKeys, fileResp.Key)

		if err := tx.AddAttachment(ctx, &database.Attachment{
			ID:     uuid.New(),
			PostID: postID,
			Link:   fileResp.Link,
			Name:   attachment.Name,
			Type:   attachment.Type,
		}); err != nil {
			log.Printf("error saving attachment %s to db: %v", fileResp.Link, err)
			return err
		}
	}

	if !post.Sage {
		shouldBump, _ := tx.CheckIfThreadBelowBumpLimit(ctx, threadID)
		if shouldBump {
			return tx.BumpThread(ctx, threadID)
		}
	}
	return nil
}

func (s *service) CreatePost(ctx context.Context, threadID uuid.UUID, post *Post) error {
	postID := uuid.New()
	var uploadedKeys []string
	err := s.repo.InTx(ctx, func(tx database.Repository) error {
		return s.persistPostInTx(ctx, tx, threadID, postID, post, &uploadedKeys)
	})
	if err != nil {
		s.cleanupUploads(ctx, uploadedKeys)
		return err
	}
	return nil
}

func (s *service) CreateThread(ctx context.Context, thread *Thread) (*Thread, error) {
	if thread == nil {
		return nil, errors.New("thread is required")
	}
	if thread.OriginalPost == nil {
		return nil, errors.New("original_post is required")
	}

	board, err := s.repo.GetBoardById(ctx, thread.BoardID)
	if err != nil {
		return nil, err
	}
	if board == nil {
		return nil, errors.New("board not found")
	}

	if board.Locked && !thread.IsAdmin {
		return nil, ErrBoardLocked
	}

	threadDB := &database.Thread{
		ID:      uuid.New(),
		BoardID: thread.BoardID,
		Title:   thread.Title,
	}

	postID := uuid.New()
	var uploadedKeys []string
	err = s.repo.InTx(ctx, func(tx database.Repository) error {
		if err := tx.AddThread(ctx, threadDB); err != nil {
			return err
		}
		return s.persistPostInTx(ctx, tx, threadDB.ID, postID, thread.OriginalPost, &uploadedKeys)
	})
	if err != nil {
		s.cleanupUploads(ctx, uploadedKeys)
		return nil, err
	}

	return &Thread{
		ID:           threadDB.ID,
		BoardID:      threadDB.BoardID,
		Title:        threadDB.Title,
		OriginalPost: nil,
		Locked:       threadDB.Locked,
		UpdatedAt:    "",
		Posts:        []*Post{},
		RepliesCount: 0,
	}, nil
}

func (s *service) DeleteBoard(ctx context.Context, id uuid.UUID) error {
	return ErrNotImplemented
}

func (s *service) DeletePost(ctx context.Context, id uuid.UUID) error {
	return ErrNotImplemented
}

func (s *service) DeleteThread(ctx context.Context, id uuid.UUID) error {
	return ErrNotImplemented
}

func (s *service) GetBoardWithThreads(ctx context.Context, alias string, limit, offset int) (*Board, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	boardDB, err := s.repo.GetBoardByAlias(ctx, alias)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.CountThreads(ctx, boardDB.ID)
	if err != nil {
		return nil, err
	}

	catalog, err := s.repo.GetBoardCatalog(ctx, boardDB.ID, limit, offset)
	if err != nil {
		return nil, err
	}

	opIDs := make([]uuid.UUID, 0, len(catalog))
	for _, ct := range catalog {
		if ct.OP.HasAttachment && ct.OP.ID != uuid.Nil {
			opIDs = append(opIDs, ct.OP.ID)
		}
	}

	attsByPost := map[uuid.UUID][]Attachment{}
	if len(opIDs) > 0 {
		atts, err := s.repo.GetAttachmentsByPostIDs(ctx, opIDs)
		if err != nil {
			log.Printf("error while batching attachments: %s", err)
		} else {
			for _, a := range atts {
				attsByPost[a.PostID] = append(attsByPost[a.PostID], Attachment{
					ID:   a.ID,
					Name: a.Name,
					Type: a.Type,
					Link: a.Link,
				})
			}
		}
	}

	threads := make([]*Thread, 0, len(catalog))
	for _, ct := range catalog {
		var attachments []Attachment
		if ct.OP.HasAttachment {
			attachments = attsByPost[ct.OP.ID]
		}
		threads = append(threads, &Thread{
			ID:      ct.ID,
			BoardID: ct.BoardID,
			Title:   ct.Title,
			Locked:  ct.Locked,
			OriginalPost: &Post{
				ID:                 ct.OP.ID,
				Number:             ct.OP.Number,
				Text:               ct.OP.Text,
				ThreadID:           ct.OP.ThreadID,
				Sage:               ct.OP.Sage,
				OpMarker:           ct.OP.OpMarker,
				BrowserFingerprint: ct.OP.BrowserFingerprint,
				IP:                 ct.OP.IP,
				CreatedAt:          ct.OP.CreatedAt,
				Attachments:        attachments,
			},
			RepliesCount: ct.RepliesCount,
		})
	}

	return &Board{
		ID:           boardDB.ID,
		Alias:        boardDB.Alias,
		Name:         boardDB.Name,
		Locked:       boardDB.Locked,
		Description:  boardDB.Description,
		Threads:      threads,
		TotalThreads: total,
		Limit:        limit,
		Offset:       offset,
	}, nil
}

func (s *service) GetBoards(ctx context.Context) ([]*Board, error) {
	boardsDB, err := s.repo.GetBoards(ctx)
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

func (s *service) GetThreadWithPosts(ctx context.Context, id uuid.UUID) (*Thread, error) {
	threadDB, err := s.repo.GetThread(ctx, id)
	if err != nil {
		return nil, err
	}
	postsDB, err := s.repo.GetPosts(ctx, id)
	if err != nil {
		return nil, err
	}
	var op *Post
	posts := make([]*Post, 0, len(postsDB))
	for i, postDB := range postsDB {
		var attachments []Attachment
		if postDB.HasAttachment {
			attachments, err = s.getAttachmentsForPost(ctx, postDB.ID)
			if err != nil {
				log.Printf("error while getting attachments for post %s: %s", postDB.ID, err)
			}
		}

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
				CreatedAt:          postDB.CreatedAt,
				Attachments:        attachments,
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
			CreatedAt:          postDB.CreatedAt,
			Attachments:        attachments,
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

func (s *service) getAttachmentsForPost(ctx context.Context, id uuid.UUID) ([]Attachment, error) {
	attachmentsDB, err := s.repo.GetAttachments(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var attachments []Attachment
	for _, attachmentDB := range attachmentsDB {
		attachments = append(attachments, Attachment{
			ID:   attachmentDB.ID,
			Name: attachmentDB.Name,
			Type: attachmentDB.Type,
			Link: attachmentDB.Link,
		})
	}
	return attachments, nil
}

func (s *service) UpdateBoard(ctx context.Context, board *Board) error {
	return ErrNotImplemented
}

func attachmentBytes(a Attachment) ([]byte, error) {
	if len(a.Data) > 0 {
		return a.Data, nil
	}
	if a.Body == "" {
		return nil, errors.New("empty attachment body")
	}
	return base64.StdEncoding.DecodeString(a.Body)
}

func NewService(repo database.Repository, file file.Service, conf *config.Config) Service {
	return &service{repo: repo, file: file, conf: conf}
}
