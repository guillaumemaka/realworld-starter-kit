package handlerfn

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chilledoj/realworld-starter-kit/models"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const (
	jsn = "application/json"
)

var (
	user1      models.User
	user1Token string
	//user2     models.User
	ae        *AppEnvironment
	mock      sqlmock.Sqlmock
	jsonTests map[string]string
)

func TestMain(m *testing.M) {
	// SETUP
	flag.Parse()

	logger := log.New(os.Stdout, "[app] ", log.LstdFlags)
	var db *sql.DB
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		panic(fmt.Sprintf("An error '%s' was not expected when opening a stub database connection", err))
	}
	ae = &AppEnvironment{Logger: logger, DB: &models.AppDB{DB: db}}
	// DATA
	user1 = models.User{ID: 1, Username: "user1", Email: "test1@test.com"}
	user1Token, err = models.NewToken(&user1)
	if err != nil {
		panic(err)
	}
	//user2 = models.User{ID: 2, Username: "user2", Email: "test1@test.com"}

	jsonTests = make(map[string]string, 7)
	jsonTests["Empty"] = ""
	jsonTests["Valid"] = `{"article": {"title":"title","description":"desc","body":"body"}}`
	jsonTests["Valid_Tags"] = `{"article": {"title":"title","description":"desc","body":"body", "tagList":["testing","dragons"]}}`
	jsonTests["MissingTitle"] = `{"article": {"description":"desc","body":"body"}}`
	jsonTests["MissingDesc"] = `{"article": {"title":"title","body":"body"}}`
	jsonTests["MissingBody"] = `{"article": {"title":"title","description":"description"}}`
	jsonTests["EmptyArticle"] = `{"article":{}}`

	// RUN
	rc := m.Run()

	// TEARDOWN
	db.Close()
	os.Exit(rc)
}

type fnTest struct {
	name           string
	reqMethod      string
	reqURL         string
	reqBody        string
	reqUser        *models.User
	expStatus      int
	expErr         bool
	expContentType string
	setupMock      func()
	bodyTest       func(reqBody, resBody string) bool
}

func handlerTest(t *testing.T, handler http.Handler, tests []fnTest) {
	for _, v := range tests {
		t.Run(v.name, func(t2 *testing.T) {
			body := strings.NewReader(v.reqBody)
			r := httptest.NewRequest(v.reqMethod, v.reqURL, body)

			//t2.Logf("User provided=%v", v.reqUser)
			if v.reqUser != nil {
				t2.Logf("A user is provided")
				ctx, err := storeJWTUserCtx(user1Token, r)
				if err != nil {
					t2.Error(err)
					return
				}

				r = r.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			v.setupMock()
			handler.ServeHTTP(w, r)

			resp := w.Result()
			resbody, _ := ioutil.ReadAll(resp.Body)
			if resp.StatusCode != v.expStatus {
				t2.Errorf("Unexpected status code: Got (%d), want (%d)", resp.StatusCode, v.expStatus)
			}
			if resp.Header.Get("Content-Type") != v.expContentType {
				t2.Errorf("Unexpected status code: Got(%s), want(%s)", resp.Header.Get("Content-Type"), v.expContentType)
			}
			if strings.Contains(string(resbody), "error") != v.expErr {
				t2.Errorf("Unexpected error found: %s", resbody)
			}
			if !v.bodyTest(v.reqBody, string(resbody)) {
				t2.Errorf("Unexpected response body: Got::\n%s", resbody)
			}

		})
	}
}

func TestCreateArticle(t *testing.T) {
	//t.SkipNow()
	tests := []fnTest{
		{"Valid", "POST", "/", jsonTests["Valid"], &user1, http.StatusOK, false, jsn,
			func() {
				mock.ExpectPrepare("INSERT INTO articles").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
			}, func(reqBody, resBody string) bool {
				return !strings.Contains(resBody, "article") || !strings.Contains(resBody, `"id":1`)
			}},
		{"Valid_Tags", "POST", "/", jsonTests["Valid_Tags"], &user1, http.StatusOK, false, jsn,
			func() {
				mock.ExpectPrepare("INSERT INTO articles").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM art_tags").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectPrepare("INSERT INTO art_tags").ExpectExec().WithArgs("dragons", 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO art_tags").WithArgs("testing", 1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "article") || strings.Contains(resBody, `"id":1`)
			}},
		{"MissingTitle", "POST", "/", jsonTests["MissingTitle"], &user1, http.StatusUnprocessableEntity, true, jsn,
			func() {
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "Title is not set")
			}},
		{"MissingDesc", "POST", "/", jsonTests["MissingDesc"], &user1, http.StatusUnprocessableEntity, true, jsn,
			func() {
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "Description is not set")
			}},
		{"MissingBody", "POST", "/", jsonTests["MissingBody"], &user1, http.StatusUnprocessableEntity, true, jsn,
			func() {
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "Body is not set")
			}},
		{"EmptyArticle", "POST", "/", jsonTests["EmptyArticle"], &user1, http.StatusUnprocessableEntity, true, jsn,
			func() {
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "Title is not set") && strings.Contains(resBody, "Description is not set") && strings.Contains(resBody, "Body is not set")
			}},
		{"InvalidJSON", "POST", "/", `This is not valid JSON`, &user1, http.StatusUnprocessableEntity, true, jsn,
			func() {
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "Invalid JSON provided")
			}},
	}
	handler := CreateArticle(ae)
	handlerTest(t, handler, tests)
}

