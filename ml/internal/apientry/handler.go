package apientry

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/repofuel/repofuel/ml/internal/rest"
	"net/http"
)

type Handler struct {
	rest *rest.Handler
}

func NewHandler(rest *rest.Handler) *Handler {
	return &Handler{
		rest: rest,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
	)

	r.Mount("/ai", h.rest.Routes())

	r.Mount("/debug", middleware.Profiler())

	return r
}
