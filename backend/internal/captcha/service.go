package captcha

import (
	"context"
	"image/color"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	generator "github.com/steambap/captcha"
)

type service struct {
	expires time.Duration
	store   TokenStore
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
	err = s.store.Set(ctx, token, data.Text, s.expires)
	if err != nil {
		return nil, "", err
	}

	return data, token, nil
}

func (s *service) Validate(ctx context.Context, input, token string) (bool, error) {
	token = strings.TrimSpace(token)
	input = strings.TrimSpace(input)
	if token == "" {
		return false, nil
	}

	text, err := s.store.GetDel(ctx, token)
	if err != nil {
		return false, err
	}

	if input == strings.TrimSpace(text) {
		return true, nil
	}

	return false, nil
}

func NewService(rdb *redis.Client, expires time.Duration) Service {
	return NewServiceWithStore(&redisStore{rdb: rdb}, expires)
}

func NewServiceWithStore(store TokenStore, expires time.Duration) Service {
	return &service{expires: expires, store: store}
}
