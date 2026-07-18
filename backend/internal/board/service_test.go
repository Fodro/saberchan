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

func TestCreateThread_RequiresOriginalPost(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	_, err := svc.CreateThread(&Thread{
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
	repo.EXPECT().GetBoardById(boardID).Return(&database.Board{ID: boardID, Locked: true}, nil)

	_, err := svc.CreateThread(&Thread{
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
	repo.EXPECT().GetBoardById(boardID).Return(&database.Board{ID: boardID, Locked: true}, nil)
	repo.EXPECT().AddThread(gomock.Any()).DoAndReturn(func(thread *database.Thread) error {
		if thread.BoardID != boardID || thread.Title != "admin thread" {
			t.Fatalf("unexpected thread: %+v", thread)
		}
		return nil
	})
	repo.EXPECT().AddPost(gomock.Any()).DoAndReturn(func(post *database.Post) error {
		if post.Text != "op" || !post.OpMarker {
			t.Fatalf("unexpected post: %+v", post)
		}
		return nil
	})
	repo.EXPECT().CheckIfThreadBelowBumpLimit(gomock.Any()).Return(true, nil)
	repo.EXPECT().BumpThread(gomock.Any()).Return(nil)

	res, err := svc.CreateThread(&Thread{
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
	repo.EXPECT().AddPost(gomock.Any()).Return(nil)

	if err := svc.CreatePost(threadID, &Post{Text: "sage", Sage: true, IP: "0.0.0.0"}); err != nil {
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
	repo.EXPECT().AddPost(gomock.Any()).Return(nil)
	repo.EXPECT().CheckIfThreadBelowBumpLimit(threadID).Return(true, nil)
	repo.EXPECT().BumpThread(threadID).Return(nil)

	if err := svc.CreatePost(threadID, &Post{Text: "bump", Sage: false, IP: "0.0.0.0"}); err != nil {
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
	repo.EXPECT().AddPost(gomock.Any()).DoAndReturn(func(post *database.Post) error {
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
			return &file.FileResp{Link: link}, nil
		})
	repo.EXPECT().AddAttachment(gomock.Any()).DoAndReturn(func(a *database.Attachment) error {
		if a.Link != link || a.Name != "pic.jpg" {
			t.Fatalf("unexpected attachment: %+v", a)
		}
		return nil
	})
	repo.EXPECT().CheckIfThreadBelowBumpLimit(gomock.Any()).Return(false, nil)

	err := svc.CreatePost(uuid.New(), &Post{
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
	repo.EXPECT().AddPost(gomock.Any()).Return(nil)
	files.EXPECT().UploadFile(gomock.Any(), gomock.Any()).Return(nil, uploadErr)

	err := svc.CreatePost(uuid.New(), &Post{
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

func TestGetBoards_MapsFields(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{})

	id := uuid.New()
	repo.EXPECT().GetBoards().Return([]database.Board{{
		ID: id, Alias: "b", Name: "Random", Description: "desc", Locked: true,
	}}, nil)

	boards, err := svc.GetBoards()
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
