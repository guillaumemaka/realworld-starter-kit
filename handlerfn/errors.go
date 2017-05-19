package handlerfn

import (
	"fmt"
	"net/http"
)

type statusCoder interface {
	StatusCode() int
	error
}
type validationError interface {
	Messages() map[string][]string
}

type invalidInputError struct {
	Errs map[string][]string `json:"errors"`
}

func (iie invalidInputError) Error() string {
	return fmt.Sprintf("Validation error: %+v", iie.Errs)
}
func (iie invalidInputError) Messages() map[string][]string {
	return iie.Errs
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

type internalError struct{}

func (internalError) Error() string {
	return http.StatusText(http.StatusInternalServerError)
}
func (internalError) StatusCode() int {
	return http.StatusInternalServerError
}
