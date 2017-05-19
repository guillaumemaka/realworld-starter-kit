package models

import (
	"sort"
)

// Tag for articles
type Tag string

func (t Tag) String() string {
	return string(t)
}

// TagJSONResponse is a wrapper for returning a list of Tags
type TagJSONResponse struct {
	Tags []Tag `json:"tags"`
}

const (
	qCreateTag      = "INSERT INTO art_tags (tag,art_id) VALUES (?,?)"
	qReadTags       = "SELECT DISTINCT tag from art_tags"
	qGetArticleTags = "SELECT tag FROM art_tags WHERE art_id=?"
	qDeleteTags     = "DELETE FROM art_tags WHERE art_id = ?"
)

// AddTags persists a slice of tags to the DB for a given article
// As a simple implementation, we will ensure that all existing
// tags for an article are deleted first, and then re-insert all
// provided tags. This could result in smashing/fragmenting the
// DB indexes, but nevermind.
// At a later date it might be best to look into JSON column type
// or even just a simple TEXT column with a delimiter and FULLTEXT Search
func (adb *AppDB) AddTags(a *Article, tags []Tag) (ts []Tag, err error) {
	//ts = make([]Tag, 0)
	tx, err := adb.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec(qDeleteTags, a.ID); err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare(qCreateTag)
	if err != nil {
		return nil, err
	}
	ts = removeDuplicateTags(tags)
	for _, tag := range ts {
		if _, err = stmt.Exec(string(tag), a.ID); err != nil {
			return
		}
	}
	return
}

// Rough function for removing duplicates
func removeDuplicateTags(l []Tag) []Tag {
	sort.Slice(l, func(i, j int) bool {
		return l[i].String() < l[j].String()
	})
	var prevValue Tag
	newList := make([]Tag, 0)
	for _, v := range l {
		if prevValue == "" || prevValue != v {
			newList = append(newList, v)
			prevValue = v
			continue
		}
	}
	return newList
}

// GetAllTags returns the distinct list of tags
func (adb *AppDB) GetAllTags() (tags []Tag, err error) {
	tags = make([]Tag, 0)
	rows, err := adb.DB.Query(qReadTags)
	if err != nil {
		return tags, nil
	}
	for rows.Next() {
		str := ""
		err = rows.Scan(&str)
		if err != nil {
			return
		}
		tags = append(tags, Tag(str))
	}
	return
}
