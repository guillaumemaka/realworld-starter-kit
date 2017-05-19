package handlerfn

import (
	"net/http"
)

// MustHaveUser checks the request context for the presence of a user
func MustHaveUser(ah AppHandler) AppHandler {
	return AppHandler{env: ah.env, fn: func(env *AppEnvironment, w http.ResponseWriter, r *http.Request) error {
		if _, err := getUserFromContext(r); err != nil {
			return forbidden{}
		}
		return ah.fn(ah.env, w, r)
	},
	}
}
