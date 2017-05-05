package models

import (
	"time"

	"github.com/Machiel/slugify"
	"github.com/jinzhu/gorm"
)

type ScopeHandler func(db *gorm.DB) *gorm.DB

type ArticleStorer interface {
	CreateArticle(*Article) error
	GetAllArticles() *gorm.DB
	GetAllArticlesAuthoredBy(string) ([]Article, error)
	GetAllArticlesFavoritedBy(string) ([]Article, error)
	GetAllArticlesWithTag(string) ([]Article, error)
	GetArticle(string) (Article, error)
	GetUser(*User) *gorm.DB
	IsFavorited(*User, *Article) bool
	IsFollowing(*User, *User) bool
	SaveArticle(*Article) error
}

// Article article model
type Article struct {
	ID             uint
	Slug           string
	Title          string
	Description    string
	Body           string
	User           User
	UserID         uint
	Tags           []Tag `gorm:"many2many:taggings;"`
	Favorites      []Favorite
	FavoritesCount uint
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewArticle returns a new Article instance.
func NewArticle(title string, description string, body string, user User) *Article {
	return &Article{
		Title:       title,
		Description: description,
		Body:        body,
		User:        user,
	}
}

// CreateArticle persist a new article
func (db *DB) CreateArticle(article *Article) (err error) {
	err = db.Create(&article).Error
	return
}

func (db *DB) SaveArticle(article *Article) (err error) {
	err = db.Save(&article).Error
	return
}

// GetArticle retrieve an article by it slug
func (db *DB) GetArticle(slug string) (article Article, err error) {
	err = db.DB.Scopes(defaultScope).First(&article, "slug = ?", slug).Error
	return
}

// GetAllArticles returns all articles.
func (db *DB) GetAllArticles() *gorm.DB {
	return db.DB.Scopes(defaultScope)
}

func (db *DB) GetAllArticlesWithTag(tagName string) (articles []Article, err error) {
	//tag := Tag{Name: tagName}
	//db.FindTag(&tag)
	err = db. //DB.Model(&tag).
			Scopes(defaultScope).
			Joins("JOIN taggings ON taggings.article_id = articles.id").
			Joins("JOIN tags ON tags.id = taggings.tag_id").
			Where("tags.name = ?", tagName).
			Find(&articles).Error
	//Related(&articles, "Articles").Error
	return
}

func (db *DB) GetAllArticlesAuthoredBy(username string) (articles []Article, err error) {
	//user := User{Username: username}
	//db.GetUser(&user)
	err = db.Scopes(defaultScope).
		Joins("JOIN users ON users.id = articles.user_id").
		Where("users.username = ?", username).
		Find(&articles).Error
	return
}

func (db *DB) GetAllArticlesFavoritedBy(username string) (articles []Article, err error) {
	user := User{Username: username}
	db.GetUser(&user)
	err = db.Scopes(defaultScope).
		Joins("JOIN favorites ON favorites.article_id = articles.id").
		Where("favorites.user_id = ?", user.ID).
		Find(&articles).Error
	return
}

func (db *DB) IsFavorited(user *User, article *Article) bool {
	f := Favorite{ArticleID: article.ID, UserID: uint(user.ID)}

	if db.Where(f).First(&f).RecordNotFound() {
		return false
	}

	return true
}

func (db *DB) IsFollowing(userFrom *User, userTo *User) bool {
	return false
}

func (db *DB) GetUser(user *User) *gorm.DB {
	return db.Where(user).First(&user)
}

// Callbacks

// BeforeCreate gorm callback
func (a *Article) BeforeCreate() (err error) {
	a.Slug = slugify.Slugify(a.Title)
	return
}

func (a *Article) BeforeUpdate() (err error) {
	a.Slug = slugify.Slugify(a.Title)
	return
}

// Scopes

// Order articles by created_at DESC eager loading Tags and User
func defaultScope(db *gorm.DB) *gorm.DB {
	return db.Order("articles.created_at desc").
		Preload("Tags").
		Preload("User")
}
