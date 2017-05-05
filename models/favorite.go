package models

type Favorite struct {
	ID        uint
	User      User
	UserID    uint
	Article   Article
	ArticleID uint
}
