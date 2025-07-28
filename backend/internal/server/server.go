package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Fodro/saberchan/config"
	"github.com/Fodro/saberchan/internal/board"
	"github.com/Fodro/saberchan/internal/health"
	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Server struct {
	srv    *http.Server
	conf   *config.Config
	board  board.Service
	health health.Service
}

func NewServer(conf *config.Config, board board.Service, health health.Service) *Server {
	return &Server{
		srv: &http.Server{
			Addr: ":" + conf.Port,
		},
		conf:   conf,
		board:  board,
		health: health,
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
		})
	})

	return s.srv.ListenAndServe()
}


func (s *Server) CreateBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	var board board.BoardInput
	if err := json.NewDecoder(r.Body).Decode(&board); err != nil {
		log.Printf("failed to decode board: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := s.board.CreateBoard(&board)
	if err != nil {
		log.Printf("failed to create board: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(boards); err != nil {
		log.Printf("failed to encode boards: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetBoardByAlias(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	alias := chi.URLParam(r, "alias")
	board, err := s.board.GetBoardWithThreads(alias)
	if err != nil {
		log.Printf("failed to get board: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(board); err != nil {
		log.Printf("failed to encode board: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreateThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	var thread board.Thread
	if err := json.NewDecoder(r.Body).Decode(&thread); err != nil {
		log.Printf("failed to decode thread: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := s.board.CreateThread(&thread)
	if err != nil {
		log.Printf("failed to create thread: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("failed to encode response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetThread(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	id := chi.URLParam(r, "id")
	convertedId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("failed to parse id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	thread, err := s.board.GetThreadWithPosts(convertedId)
	if err != nil {
		log.Printf("failed to get thread: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(thread); err != nil {
		log.Printf("failed to encode thread: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	threadID := chi.URLParam(r, "thread_id")
	convertedThreadID, err := uuid.Parse(threadID)
	if err != nil {
		log.Printf("failed to parse thread_id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var post board.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		log.Printf("failed to decode post: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.board.CreatePost(convertedThreadID, &post); err != nil {
		log.Printf("failed to create post: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
