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
		Name   string    `json:"name"`
		Type   string    `json:"type"`
		// Data is raw file bytes (multipart path or decoded base64).
		Data []byte `json:"-"`
	}

	FileResp struct {
		Link string `json:"link"`
	}
)
