package handlerfn

import (
	"encoding/json"
	"net/http"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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
func favArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Parse
	vars := mux.Vars(r)
	slug := vars["slug"]
	ae.Logger.Println(slug)
	// GetUser
	u, err := getUserFromContext(r)
	if err != nil {
		return errors.Wrap(err, "favArticle:: getUserFromContext()")
	}

	// Get Article
	a, err := ae.DB.GetArticle(slug, u.ID)
	if err != nil {
		return errors.Wrap(err, "favArticle:: DB.GetArticle()")
	}
	// Favourite Article
	ae.Logger.Printf("Trying to favourite article %d for user %d", a.ID, u.ID)
	if err := ae.DB.FavouriteArticle(a, u); err != nil {
		return errors.Wrap(err, "favArticle:: DB.FavouriteArticle()")
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SingleArtJSONResponse{Article: a})
	return nil
}

// DELETE /articles/{slug}/favourite
func unfavArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Parse
	vars := mux.Vars(r)
	slug := vars["slug"]

	// GetUser
	u, err := getUserFromContext(r)
	if err != nil {
		return errors.Wrap(err, "unfavArticle:: getUserFromContext()")
	}
	if u == nil {
		return unAuthorised{}
	}
	// Get Article
	a, err := ae.DB.GetArticle(slug, u.ID)
	if err != nil {
		return errors.Wrap(err, "unfavArticle:: DB.GetArticle()")
	}
	// Unfavourite Article
	if err := ae.DB.UnfavouriteArticle(a, u); err != nil {
		return errors.Wrap(err, "unfavArticle:: DB.UnfavouriteArticle()")
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SingleArtJSONResponse{Article: a})
	return nil
}
