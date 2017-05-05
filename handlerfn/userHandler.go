package handlerfn

import (
	"chilledoj/myreal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GetCurrentUser route handler (convert to AppHandler)
func GetCurrentUser(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, getUser}
}

func getUser(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Get Data
	u, err := getUserFromContext(r)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}
	token, err := getTokenFromContext(r)
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

type updUser struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}
type updateJSONPost struct {
	User *updUser `json:"user"`
}

// UpdateUser allows the user to update their details
func UpdateUser(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, updateUser}
}

func updateUser(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Parse
	updU := updateJSONPost{}
	err := json.NewDecoder(r.Body).Decode(&updU)
	if err != nil {
		return &AppError{Err: []error{err}}
	}
	defer r.Body.Close()

	// Get Current User
	reqU, err := getUserFromContext(r)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}
	u, err := ae.DB.UserByID(reqU.ID)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}

	// Check for Changes - if present in provided data, apply to model
	changed := false
	if updU.User.Username != "" && len(updU.User.Username) >= models.UsernameLengthRequirement {
		u.Username = updU.User.Username
		changed = true
	}
	if updU.User.Email != "" {
		u.Email = strings.ToLower(updU.User.Email)
		changed = true
	}
	if updU.User.Password != "" && len(updU.User.Password) >= models.PasswordLengthRequirement {
		u.SetPassword(updU.User.Password)
		changed = true
	}
	if updU.User.Bio != "" {
		u.Bio = updU.User.Bio
		changed = true
	}
	if updU.User.Image != "" {
		u.Image = updU.User.Image
		changed = true
	}
	if errs := u.ValidateValues(); len(errs) != 0 {
		return &AppError{StatusCode: http.StatusBadRequest, Err: errs}
	}

	// Get Token for response
	token, err := getTokenFromContext(r)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{fmt.Errorf("Problem getting user details")}}
	}

	u.Token = token

	if !changed {
		return &AppError{Err: []error{fmt.Errorf("No updates found")}}
	}

	// Persist
	if err := ae.DB.UpdateUser(*u); err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{fmt.Errorf("Error saving updates"), err}}
	}

	// Response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(models.UserResponse{User: u})

	return nil
}
