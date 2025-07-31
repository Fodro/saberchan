package file

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	UploadFile(ctx context.Context, file *FileReq) (*FileResp, error)
}

type (
	FileReq struct {
		PostID uuid.UUID `json:"post_id"`
		Name string `json:"name"`
		Body string `json:"body"`
	}

	FileResp struct {
		Link string `json:"link"`
	}

)