package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

type Route struct {
	Pattern        string
	ActionHandlers map[string]http.HandlerFunc
}

type Router struct {
	routes []Route
	logger *log.Logger
	debug  bool
}

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

func NewRouter(logger *log.Logger) *Router {
	return &Router{
		routes: make([]Route, 0),
		logger: logger,
		debug:  false,
	}
}

// AddRoute add a new route to the router
func (r *Router) AddRoute(pattern string, method string, handler http.HandlerFunc) {
	var found = false
	for _, route := range r.routes {
		if route.Pattern == pattern {
			found = true
			route.ActionHandlers[method] = handler
		}
	}

	if !found {
		r.routes = append(r.routes, Route{
			Pattern: pattern,
			ActionHandlers: map[string]http.HandlerFunc{
				method: handler,
			},
		})
	}
}

func (router *Router) DebugMode(enabled bool) {
	router.debug = enabled
}

func (router *Router) Dispatch(w http.ResponseWriter, r *http.Request) {
	for _, route := range router.routes {
		if matched, _ := regexp.MatchString(route.Pattern, r.URL.Path); matched {
			if h, registered := route.ActionHandlers[r.Method]; registered {
				if router.debug {
					router.logger.Println(r.Method, r.URL.Path)
				}
				r = r.WithContext(buildContext(route.Pattern, r))
				h(w, r)
			} else {
				http.NotFound(w, r)
			}
			return
		}
	}
}

func buildContext(pattern string, r *http.Request) context.Context {
	re := regexp.MustCompile(pattern)
	n1 := re.SubexpNames()
	r2 := re.FindAllStringSubmatch(r.URL.Path, -1)

	ctx := r.Context()

	if len(r2) > 0 {
		for i, n := range r2[0] {
			if n1[i] != "" {
				ctx = context.WithValue(ctx, n1[i], n)
				fmt.Println("Context: ", ctx)
			}

			fmt.Printf("%d. match='%s'\tname='%s'\n", i, n, n1[i])
		}
	}
	return ctx
}
