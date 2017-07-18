package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JackyChiu/realworld-starter-kit/models"
)

type Article struct {
	Slug           string   `json:"slug"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Body           string   `json:"body"`
	Favorited      bool     `json:"favorited"`
	FavoritesCount int      `json:"favoritesCount"`
	TagList        []string `json:"tagList"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
	Author         Author   `json:"user"`
}

type Author struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

type ArticleJSON struct {
	Article `json:"article"`
}

type ArticlesJSON struct {
	Articles      []Article `json:"articles"`
	ArticlesCount int       `json:"articlesCount"`
}

// ArticlesHandler handle /api/articles
func (h *Handler) ArticlesHandler(w http.ResponseWriter, r *http.Request) {
	router := NewRouter(h.Logger)

	// Unprotected routes
	router.AddRoute(
		`articles\/?$`,
		http.MethodGet,
		h.getCurrentUser(h.getArticles),
	)

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)$`,
		http.MethodGet,
		h.getCurrentUser(h.extractArticle(h.getArticle)),
	)

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)\/comments$`,
		http.MethodGet,
		h.getCurrentUser(h.extractArticle(h.getComments)),
	)

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)\/comments\/(?P<commentID>[0-9]+)$`,
		http.MethodGet,
		h.getCurrentUser(h.extractArticle(h.getComment)),
	)

	// Protected routes
	router.AddRoute(
		`articles\/?$`,
		http.MethodPost,
		h.getCurrentUser(h.authorize(h.createArticle)),
	)

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)$`,
		http.MethodPut,
		h.getCurrentUser(h.authorize(h.extractArticle(h.updateArticle))),
	)

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)$`,
		http.MethodDelete,
		h.getCurrentUser(h.authorize(h.extractArticle(h.deleteArticle))),
	)

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)\/favorite$`,
		http.MethodPost,
		h.getCurrentUser(h.authorize(h.extractArticle(h.favoriteArticle))),
	)

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)\/favorite$`,
		http.MethodDelete,
		h.getCurrentUser(h.authorize(h.extractArticle(h.unFavoriteArticle))),
	)

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)\/comments$`,
		http.MethodPost,
		h.getCurrentUser(h.authorize(h.extractArticle(h.addComment))),
	)

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)\/comments\/(?P<commentID>[0-9]+)$`,
		http.MethodDelete,
		h.getCurrentUser(h.authorize(h.extractArticle(h.deleteComment))),
	)

	router.ServeHTTP(w, r)
}

func (h *Handler) extractArticle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if slug, ok := ctx.Value("slug").(string); ok {
			a, err := h.DB.GetArticle(slug)

			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			if a != nil {
				ctx := r.Context()
				ctx = context.WithValue(ctx, fetchedArticleKey, a)
				r = r.WithContext(ctx)
			}
		}
		next(w, r)
	}
}

func (h *Handler) getArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	a := ctx.Value(fetchedArticleKey).(*models.Article)
	u := ctx.Value(currentUserKey).(*models.User)

	articleJSON := ArticleJSON{
		Article: h.buildArticleJSON(a, u),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(articleJSON)
}

// getArticles handle GET /api/articles
func (h *Handler) getArticles(w http.ResponseWriter, r *http.Request) {
	var err error
	var articles = []models.Article{}
	var articlesJSON ArticlesJSON

	query := h.DB.GetAllArticles()

	r.ParseForm()

	query = h.DB.FilterByTag(query, r.Form)
	query = h.DB.FilterAuthoredBy(query, r.Form)
	query = h.DB.FilterFavoritedBy(query, r.Form)

	query.Count(&articlesJSON.ArticlesCount)

	query = h.DB.Limit(query, r.Form)
	query = h.DB.Offset(query, r.Form)

	err = query.Find(&articles).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(articles) == 0 {
		json.NewEncoder(w).Encode(ArticlesJSON{})
		return
	}

	var u = r.Context().Value(currentUserKey).(*models.User)

	for i := range articles {
		a := &articles[i]
		articlesJSON.Articles = append(articlesJSON.Articles, h.buildArticleJSON(a, u))
	}

	json.NewEncoder(w).Encode(articlesJSON)
}

// createArticle handle POST /api/articles
func (h *Handler) createArticle(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Article struct {
			Title       string   `json:"title"`
			Description string   `json:"description"`
			Body        string   `json:"body"`
			TagList     []string `json:"tagList"`
		} `json:"article"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	u := r.Context().Value(currentUserKey).(*models.User)

	a := models.NewArticle(body.Article.Title, body.Article.Description, body.Article.Body, u)

	if valid, errs := a.IsValid(); !valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		errorJSON := errorJSON{errs}
		json.NewEncoder(w).Encode(errorJSON)
		return
	}

	for _, tagName := range body.Article.TagList {
		tag, _ := h.DB.FindTagOrInit(tagName)
		a.Tags = append(a.Tags, tag)
	}

	if err := h.DB.CreateArticle(a); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	articleJSON := ArticleJSON{
		Article: h.buildArticleJSON(a, u),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(articleJSON)
}

