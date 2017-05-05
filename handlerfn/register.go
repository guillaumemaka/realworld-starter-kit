package handlerfn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/chilledoj/realworld-starter-kit/models"
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

func (rfp *registerFormPost) Validate() error {
	_, err := mail.ParseAddress(rfp.User.Email)
	if err != nil {
		return err
	}
	rfp.User.Email = strings.ToLower(rfp.User.Email)

	if len(rfp.User.Username) < models.UsernameLengthRequirement {
		return fmt.Errorf("Username should be at least %d characters long", models.UsernameLengthRequirement)
	}
	if len(rfp.User.Password) < models.PasswordLengthRequirement {
		return fmt.Errorf("Password should be at least %d characters long", models.PasswordLengthRequirement)
	}
	return nil
}

func register(env *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Parse
	regUsr := registerFormPost{}
	err := json.NewDecoder(r.Body).Decode(&regUsr)
	if err != nil {
		return &AppError{Err: []error{err}}
	}
	defer r.Body.Close()

	// Validate
	if err = regUsr.Validate(); err != nil {
		return &AppError{StatusCode: http.StatusBadRequest, Err: []error{err}}
	}

	// Persist
	u, err := env.DB.CreateUser(regUsr.User.Email, regUsr.User.Username, regUsr.User.Password)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{fmt.Errorf("Unable to create user"), err}}
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
