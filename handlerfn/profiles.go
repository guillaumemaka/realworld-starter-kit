package handlerfn

import (
	"encoding/json"
	"net/http"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// GetProfile is the handler for the Get Profile route
func GetProfile(ae *AppEnvironment) http.Handler {
	return AppHandler{ae, getProfile}
}

func getProfile(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	// Parse
	vars := mux.Vars(r)
	username := vars["username"]

	// This should return a nil pointer if the user is not authenticated
	// and we can ignore the error here.
	u, _ := getUserFromContext(r)
	var id uint
	if u != nil {
		id = u.ID
	}
	// id could be zero value here, which is fine based on the way the query has been
	// written. The logic to find if FOLLOWING should be true or not is a LEFT OUTER
	// JOIN and so if a value of zero is passed in, then FOLLOWING would just be false.
	p, err := ae.DB.GetProfileByUsername(username, id)
	if err != nil {
		return errors.Wrap(err, "getProfile:: DB.GetProfileByUsername()")
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

func followUser(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	u, p, err := getUserAndProfile(ae, r)
	if err != nil {
		return err
	}
	if !p.Following {
		// Follow
		if err := ae.DB.FollowUser(u.ID, p.ID); err != nil {
			return errors.Wrap(err, "followUser:: DB.FollowUser()")
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

func unfollowUser(ae *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
	u, p, err := getUserAndProfile(ae, r)
	if err != nil {
		return err
	}
	if p.Following {
		// Unfollow
		if err := ae.DB.UnfollowUser(u.ID, p.ID); err != nil {
			return errors.Wrap(err, "unfollowUser:: DB.UnfollowUser()")
		}
		p.Following = false
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.ProfileResponse{Profile: p})
	return nil
}

func getUserAndProfile(ae *AppEnvironment, r *http.Request) (*models.User, *models.Profile, error) {
	// Parse
	vars := mux.Vars(r)
	username := vars["username"]

	u, err := getUserFromContext(r)
	if err != nil {
		return nil, nil, errors.Wrap(err, "getUserAndProfile:: getUserFromContext()")
	}

	// Get Profile
	p, err := ae.DB.GetProfileByUsername(username, u.ID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "getUserAndProfile:: DB.GetProfileByUsername()")
	}
	return u, p, nil
}
