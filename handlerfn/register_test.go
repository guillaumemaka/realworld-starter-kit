package handlerfn

import (
	"encoding/json"
	"net/http"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const registerURL = "/api/register"

func TestRegister(t *testing.T) {
	tests := []fnTest{
		{"Register:NoBody", http.MethodPost, registerURL, "", nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {},
			func(reqBody, resBody string) bool {
				return len(resBody) == 0
			}},
		{"Register:EmptyJSON", http.MethodPost, registerURL, "{}", nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {},
			func(reqBody, resBody string) bool {
				invalid := invalidInputError{}
				if err := json.Unmarshal([]byte(resBody), &invalid); err != nil {
					t.Fatal(err)
				}
				return len(invalid.Errs["username"]) > 0 && len(invalid.Errs["email"]) > 0 && len(invalid.Errs["password"]) > 0
			}},
		{"Register:OnlyEmail", http.MethodPost, registerURL, `{"user":{"email":"test.com"}}`, nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {},
			func(reqBody, resBody string) bool {
				invalid := invalidInputError{}
				if err := json.Unmarshal([]byte(resBody), &invalid); err != nil {
					t.Fatal(err)
				}
				return len(invalid.Errs["username"]) > 0 && len(invalid.Errs["password"]) > 0
			}},
		{"Register:OnlyUsername", http.MethodPost, registerURL, `{"user":{"username":"a"}}`, nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {},
			func(reqBody, resBody string) bool {
				invalid := invalidInputError{}
				if err := json.Unmarshal([]byte(resBody), &invalid); err != nil {
					t.Fatal(err)
				}
				return len(invalid.Errs["password"]) > 0 && len(invalid.Errs["email"]) > 0
			}},
		{"Register:OnlyPassword", http.MethodPost, registerURL, `{"user":{"password":"password"}}`, nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {},
			func(reqBody, resBody string) bool {
				invalid := invalidInputError{}
				if err := json.Unmarshal([]byte(resBody), &invalid); err != nil {
					t.Fatal(err)
				}
				return len(invalid.Errs["username"]) > 0 && len(invalid.Errs["email"]) > 0
			}},
		{"Register:NoEmail", http.MethodPost, registerURL, `{"user":{"username":"a","password":"password"}}`, nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {},
			func(reqBody, resBody string) bool {
				invalid := invalidInputError{}
				if err := json.Unmarshal([]byte(resBody), &invalid); err != nil {
					t.Fatal(err)
				}
				return len(invalid.Errs["email"]) > 0
			}},
		{"Register:NoUsername", http.MethodPost, registerURL, `{"user":{"email":"email@test.com","password":"password"}}`, nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {},
			func(reqBody, resBody string) bool {
				invalid := invalidInputError{}
				if err := json.Unmarshal([]byte(resBody), &invalid); err != nil {
					t.Fatal(err)
				}
				return len(invalid.Errs["username"]) > 0
			}},
		{"Register:NoPassword", http.MethodPost, registerURL, `{"user":{"email":"email@test.com","username":"a"}}`, nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {},
			func(reqBody, resBody string) bool {
				invalid := invalidInputError{}
				if err := json.Unmarshal([]byte(resBody), &invalid); err != nil {
					t.Fatal(err)
				}
				return len(invalid.Errs["password"]) > 0
			}},
		{"Register:InvalidEmail", http.MethodPost, registerURL, `{"user":{"email":"emailtest.com","username":"a","password":"password"}}`, nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {},
			func(reqBody, resBody string) bool {
				invalid := invalidInputError{}
				if err := json.Unmarshal([]byte(resBody), &invalid); err != nil {
					t.Fatal(err)
				}
				return len(invalid.Errs["email"]) > 0
			}},
		{"Register:InvalidPassword", http.MethodPost, registerURL, `{"user":{"email":"email@test.com","username":"a","password":"test"}}`, nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {},
			func(reqBody, resBody string) bool {
				invalid := invalidInputError{}
				if err := json.Unmarshal([]byte(resBody), &invalid); err != nil {
					t.Fatal(err)
				}
				return len(invalid.Errs["password"]) > 0
			}},
		{"Register:ExistingUsernameClash", http.MethodPost, registerURL, `{"user":{"email":"email@test.com","username":"a","password":"password"}}`, nil, http.StatusUnprocessableEntity, true, jsn,
			func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"num"}).AddRow(1)
				mock.ExpectQuery("SELECT count(.)+ FROM users WHERE username=?").WithArgs("a").WillReturnRows(rows)
			}, func(reqBody, resBody string) bool {
				invalid := invalidInputError{}
				if err := json.Unmarshal([]byte(resBody), &invalid); err != nil {
					t.Fatal(err)
				}
				return len(invalid.Errs["username"]) > 0
			}},
		{"Register:ValidEntry", http.MethodPost, registerURL, `{"user":{"email":"email@test.com","username":"b","password":"password"}}`, nil, http.StatusOK, false, jsn,
			func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"num"}).AddRow(0)
				mock.ExpectQuery("SELECT count(.)+ FROM users WHERE username=?").WithArgs("b").WillReturnRows(rows)
				mock.ExpectPrepare("INSERT INTO users (.)+").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
			},
			func(reqBody, resBody string) bool {
				return true
			}},
	}
	//handler := Register(ae)
	handlerTest(t, Register, tests)

}
