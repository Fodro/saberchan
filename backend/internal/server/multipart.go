package server

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Fodro/saberchan/internal/board"
	"github.com/google/uuid"
)

const (
	maxMultipartMemory = 10 << 20 // 10 MiB
	maxUploadBytes     = 2 << 20  // 2 MiB per file
	maxUploadFiles     = 4
)

var allowedImageMIME = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

func isMultipart(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, "multipart/form-data")
}

func formBool(r *http.Request, key string) bool {
	v := strings.ToLower(strings.TrimSpace(r.FormValue(key)))
	return v == "true" || v == "1" || v == "on" || v == "yes"
}

func parseMultipartFiles(r *http.Request) ([]board.Attachment, error) {
	if r.MultipartForm == nil {
		return nil, fmt.Errorf("multipart form not parsed")
	}
	headers := r.MultipartForm.File["files"]
	if len(headers) == 0 {
		// Also accept singular field name used by some clients.
		headers = r.MultipartForm.File["file"]
	}
	if len(headers) > maxUploadFiles {
		return nil, fmt.Errorf("maximum file count is %d", maxUploadFiles)
	}

	out := make([]board.Attachment, 0, len(headers))
	for _, fh := range headers {
		if fh.Size > maxUploadBytes {
			return nil, fmt.Errorf("maximum file size is 2MB")
		}
		ctype := fh.Header.Get("Content-Type")
		if ctype == "" {
			ctype = mime.TypeByExtension(filepath.Ext(fh.Filename))
		}
		if mediaType, _, err := mime.ParseMediaType(ctype); err == nil {
			ctype = mediaType
		}
		if !allowedImageMIME[strings.ToLower(ctype)] {
			return nil, fmt.Errorf("unsupported file type %q", ctype)
		}

		f, err := fh.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(f, maxUploadBytes+1))
		_ = f.Close()
		if err != nil {
			return nil, err
		}
		if len(data) > maxUploadBytes {
			return nil, fmt.Errorf("maximum file size is 2MB")
		}

		out = append(out, board.Attachment{
			Name: fh.Filename,
			Type: ctype,
			Data: data,
		})
	}
	return out, nil
}

func parseMultipartPost(r *http.Request) (*board.Post, error) {
	if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
		return nil, err
	}
	atts, err := parseMultipartFiles(r)
	if err != nil {
		return nil, err
	}
	return &board.Post{
		Text:               r.FormValue("text"),
		Sage:               formBool(r, "sage"),
		OpMarker:           formBool(r, "op_marker"),
		BrowserFingerprint: r.FormValue("browser_fingerprint"),
		Attachments:        atts,
	}, nil
}

func parseMultipartThread(r *http.Request) (*board.Thread, error) {
	if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
		return nil, err
	}
	boardID, err := uuid.Parse(r.FormValue("board_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid board_id")
	}
	atts, err := parseMultipartFiles(r)
	if err != nil {
		return nil, err
	}
	isAdmin, _ := strconv.ParseBool(r.FormValue("is_admin"))
	return &board.Thread{
		BoardID: boardID,
		Title:   r.FormValue("title"),
		IsAdmin: isAdmin || formBool(r, "is_admin"),
		OriginalPost: &board.Post{
			Text:               r.FormValue("text"),
			Sage:               false,
			OpMarker:           true,
			BrowserFingerprint: r.FormValue("browser_fingerprint"),
			Attachments:        atts,
		},
	}, nil
}
