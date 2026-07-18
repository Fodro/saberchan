package server

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Fodro/saberchan/internal/board"
	"github.com/google/uuid"
)

const (
	maxMultipartMemory = 40 << 20 // headroom for up to 4×10 MiB videos
	maxImageBytes      = 5 << 20  // 5 MiB
	maxVideoBytes      = 10 << 20 // 10 MiB
	maxUploadFiles     = 4
)

var allowedImageMIME = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var allowedVideoMIME = map[string]bool{
	"video/webm": true,
	"video/mp4":  true,
}

func isAllowedUploadMIME(ctype string) bool {
	return allowedImageMIME[ctype] || allowedVideoMIME[ctype]
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
		ctype := fh.Header.Get("Content-Type")
		if ctype == "" {
			ctype = mime.TypeByExtension(filepath.Ext(fh.Filename))
		}
		if mediaType, _, err := mime.ParseMediaType(ctype); err == nil {
			ctype = mediaType
		}
		ctype = strings.ToLower(ctype)
		limit := int64(maxImageBytes)
		limitMsg := "maximum image size is 5MB"
		if allowedVideoMIME[ctype] {
			limit = int64(maxVideoBytes)
			limitMsg = "maximum video size is 10MB"
		}
		if fh.Size > limit {
			return nil, fmt.Errorf("%s", limitMsg)
		}
		if !isAllowedUploadMIME(ctype) {
			return nil, fmt.Errorf("unsupported file type %q", ctype)
		}

		f, err := fh.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(f, limit+1))
		_ = f.Close()
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > limit {
			return nil, fmt.Errorf("%s", limitMsg)
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
	return &board.Thread{
		BoardID: boardID,
		Title:   r.FormValue("title"),
		IsAdmin: false, // set by handler via admin token only
		OriginalPost: &board.Post{
			Text:               r.FormValue("text"),
			Sage:               false,
			OpMarker:           true,
			BrowserFingerprint: r.FormValue("browser_fingerprint"),
			Attachments:        atts,
		},
	}, nil
}
