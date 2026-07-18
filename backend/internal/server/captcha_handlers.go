package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Fodro/saberchan/internal/captcha"
)

func (s *Server) GenerateCaptcha(w http.ResponseWriter, r *http.Request) {
	data, token, err := s.captcha.Generate(r.Context())
	if err != nil {
		log.Printf("failed to generate captcha: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Add("x-captcha-token", token)
	data.WriteImage(w)
}

func (s *Server) ValidateCaptcha(w http.ResponseWriter, r *http.Request) {
	var captchaReq captcha.CaptchaValidateReq
	if err := json.NewDecoder(r.Body).Decode(&captchaReq); err != nil {
		log.Printf("failed to decode captcha req: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}

	passed, _ := s.captcha.Validate(r.Context(), captchaReq.Input, captchaReq.Token)

	resp := captcha.CaptchaValidateResp{
		Passed: passed,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
