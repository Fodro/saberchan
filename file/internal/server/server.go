package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Fodro/saberchan/file/config"
	"github.com/Fodro/saberchan/file/internal/file"
	"github.com/Fodro/saberchan/file/internal/health"
	chi "github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Server struct {
	srv  *http.Server
	conf *config.Config

	health health.Service
	file   file.Service
}

func NewServer(conf *config.Config, health health.Service, file file.Service) *Server {
	return &Server{
		srv: &http.Server{
			Addr: ":" + conf.Port,
		},
		conf:   conf,
		health: health,
		file:   file,
	}
}

func (s *Server) Start() error {
	r := chi.NewRouter()
	s.srv.Handler = r

	// healtcheck
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
			r.Route("/file", func(r chi.Router) {
				r.Put("/", s.UploadFile)
				r.Delete("/{post_id}", s.DeleteFilesForPost)
			})
		})
	})

	return s.srv.ListenAndServe()
}

func (s *Server) UploadFile(w http.ResponseWriter, r *http.Request) {
	var f file.FileReq
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("failed to decode file req: %s", err)
		return
	}

	resp, err := s.file.UploadFile(r.Context(), &f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to upload file: %s", err)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to encode file resp: %s", err)
		return
	}
}

func (s *Server) DeleteFilesForPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "post_id")
	convertedId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("failed to parse id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.file.ClearFilesForPost(r.Context(), convertedId); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to delete files for post: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
