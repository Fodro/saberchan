package server

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	errCaptchaRequired = errors.New("captcha required")
	errCaptchaFailed   = errors.New("captcha failed")
)

// requireCaptcha validates a one-time captcha token. Call after the request
// body/form has been parsed so multipart FormValue works. Returns false (and
// writes the error response) when validation fails.
func (s *Server) requireCaptcha(w http.ResponseWriter, r *http.Request, input, token string) bool {
	if s.captcha == nil {
		writeJSONError(w, http.StatusInternalServerError, errors.New("captcha unavailable"), "internal_error")
		return false
	}
	if input == "" || token == "" {
		writeJSONError(w, http.StatusBadRequest, errCaptchaRequired, "captcha_required")
		return false
	}
	ok, err := s.captcha.Validate(r.Context(), input, token)
	if err != nil || !ok {
		writeJSONError(w, http.StatusForbidden, errCaptchaFailed, "captcha_failed")
		return false
	}
	return true
}

type captchaJSON struct {
	CaptchaInput string `json:"captcha_input"`
	CaptchaToken string `json:"captcha_token"`
}

// decodeJSONWithCaptcha unmarshals JSON that embeds captcha_* next to the payload.
func decodeJSONWithCaptcha(r *http.Request, dest any) (input, token string, err error) {
	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return "", "", err
	}
	var c captchaJSON
	if err := json.Unmarshal(raw, &c); err != nil {
		return "", "", err
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return "", "", err
	}
	return c.CaptchaInput, c.CaptchaToken, nil
}
