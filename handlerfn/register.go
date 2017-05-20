package handlerfn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/pkg/errors"
)

// Register route handler (convert to AppHandler)
func Register(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, register}
}

type registerUser struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type registerFormPost struct {
	User *registerUser `json:"user"`
}

func (ru *registerUser) Validate() error {
	validations := invalidInputError{Errs: make(map[string][]string)}
	if ru.Email == "" {
		validations.Errs["email"] = []string{"can't be blank"}
	} else if _, err := mail.ParseAddress(ru.Email); err != nil {
		validations.Errs["email"] = []string{"Invalid email address provided"}
	}
	ru.Email = strings.ToLower(ru.Email)
	if ru.Username == "" {
		validations.Errs["username"] = []string{"can't be blank"}
	}

	if ru.Password == "" {
		validations.Errs["password"] = []string{"can't be blank"}
	} else if len(ru.Password) < models.PasswordLengthRequirement {
		validations.Errs["password"] = []string{fmt.Sprintf("is too short (minimum is %d characters)", models.PasswordLengthRequirement)}
	}

	if len(validations.Errs) > 0 {
		return validations
	}
	return nil
}

func register(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Parse
	regUsr := registerFormPost{}
	err := json.NewDecoder(r.Body).Decode(&regUsr)
	if err != nil {
		return badRequest{err: errors.Wrap(err, "register:: jsonDecode")}
	}
	defer r.Body.Close()
	// Validate
	if regUsr.User == nil {
		e := registerUser{}
		return e.Validate() //shortcut
	}
	validationErrors := invalidInputError{Errs: make(map[string][]string)}
	errs := regUsr.User.Validate()
	if vals, ok := errs.(invalidInputError); ok {
		validationErrors = vals
	}
	// CheckUsername
	if len(validationErrors.Errs["username"]) == 0 {
		n, _ := ae.DB.CountUsername(regUsr.User.Username)
		if n > 0 {
			validationErrors.Errs["username"] = []string{"username already taken"}
			return validationErrors
		}
	}

	if len(validationErrors.Errs) > 0 {
		return validationErrors
	}

	// Persist
	u, err := ae.DB.CreateUser(regUsr.User.Email, regUsr.User.Username, regUsr.User.Password)
	if err != nil {
		return errors.Wrap(err, "register:: DB.CreateUser()")
	}

	// JWT Token / Login
	token, err := models.NewToken(u)
	if err != nil {
		return errors.Wrap(err, "register:: NewToken()")
	}
	u.Token = token

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.UserResponse{User: u})
	return nil
}
