package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/JackyChiu/realworld-starter-kit/auth"
	"github.com/JackyChiu/realworld-starter-kit/models"
)

type articleEntity struct {
	Article article `json:"article"`
}

type article struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagsList    []string `json:"tagsList"`
}

func NewTestHandler() *Handler {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	db, err := models.NewDB("sqlite3", "../conduit_test.db")
	if err != nil {
		logger.Fatal(err)
	}

	db.InitSchema()

	db.Seed()

	j := auth.NewJWT()
	return New(db, j, logger)
}

var (
	h *Handler
)

func TestMain(m *testing.M) {
	h = NewTestHandler()
	os.Exit(m.Run())
}

func TestArticlesHandler_Index(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/articles", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(h.ArticlesHandler)

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var articles ArticlesJSON
	json.NewDecoder(recorder.Body).Decode(&articles)
	expected := 5
	if len(articles.Articles) != expected {
		t.Errorf("should return a list of articles: got %v want %v", len(articles.Articles), expected)
	}

	expectedUsername := "user1"
	if article1 := articles.Articles[0]; article1.Author.Username != expectedUsername {
		t.Errorf("should return the correct author username: got %v want %v", article1.Author.Username, expectedUsername)
	}

	expectedUsername = "user2"
	if article2 := articles.Articles[1]; article2.Author.Username != expectedUsername {
		t.Errorf("should return the correct author username: got %v want %v", article2.Author.Username, expectedUsername)
	}
}

func TestArticlesHandler_Read(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/articles/title-5", nil)

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(h.ArticlesHandler)

	handler.ServeHTTP(recorder, req)
	var article ArticleJSON
	json.NewDecoder(recorder.Body).Decode(&article)

	if article.Article.Title != "Title 5" {
		t.Errorf("should return the correct article title: got %v want %v", article.Article.Title, "Title 5")
	}

	if article.Article.Description != "Description 5" {
		t.Errorf("should return the correct article description: got %v want %v", article.Article.Description, "Description 5")
	}

	if article.Article.Body != "Body 5" {
		t.Errorf("should return the correct article boy: got %v want %v", article.Article.Body, "Body 5")
	}

	if article.Article.Author.Username != "user1" {
		t.Errorf("should return the correct article author username: got %v want %v", article.Article.Author.Username, "user1")
	}
}

func TestArticlesHandler_FilterByTag(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/articles?tag=tag1", nil)

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(h.ArticlesHandler)
	handler.ServeHTTP(recorder, req)

	var articles ArticlesJSON
	json.NewDecoder(recorder.Body).Decode(&articles)

	if len(articles.Articles) != 1 {
		t.Errorf("should return the correct number article: got %v want %v", len(articles.Articles), 1)
	}

	if article := articles.Articles[0]; article.Title != "Title 1" {
		t.Errorf("should return the correct article title: got %v want %v", article.Title, "Title 1")
	}

	if article := articles.Articles[0]; article.Author.Username != "user1" {
		t.Errorf("should return the correct article author username: got %v want %v", article.Author.Username, "user1")
	}
}

func TestArticlesHandler_FilterByAuthor(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/articles?author=user1", nil)

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(h.ArticlesHandler)
	handler.ServeHTTP(recorder, req)

	var articles ArticlesJSON
	json.NewDecoder(recorder.Body).Decode(&articles)

	if len(articles.Articles) != 3 {
		t.Errorf("should return the correct number article: got %v want %v", len(articles.Articles), 3)
	}

	if article := articles.Articles[0]; article.Author.Username != "user1" {
		t.Errorf("should return the correct article author username: got %v want %v", article.Author.Username, "user1")
	}

	if article := articles.Articles[0]; article.Title != "Title 5" {
		t.Errorf("should return the correct article title: got %v want %v", article.Title, "Title 5")
	}
}

func TestArticlesHandler_FilterByFavorited(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/articles?favorited=user1", nil)

	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(h.ArticlesHandler)
	handler.ServeHTTP(recorder, req)

	var articles ArticlesJSON
	json.NewDecoder(recorder.Body).Decode(&articles)

	if len(articles.Articles) != 3 {
		t.Errorf("should return the correct number article: got %v want %v", len(articles.Articles), 3)
	}

	if article := articles.Articles[0]; article.Title != "Title 5" {
		t.Errorf("should return the correct article title: got %v want %v", article.Title, "Title 5")
	}

	if article := articles.Articles[0]; article.Author.Username != "user1" {
		t.Errorf("should return the correct article author username: got %v want %v", article.Author.Username, "user1")
	}
}

func TestArticlesHandler_CreateUnauthorized(t *testing.T) {
	a := Article{
		Title:       "GoLang Web Services",
		Description: "GoLang Web Services description",
		Body:        "GoLang Web Services",
		TagsList:    []string{"Go"},
	}

	json, _ := json.Marshal(a)
	req, err := http.NewRequest("POST", "/api/articles", bytes.NewBuffer(json))

	if err != nil {
		t.Fatal(err)
	}

	recoder := httptest.NewRecorder()
	handler := http.HandlerFunc(h.ArticlesHandler)

	handler.ServeHTTP(recoder, req)

	if Code := recoder.Code; Code != http.StatusUnauthorized {
		t.Errorf("should get an unauthorized status code: got %v wamt %v", Code, http.StatusUnauthorized)
	}
}

func TestArticlesHandler_Create(t *testing.T) {
	a := articleEntity{
		Article: article{
			Title:       "GoLang Web Services",
			Description: "GoLang Web Services description",
			Body:        "GoLang Web Services",
			TagsList:    []string{"Go", "Web Services"},
		},
	}

	jsonBody, _ := json.Marshal(a)
	req, err := http.NewRequest("POST", "/api/articles", bytes.NewBuffer(jsonBody))

	jwt := auth.NewJWT().NewToken("user1")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))

	if err != nil {
		t.Fatal(err)
	}

	recoder := httptest.NewRecorder()
	handler := http.HandlerFunc(h.ArticlesHandler)

	handler.ServeHTTP(recoder, req)

	if Code := recoder.Code; Code != http.StatusCreated {
		t.Errorf("should get an 201 status code: got %v wamt %v", Code, http.StatusCreated)
	}

	var articleResponse ArticleJSON
	json.NewDecoder(recoder.Body).Decode(&articleResponse)

	if article := articleResponse.Article; article.Title != "GoLang Web Services" {
		t.Errorf("should get the correct article title: got %v wamt %v", article.Title, "GoLang Web Services")
	}
}
