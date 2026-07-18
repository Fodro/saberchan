package server

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
)

func TestFormBool(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Form = map[string][]string{
		"sage": {"true"},
		"off":  {"no"},
	}
	if !formBool(req, "sage") {
		t.Fatal("expected sage true")
	}
	if formBool(req, "off") {
		t.Fatal("expected off false")
	}
}

func TestParseMultipartPost_RejectsBadMIME(t *testing.T) {
	t.Parallel()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("text", "hello")
	part, err := w.CreateFormFile("files", "evil.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("not an image"))
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	_, err = parseMultipartPost(req)
	if err == nil {
		t.Fatal("expected MIME rejection")
	}
}

func TestParseMultipartPost_AcceptsJPEG(t *testing.T) {
	t.Parallel()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("text", "hello")
	_ = w.WriteField("sage", "true")
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="files"; filename="pic.jpg"`)
	h.Set("Content-Type", "image/jpeg")
	part, err := w.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte{0xff, 0xd8, 0xff})
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	post, err := parseMultipartPost(req)
	if err != nil {
		t.Fatal(err)
	}
	if post.Text != "hello" || !post.Sage {
		t.Fatalf("unexpected post: %+v", post)
	}
	if len(post.Attachments) != 1 || post.Attachments[0].Type != "image/jpeg" {
		t.Fatalf("unexpected attachments: %+v", post.Attachments)
	}
}
