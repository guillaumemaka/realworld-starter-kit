package handlerfn

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/chilledoj/realworld-starter-kit/models"
)

const (
	jsn = "application/json"
)

var (
	user1      models.User
	user1Token string
)

func TestMain(m *testing.M) {
	// SETUP
	flag.Parse()

	// DATA
	user1 = models.User{ID: 1, Username: "user1", Email: "test1@test.com"}
	var err error
	user1Token, err = models.NewToken(&user1)
	if err != nil {
		panic(err)
	}
	//user2 = models.User{ID: 2, Username: "user2", Email: "test1@test.com"}

	// RUN
	rc := m.Run()

	// TEARDOWN
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
	setupMock      func(sqlmock.Sqlmock)
	bodyTest       func(reqBody, resBody string) bool
}

func handlerTest(t *testing.T, fn func(*AppEnvironment) http.Handler, tests []fnTest) {

	for _, v := range tests {
		t.Run(v.name, func(t2 *testing.T) {
			body := strings.NewReader(v.reqBody)
			r := httptest.NewRequest(v.reqMethod, v.reqURL, body)

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

			logger := log.New(os.Stdout, fmt.Sprintf("[%s] ", v.name), log.LstdFlags)
			var db *sql.DB
			var err error
			db, mock, err := sqlmock.NewWithDSN(v.name)
			if err != nil {
				panic(fmt.Sprintf("An error '%s' was not expected when opening a stub database connection", err))
			}
			ae := &AppEnvironment{Logger: logger, DB: &models.AppDB{DB: db}}
			//mock.MatchExpectationsInOrder(true)
			v.setupMock(mock)
			//fmt.Println(mock)
			handler := fn(ae)
			handler.ServeHTTP(w, r)

			resp := w.Result()
			resbody, _ := ioutil.ReadAll(resp.Body)
			if resp.StatusCode != v.expStatus || (v.expErr && resp.StatusCode < 400) {
				t2.Errorf("Unexpected status code: Got (%d), want (%d)", resp.StatusCode, v.expStatus)
			}
			if resp.Header.Get("Content-Type") != v.expContentType {
				t2.Errorf("Unexpected status code: Got(%s), want(%s)", resp.Header.Get("Content-Type"), v.expContentType)
			}
			if !v.bodyTest(v.reqBody, string(resbody)) {
				t2.Errorf("Unexpected response body: Got::\n%s", resbody)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t2.Errorf("Expectations not met: %+v", err)
			}
			mock.ExpectClose()
			db.Close()
		})
	}
}
