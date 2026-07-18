package captcha

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Fodro/saberchan/internal/captcha/mocks"
	"go.uber.org/mock/gomock"
)

func TestValidate_SuccessAndConsume(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store := mocks.NewMockTokenStore(ctrl)
	svc := NewServiceWithStore(store, time.Minute)
	ctx := context.Background()

	store.EXPECT().GetDel(ctx, "tok").Return("42", nil)

	ok, err := svc.Validate(ctx, "42", "tok")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !ok {
		t.Fatal("expected captcha to pass")
	}
}

func TestValidate_WrongAnswer(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store := mocks.NewMockTokenStore(ctrl)
	svc := NewServiceWithStore(store, time.Minute)
	ctx := context.Background()

	store.EXPECT().GetDel(ctx, "tok").Return("7", nil)

	ok, err := svc.Validate(ctx, "8", "tok")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ok {
		t.Fatal("expected captcha to fail")
	}
}

func TestValidate_StoreError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store := mocks.NewMockTokenStore(ctrl)
	svc := NewServiceWithStore(store, time.Minute)
	ctx := context.Background()

	store.EXPECT().GetDel(ctx, "missing").Return("", errors.New("token not found"))

	ok, err := svc.Validate(ctx, "1", "missing")
	if err == nil || ok {
		t.Fatal("expected store error")
	}
}

func TestGenerate_StoresAnswer(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	store := mocks.NewMockTokenStore(ctrl)
	svc := NewServiceWithStore(store, time.Minute)
	ctx := context.Background()

	var savedToken, savedAnswer string
	store.EXPECT().
		Set(ctx, gomock.Any(), gomock.Any(), time.Minute).
		DoAndReturn(func(_ context.Context, token, answer string, _ time.Duration) error {
			savedToken = token
			savedAnswer = answer
			return nil
		})

	data, token, err := svc.Generate(ctx)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if token == "" || token != savedToken {
		t.Fatalf("token mismatch: got %q saved %q", token, savedToken)
	}
	if data == nil || data.Text == "" || data.Text != savedAnswer {
		t.Fatalf("answer mismatch: data=%q saved=%q", data.Text, savedAnswer)
	}
}
