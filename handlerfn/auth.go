package handlerfn

import (
	"fmt"
	"net/http"
)

var notAuthenticated = &AppError{StatusCode: http.StatusForbidden, Err: []error{fmt.Errorf("Not Authenticated")}}

// MustHaveUser checks the request context for the presence of a user
func MustHaveUser(ah AppHandler) AppHandler {
	return AppHandler{env: ah.env, fn: func(env *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError {
		if _, err := getUserFromContext(r); err != nil {
			return notAuthenticated
		}
		return ah.fn(ah.env, w, r)
	},
	}
}
