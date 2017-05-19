package handlerfn

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// AddComment route
func AddComment(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, addComment}
}

// GetComments route
func GetComments(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, getComments}
}

// DeleteComment route
func DeleteComment(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, deleteComment}
}

type commentJSONPost struct {
	Comment map[string]string `json:"comment"`
}

func addComment(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// GetUser
	// This should return a nil pointer if the user is not authenticated
	u, err := getUserFromContext(r)
	if err != nil {
		return unAuthorised{}
	}
	p := models.ProfileFromUser(*u)

	// Parse
	postedComment := commentJSONPost{}
	err = json.NewDecoder(r.Body).Decode(&postedComment)
	if err != nil {
		return errors.Wrap(err, "addComment:: jsonDecode")
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	slug := vars["slug"]

	// Validate
	body, ok := postedComment.Comment["body"]
	if !ok || body == "" {
		return invalidInputError{
			Errs: map[string][]string{
				"body": []string{"no comment provided"},
			},
		}
	}
	comment := models.NewComment(body, &p)
	if err := ae.DB.AddComment(comment, slug, u.ID); err != nil {
		return errors.Wrap(err, "addComment:: DB.AddComment()")
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SingleComJSONResponse{Comment: comment})
	return nil
}
func getComments(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Parse
	vars := mux.Vars(r)
	slug := vars["slug"]

	// GetUser
	// This should return a nil pointer if the user is not authenticated
	u, _ := getUserFromContext(r)
	var id uint
	if u != nil {
		id = u.ID
	}

	// Get Comments
	comments, err := ae.DB.GetComments(slug, id)
	if err != nil {
		return errors.Wrap(err, "getComments:: DB.GetComments()")
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.MultipleComJSONResponse{Comments: comments})
	return nil
}
func deleteComment(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// GetUser
	// This should return a nil pointer if the user is not authenticated
	u, err := getUserFromContext(r)
	if err != nil {
		return unAuthorised{}
	}

	// Parse
	vars := mux.Vars(r)
	slug := vars["slug"]
	commentID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil || commentID == 0 {
		return invalidInputError{
			Errs: map[string][]string{
				"comment": []string{"invalid id"},
			},
		}
	}

	comment, err := ae.DB.GetCommentByID(uint(commentID), slug, u.ID)
	if err != nil {
		return errors.Wrap(err, "deleteComment:: DB.GetCommentByID()")
	}
	if u.ID != comment.Author.ID {
		return forbidden{}
	}
	if err := ae.DB.DeleteComment(comment); err != nil {
		return errors.Wrap(err, "deleteComment:: DB.DeleteComment()")
	}

	return nil
}
