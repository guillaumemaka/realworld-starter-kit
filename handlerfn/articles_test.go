package handlerfn

import (
	"encoding/json"
	"flag"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/chilledoj/realworld-starter-kit/models"
)

var (
	user1 models.User
	user2 models.User
)

func TestMain(m *testing.M) {
	flag.Parse()
	user1 = models.User{ID: 1, Username: "user1"}
	user2 = models.User{ID: 2, Username: "user2"}
	os.Exit(m.Run())
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
		{"SillyLimitSet", "GET", "/api/articles?limit=10000", 10000, 0, empSl, empSl, empSl},
		{"InvalidLimit", "GET", "/api/articles?limit=ABCDE", 20, 0, empSl, empSl, empSl},
		{"OffsetSet", "GET", "/api/articles?offset=10", 20, 10, empSl, empSl, empSl},
		{"SillyOffsetSet", "GET", "/api/articles?offset=20000", 20, 20000, empSl, empSl, empSl},
		{"InvalidOffset", "GET", "/api/articles?offset=ABCDE", 20, 0, empSl, empSl, empSl},
		{"BothLimit+Offset", "GET", "/api/articles?limit=10&offset=10", 10, 10, empSl, empSl, empSl},
		{"SingleTag", "GET", "/api/articles?tag=testing", 20, 0, []string{"testing"}, empSl, empSl},
		{"MultipleTag", "GET", "/api/articles?tag=testing&tag=chilledoj&tag=johnjacob", 20, 0, []string{"testing", "chilledoj", "johnjacob"}, empSl, empSl},
		{"SingleAuthor", "GET", "/api/articles?author=testing", 20, 0, empSl, []string{"testing"}, empSl},
		{"MultipleAuthor", "GET", "/api/articles?author=testing&author=chilledoj&author=johnjacob", 20, 0, empSl, []string{"testing", "chilledoj", "johnjacob"}, empSl},
		{"SingleFavorite", "GET", "/api/articles?favorite=testing", 20, 0, empSl, empSl, []string{"testing"}},
		{"MultipleFavorite", "GET", "/api/articles?favorite=testing&favorite=chilledoj&favorite=johnjacob", 20, 0, empSl, empSl, []string{"testing", "chilledoj", "johnjacob"}},
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
