package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/board"
	"github.com/Fodro/saberchan/internal/captcha"
	"github.com/Fodro/saberchan/internal/health"
	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Server struct {
	srv     *http.Server
	conf    *config.Config
	board   board.Service
	captcha captcha.Service
	health  health.Service
}

func NewServer(conf *config.Config, board board.Service, captcha captcha.Service, health health.Service) *Server {
	return &Server{
		srv: &http.Server{
			Addr: ":" + conf.Port,
		},
		conf:    conf,
		board:   board,
		captcha: captcha,
		health:  health,
	}
}

func writeJSONError(w http.ResponseWriter, status int, err error, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
		"code":  code,
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Server) applyClientIP(r *http.Request, post *board.Post) {
	if post == nil {
		return
	}
	if s.conf.StoreIp {
		post.IP = clientIP(r)
	} else {
		post.IP = ""
	}
}

func (s *Server) Start() error {
	r := chi.NewRouter()
	s.srv.Handler = r

	// healtcheck
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/readiness", func(w http.ResponseWriter, r *http.Request) {
		if err := s.health.Readiness(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {

			r.Route("/board", func(r chi.Router) {
				r.Post("/", s.CreateBoard)
				r.Get("/", s.GetBoards)
				r.Get("/{alias}", s.GetBoardByAlias)
			})

			r.Route("/thread", func(r chi.Router) {
				r.Post("/", s.CreateThread)
				r.Get("/{id}", s.GetThread)
			})

			r.Route("/post", func(r chi.Router) {
				r.Post("/{thread_id}", s.CreatePost)
			})

			r.Route("/captcha", func(r chi.Router) {
				r.Get("/", s.GenerateCaptcha)
				r.Post("/", s.ValidateCaptcha)
			})
		})
	})

	return s.srv.ListenAndServe()
}

func (s *Server) CreateBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	var boardIn board.BoardInput
	if err := json.NewDecoder(r.Body).Decode(&boardIn); err != nil {
		log.Printf("failed to decode board: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	err := s.board.CreateBoard(&boardIn)
	if err != nil {
		log.Printf("failed to create board: %v", err)
		if errors.Is(err, board.ErrNotImplemented) {
			writeJSONError(w, http.StatusNotImplemented, err, "not_implemented")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetBoards(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	boards, err := s.board.GetBoards()
	if err != nil {
		log.Printf("failed to get boards: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(boards); err != nil {
		log.Printf("failed to encode boards: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
}

func (s *Server) GetBoardByAlias(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	alias := chi.URLParam(r, "alias")

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	boardRes, err := s.board.GetBoardWithThreads(alias, limit, offset)
	if err != nil {
		log.Printf("failed to get board: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(boardRes); err != nil {
		log.Printf("failed to encode board: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
}

func (s *Server) CreateThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	var thread board.Thread
	if isMultipart(r) {
		parsed, err := parseMultipartThread(r)
		if err != nil {
			log.Printf("failed to parse multipart thread: %v", err)
			writeJSONError(w, http.StatusBadRequest, err, "bad_request")
			return
		}
		thread = *parsed
	} else if err := json.NewDecoder(r.Body).Decode(&thread); err != nil {
		log.Printf("failed to decode thread: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	s.applyClientIP(r, thread.OriginalPost)
	res, err := s.board.CreateThread(&thread)
	if err != nil {
		log.Printf("failed to create thread: %v", err)
		status := http.StatusInternalServerError
		code := "internal_error"
		if errors.Is(err, board.ErrBoardLocked) {
			status = http.StatusForbidden
			code = "board_locked"
		} else if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusBadRequest
			code = "not_found"
		} else if errors.Is(err, board.ErrNotImplemented) {
			status = http.StatusNotImplemented
			code = "not_implemented"
		}
		writeJSONError(w, status, err, code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func (s *Server) GetThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	id := chi.URLParam(r, "id")
	convertedId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("failed to parse id: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	thread, err := s.board.GetThreadWithPosts(convertedId)
	if err != nil {
		log.Printf("failed to get thread: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(thread); err != nil {
		log.Printf("failed to encode thread: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err, "internal_error")
		return
	}
}

func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	threadID := chi.URLParam(r, "thread_id")
	convertedThreadID, err := uuid.Parse(threadID)
	if err != nil {
		log.Printf("failed to parse thread_id: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	var post board.Post
	if isMultipart(r) {
		parsed, err := parseMultipartPost(r)
		if err != nil {
			log.Printf("failed to parse multipart post: %v", err)
			writeJSONError(w, http.StatusBadRequest, err, "bad_request")
			return
		}
		post = *parsed
	} else if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		log.Printf("failed to decode post: %v", err)
		writeJSONError(w, http.StatusBadRequest, err, "bad_request")
		return
	}
	s.applyClientIP(r, &post)
	if err := s.board.CreatePost(convertedThreadID, &post); err != nil {
		log.Printf("failed to create post: %v", err)
		status := http.StatusInternalServerError
		code := "internal_error"
		if errors.Is(err, board.ErrBoardLocked) {
			status = http.StatusForbidden
			code = "board_locked"
		} else if errors.Is(err, board.ErrNotImplemented) {
			status = http.StatusNotImplemented
			code = "not_implemented"
		}
		writeJSONError(w, status, err, code)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

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

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