func TestDeleteArticle(t *testing.T) {
	/*
		SQL Select Columns
		a.id,slug,title,description,body,created,updated,
		u.username as author_username, u.bio, u.image as author_image
		,CASE WHEN uf.usr_following_id IS null THEN 0 ELSE 1 END AS following
		,t.tag
	*/
	tests := []fnTest{
		{"Valid", "POST", "/articles/test", jsonTests["Empty"], &user1, http.StatusOK, false, jsn,
			func() {
				rows := sqlmock.NewRows([]string{"id", "slug", "title", "description", "body", "created", "updated",
					"author_username", "bio", "author_image", "following", "tag"})
				rows = rows.AddRow(1, "test", "title", "desc", "body", time.Now(), time.Now(),
					user1.Username, "", "", 0, "dragons||test")
				mock.ExpectQuery("SELECT (.)+ FROM articles (.)+").WithArgs(user1.ID, "").WillReturnRows(rows)
				mock.ExpectPrepare("DELETE (.)").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
			}, func(reqBody, resBody string) bool {
				return resBody == "{}"
			}},
		{"NotFound", "POST", "/articles/test", jsonTests["Empty"], &user1, http.StatusUnprocessableEntity, true, jsn,
			func() {
				mock.ExpectQuery("SELECT (.)+ FROM articles (.)+").WithArgs(user1.ID, "").WillReturnError(fmt.Errorf("Not found"))
			}, func(reqBody, resBody string) bool {
				return strings.Contains(resBody, "Not found")
			}},
		{"NoUser", "POST", "/articles/test", jsonTests["Empty"], nil, http.StatusInternalServerError, true, jsn,
			func() {
			}, func(reqBody, resBody string) bool {
				return true // strings.Contains(resBody, "Not found")
			}}}
	handler := DeleteArticle(ae)
	handlerTest(t, handler, tests)
}

func TestBuildQueryOptions(t *testing.T) {
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

func TestArticleFormPost_Validate(t *testing.T) {
	tests := []struct {
		name         string
		ipJSON       string
		expErrLen    int
		expErrString string
	}{
		{"Valid", `{"article": {"title":"title","description":"desc","body":"body"}}`, 0, ""},
		{"MissingTitle", `{"article": {"description":"desc","body":"body"}}`, 1, "Title"},
		{"MissingDesc", `{"article": {"title":"title","body":"body"}}`, 1, "Description"},
		{"MissingBody", `{"article": {"title":"title","body":"body"}}`, 1, "Description"},
		{"Invalid", `{"article":{}}`, 3, "Title Description Body"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := strings.NewReader(test.ipJSON)
			artPost := articleFormPost{}
			err := json.NewDecoder(r).Decode(&artPost)
			if err != nil {
				t.Error(err)
				return
			}
			errs := artPost.Validate()
			if len(errs) != test.expErrLen {
				t.Errorf("Unexpected errors: Got (%s), want %d", errs, test.expErrLen)
			}
			for _, v := range errs {
				if !strings.ContainsAny(v.Error(), test.expErrString) {
					t.Errorf("Can't find %s in %s", v.Error(), test.expErrString)
				}
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
