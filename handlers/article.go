package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/JackyChiu/realworld-starter-kit/models"
)

type Article struct {
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Body           string    `json:"body"`
	Favorited      bool      `json:"favorited"`
	FavoritesCount int       `json:"favoritesCount"`
	TagsList       []string  `json:"tagsList"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Author         struct {
		Username  string `json:"username"`
		Bio       string `json:"bio"`
		Image     string `json:"image"`
		Following bool   `json:"following"`
	} `json:"user"`
}

type ArticleJSON struct {
	Article `json:"article"`
}

type ArticlesJSON struct {
	Articles      []Article `json:"articles"`
	ArticlesCount int       `json:"articlesCount"`
}

type Author struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

const (
	contextKeyCurrentUser = contextKey("current_user")
	contextKeyArticle     = contextKey("article")
	contextKeyLoggedIn    = contextKey("logged_in")
)

// ArticlesHandler handle /api/articles
func (h *Handler) ArticlesHandler(w http.ResponseWriter, r *http.Request) {
	//h.Logger.Println(r.Method, r.URL.Path)

	// Unprotected routes
	router := NewRouter(h.Logger)
	router.AddRoute(`articles\/?$`, "GET", h.getArticles)

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)$`,
		"GET", h.getCurrentUser(h.extractArticle(h.getArticle)))

	// Protected routes
	router.AddRoute(
		`articles\/?$`,
		"POST", h.getCurrentUser(h.authorize(h.createArticle)))

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)$`,
		"PUT", h.getCurrentUser(h.authorize(h.updateArticle)))

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)$`,
		"DELETE", h.getCurrentUser(h.authorize(h.deleteArticle)))

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)\/favorite$`,
		"POST", h.getCurrentUser(h.authorize(h.favoriteArticle)))

	router.AddRoute(
		`articles\/(?P<slug>[0-9a-zA-Z\-]+)\/favorite$`,
		"DELETE", h.getCurrentUser(h.authorize(h.unFavoriteArticle)))

	router.DebugMode(true)

	router.Dispatch(w, r)
}

func (h *Handler) getCurrentUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var u *models.User

		if claim, _ := h.JWT.CheckRequest(r); claim != nil {
			u, _ = h.DB.FindUserByUsername(claim.Username)
			h.Logger.Println(contextKeyCurrentUser, u)
		}

		if u != nil {
			ctx = context.WithValue(ctx, contextKeyCurrentUser, u)
			ctx = context.WithValue(ctx, contextKeyLoggedIn, true)
		} else {
			ctx = context.WithValue(ctx, contextKeyCurrentUser, &models.User{})
			ctx = context.WithValue(ctx, contextKeyLoggedIn, false)
		}

		r = r.WithContext(ctx)
		next(w, r)
	}
}

func (h *Handler) extractArticle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if slug, ok := ctx.Value("slug").(string); ok {
			a, _ := h.DB.GetArticle(slug)
			if a != nil {
				ctx := r.Context()
				ctx = context.WithValue(ctx, contextKeyArticle, a)
				r = r.WithContext(ctx)
			}
		}
		next(w, r)
	}
}

func (h *Handler) authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if loggedIn, _ := r.Context().Value(contextKeyLoggedIn).(bool); !loggedIn {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// getArticle handle GET /api/articles/:slug
func (h *Handler) getArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slug, _ := ctx.Value("slug").(string)

	h.Logger.Println("Slug: ", slug)

	a, ok := ctx.Value(contextKeyArticle).(*models.Article)

	if !ok {
		http.Error(w, fmt.Errorf("Article not found with slug %s", slug).Error(), http.StatusNotFound)
		return
	}

	u, _ := ctx.Value(contextKeyCurrentUser).(*models.User)

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

	query := h.DB.GetAllArticles()

	r.ParseForm()
	queryParams := r.Form

	if limit, ok := queryParams["limit"]; ok {
		query = query.Limit(limit[0])
	} else {
		query = query.Limit(20)
	}

	if offset, ok := queryParams["offset"]; ok {
		query = query.Offset(offset[0])
	} else {
		query = query.Offset(0)
	}

	err = query.Find(&articles).Error

	if tags, ok := queryParams["tag"]; ok {
		articles, err = h.DB.GetAllArticlesWithTag(tags[0])
	}

	if authorBy, ok := queryParams["author"]; ok {
		articles, err = h.DB.GetAllArticlesAuthoredBy(authorBy[0])
	}

	if favoritedBy, ok := queryParams["favorited"]; ok {
		articles, err = h.DB.GetAllArticlesFavoritedBy(favoritedBy[0])
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(articles) == 0 {
		json.NewEncoder(w).Encode(ArticlesJSON{})
		return
	}

	var u = &models.User{}
	if claim, _ := h.JWT.CheckRequest(r); claim != nil {
		u, err = h.DB.FindUserByUsername(claim.Username)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var articlesJSON ArticlesJSON
	for i := range articles {
		a := &articles[i]
		articlesJSON.Articles = append(articlesJSON.Articles, h.buildArticleJSON(a, u))
	}

	articlesJSON.ArticlesCount = len(articles)

	json.NewEncoder(w).Encode(articlesJSON)
}

// createArticle handle POST /api/articles
func (h *Handler) createArticle(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Article struct {
			Title       string   `json:"title"`
			Description string   `json:"description"`
			Body        string   `json:"body"`
			TagsList    []string `json:"tagsList"`
		} `json:"article"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	u, _ := r.Context().Value(contextKeyCurrentUser).(*models.User)

	if u == nil {
		http.Error(w, fmt.Errorf("User not found").Error(), http.StatusNotFound)
		return
	}

	a := models.NewArticle(body.Article.Title, body.Article.Description, body.Article.Body, u)

	for _, tagName := range body.Article.TagsList {
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
	claim, err := h.JWT.CheckRequest(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()

	slug := path.Base(r.URL.Path)

	a, err := h.DB.GetArticle(slug)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !a.CanUpdate(claim.Username) {
		http.Error(w, fmt.Errorf("You don't have the authorization to edit this article").Error(), http.StatusUnauthorized)
		return
	}

	u, err := h.DB.FindUserByUsername(claim.Username)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var body map[string]map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var article map[string]interface{}

	if _, present := body["article"]; !present {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	article = body["article"]

	if title, present := article["title"]; present && a.Title != title {
		a.Title = title.(string)
	}

	if description, present := article["description"]; present && a.Description != description {
		a.Description = description.(string)
	}

	if body, present := article["body"]; present && a.Body != body {
		a.Body = body.(string)
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

}

// favoriteArticle handle POST /api/articles/:slug/favorite
func (h *Handler) favoriteArticle(w http.ResponseWriter, r *http.Request) {
	claim, err := h.JWT.CheckRequest(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	slug := strings.Split(r.URL.Path, "/")[3]
	h.Logger.Println("Slug: ", slug)

	a, err := h.DB.GetArticle(slug)

	if err != nil {
		h.Logger.Println("Err: ", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	u, err := h.DB.FindUserByUsername(claim.Username)

	if err != nil {
		h.Logger.Println("Err: ", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = h.DB.FavoriteArticle(u.ID, a.ID)

	articleJSON := ArticleJSON{
		Article: h.buildArticleJSON(a, u),
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		h.Logger.Println("Err: ", err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(articleJSON)
}

// unFavoriteArticle handle DELETE /api/articles/:slug/favorite
func (h *Handler) unFavoriteArticle(w http.ResponseWriter, r *http.Request) {
	claim, err := h.JWT.CheckRequest(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	slug := strings.Split(r.URL.Path, "/")[3]
	h.Logger.Println("Slug: ", slug)

	a, err := h.DB.GetArticle(slug)

	if err != nil {
		h.Logger.Println("Err: ", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	u, err := h.DB.FindUserByUsername(claim.Username)

	if err != nil {
		h.Logger.Println("Err: ", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = h.DB.UnfavoriteArticle(u.ID, a.ID)

	articleJSON := ArticleJSON{
		Article: h.buildArticleJSON(a, u),
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		h.Logger.Println("Err: ", err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(articleJSON)
}

func (h *Handler) buildArticleJSON(a *models.Article, u *models.User) Article {
	following := false
	favorited := false
	//h.Logger.Println("\nArticle: ", a, "\nAuthor: ", a.User, "\nLogged User: ", u)
	if (u != &models.User{}) {
		following = h.DB.IsFollowing(u.ID, a.User.ID)
		favorited = h.DB.IsFavorited(u.ID, a.ID)
		h.Logger.Println("isFavorited: ", favorited)
	}

	article := Article{
		Slug:           a.Slug,
		Title:          a.Title,
		Description:    a.Description,
		Body:           a.Body,
		Favorited:      favorited,
		FavoritesCount: a.FavoritesCount,
		Author: Author{
			Username:  a.User.Username,
			Bio:       a.User.Bio,
			Image:     a.User.Image,
			Following: following,
		},
	}

	for _, t := range a.Tags {
		article.TagsList = append(article.TagsList, t.Name)
	}

	return article
}
