package models

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	sq "gopkg.in/Masterminds/squirrel.v1"
)

// Article holds the article data
type Article struct {
	ID              uint      `db:"id" json:"-"`
	Slug            string    `db:"slug" json:"slug"`
	Title           string    `db:"title" json:"title"`
	Description     string    `db:"description" json:"description"`
	Body            string    `db:"body" json:"body"`
	CreatedAt       time.Time `db:"created" json:"createdAt"`
	UpdatedAt       time.Time `db:"updated" json:"updatedAt"`
	Favourited      bool      `db:"fav" json:"favorited"`
	FavouritesCount uint      `db:"favCount" json:"favoritesCount"`
	Author          *Profile  `db:"author_id" json:"author"`
	TagList         []Tag     `json:"tagList"`
}

// MarshalJSON implements JSON encoding
func (a *Article) MarshalJSON() ([]byte, error) {
	type Art Article
	return json.Marshal(&struct {
		*Art
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}{
		Art:       (*Art)(a),
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		UpdatedAt: a.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
	})
}

// SingleArtJSONResponse is the container for a single article response
type SingleArtJSONResponse struct {
	Article *Article `json:"article"`
}

// MultipleArtJSONResponse is the container for a multiple article response
type MultipleArtJSONResponse struct {
	Articles      []*Article `json:"articles"`
	ArticlesCount int        `json:"articlesCount"`
}

const (
	// CREATESQL
	qCreateArticle = "INSERT INTO articles (slug,title,description,body,created,updated,author_id) VALUES (?,?,?,?,?,?,?)"
	// READSQL
	qArticleDetailsPart = `SELECT a.id,slug,title,description,body,created,updated,
	u.username as author_username, u.bio, u.image as author_image
	,CASE WHEN uf.usr_following_id IS null THEN 0 ELSE 1 END AS following
	, t.tags
	FROM articles a
	JOIN users u on a.author_id = u.id
	LEFT OUTER JOIN (SELECT art_id,group_concat(tag SEPARATOR '||') as tags FROM art_tags GROUP BY art_id) t on a.id = t.art_id
	`
	qGetArticle = qArticleDetailsPart + `LEFT OUTER JOIN usr_following uf
		ON u.id = uf.usr_following_id
		and uf.usr_id = ?
	WHERE a.slug = ?`
	qGetFeedArticles = qArticleDetailsPart + `JOIN usr_following uf
		ON u.id = uf.usr_following_id
		and uf.usr_id = ?
	`
	// UPDATESQL
	qUpdateArticle = `UPDATE articles SET slug=?, title=?, description=?,body=?,updated=? WHERE id=?`
	// DELETESQL
	qDeleteArticle = `DELETE FROM articles WHERE slug = ?`
)

// NewArticle creates a new article object (not persisted to DB)
func NewArticle(title, description, body string, author *Profile) (*Article, error) {
	a := &Article{
		Title: title, Description: description, Body: body, Author: author,
		Favourited: false, FavouritesCount: 0,
	}

	if err := a.CreateSlug(); err != nil {
		return nil, err
	}
	// Let's save the create timestamp here instead of leaving to DB
	n := time.Now().UTC()
	a.CreatedAt = n
	a.UpdatedAt = n
	return a, nil
}

// CreateSlug adds a slug to the article
func (a *Article) CreateSlug() error {
	if a.Title == "" {
		return fmt.Errorf("Title has not been set")
	}
	// Random 6 char code at end of slug to allow duplicate titles
	buf := make([]byte, 3)
	_, err := rand.Read(buf)
	if err != nil {
		return err
	}
	hx := hex.EncodeToString(buf)
	a.Slug = slug(a.Title + "-" + hx)
	return nil
}
func slug(t string) string {
	retStr := strings.TrimSpace(t)
	retStr = strings.ToLower(retStr)
	re := regexp.MustCompile(`/[^a-zA-Z0-9 -]/g`)
	retStr = string(re.ReplaceAll([]byte(retStr), []byte("")))
	rep := strings.NewReplacer(" ", "-")
	return rep.Replace(retStr)
}

// CreateArticle inserts the article into DB and updates the ID
func (adb *AppDB) CreateArticle(a *Article) error {
	stmt, err := adb.DB.Prepare(qCreateArticle)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(a.Slug, a.Title, a.Description, a.Body, a.CreatedAt, a.UpdatedAt, a.Author.ID)
	if err != nil {
		return err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	// update ID on article
	a.ID = uint(lastID)
	return nil
}

// GetArticle gets a single article by slug
func (adb *AppDB) GetArticle(slug string, whosasking uint) (*Article, error) {
	a := &Article{}
	p := Profile{}
	/*
		a.id,slug,title,description,body,created,updated,
		u.username as author_username, u.bio, u.image as author_image
		,CASE WHEN uf.usr_following_id IS null THEN 0 ELSE 1 END AS following
		,t.tag
	*/
	var tags string
	if err := adb.DB.QueryRow(qGetArticle, whosasking, slug).Scan(&a.ID, &a.Slug, &a.Title, &a.Description,
		&a.Body, &a.CreatedAt, &a.UpdatedAt, &p.Username, &p.Bio, &p.Image, &p.Following, &tags); err != nil {
		return nil, err
	}
	a.Author = &p
	a.TagList = splitTagString(tags)
	return a, nil
}

func splitTagString(tags string) (tagList []Tag) {
	tagList = make([]Tag, 0)
	for _, v := range strings.Split(tags, "||") {
		tagList = append(tagList, Tag(v))
	}
	return
}

// UpdateArticle does exactly what it says on the tin
func (adb *AppDB) UpdateArticle(a *Article) error {
	//qUpdateArticle = `UPDATE articles SET slug=?, title=?, description=?,body=?,updated=? WHERE id=?`
	stmt, err := adb.DB.Prepare(qUpdateArticle)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(a.Slug, a.Title, a.Description, a.Body, a.UpdatedAt, a.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("Could not Update. Article not found")
	}
	return nil
}

// DeleteArticle does exactly what it says on the tin
func (adb *AppDB) DeleteArticle(slug string) error {
	stmt, err := adb.DB.Prepare(qDeleteArticle)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(slug)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("Could not delete. Article not found")
	}
	return nil
}

