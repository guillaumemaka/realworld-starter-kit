package handlerfn

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/gorilla/mux"
)

// CreateArticle handler
func CreateArticle(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, createArticle}
}

// GetArticle handler
func GetArticle(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, getArticle}
}

// ListArticles handler
func ListArticles(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, listArticles}
}

// UpdateArticle handler
func UpdateArticle(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, updateArticle}
}

// DeleteArticle handler
func DeleteArticle(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, deleteArticle}
}

type articlePost struct {
	Title       *string       `json:"title"`
	Description *string       `json:"description"`
	Body        *string       `json:"body"`
	Tags        *[]models.Tag `json:"tagList"`
}
type articleFormPost struct {
	Article articlePost `json:"article"`
}

// UnmarshalJSON implements json decoding
func (ap *articlePost) UnmarshalJSON(data []byte) error {
	type ArticleAlias articlePost
	aux := &struct {
		ArticleAlias
		Tags *[]string `json:"tagList"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return nil
	}
	ap.Body = aux.Body
	ap.Title = aux.Title
	ap.Description = aux.Description

	if aux.Tags == nil {
		return nil
	}
	list := make([]models.Tag, len(*aux.Tags))
	for i, v := range *aux.Tags {
		list[i] = models.Tag(v)
	}
	ap.Tags = &list
	return nil
}

func (ap articleFormPost) Validate() []error {
	var errs []error
	if ap.Article.Title == nil || *ap.Article.Title == "" {
		errs = append(errs, fmt.Errorf("Title is not set"))
	}
	if ap.Article.Description == nil || *ap.Article.Description == "" {
		errs = append(errs, fmt.Errorf("Description is not set"))
	}
	if ap.Article.Body == nil || *ap.Article.Body == "" {
		errs = append(errs, fmt.Errorf("Body is not set"))
	}
	return errs
}

// C - CREATE
func createArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Parse
	artPost := articleFormPost{}
	err := json.NewDecoder(r.Body).Decode(&artPost)
	if err != nil {
		return &AppError{Err: []error{err}}
	}
	defer r.Body.Close()

	// Validate
	if errs := artPost.Validate(); len(errs) > 0 {
		return &AppError{Err: errs}
	}
	// Get user from request and convert to Profile
	u, err := getUserFromContext(r)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}
	if u == nil { // Really need to get a user
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{fmt.Errorf("Could not retrieve User")}}
	}
	p := models.ProfileFromUser(*u)

	// Create
	a, err := models.NewArticle(*artPost.Article.Title, *artPost.Article.Description, *artPost.Article.Body, &p)
	if err != nil {
		return &AppError{Err: []error{err}}
	}

	// Persist to DB
	if err := ae.DB.CreateArticle(a); err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}

	// Tags
	if artPost.Article.Tags != nil {
		tags, err := ae.DB.AddTags(a, *artPost.Article.Tags)
		if err != nil {
			// Let's just log for the moment. Article has been created.
			ae.Logger.Printf("Error adding tags\n%s\n%s", tags, err)
		} else {
			a.TagList = tags
		}
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SingleArtJSONResponse{Article: a})
	return nil

}

// R - READ
func getArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Parse
	vars := mux.Vars(r)
	slug := vars["slug"]

	// GetUser
	// This should return a nil pointer if the user is not authenticated
	u, err := getUserFromContext(r)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}
	var id uint
	if u != nil {
		id = u.ID
	}
	// Get Article
	a, err := ae.DB.GetArticle(slug, id)
	if err != nil {
		return &AppError{Err: []error{err}}
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SingleArtJSONResponse{Article: a})
	return nil
}

// R - READ
func listArticles(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Parse
	opts := buildQueryOptions(r)

	// GetUser
	// This should return a nil pointer if the user is not authenticated
	u, err := getUserFromContext(r)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}
	var id uint
	if u != nil {
		id = u.ID
	}

	ae.Logger.Printf("Getting %d records offset by %d", opts.Limit, opts.Offset)
	// Query
	articles, err := ae.DB.ListArticles(opts, id)
	if err != nil {
		return &AppError{Err: []error{err}}
	}
	ae.Logger.Printf("Got %d records", len(articles))

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.MultipleArtJSONResponse{Articles: articles, ArticlesCount: len(articles)})
	return nil
}

func buildQueryOptions(r *http.Request) models.ListArticleOptions {
	qVals := r.URL.Query()
	parsedOptions := make(map[string]interface{})
	if v, err := strconv.ParseUint(qVals.Get("limit"), 10, 32); err == nil && v > 0 {
		parsedOptions["limit"] = uint(v)
	}
	if v, err := strconv.ParseUint(qVals.Get("offset"), 10, 32); err == nil && v > 0 {
		parsedOptions["offset"] = uint(v)
	}
	filters := make(map[string][]string)
	filters["tag"] = qVals["tag"]
	filters["author"] = qVals["author"]
	filters["favorite"] = qVals["favorite"]

	parsedOptions["filters"] = filters
	log.Println(parsedOptions)
	return models.NewListOptions(parsedOptions)
}

// U - UPDATE
func updateArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Parse
	vars := mux.Vars(r)
	slug := vars["slug"]

	artPost := articleFormPost{}
	err := json.NewDecoder(r.Body).Decode(&artPost)
	if err != nil {
		return &AppError{Err: []error{err}}
	}
	defer r.Body.Close()

	// GetUser
	// This should return a nil pointer if the user is not authenticated
	u, err := getUserFromContext(r)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}
	// Get Article
	a, err := ae.DB.GetArticle(slug, u.ID)
	if err != nil {
		return &AppError{Err: []error{err}}
	}
	// Update attributes
	changed := false
	if artPost.Article.Title != nil && len(*artPost.Article.Title) > 0 {
		a.Title = *artPost.Article.Title
		a.CreateSlug()
		changed = true
	}
	if artPost.Article.Description != nil && len(*artPost.Article.Description) > 0 {
		a.Description = *artPost.Article.Description
		changed = true
	}
	if artPost.Article.Body != nil && len(*artPost.Article.Body) > 0 {
		a.Body = *artPost.Article.Body
		changed = true
	}
	if artPost.Article.Tags != nil {
		a.TagList = *artPost.Article.Tags
		changed = true
	}

	if !changed {
		return &AppError{Err: []error{fmt.Errorf("No relevant fields sent")}}
	}
	a.UpdatedAt = time.Now().UTC()

	// Update
	if err := ae.DB.UpdateArticle(a); err != nil {
		return &AppError{Err: []error{err}}
	}
	if artPost.Article.Tags != nil {
		tags, err := ae.DB.AddTags(a, a.TagList)
		if err != nil {
			return &AppError{Err: []error{fmt.Errorf("Error adding tags\n%s\n%s", tags, err)}}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a)
	return nil
}

// D - DELETE
func deleteArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
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
	if u.Username != a.Author.Username {
		return &AppError{StatusCode: http.StatusForbidden, Err: []error{fmt.Errorf("You may only delete articles you are the author of")}}
	}

	// Delete
	if err := ae.DB.DeleteArticle(a.Slug); err != nil {
		return &AppError{Err: []error{err}}
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
	return nil
}
