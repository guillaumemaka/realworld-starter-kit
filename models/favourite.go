package models

import "fmt"

const (
	qFavArticle   = `INSERT INTO usr_art_favourite (usr_id, art_id) VALUES (?,?)`
	qUnfavArticle = `DELETE FROM usr_art_favourite WHERE usr_id=? AND art_id=?`
)

// FavouriteArticle allows a user to favourite an article
func (adb *AppDB) FavouriteArticle(a *Article, u *User) error {
	stmt, err := adb.DB.Prepare(qFavArticle)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(u.ID, a.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("Did not insert row for %s to favourite %s", u.Username, a.Slug)
	}
	a.Favourited = true
	a.FavouritesCount++
	return nil
}

// UnfavouriteArticle allows a user to favourite an article
func (adb *AppDB) UnfavouriteArticle(a *Article, u *User) error {
	stmt, err := adb.DB.Prepare(qUnfavArticle)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(u.ID, a.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("Did not delete row for %s to follow %s", u.Username, a.Slug)
	}
	a.FavouritesCount--
	a.Favourited = a.FavouritesCount > 0
	return nil
}
