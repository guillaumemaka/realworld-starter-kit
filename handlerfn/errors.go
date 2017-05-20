package handlerfn

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type statusCoder interface {
	StatusCode() int
	error
}
type dataError interface {
	Data() []byte
	error
}

type unAuthorised struct{}

func (e unAuthorised) Error() string {
	return http.StatusText(http.StatusUnauthorized)
}
func (e unAuthorised) StatusCode() int {
	return http.StatusUnauthorized
}

type forbidden struct{}

func (forbidden) Error() string {
	return http.StatusText(http.StatusForbidden)
}
func (forbidden) StatusCode() int {
	return http.StatusForbidden
}

type invalidInputError struct {
	Errs map[string][]string `json:"errors"`
}

func (iie invalidInputError) Error() string {
	return fmt.Sprintf("Validation error: %+v", iie.Errs)
}
func (iie invalidInputError) StatusCode() int {
	return http.StatusUnprocessableEntity
}
func (iie invalidInputError) Data() []byte {
	d, err := json.Marshal(iie)
	if err != nil {
		fmt.Printf("OH DEAR AN ERROR ON AN ERROR: %+v", err)
		return []byte("")
	}
	return d
}

type badRequest struct{ err error }

func (br badRequest) Error() string {
	return br.err.Error()
}
func (badRequest) StatusCode() int {
	return http.StatusBadRequest
}
func (badRequest) Data() []byte {
	return []byte(`{"status": "400","error": "Bad Request"}`)
}
