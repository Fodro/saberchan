package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Fodro/saberchan/internal/board"
)

func TestDecodeJSONWithCaptcha(t *testing.T) {
	t.Parallel()
	body := `{"text":"hi","captcha_input":"4","captcha_token":"tok","sage":false}`
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	var post board.Post
	input, token, err := decodeJSONWithCaptcha(r, &post)
	if err != nil {
		t.Fatal(err)
	}
	if input != "4" || token != "tok" {
		t.Fatalf("captcha = %q/%q", input, token)
	}
	if post.Text != "hi" {
		t.Fatalf("text = %q", post.Text)
	}
}
