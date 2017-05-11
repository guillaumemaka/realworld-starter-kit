package handlerfn

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/gorilla/mux"
)

// GetProfile is the handler for the Get Profile route
func GetProfile(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, getProfile}
}

func getProfile(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	// Parse
	vars := mux.Vars(r)
	username := vars["username"]

	// This should return a nil pointer if the user is not authenticated
	u, err := getUserFromContext(r)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}
	var id uint
	if u != nil {
		id = u.ID
	}
	// id could be zero value here, which is fine based on the way the query has been
	// written. The logic to find if FOLLOWING should be true or not is a LEFT OUTER
	// JOIN and so if a value of zero is passed in, then FOLLOWING would just be false.
	p, err := ae.DB.GetProfileByUsername(username, id)
	if err != nil {
		return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.ProfileResponse{Profile: p})
	return nil
}

// FollowUser is the Follow User route
func FollowUser(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, followUser}
}

func followUser(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	u, p, err := getUserAndProfile(ae, r)
	if err != nil {
		return err
	}
	if !p.Following {
		// Follow
		if err := ae.DB.FollowUser(u.ID, p.ID); err != nil {
			return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
		}
		p.Following = true
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.ProfileResponse{Profile: p})
	return nil
}

// UnfollowUser is the Unfollow User route
func UnfollowUser(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, unfollowUser}
}

func unfollowUser(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
	u, p, err := getUserAndProfile(ae, r)
	if err != nil {
		return err
	}
	if p.Following {
		// Unfollow
		if err := ae.DB.UnfollowUser(u.ID, p.ID); err != nil {
			return &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
		}
		p.Following = false
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.ProfileResponse{Profile: p})
	return nil
}

func getUserAndProfile(ae *AppEnvironment, r *http.Request) (*models.User, *models.Profile, *AppError) {
	// Parse
	vars := mux.Vars(r)
	username := vars["username"]

	u, err := getUserFromContext(r)
	if err != nil {
		return nil, nil, &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}
	if u == nil {
		return nil, nil, &AppError{StatusCode: http.StatusForbidden, Err: []error{fmt.Errorf("Not authenticated")}}
	}

	// Get Profile
	p, err := ae.DB.GetProfileByUsername(username, u.ID)
	if err != nil {
		return nil, nil, &AppError{StatusCode: http.StatusInternalServerError, Err: []error{err}}
	}
	return u, p, nil
}
