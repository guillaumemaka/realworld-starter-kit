package handlerfn

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chilledoj/realworld-starter-kit/models"
)

// AppEnvironment holds the database connection pool and logger
type AppEnvironment struct {
	DB     *models.AppDB
	Logger *log.Logger
}

// AppError is a struct to manage return codes in errors
type AppError struct {
	StatusCode int
	Err        []error
}

// MarshalJSON implements JSON encoding on AppError
func (ae AppError) MarshalJSON() ([]byte, error) {
	type errSlice struct {
		Body []string `json:"body"`
	}

	errs := make([]string, len(ae.Err))
	for i, v := range ae.Err {
		errs[i] = v.Error()
	}
	c := struct {
		Errors errSlice `json:"errors"`
	}{
		Errors: errSlice{Body: errs},
	}
	return json.Marshal(c)
}

// AppHandler is a struct to manage error providing handlers
type AppHandler struct {
	env *AppEnvironment
	fn  func(e *AppEnvironment, w http.ResponseWriter, r *http.Request) *AppError
}

func (ah AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := ctx.Err(); err != nil {
		return
	}
	if e := ah.fn(ah.env, w, r); e != nil { // e is *AppError, not os.Error.
		ah.env.Logger.Printf("%+v", e.Err)
		c := http.StatusUnprocessableEntity
		if e.StatusCode != 0 {
			c = e.StatusCode
		}
		w.WriteHeader(c)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(e)

	}
}
