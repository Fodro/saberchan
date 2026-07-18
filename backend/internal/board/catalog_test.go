package board

import (
	"context"
	"testing"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/database"
	dbmocks "github.com/Fodro/saberchan/internal/database/mocks"
	filemocks "github.com/Fodro/saberchan/internal/file/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestGetBoardWithThreads_UsesCatalogAndClampsLimit(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	repo := dbmocks.NewMockRepository(ctrl)
	files := filemocks.NewMockService(ctrl)
	svc := NewService(repo, files, &config.Config{}, nil)

	boardID := uuid.New()
	opID := uuid.New()
	threadID := uuid.New()

	repo.EXPECT().GetBoardByAlias(gomock.Any(), "b", false).Return(&database.Board{
		ID: boardID, Alias: "b", Name: "Board", Description: "d",
	}, nil)
	repo.EXPECT().CountThreads(gomock.Any(), boardID, false).Return(uint64(50), nil)
	repo.EXPECT().GetBoardCatalog(gomock.Any(), boardID, 100, 0, false).Return([]database.CatalogThread{{
		Thread: database.Thread{ID: threadID, BoardID: boardID, Title: "t"},
		OP: database.Post{
			ID: opID, ThreadID: threadID, Number: 1, Text: "op", HasAttachment: true,
		},
		RepliesCount: 3,
	}}, nil)
	repo.EXPECT().GetAttachmentsByPostIDs(gomock.Any(), []uuid.UUID{opID}).Return([]database.Attachment{{
		ID: uuid.New(), PostID: opID, Link: "http://x/y", Name: "a.jpg", Type: "image",
	}}, nil)

	// limit 500 should clamp to 100
	got, err := svc.GetBoardWithThreads(context.Background(), "b", 500, -5, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Limit != 100 || got.Offset != 0 {
		t.Fatalf("limit/offset = %d/%d, want 100/0", got.Limit, got.Offset)
	}
	if got.TotalThreads != 50 {
		t.Fatalf("total = %d", got.TotalThreads)
	}
	if len(got.Threads) != 1 || got.Threads[0].RepliesCount != 3 {
		t.Fatalf("threads = %+v", got.Threads)
	}
	if len(got.Threads[0].OriginalPost.Attachments) != 1 {
		t.Fatalf("attachments = %+v", got.Threads[0].OriginalPost.Attachments)
	}
}
