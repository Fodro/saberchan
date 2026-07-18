package board

import (
	"context"
	"errors"
	"testing"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/database"
	dbmocks "github.com/Fodro/saberchan/internal/database/mocks"
	"github.com/Fodro/saberchan/internal/file"
	filemocks "github.com/Fodro/saberchan/internal/file/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func expectInTx(repo *dbmocks.MockRepository) {
	repo.EXPECT().InTx(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, fn func(database.Repository) error) error {
		return fn(repo)
	}).AnyTimes()
}

func TestCreateThread_RequiresOriginalPost(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	_, err := svc.CreateThread(context.Background(), &Thread{
		BoardID: uuid.New(),
		Title:   "no op",
	})
	if err == nil {
		t.Fatal("expected error when original_post is missing")
	}
}

func TestCreateThread_LockedBoard_NonAdmin(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	boardID := uuid.New()
	repo.EXPECT().GetBoardById(gomock.Any(), boardID, false).Return(&database.Board{ID: boardID, Locked: true}, nil)

	_, err := svc.CreateThread(context.Background(), &Thread{
		BoardID:      boardID,
		Title:        "nope",
		IsAdmin:      false,
		OriginalPost: &Post{Text: "hi", IP: "1.1.1.1"},
	})
	if !errors.Is(err, ErrBoardLocked) {
		t.Fatalf("got %v, want ErrBoardLocked", err)
	}
}

func TestCreateThread_LockedBoard_Admin(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	boardID := uuid.New()
	expectInTx(repo)
	repo.EXPECT().GetBoardById(gomock.Any(), boardID, false).Return(&database.Board{ID: boardID, Locked: true}, nil)
	repo.EXPECT().AddThread(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, thread *database.Thread) error {
		if thread.BoardID != boardID || thread.Title != "admin thread" {
			t.Fatalf("unexpected thread: %+v", thread)
		}
		return nil
	})
	repo.EXPECT().AddPost(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, post *database.Post) error {
		if post.Text != "op" || !post.OpMarker {
			t.Fatalf("unexpected post: %+v", post)
		}
		return nil
	})
	repo.EXPECT().CheckIfThreadBelowBumpLimit(gomock.Any(), gomock.Any()).Return(true, nil)
	repo.EXPECT().BumpThread(gomock.Any(), gomock.Any()).Return(nil)

	res, err := svc.CreateThread(context.Background(), &Thread{
		BoardID:      boardID,
		Title:        "admin thread",
		IsAdmin:      true,
		OriginalPost: &Post{Text: "op", IP: "1.1.1.1", OpMarker: true},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res == nil || res.Title != "admin thread" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestCreatePost_SageSkipsBump(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	threadID := uuid.New()
	expectInTx(repo)
	repo.EXPECT().AddPost(gomock.Any(), gomock.Any()).Return(nil)

	if err := svc.CreatePost(context.Background(), threadID, &Post{Text: "sage", Sage: true, IP: "0.0.0.0"}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestCreatePost_BumpsWhenAllowed(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	threadID := uuid.New()
	expectInTx(repo)
	repo.EXPECT().AddPost(gomock.Any(), gomock.Any()).Return(nil)
	repo.EXPECT().CheckIfThreadBelowBumpLimit(gomock.Any(), threadID).Return(true, nil)
	repo.EXPECT().BumpThread(gomock.Any(), threadID).Return(nil)

	if err := svc.CreatePost(context.Background(), threadID, &Post{Text: "bump", Sage: false, IP: "0.0.0.0"}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestCreatePost_UploadsAttachment(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	link := "http://localhost:9000/saberchan/abc.jpg"
	expectInTx(repo)
	repo.EXPECT().AddPost(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, post *database.Post) error {
		if !post.HasAttachment {
			t.Fatal("HasAttachment should be true")
		}
		return nil
	})
	files.EXPECT().
		UploadFile(gomock.Any(), gomock.AssignableToTypeOf(&file.FileReq{})).
		DoAndReturn(func(_ context.Context, req *file.FileReq) (*file.FileResp, error) {
			if req.Name != "pic.jpg" {
				t.Fatalf("name=%s", req.Name)
			}
			return &file.FileResp{Link: link, Key: "abc.jpg"}, nil
		})
	repo.EXPECT().AddAttachment(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, a *database.Attachment) error {
		if a.Link != link || a.Name != "pic.jpg" {
			t.Fatalf("unexpected attachment: %+v", a)
		}
		return nil
	})
	repo.EXPECT().CheckIfThreadBelowBumpLimit(gomock.Any(), gomock.Any()).Return(false, nil)

	err := svc.CreatePost(context.Background(), uuid.New(), &Post{
		Text: "with file",
		IP:   "0.0.0.0",
		Attachments: []Attachment{{
			Name: "pic.jpg",
			Type: "image/jpeg",
			Data: []byte("fake-image-bytes"),
		}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestCreatePost_UploadFailurePropagates(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	uploadErr := errors.New("minio down")
	expectInTx(repo)
	repo.EXPECT().AddPost(gomock.Any(), gomock.Any()).Return(nil)
	files.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return(nil, uploadErr)

	err := svc.CreatePost(context.Background(), uuid.New(), &Post{
		Text: "x",
		IP:   "0.0.0.0",
		Attachments: []Attachment{{
			Name: "a.jpg",
			Type: "image/jpeg",
			Body: "QQ==",
		}},
	})
	if !errors.Is(err, uploadErr) {
		t.Fatalf("got %v, want %v", err, uploadErr)
	}
}

func TestCreatePost_CleansUpS3WhenAttachmentInsertFails(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	expectInTx(repo)
	repo.EXPECT().AddPost(gomock.Any(), gomock.Any()).Return(nil)
	files.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return(&file.FileResp{
		Link: "http://x/k.jpg", Key: "k.jpg",
	}, nil)
	repo.EXPECT().AddAttachment(gomock.Any(), gomock.Any()).Return(errors.New("db down"))
	files.EXPECT().DeleteFile(gomock.Any(), "k.jpg").Return(nil)

	err := svc.CreatePost(context.Background(), uuid.New(), &Post{
		Text: "x",
		IP:   "0.0.0.0",
		Attachments: []Attachment{{
			Name: "a.jpg",
			Type: "image/jpeg",
			Data: []byte("x"),
		}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetBoards_MapsFields(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	id := uuid.New()
	repo.EXPECT().GetBoards(gomock.Any(), false).Return([]database.Board{{
		ID: id, Alias: "b", Name: "Random", Description: "desc", Locked: true,
	}}, nil)

	boards, err := svc.GetBoards(context.Background(), false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(boards) != 1 {
		t.Fatalf("len=%d", len(boards))
	}
	got := boards[0]
	if got.ID != id || got.Alias != "b" || got.Name != "Random" || !got.Locked {
		t.Fatalf("bad mapping: %+v", got)
	}
}