// updateArticle handle PUT /api/articles/:slug
func (h *Handler) updateArticle(w http.ResponseWriter, r *http.Request) {
	var err error
	a := r.Context().Value(fetchedArticleKey).(*models.Article)
	u := r.Context().Value(currentUserKey).(*models.User)

	if !a.IsOwnedBy(u.Username) {
		err = fmt.Errorf("You don't have the permission to edit this article")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	var body map[string]map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	defer r.Body.Close()

	if _, present := body["article"]; !present {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var article map[string]interface{}

	article = body["article"]

	if title, present := article["title"]; present {
		a.Title = title.(string)
	}

	if description, present := article["description"]; present {
		a.Description = description.(string)
	}

	if body, present := article["body"]; present {
		a.Body = body.(string)
	}

	if valid, errs := a.IsValid(); !valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		errorJSON := errorJSON{errs}
		json.NewEncoder(w).Encode(errorJSON)
		return
	}

	if err := h.DB.SaveArticle(a); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	articleJSON := ArticleJSON{
		Article: h.buildArticleJSON(a, u),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(articleJSON)
}

// deleteArticle handle DELETE /api/articles/:slug
func (h *Handler) deleteArticle(w http.ResponseWriter, r *http.Request) {
	var err error
	a := r.Context().Value(fetchedArticleKey).(*models.Article)
	u := r.Context().Value(currentUserKey).(*models.User)

	if !a.IsOwnedBy(u.Username) {
		err = fmt.Errorf("You don't have the permission to delete this article")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	err = h.DB.DeleteArticle(a)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// favoriteArticle handle POST /api/articles/:slug/favorite
func (h *Handler) favoriteArticle(w http.ResponseWriter, r *http.Request) {
	a := r.Context().Value(fetchedArticleKey).(*models.Article)
	u := r.Context().Value(currentUserKey).(*models.User)

	err := h.DB.FavoriteArticle(u, a)

	articleJSON := ArticleJSON{
		Article: h.buildArticleJSON(a, u),
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(articleJSON)
}

// unFavoriteArticle handle DELETE /api/articles/:slug/favorite
func (h *Handler) unFavoriteArticle(w http.ResponseWriter, r *http.Request) {
	a := r.Context().Value(fetchedArticleKey).(*models.Article)
	u := r.Context().Value(currentUserKey).(*models.User)

	err := h.DB.UnfavoriteArticle(u, a)

	articleJSON := ArticleJSON{
		Article: h.buildArticleJSON(a, u),
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(articleJSON)
}

func (h *Handler) buildArticleJSON(a *models.Article, u *models.User) Article {
	following := false
	favorited := false

	if (u != &models.User{}) {
		following = h.DB.IsFollowing(u.ID, a.User.ID)
		favorited = h.DB.IsFavorited(u.ID, a.ID)
	}

	article := Article{
		Slug:           a.Slug,
		Title:          a.Title,
		Description:    a.Description,
		Body:           a.Body,
		Favorited:      favorited,
		FavoritesCount: a.FavoritesCount,
		CreatedAt:      a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      a.UpdatedAt.Format(time.RFC3339),
		Author: Author{
			Username:  a.User.Username,
			Bio:       a.User.Bio,
			Image:     a.User.Image,
			Following: following,
		},
	}

	for _, t := range a.Tags {
		article.TagList = append(article.TagList, t.Name)
	}

	return article
}
