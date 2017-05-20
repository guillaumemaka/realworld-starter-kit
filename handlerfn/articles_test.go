package handlerfn

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chilledoj/realworld-starter-kit/models"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestCreateArticle(t *testing.T) {
	jsonTests := make(map[string]string, 7)
	jsonTests["Empty"] = ""
	jsonTests["Valid"] = `{"article": {"title":"title","description":"desc","body":"body"}}`
	jsonTests["Valid_Tags"] = `{"article": {"title":"title","description":"desc","body":"body", "tagList":["testing","dragons"]}}`
	jsonTests["MissingTitle"] = `{"article": {"description":"desc","body":"body"}}`
	jsonTests["MissingDesc"] = `{"article": {"title":"title","body":"body"}}`
	jsonTests["MissingBody"] = `{"article": {"title":"title","description":"description"}}`
	jsonTests["EmptyArticle"] = `{"article":{}}`
	path := "/api/articles"
	tests := []fnTest{
		{"CreateArticle:Valid", "POST", path, jsonTests["Valid"], &user1, http.StatusOK, false, jsn,
			func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INSERT INTO articles").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
			}, func(reqBody, resBody string) bool {
				return !strings.Contains(resBody, "article") || !strings.Contains(resBody, `"id":1`)
			}},
		{"CreateArticle:Valid_Tags", "POST", path, jsonTests["Valid_Tags"], &user1, http.StatusOK, false, jsn,
			func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INSERT INTO articles").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM art_tags").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectPrepare("INSERT INTO art_tags").ExpectExec().WithArgs("dragons", 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO art_tags").WithArgs("testing", 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "article") || strings.Contains(resBody, `"id":1`)
			}},
		{"CreateArticle:MissingTitle", "POST", path, jsonTests["MissingTitle"], &user1, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "Title is not set")
			}},
		{"CreateArticle:MissingDesc", "POST", path, jsonTests["MissingDesc"], &user1, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "Description is not set")
			}},
		{"CreateArticle:MissingBody", "POST", path, jsonTests["MissingBody"], &user1, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "Body is not set")
			}},
		{"CreateArticle:EmptyArticle", "POST", path, jsonTests["EmptyArticle"], &user1, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "Title is not set") && strings.Contains(resBody, "Description is not set") && strings.Contains(resBody, "Body is not set")
			}},
		{"CreateArticle:InvalidJSON", "POST", path, `This is not valid JSON`, &user1, http.StatusBadRequest, true, jsn,
			func(mock sqlmock.Sqlmock) {
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, `"error": "Bad Request"`)
			}},
	}
	handlerTest(t, CreateArticle, tests)
}

func TestDeleteArticle(t *testing.T) {
	t.SkipNow()
	/*
		SQL Select Columns
		a.id,slug,title,description,body,created,updated,
		u.username as author_username, u.bio, u.image as author_image
		,CASE WHEN uf.usr_following_id IS null THEN 0 ELSE 1 END AS following
		,t.tag
	*/
}

