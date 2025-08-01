package captcha

import (
	"context"
	"image/color"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	generator "github.com/steambap/captcha"
)

type service struct {
	expires time.Duration
	rdb     *redis.Client
}

func (s *service) Generate(ctx context.Context) (*generator.Data, string, error) {
	data, err := generator.NewMathExpr(150, 50, func(o *generator.Options) {
		o.BackgroundColor = color.White
		o.TextLength = 6
	})

	if err != nil {
		return nil, "", err
	}

	token := uuid.NewString()
	err = s.rdb.Set(ctx, token, data.Text, s.expires).Err()
	if err != nil {
		return nil, "", err
	}

	return data, token, nil
}

func (s *service) Validate(ctx context.Context, input, token string) (bool, error) {
	text, err := s.rdb.GetDel(ctx, token).Result()
	if err != nil {
		return false, err
	}

	if input == text {
		return true, nil
	}

	return false, nil
}

func NewService(rdb *redis.Client, expires time.Duration) Service {
	return &service{expires, rdb}
}
