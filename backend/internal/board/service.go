package board

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/database"
	"github.com/Fodro/saberchan/internal/file"
	"github.com/Fodro/saberchan/internal/file/s3service"
	"github.com/Fodro/saberchan/internal/follow"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const restoreWindow = 24 * time.Hour

type service struct {
	repo        database.Repository
	file        file.Service
	conf        *config.Config
	follow      follow.Service // optional; nil disables follow refresh on bump
	mediaPrefix string         // public object URL prefix (S3_PUBLIC_URL/…/bucket)
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
	bumped *bool,
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
			Key:    fileResp.Key,
		}); err != nil {
			log.Printf("error saving attachment %s to db: %v", fileResp.Link, err)
			return err
		}
	}

	if !post.Sage {
		shouldBump, _ := tx.CheckIfThreadBelowBumpLimit(ctx, threadID)
		if shouldBump {
			if err := tx.BumpThread(ctx, threadID); err != nil {
				return err
			}
			if bumped != nil {
				*bumped = true
			}
		}
	}
	return nil
}

func (s *service) refreshFollowOnBump(ctx context.Context, threadID uuid.UUID) {
	if s.follow == nil {
		return
	}
	if err := s.follow.RefreshOnBump(ctx, threadID); err != nil {
		log.Printf("follow: refresh on bump %s: %v", threadID, err)
	}
}

func (s *service) CreatePost(ctx context.Context, threadID uuid.UUID, post *Post) error {
	postID := uuid.New()
	var uploadedKeys []string
	var bumped bool
	err := s.repo.InTx(ctx, func(tx database.Repository) error {
		return s.persistPostInTx(ctx, tx, threadID, postID, post, &uploadedKeys, &bumped)
	})
	if err != nil {
		s.cleanupUploads(ctx, uploadedKeys)
		return err
	}
	if bumped {
		s.refreshFollowOnBump(ctx, threadID)
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

	board, err := s.repo.GetBoardById(ctx, thread.BoardID, false)
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
	var bumped bool
	err = s.repo.InTx(ctx, func(tx database.Repository) error {
		if err := tx.AddThread(ctx, threadDB); err != nil {
			return err
		}
		return s.persistPostInTx(ctx, tx, threadDB.ID, postID, thread.OriginalPost, &uploadedKeys, &bumped)
	})
	if err != nil {
		s.cleanupUploads(ctx, uploadedKeys)
		return nil, err
	}
	if bumped {
		s.refreshFollowOnBump(ctx, threadDB.ID)
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

// DeleteBoard soft-deletes a board and cascades the soft-delete to its
// threads and posts.
func (s *service) DeleteBoard(ctx context.Context, id uuid.UUID) error {
	b, err := s.repo.GetBoardById(ctx, id, true)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if b.DeletedAt != nil {
		return ErrAlreadyDeleted
	}
	return s.repo.SoftDeleteBoard(ctx, id)
}

	// RestoreBoard undoes a soft-delete within the 24h grace window and cascades
	// restore only to threads/posts that share the board's deleted_at timestamp.
	func (s *service) RestoreBoard(ctx context.Context, id uuid.UUID) error {
	b, err := s.repo.GetBoardById(ctx, id, true)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if b.DeletedAt == nil {
		return ErrNotFound
	}
	if b.DeletedAt.Before(time.Now().Add(-restoreWindow)) {
		return ErrRestoreExpired
	}
	return s.repo.RestoreBoard(ctx, id)
}

// DeleteThread soft-deletes a thread and cascades the soft-delete to its
// posts.
func (s *service) DeleteThread(ctx context.Context, id uuid.UUID) error {
	t, err := s.repo.GetThread(ctx, id, true)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if t.DeletedAt != nil {
		return ErrAlreadyDeleted
	}
	return s.repo.SoftDeleteThread(ctx, id)
}

	// RestoreThread undoes a soft-delete within the 24h grace window and
	// cascades restore only to posts that share the thread's deleted_at timestamp.
	func (s *service) RestoreThread(ctx context.Context, id uuid.UUID) error {
	t, err := s.repo.GetThread(ctx, id, true)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if t.DeletedAt == nil {
		return ErrNotFound
	}
	if t.DeletedAt.Before(time.Now().Add(-restoreWindow)) {
		return ErrRestoreExpired
	}
	return s.repo.RestoreThread(ctx, id)
}

// DeletePost soft-deletes a single post. Posts have no children.
func (s *service) DeletePost(ctx context.Context, id uuid.UUID) error {
	p, err := s.repo.GetPost(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if p.DeletedAt != nil {
		return ErrAlreadyDeleted
	}
	return s.repo.SoftDeletePost(ctx, id)
}

// RestorePost undoes a soft-delete within the 24h grace window.
func (s *service) RestorePost(ctx context.Context, id uuid.UUID) error {
	p, err := s.repo.GetPost(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if p.DeletedAt == nil {
		return ErrNotFound
	}
	if p.PurgedAt != nil {
		return ErrRestoreExpired
	}
	if p.DeletedAt.Before(time.Now().Add(-restoreWindow)) {
		return ErrRestoreExpired
	}
	return s.repo.RestorePost(ctx, id)
}

func (s *service) GetBoardWithThreads(ctx context.Context, alias string, limit, offset int, includeDeleted bool) (*Board, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	boardDB, err := s.repo.GetBoardByAlias(ctx, alias, includeDeleted)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.CountThreads(ctx, boardDB.ID, includeDeleted)
	if err != nil {
		return nil, err
	}

	catalog, err := s.repo.GetBoardCatalog(ctx, boardDB.ID, limit, offset, includeDeleted)
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
					Link: s.publicAttachmentLink(a),
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
			DeletedAt:    ct.DeletedAt,
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
		DeletedAt:    boardDB.DeletedAt,
	}, nil
}

func (s *service) GetBoards(ctx context.Context, includeDeleted bool) ([]*Board, error) {
	boardsDB, err := s.repo.GetBoards(ctx, includeDeleted)
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
			DeletedAt:   boardDB.DeletedAt,
		})
	}
	return boards, nil
}

func (s *service) GetThreadWithPosts(ctx context.Context, id uuid.UUID, includeDeleted bool) (*Thread, error) {
	threadDB, err := s.repo.GetThread(ctx, id, includeDeleted)
	if err != nil {
		return nil, err
	}
	postsDB, err := s.repo.GetPosts(ctx, id, includeDeleted)
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
				DeletedAt:          postDB.DeletedAt,
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
			DeletedAt:          postDB.DeletedAt,
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
		DeletedAt:    threadDB.DeletedAt,
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
			Link: s.publicAttachmentLink(attachmentDB),
		})
	}
	return attachments, nil
}

