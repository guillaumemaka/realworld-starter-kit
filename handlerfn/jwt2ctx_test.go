package handlerfn

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/chilledoj/realworld-starter-kit/models"
)

var user models.User

func init() {
	user = models.User{ID: 1, Username: "Testing", Email: "test@mctesterton.com"}
}

func Test_storeJWTUserCtx(t *testing.T) {

	token, err := models.NewToken(&user)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"ValidToken", token, false},
		{"InvalidToken", "XXX", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			ctx, err := storeJWTUserCtx(tt.token, r)
			if (err != nil) != tt.wantErr {
				t.Errorf("storeJWTUserCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			gotToken, ok := ctx.Value(tokenKey).(string)
			if !ok || gotToken != tt.token {
				t.Errorf("storeJWTUserCtx() did not store token correctly = %s, want %s", token, tt.token)
				return
			}
			usr := ctx.Value(userKey).(*models.User)
			if usr.ID != user.ID || usr.Username != user.Username {
				t.Errorf("storeJWTUserCtx() did not store user correctly %v, want %v", usr, user)
			}
		})
	}
}
func Test_getUserFromContext(t *testing.T) {
	token, err := models.NewToken(&user)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		storeUser bool
		wantErr   bool
	}{
		{"Authorised", true, false},
		{"Unauthorised", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			ctx := r.Context()
			if !tt.storeUser && tt.wantErr {

				if u, err := getUserFromContext(r); err == nil {
					t.Errorf("getUserFromContext() did not return an error %v", u)
				}
				return
			}

			claims, err := models.ValidateToken(token)
			if err != nil {
				t.Errorf("getUserFromContext::ValidateToken() return an unexpected error %v", err)
				return
			}
			ctx = context.WithValue(ctx, userKey, claims.User)
			ctx = context.WithValue(ctx, tokenKey, token)

			r = r.WithContext(ctx)
			u, err := getUserFromContext(r)

			if u.ID != user.ID || u.Username != user.Username {
				t.Errorf("getUserFromContext() did not store user correctly %v, want %v", claims.User, user)
			}

		})
	}
}