func Test_buildQueryOptions(t *testing.T) {
	buf := strings.NewReader("TEST")
	empSl := []string{}
	tests := []struct {
		name     string
		method   string
		url      string
		limit    uint
		offset   uint
		tags     []string
		author   []string
		favorite []string
	}{
		{"NoQueryValues", "GET", "/api/articles", 20, 0, empSl, empSl, empSl},
		{"LimitSet", "GET", "/api/articles?limit=10", 10, 0, empSl, empSl, empSl},
		{"SillyLimitSet", "GET", "/api/articles?limit=10000", 100, 0, empSl, empSl, empSl},
		{"InvalidLimit", "GET", "/api/articles?limit=ABCDE", 20, 0, empSl, empSl, empSl},
		{"OffsetSet", "GET", "/api/articles?offset=10", 20, 10, empSl, empSl, empSl},
		{"SillyOffsetSet", "GET", "/api/articles?offset=20000", 20, 20000, empSl, empSl, empSl},
		{"InvalidOffset", "GET", "/api/articles?offset=ABCDE", 20, 0, empSl, empSl, empSl},
		{"BothLimit+Offset", "GET", "/api/articles?limit=10&offset=10", 10, 10, empSl, empSl, empSl},
		{"SingleTag", "GET", "/api/articles?tag=testing", 20, 0, []string{"testing"}, empSl, empSl},
		{"MultipleTag", "GET", "/api/articles?tag=testing&tag=chilledoj&tag=johnjacob", 20, 0, []string{"testing", "chilledoj", "johnjacob"}, empSl, empSl},
		{"SingleAuthor", "GET", "/api/articles?author=testing", 20, 0, empSl, []string{"testing"}, empSl},
		{"MultipleAuthor", "GET", "/api/articles?author=testing&author=chilledoj&author=johnjacob", 20, 0, empSl, []string{"testing", "chilledoj", "johnjacob"}, empSl},
		{"SingleFavorite", "GET", "/api/articles?favorited=testing", 20, 0, empSl, empSl, []string{"testing"}},
		{"MultipleFavorite", "GET", "/api/articles?favorited=testing&favorited=chilledoj&favorited=johnjacob", 20, 0, empSl, empSl, []string{"testing", "chilledoj", "johnjacob"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, test.url, buf)
			opts := buildQueryOptions(r)
			if opts.Limit != test.limit {
				t.Errorf("Unexpected limit: Got %d, want %d", opts.Limit, test.limit)
			}
			if opts.Offset != test.offset {
				t.Errorf("Unexpected offset: Got %d, want %d", opts.Offset, test.offset)
			}
			if len(opts.Filters["tag"]) != len(test.tags) {
				t.Errorf("Unexpected tag filter: Got %s, want %s", opts.Filters["tag"], test.tags)
			}
			if len(opts.Filters["author"]) != len(test.author) {
				t.Errorf("Unexpected author filter: Got %s, want %s", opts.Filters["author"], test.author)
			}
			if len(opts.Filters["favorite"]) != len(test.favorite) {
				t.Errorf("Unexpected favorite filter: Got %s, want %s", opts.Filters["favorite"], test.favorite)
			}
		})

	}

}

func TestArticlePost_UnmarshalJSON(t *testing.T) {
	testString := "TESTing"
	test1Tag := models.Tag("test1")
	test2Tag := models.Tag("test2")
	oneTag := []models.Tag{test1Tag}
	twoTag := []models.Tag{test1Tag, test2Tag}
	tests := []struct {
		name     string
		json     string
		expected articlePost
		expTags  bool
	}{
		{"EmptyJSON", "{}", articlePost{}, false},
		{"JustTitle", `{"title":"TESTing"}`, articlePost{Title: &testString}, false},
		{"JustDescription", `{"description":"TESTing"}`, articlePost{Description: &testString}, false},
		{"JustBody", `{"body":"TESTing"}`, articlePost{Body: &testString}, false},
		{"JustOneTag", `{"tagList":["test1"]}`, articlePost{Tags: &oneTag}, true},
		{"JustTwoTags", `{"tagList":["test1","test2"]}`, articlePost{Tags: &twoTag}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ap := articlePost{}
			r := strings.NewReader(test.json)
			json.NewDecoder(r).Decode(&ap)
			if !checkStringPointers(ap.Title, test.expected.Title) {
				t.Errorf("Unexpected title: Got %v, want %v", ap.Title, test.expected.Title)
			}
			if !checkStringPointers(ap.Description, test.expected.Description) {
				t.Errorf("Unexpected Description: Got %v, want %v", ap.Description, test.expected.Description)
			}
			if !checkStringPointers(ap.Body, test.expected.Body) {
				t.Errorf("Unexpected Description: Got %v, want %v", ap.Body, test.expected.Body)
			}
			if !test.expTags {
				return
			}

			if len(*ap.Tags) != len(*test.expected.Tags) {
				t.Errorf("Unexpected tags: Got %v, want %v", ap.Tags, test.expected.Tags)
			}
		})
	}
}

func checkStringPointers(s1, s2 *string) bool {
	if (s1 == nil && s2 != nil) || (s1 != nil && s2 == nil) {
		return false
	}
	if s1 != nil && s2 != nil && *s1 != *s2 {
		return false
	}
	return true
}