func (s *service) publicAttachmentLink(a database.Attachment) string {
	return s3service.ObjectPublicURL(s.mediaPrefix, a.Key, a.Link)
}

func (s *service) UpdateBoard(ctx context.Context, board *Board) error {
	return ErrNotImplemented
}

func (s *service) GetBoardMetrics(ctx context.Context, from, to time.Time) ([]*BoardMetrics, error) {
	bm, err := s.repo.GetBoardMetrics(ctx, from, to)
	if err != nil {
		return nil, err
	}
	result := make([]*BoardMetrics, len(bm))
	for i := range bm {
		result[i] = &BoardMetrics{
			BoardID:      bm[i].BoardID,
			BoardAlias:   bm[i].BoardAlias,
			PostCount:    bm[i].PostCount,
			DeletedCount: bm[i].DeletedCount,
			SageCount:    bm[i].SageCount,
			ThreadCount:  bm[i].ThreadCount,
		}
	}
	return result, nil
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

func NewService(repo database.Repository, fileSvc file.Service, conf *config.Config, followSvc follow.Service) Service {
	var mediaPrefix string
	if conf != nil && conf.S3 != nil {
		mediaPrefix = s3service.ResolveLinkPrefix(
			conf.S3.Bucket,
			conf.S3.Url,
			conf.S3.PublicURL,
			conf.S3.UseSSL,
			conf.S3.ForcePathStyle,
		)
	}
	return &service{
		repo:        repo,
		file:        fileSvc,
		conf:        conf,
		follow:      followSvc,
		mediaPrefix: mediaPrefix,
	}
}
