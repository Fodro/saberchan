package file

import (
	"context"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=mocks/mock_service.go -package=mocks github.com/Fodro/saberchan/internal/file Service

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