// ListArticles returns a list of articles
// The list will be filtered by the passed in options
// If feed is true, the list of articles will only contain those articles by authors
// who the user (whosasking) is following
func (adb *AppDB) ListArticles(opt ListArticleOptions, whosasking uint, feed bool) ([]*Article, error) {
	articles := make([]*Article, 0)
	sql, args, err := opt.BuildArticleQuery(whosasking, feed)
	if err != nil {
		return nil, err
	}
	rows, err := adb.DB.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		a := &Article{}
		p := Profile{}
		var tags string
		if err := rows.Scan(&a.ID, &a.Slug, &a.Title, &a.Description,
			&a.Body, &a.CreatedAt, &a.UpdatedAt, &p.Username, &p.Bio, &p.Image, &p.Following, &tags); err != nil {
			return nil, err
		}
		a.Author = &p
		a.TagList = splitTagString(tags)
		articles = append(articles, a)
	}
	return articles, nil
}

// ListArticleOptions holds the select clause details for listing articles
type ListArticleOptions struct {
	Limit   uint
	Offset  uint
	Filters map[string][]string
}

// NewListOptions instantiates a new struct with defaults
// Pass in all options as a map
// e.g.
//      opts:= make([string]interface{},0)
// 		opts["limit"] = 30
// 		opts["filters"] = map[string][]string{   // This is equivalent to url.Values
//			"tag": []string{"AngularJS"},
//		}
func NewListOptions(args map[string]interface{}) ListArticleOptions {
	// Defaults
	opts := ListArticleOptions{
		Limit:   20,
		Offset:  0,
		Filters: map[string]([]string){},
	}
	if len(args) == 0 {
		return opts
	}
	if v, ok := args["limit"].(uint); ok && v > 0 {
		opts.Limit = uint(v)
	}
	if v, ok := args["offset"].(uint); ok && v > 0 {
		opts.Offset = uint(v)
	}
	if v, ok := args["filters"].(map[string][]string); ok {
		opts.Filters = v
	}
	return opts
}

// BuildArticleQuery builds the list article SQL
func (opts ListArticleOptions) BuildArticleQuery(whoisasking uint, feed bool) (string, []interface{}, error) {
	qry := sq.Select(`a.id,slug,title,description,body,created,updated,
	u.username as author_username, u.bio, u.image as author_image
	,CASE WHEN uf.usr_following_id IS null THEN 0 ELSE 1 END AS following
	, t.tags`).From("articles a").
		Join("users u on a.author_id = u.id").
		LeftJoin("(SELECT art_id,group_concat(tag SEPARATOR '||') as tags FROM art_tags GROUP BY art_id) t on a.id = t.art_id")
	if feed {
		qry = qry.Join("usr_following uf ON u.id = uf.usr_following_id and uf.usr_id = ?", whoisasking)
	} else {
		qry = qry.LeftJoin("usr_following uf ON u.id = uf.usr_following_id and uf.usr_id = ?", whoisasking)
	}

	sql, args, err := qry.ToSql()
	if err != nil {
		return "", nil, err
	}
	where := ""

	for filterType, filterValues := range opts.Filters {
		if len(filterValues) == 0 {
			continue
		}
		if where == "" {
			where = "WHERE "
		} else {
			where += "\nAND "
		}
		switch filterType {
		case "author":
			authors, vals, err := sq.Eq{"u.username": filterValues}.ToSql()
			if err != nil {
				return "", nil, err
			}
			where += authors
			args = append(args, vals...)
		case "favorite":
			favs, vals, err := sq.Select("art_id").From("usr_art_favourite fav").
				Join("users u on fav.usr_id = u.id").
				Where(sq.Eq{"u.username": filterValues}).ToSql()
			if err != nil {
				log.Println(err)
			}
			where += fmt.Sprintf("a.id in (%s)", favs)
			args = append(args, vals...)
		case "tag":
			tagsql, vals, err := sq.Select("art_id").From("art_tags").Where(sq.Eq{"tag": filterValues}).ToSql()
			if err != nil {
				return "", nil, err
			}
			where += fmt.Sprintf("a.id IN (%s)", tagsql)
			args = append(args, vals...)

		}
	}
	sql += "\n" + where
	sql += "\nORDER BY created desc"
	sql += "\n" + fmt.Sprintf("LIMIT %d", opts.Limit)
	if opts.Offset > 0 {
		sql += "\n" + fmt.Sprintf("OFFSET %d", opts.Offset)
	}

	return sql, args, nil
}
