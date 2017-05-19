package handlerfn

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/chilledoj/realworld-starter-kit/models"
	"github.com/pkg/errors"
)

// Jwt2Ctx is effectively a middleware struct in http.Handler form that puts
// the user stored in JWT claims into the request context (GO >=1.7)
type Jwt2Ctx struct {
	Env *AppEnvironment
	Fn  http.Handler
}

// ServeHTTP implements http.Handler interface
func (m Jwt2Ctx) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		token := strings.TrimPrefix(authHeader, "Token ") // API Spec mentions Token instead of Bearer
		ctx, err := storeJWTUserCtx(token, r)
		if err != nil {
			m.Env.Logger.Printf("JWT Validation err: %+v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct {
				Errors string `json:"error"`
			}{Errors: "Invalid Token"})
		}
		r = r.WithContext(ctx)
	}

	m.Fn.ServeHTTP(w, r)
	return
}

type key int

const userKey key = 0
const tokenKey key = 1

func storeJWTUserCtx(token string, r *http.Request) (context.Context, error) {
	ctx := r.Context()
	claims, err := models.ValidateToken(token)
	if err != nil {
		return ctx, err
	}
	ctx = context.WithValue(ctx, userKey, claims.User)
	ctx = context.WithValue(ctx, tokenKey, token)

	return ctx, nil
}

func getUserFromContext(r *http.Request) (*models.User, error) {
	ctx := r.Context()
	u, ok := ctx.Value(userKey).(*models.User)
	// check if u==nil first because we don't care if ok is true/false if u is nil
	if u == nil || !ok {
		return nil, errors.New("No User in context")
	}
	return u, nil
}

func getTokenFromContext(r *http.Request) (string, error) {
	ctx := r.Context()
	token, ok := ctx.Value(tokenKey).(string)
	// check if token=="" first because we don't care if ok is true/false if token is ""
	if token == "" || !ok {
		return "", errors.New("No Token in context")
	}
	return token, nil
}
