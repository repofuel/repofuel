package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/repofuel/repofuel/accounts/internal/entity"
	"github.com/repofuel/repofuel/accounts/pkg/jwtauth"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	authCheck *jwtauth.AuthCheck
	usersDB   entity.UsersDataSource
}

func NewHandler(authCheck *jwtauth.AuthCheck, usersDB entity.UsersDataSource) *Handler {
	return &Handler{authCheck: authCheck, usersDB: usersDB}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Route("/users/{user_id}", func(r chi.Router) {
		r.Use(h.authCheck.Middleware, permission.OnlyServiceAccounts)
		r.Get("/", h.UserByID)
		r.Get("/providers/{provider_id}/token", h.ProviderToken)

	})

	r.Get("/me", h.Me)

	return r
}

func (h *Handler) ProviderToken(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider_id")
	userID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "user_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := h.usersDB.ProviderToken(r.Context(), userID, providerID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, err := h.authCheck.ClaimsFromToken(jwtauth.StripBearerToken(r.Header.Get("Authorization")))
	if err != nil || claims.UserInfo == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := h.usersDB.Find(r.Context(), claims.UserID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Println(err)
	}

}

func (h *Handler) UserByID(w http.ResponseWriter, r *http.Request) {
	userID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "user_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.usersDB.Find(r.Context(), userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Println(err)
	}

}
