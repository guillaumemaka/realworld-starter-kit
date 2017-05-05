package handlerfn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/chilledoj/realworld-starter-kit/models"
)

// Login returns JWT on successful validation of provided credentials
func Login(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, login}
}

type loginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type loginJSONPost struct {
	User *loginUser `json:"user"`
}

func (login *loginJSONPost) Validate() error {
	_, err := mail.ParseAddress(login.User.Email)
	if err != nil {
		return err
	}
	login.User.Email = strings.ToLower(login.User.Email)

	if len(login.User.Password) < models.PasswordLengthRequirement {
		return fmt.Errorf("Password should be at least %d characters long", models.PasswordLengthRequirement)
	}
	return nil
}

func login(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Parse
	loginUsr := loginJSONPost{}
	err := json.NewDecoder(r.Body).Decode(&loginUsr)
	if err != nil {
		return &AppError{Err: []error{err}}
	}
	defer r.Body.Close()

	// Validate Post
	if err = loginUsr.Validate(); err != nil {
		return &AppError{Err: []error{err}}
	}

	// Check Credentials
	u, err := ae.DB.UserByEmail(loginUsr.User.Email)
	if err != nil {
		return &AppError{StatusCode: http.StatusBadRequest, Err: []error{err}}
	}
	if u == nil {
		return &AppError{StatusCode: http.StatusBadRequest, Err: []error{err}}
	}
	if err = u.ValidatePassword(loginUsr.User.Password); err != nil {
		return &AppError{StatusCode: http.StatusBadRequest, Err: []error{err}}
	}

	// JWT Token / Login
	token, err := models.NewToken(u)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}
	u.Token = token

	// Response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(models.UserResponse{User: u})
	return nil
}
