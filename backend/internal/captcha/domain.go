package captcha

import (
	"context"

	generator "github.com/steambap/captcha"
)

type Service interface {
	Generate(ctx context.Context) (*generator.Data, string, error)
	Validate(ctx context.Context, input, token string) (bool, error)
}

type (
	CaptchaValidateReq struct {
		Input string `json:"input"`
		Token string `json:"token"`
	}

	CaptchaValidateResp struct {
		Passed bool `json:"passed"`
	}
)