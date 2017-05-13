package models

import "github.com/jinzhu/gorm"

type Favorite struct {
	ID        int
	User      User
	UserID    int
	Article   Article
	ArticleID int
}

func (f *Favorite) AfterCreate(db *gorm.DB) (err error) {
	var a = &Article{ID: f.ArticleID}
	err = db.First(&a).Update("favorites_count", gorm.Expr("favorites_count + ?", 1)).Error
	return
}

func (f *Favorite) AfterDelete(db *gorm.DB) (err error) {
	var a = &Article{ID: f.ArticleID}
	err = db.First(&a).Update("favorites_count", gorm.Expr("favorites_count - ?", 1)).Error
	return
}
