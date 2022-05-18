package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ml/internal/entity"
	"github.com/repofuel/repofuel/ml/pkg/ml"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	mlServer *ml.ModelServer
	auth     *jwtauth.AuthCheck
	modelsDB entity.ModelDataSource
}

func NewHandler(auth *jwtauth.AuthCheck, mlServer *ml.ModelServer, m entity.ModelDataSource) *Handler {
	return &Handler{
		mlServer: mlServer,
		auth:     auth,
		modelsDB: m,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(h.auth.Middleware)

	r.With(permission.OnlyServiceAccounts).Get("/repositories/{repo_id}/jobs/{job}/prediction", h.Prediction)
	r.With(permission.OnlyAdmin).Get("/repositories/{repo_id}/models", h.ModelsList)

	return r
}

var findModelsOpts = options.Find().SetLimit(50).SetSort(bson.M{"created_at": -1})

func (h *Handler) ModelsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repoId := chi.URLParam(r, "repo_id")
	itr, err := h.modelsDB.FindRepoModels(ctx, repoId, findModelsOpts)
	if err != nil {
		internalError(w, err)
		return
	}

	encoder := json.NewEncoder(w)
	err = itr.ForEach(ctx, func(model *entity.Model) error {
		return encoder.Encode(model)
	})
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) Prediction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repoId := chi.URLParam(r, "repo_id")
	currentJob := chi.URLParam(r, "job")
	oldestJob := r.FormValue("oldest_job")

	log.Println("received prediction request for repo: ", repoId)

	if oldestJob == "" {
		oldestJob = currentJob
	}

	res, err := h.mlServer.Predict(ctx, repoId, oldestJob, currentJob)
	if err != nil {
		internalError(w, err)
		return
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println(err)
	}
}

// todo: should be extracted and reused in other services
func internalError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	_, wErr := fmt.Fprintf(w, "{\"message\": %q}", err.Error())
	if wErr != nil {
		log.Println(err)
	}
}
