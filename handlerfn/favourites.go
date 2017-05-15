package handlerfn

import (
	"encoding/json"
	"net/http"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/gorilla/mux"
)

// FavouriteArticle handler
func FavouriteArticle(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, favArticle}
}

// UnfavouriteArticle handler
func UnfavouriteArticle(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, unfavArticle}
}

// POST /articles/{slug}/favourite
func favArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Parse
	vars := mux.Vars(r)
	slug := vars["slug"]

	// GetUser
	u, err := getUserFromContext(r)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}

	// Get Article
	a, err := ae.DB.GetArticle(slug, u.ID)
	if err != nil {
		return &AppError{Err: []error{err}}
	}
	// Favourite Article
	if err := ae.DB.FavouriteArticle(a, u); err != nil {
		return &AppError{Err: []error{err}}
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SingleArtJSONResponse{Article: a})
	return nil
}

// DELETE /articles/{slug}/favourite
func unfavArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Parse
	vars := mux.Vars(r)
	slug := vars["slug"]

	// GetUser
	u, err := getUserFromContext(r)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}

	// Get Article
	a, err := ae.DB.GetArticle(slug, u.ID)
	if err != nil {
		return &AppError{Err: []error{err}}
	}
	// Unfavourite Article
	if err := ae.DB.UnfavouriteArticle(a, u); err != nil {
		return &AppError{Err: []error{err}}
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SingleArtJSONResponse{Article: a})
	return nil
}
