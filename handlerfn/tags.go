package handlerfn

import (
	"encoding/json"
	"net/http"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/pkg/errors"
)

// GetTags handler
func GetTags(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, getTags}
}

func getTags(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	tags, err := ae.DB.GetAllTags()
	if err != nil {
		return errors.Wrap(err, "getTags:: DB.GetTags()")
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.TagJSONResponse{Tags: tags})
	return nil
}
