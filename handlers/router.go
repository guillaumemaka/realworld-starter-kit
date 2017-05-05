package handlers

import (
	"log"
	"net/http"
	"regexp"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

type Route struct {
	Pattern        string
	ActionHandlers map[string]ActionRoute
}

type ActionRoute struct {
	Action
	HandlerFunc
}

type Router struct {
	routes []Route
	logger *log.Logger
	debug  bool
}

func NewRouter(logger *log.Logger) *Router {
	return &Router{
		routes: make([]Route, 0),
		logger: logger,
		debug:  false,
	}
}

// AddRoute add a new route to the router
func (r *Router) AddRoute(pattern string, method string, action Action, handler HandlerFunc) {
	var found = false
	for _, route := range r.routes {
		if route.Pattern == pattern {
			found = true
			route.ActionHandlers[method] = ActionRoute{
				Action:      action,
				HandlerFunc: handler,
			}
		}
	}

	if !found {
		r.routes = append(r.routes, Route{
			Pattern: pattern,
			ActionHandlers: map[string]ActionRoute{
				method: ActionRoute{
					Action:      action,
					HandlerFunc: handler,
				},
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
			if actionRoute, registered := route.ActionHandlers[r.Method]; registered {
				if router.debug {
					router.logger.Println(r.Method, r.URL.Path)
				}
				actionRoute.HandlerFunc(w, r)
			} else {
				http.NotFound(w, r)
			}
			return
		}
	}
}
