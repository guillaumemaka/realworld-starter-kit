package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/JackyChiu/realworld-starter-kit/models"
)

type Comment struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Author    Author `json:"author"`
}

type CommentJSON struct {
	Comment `json:"comment"`
}

type CommentsJSON struct {
	Comments []Comment `json:"comments"`
}

type commentBody struct {
	Comment struct {
		Body string `json:"body"`
	} `json:"comment"`
}

func (h *Handler) getComments(w http.ResponseWriter, r *http.Request) {
	a := r.Context().Value(fetchedArticleKey).(*models.Article)
	u := r.Context().Value(currentUserKey).(*models.User)

	var comments []models.Comment
	err := h.DB.GetComments(a, &comments)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var commentsJSON = CommentsJSON{}
	for _, comment := range comments {
		commentsJSON.Comments = append(commentsJSON.Comments, h.buildCommentJSON(&comment, u))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commentsJSON)
}

func (h *Handler) getComment(w http.ResponseWriter, r *http.Request) {
	commentID, _ := strconv.Atoi(r.Context().Value("commentID").(string))
	u := r.Context().Value(currentUserKey).(*models.User)

	var comment models.Comment
	err := h.DB.GetComment(commentID, &comment)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var commentJSON = CommentJSON{
		Comment: h.buildCommentJSON(&comment, u),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commentJSON)
}

func (h *Handler) addComment(w http.ResponseWriter, r *http.Request) {
	a := r.Context().Value(fetchedArticleKey).(*models.Article)
	u := r.Context().Value(currentUserKey).(*models.User)

	var commentBody commentBody
	if err := json.NewDecoder(r.Body).Decode(&commentBody); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	defer r.Body.Close()

	c, errs := models.NewComment(a, u, commentBody.Comment.Body)

	if errs != nil {
		errorJSON := errorJSON{errs}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(errorJSON)
		return
	}

	err := h.DB.CreateComment(c)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	commentJSON := CommentJSON{
		Comment: h.buildCommentJSON(c, u),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(commentJSON)
}

func (h *Handler) deleteComment(w http.ResponseWriter, r *http.Request) {
	commentID, _ := strconv.Atoi(r.Context().Value("commentID").(string))
	u := r.Context().Value(currentUserKey).(*models.User)

	var comment = models.Comment{}
	err := h.DB.GetComment(commentID, &comment)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if canDelete := comment.CanBeDeletedBy(u); !canDelete {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	err = h.DB.DeleteComment(&comment)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) buildCommentJSON(c *models.Comment, u *models.User) Comment {
	following := false

	if (u != &models.User{}) {
		following = h.DB.IsFollowing(u.ID, c.User.ID)
	}

	return Comment{
		ID:        c.ID,
		Body:      c.Body,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
		Author: Author{
			Username:  c.User.Username,
			Bio:       c.User.Bio,
			Image:     c.User.Image,
			Following: following,
		},
	}
}
