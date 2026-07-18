package purge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Fodro/saberchan/internal/database"
	dbmocks "github.com/Fodro/saberchan/internal/database/mocks"
	filemocks "github.com/Fodro/saberchan/internal/file/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestKeyFromLink(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"http://localhost:9000/saberchan/abc123.jpg": "abc123.jpg",
		"abc123.jpg":        "abc123.jpg",
		"":                  "",
		"http://x/y/z.png/": "z.png",
	}
	for link, want := range cases {
		if got := KeyFromLink(link); got != want {
			t.Errorf("KeyFromLink(%q) = %q, want %q", link, got, want)
		}
	}
}

func TestSweep_PurgesPostWithAttachments(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)

	postID := uuid.New()
	post := database.Post{ID: postID}

	repo.EXPECT().ListPostsDueForPurge(gomock.Any(), gomock.Any()).Return([]database.Post{post}, nil)
	repo.EXPECT().GetAttachments(gomock.Any(), postID).Return([]database.Attachment{
		{ID: uuid.New(), PostID: postID, Key: "abc.jpg"},
		{ID: uuid.New(), PostID: postID, Link: "http://x/def.png"}, // empty key, must derive from link
	}, nil)
	files.EXPECT().DeleteFile(gomock.Any(), "abc.jpg").Return(nil)
	files.EXPECT().DeleteFile(gomock.Any(), "def.png").Return(nil)
	repo.EXPECT().DeleteAttachmentsByPostID(gomock.Any(), postID).Return(nil)
	repo.EXPECT().MarkPostPurged(gomock.Any(), postID).Return(nil)

	repo.EXPECT().ListThreadsDueForPurge(gomock.Any(), gomock.Any()).Return(nil, nil)
	repo.EXPECT().ListBoardsDueForPurge(gomock.Any(), gomock.Any()).Return(nil, nil)

	Sweep(context.Background(), repo, files)
}

func TestSweep_MarksThreadsAndBoardsPurged(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)

	threadID := uuid.New()
	boardID := uuid.New()

	repo.EXPECT().ListPostsDueForPurge(gomock.Any(), gomock.Any()).Return(nil, nil)
	repo.EXPECT().ListThreadsDueForPurge(gomock.Any(), gomock.Any()).Return([]database.Thread{{ID: threadID}}, nil)
	repo.EXPECT().MarkThreadPurged(gomock.Any(), threadID).Return(nil)
	repo.EXPECT().ListBoardsDueForPurge(gomock.Any(), gomock.Any()).Return([]database.Board{{ID: boardID}}, nil)
	repo.EXPECT().MarkBoardPurged(gomock.Any(), boardID).Return(nil)

	Sweep(context.Background(), repo, files)
}

func TestSweep_ContinuesAfterPostPurgeError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)

	failingPostID := uuid.New()

	repo.EXPECT().ListPostsDueForPurge(gomock.Any(), gomock.Any()).Return([]database.Post{{ID: failingPostID}}, nil)
	repo.EXPECT().GetAttachments(gomock.Any(), failingPostID).Return(nil, errors.New("db down"))
	repo.EXPECT().ListThreadsDueForPurge(gomock.Any(), gomock.Any()).Return(nil, nil)
	repo.EXPECT().ListBoardsDueForPurge(gomock.Any(), gomock.Any()).Return(nil, nil)

	// Should not panic and should not call MarkPostPurged for the failed post.
	Sweep(context.Background(), repo, files)
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)

	repo.EXPECT().ListPostsDueForPurge(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	repo.EXPECT().ListThreadsDueForPurge(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	repo.EXPECT().ListBoardsDueForPurge(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		Run(ctx, repo, files, time.Hour)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop after context cancellation")
	}
}
