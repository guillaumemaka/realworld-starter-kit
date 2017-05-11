package handlerfn

import (
	"encoding/json"
	"net/http"

	"github.com/chilledoj/realworld-starter-kit/models"
)

// GetTags handler
func GetTags(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, getTags}
}

func getTags(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	tags, err := ae.DB.GetAllTags()
	if err != nil {
		return &AppError{Err: []error{err}}
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.TagJSONResponse{Tags: tags})
	return nil
}
