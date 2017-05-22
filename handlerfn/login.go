package handlerfn

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/pkg/errors"
)

// Login returns JWT on successful validation of provided credentials
func Login(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, login}
}

var invalidLogin = invalidInputError{
	Errs: map[string][]string{"email or password": []string{"is invalid"}},
}

type loginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type loginJSONPost struct {
	User *loginUser `json:"user"`
}

func (login *loginJSONPost) Validate() error {

	if login.User.Email == "" || login.User.Password == "" {
		return invalidLogin
	}
	if _, err := mail.ParseAddress(login.User.Email); err != nil {
		return invalidLogin
	}
	login.User.Email = strings.ToLower(login.User.Email)

	if len(login.User.Password) < models.PasswordLengthRequirement {
		return invalidLogin
	}
	return nil
}

func login(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Parse
	loginUsr := loginJSONPost{}
	err := json.NewDecoder(r.Body).Decode(&loginUsr)
	if err != nil {
		return errors.Wrap(err, "login:: jsonDecode")
	}
	defer r.Body.Close()

	// Validate Post
	if err = loginUsr.Validate(); err != nil {
		return err
	}

	// Check Credentials
	u, err := ae.DB.UserByEmail(loginUsr.User.Email)
	if err == sql.ErrNoRows {
		return invalidLogin
	}
	if err != nil {
		return errors.Wrap(err, "login:: DB.UserByEmail()")
	}
	if u == nil {
		return errors.Wrap(err, "login:: DB.UserByEmail() u==nil")
	}
	if err = u.ValidatePassword(loginUsr.User.Password); err != nil {
		return invalidLogin
	}

	// JWT Token / Login
	token, err := models.NewToken(u)
	if err != nil {
		return errors.Wrap(err, "login:: models.NewToken()")
	}
	u.Token = token

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.UserResponse{User: u})
	return nil
}
