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

// GetCurrentUser route handler (convert to AppHandler)
func GetCurrentUser(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, getUser}
}

func getUser(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Get Data
	u, err := getUserFromContext(r)
	if err != nil {
		return err
	}
	token, err := getTokenFromContext(r)
	if err != nil {
		return err
	}
	u.Token = token

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.UserResponse{User: u})
	return nil
}

type updUser struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	Bio      *string `json:"bio"`
	Image    *string `json:"image"`
}
type updateJSONPost struct {
	User *updUser `json:"user"`
}

func (ujp updateJSONPost) Validate() error {
	validations := invalidInputError{}
	if ujp.User.Email != nil {
		if _, err := mail.ParseAddress(*ujp.User.Email); err != nil {
			validations.Errs["email"] = []string{"is invalid"}
		}
	}
	if ujp.User.Password != nil && len(*ujp.User.Password) < models.PasswordLengthRequirement {
		validations.Errs["password"] = []string{fmt.Sprintf("is too short (minimum is %d characters)", models.PasswordLengthRequirement)}
	}
	if len(validations.Errs) > 0 {
		return validations
	}
	return nil
}

// UpdateUser allows the user to update their details
func UpdateUser(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, updateUser}
}

func updateUser(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Parse
	updU := updateJSONPost{}
	err := json.NewDecoder(r.Body).Decode(&updU)
	if err != nil {
		return errors.Wrap(err, "updateUser:: JSON decode")
	}
	defer r.Body.Close()

	// Get Current User
	reqU, err := getUserFromContext(r)
	if err != nil {
		return err
	}
	u, err := ae.DB.UserByID(reqU.ID)
	if err != nil {
		return errors.Wrap(err, "updateUser:: DB.UserByID()")
	}

	if err = updU.Validate(); err != nil {
		return err
	}

	changed := false
	if updU.User.Username != nil {
		u.Username = *updU.User.Username
		changed = true
	}
	if updU.User.Email != nil {
		u.Email = strings.ToLower(*updU.User.Email)
		changed = true
	}
	if updU.User.Password != nil {
		u.SetPassword(*updU.User.Password)
		changed = true
	}
	if updU.User.Bio != nil {
		u.Bio = *updU.User.Bio
		changed = true
	}
	if updU.User.Image != nil {
		u.Image = *updU.User.Image
		changed = true
	}

	// Get Token for response
	token, err := getTokenFromContext(r)
	if err != nil {
		return err
	}

	u.Token = token

	if !changed {
		return invalidInputError{Errs: map[string][]string{"body": []string{"no changes detected"}}}
	}

	// Persist
	if err := ae.DB.UpdateUser(*u); err != nil {
		return errors.Wrap(err, "updateUser:: DB.UpdateUser()")
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.UserResponse{User: u})

	return nil
}
