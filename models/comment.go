package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Comment model
type Comment struct {
	ID        uint      `db:"id" json:"id"`
	Body      string    `db:"body" json:"body"`
	CreatedAt time.Time `db:"created" json:"createdAt"`
	UpdatedAt time.Time `db:"updated" json:"updatedAt"`
	Author    *Profile  `db:"author_id" json:"author"`
}

// MarshalJSON implements JSON encoding
func (c *Comment) MarshalJSON() ([]byte, error) {
	type Com Comment
	return json.Marshal(&struct {
		*Com
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}{
		Com:       (*Com)(c),
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
	})
}

// SingleComJSONResponse is the container for a single article response
type SingleComJSONResponse struct {
	Comment *Comment `json:"comment"`
}

// MultipleComJSONResponse is the container for a multiple article response
type MultipleComJSONResponse struct {
	Comments []*Comment `json:"comments"`
}

const (
	qGetCommentsForArticle = `SELECT
  c.id, c.body, c.createdAt, c.updatedAt,
  u.id, u.username, u.bio, u.image
  , CASE WHEN uf.usr_following_id IS null THEN 0 ELSE 1 END AS following
  FROM comments c
  JOIN users u on c.author_id = u.id
  JOIN articles a on c.article_id = a.id
  LEFT OUTER JOIN usr_following uf
  		ON u.id = uf.usr_following_id
  		and uf.usr_id = ?
  WHERE a.slug = ?
  `
	qGetCommentsByID = `SELECT
  c.id, c.body, c.createdAt, c.updatedAt,
  u.id, u.username, u.bio, u.image
  , CASE WHEN uf.usr_following_id IS null THEN 0 ELSE 1 END AS following
  FROM comments c
  JOIN users u on c.author_id = u.id
  JOIN articles a on c.article_id = a.id
  LEFT OUTER JOIN usr_following uf
  		ON u.id = uf.usr_following_id
  		and uf.usr_id = ?
  WHERE a.slug = ? AND c.id = ?
  `
	qAddComment = `insert into comments(body,createdAt,updatedAt,author_id,article_id)
  SELECT
  ? as body,
  ? as createdAt,
  ? as updatedAt,
  ? as author_id,
  a.id as article_id
  FROM articles a
  WHERE a.slug = ?`
	qDeleteComment = "DELETE FROM comments WHERE id=?"
)

// NewComment creates a new comment object (not persisted to DB)
func NewComment(body string, author *Profile) *Comment {
	c := &Comment{
		Body: body, Author: author,
	}
	// Let's save the create timestamp here instead of leaving to DB
	n := time.Now().UTC()
	c.CreatedAt = n
	c.UpdatedAt = n
	return c
}

// GetCommentByID returns slice of comments for a given slug, and request user id
func (adb *AppDB) GetCommentByID(id uint, slug string, whosasking uint) (*Comment, error) {
	c := &Comment{}
	p := Profile{}
	if err := adb.DB.QueryRow(qGetCommentsByID, whosasking, slug, id).Scan(&c.ID, &c.Body, &c.CreatedAt, &c.UpdatedAt,
		&p.ID, &p.Username, &p.Bio, &p.Image, &p.Following); err != nil {
		return nil, err
	}
	c.Author = &p

	return c, nil
}

// GetComments returns slice of comments for a given slug, and request user id
func (adb *AppDB) GetComments(slug string, whosasking uint) ([]*Comment, error) {
	comments := make([]*Comment, 0)
	rows, err := adb.DB.Query(qGetCommentsForArticle, whosasking, slug)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		c := &Comment{}
		p := Profile{}
		/*
		   c.id, c.body, c.createdAt, c.updatedAt,
		   u.id, u.username, u.bio, u.image
		   , CASE WHEN uf.usr_following_id IS null THEN 0 ELSE 1 END AS following
		*/
		if err := rows.Scan(&c.ID, &c.Body, &c.CreatedAt, &c.UpdatedAt,
			&p.ID, &p.Username, &p.Bio, &p.Image, &p.Following); err != nil {
			return nil, err
		}
		c.Author = &p
		comments = append(comments, c)
	}
	return comments, nil
}

// AddComment adds the provided comment to the article with slug
func (adb *AppDB) AddComment(c *Comment, slug string, whosasking uint) error {
	stmt, err := adb.DB.Prepare(qAddComment)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(c.Body, c.CreatedAt, c.UpdatedAt, whosasking, slug)
	if err != nil {
		return err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	// update ID on article
	c.ID = uint(lastID)
	return nil
}

// DeleteComment removes the commment by ID
func (adb *AppDB) DeleteComment(c *Comment) error {
	stmt, err := adb.DB.Prepare(qDeleteComment)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(c.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("Could not delete. Comment not found")
	}
	return nil
}
