package handlerfn

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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

// FeedArticles handler
func FeedArticles(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, feedArticles}
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

const snil = "<nil>"

func (ap articlePost) String() string {
	var t, d, b, tg string
	if ap.Title != nil {
		t = *ap.Title
	} else {
		t = snil
	}
	if ap.Description != nil {
		d = *ap.Description
	} else {
		d = snil
	}
	if ap.Body != nil {
		b = *ap.Body
	} else {
		b = snil
	}
	if ap.Tags != nil {
		tg = fmt.Sprintf("%v", *ap.Tags)
	} else {
		tg = snil
	}
	return fmt.Sprintf("{Title: %v, Description: %v, Body:%s, Tags: %s}", t, d, b, tg)
}

type articleFormPost struct {
	Article articlePost `json:"article"`
}

func (afp articleFormPost) String() string {
	return fmt.Sprintf("{Article: %s}", afp.Article)
}

// UnmarshalJSON implements json decoding
func (ap *articlePost) UnmarshalJSON(data []byte) error {
	type ArticleAlias articlePost
	aux := &struct {
		ArticleAlias
		Tags *[]string `json:"tagList"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return errors.Wrap(err, "articlePost::Unmarshal()")
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

func (afp articleFormPost) Validate() error {
	validations := invalidInputError{}
	if afp.Article.Title == nil || *afp.Article.Title == "" {
		validations.Errs["title"] = []string{"Title is not set"}
	}
	if afp.Article.Description == nil || *afp.Article.Description == "" {
		validations.Errs["description"] = []string{"Description is not set"}
	}
	if afp.Article.Body == nil || *afp.Article.Body == "" {
		validations.Errs["body"] = []string{"Body is not set"}
	}
	if len(validations.Errs) > 0 {
		return validations
	}
	return nil
}

// C - CREATE
func createArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Parse
	artPost := articleFormPost{}
	err := json.NewDecoder(r.Body).Decode(&artPost)
	if err != nil {
		//ae.Logger.Printf("JSON ERR: %+v", err)
		return errors.Wrap(err, "createArticle:: articleFormPost decode()")
	}
	defer r.Body.Close()
	ae.Logger.Printf("Validating %s", artPost.String())
	// Validate
	if errs := artPost.Validate(); errs != nil {
		return errs
	}

	// Get user from request and convert to Profile
	u, err := getUserFromContext(r)
	if err != nil {
		return err
	}
	p := models.ProfileFromUser(*u)

	// Create
	a, err := models.NewArticle(*artPost.Article.Title, *artPost.Article.Description, *artPost.Article.Body, &p)
	if err != nil {
		return errors.Wrap(err, "createArticle: DB.NewArticle() failure")
	}

	// Persist to DB
	if err := ae.DB.CreateArticle(a); err != nil {
		return errors.Wrap(err, "createArticle: DB.CreateArticle() failure")
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
func getArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
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
	// Get Article
	a, err := ae.DB.GetArticle(slug, id)
	if err != nil {
		return errors.Wrap(err, "getArticle:: DB.GetArticle() failure")
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.SingleArtJSONResponse{Article: a})
	return nil
}

// R - READ
func listArticles(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	return respondListOfArticles(ae, w, r, false)
}

// R - READ
func feedArticles(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	return respondListOfArticles(ae, w, r, true)
}

func respondListOfArticles(ae *AppEnvironment, w http.ResponseWriter, r *http.Request, feed bool) error {
	// Parse
	opts := buildQueryOptions(r)
	ae.Logger.Printf("Built Query Opts :\n%s\n%v\n", r.URL, opts)
	// GetUser
	// This should return a nil pointer if the user is not authenticated (in list articles route)
	u, _ := getUserFromContext(r)
	var id uint // id has zero value. Assume no user with ID=0 in DB
	if u != nil {
		id = u.ID
	}

	ae.Logger.Printf("Getting %d records offset by %d", opts.Limit, opts.Offset)
	// Query
	articles, err := ae.DB.ListArticles(opts, id, feed)
	if err != nil {
		return errors.Wrap(err, "respondListOfArticles:: DB.ListArticles failure")
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
		parsedOptions["limit"] = uint(math.Min(float64(v), 100))
	}
	if v, err := strconv.ParseUint(qVals.Get("offset"), 10, 32); err == nil && v > 0 {
		parsedOptions["offset"] = uint(v)
	}
	filters := make(map[string][]string)
	filters["tag"] = qVals["tag"]
	filters["author"] = qVals["author"]
	filters["favorite"] = qVals["favorited"]

	parsedOptions["filters"] = filters
	return models.NewListOptions(parsedOptions)
}

// U - UPDATE
func updateArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Parse
	vars := mux.Vars(r)
	slug := vars["slug"]

	artPost := articleFormPost{}
	err := json.NewDecoder(r.Body).Decode(&artPost)
	if err != nil {
		return invalidInputError{}
	}
	defer r.Body.Close()

	// GetUser
	u, err := getUserFromContext(r)
	if err != nil {
		return err
	}
	// Get Article
	a, err := ae.DB.GetArticle(slug, u.ID)
	if err != nil {
		return errors.Wrap(err, "updateArticle:: DB.GetArticle()")
	}

	if u.ID != a.Author.ID {
		return forbidden{}
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
		return invalidInputError{}
	}
	a.UpdatedAt = time.Now().UTC()

	// Update
	if err := ae.DB.UpdateArticle(a); err != nil {
		return errors.Wrap(err, "updateArticle:: DB.UpdateArticle()")
	}
	if artPost.Article.Tags != nil {
		tags, err := ae.DB.AddTags(a, a.TagList)
		if err != nil {
			return errors.Wrap(err, "updateArticle:: DB.AddTags()")
		}
		a.TagList = tags
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(a)
	return nil
}

// D - DELETE
func deleteArticle(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Parse
	vars := mux.Vars(r)
	slug := vars["slug"]

	// GetUser
	u, err := getUserFromContext(r)
	if err != nil {
		return err
	}
	// Get Article
	a, err := ae.DB.GetArticle(slug, u.ID)
	if err != nil {
		return errors.Wrap(err, "deleteArticle:: DB.GetArticle()")
	}
	if u.ID != a.Author.ID {
		return forbidden{}
	}

	// Delete
	if err := ae.DB.DeleteArticle(a.Slug); err != nil {
		return errors.Wrap(err, "deleteArticle:: DB.DeleteArticle()")
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
	return nil
}
