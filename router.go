package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	hfn "github.com/chilledoj/realworld-starter-kit/handlerfn"
	"github.com/chilledoj/realworld-starter-kit/models"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func buildRouter(db *models.AppDB, logger *log.Logger, fullyqualifiedlisteningaddress string) http.Handler {

	env := hfn.AppEnvironment{DB: db, Logger: logger}

	//na := notImplemented{}
	r := mux.NewRouter()

	/*
	 Public files handler for front end files, if required
	 Copy the compiled front end index.html into the public folder and the
	 compiled js assets into the public/static folder
	*/
	r.Handle("/", http.FileServer(http.Dir("./public")))
	r.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./public/static"))))

	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	api := r.PathPrefix("/api").Subrouter()

	// User Routes
	api.Handle("/users", hfn.Register(&env)).Methods("POST").Name("Registration")
	api.Handle("/users/login", hfn.Login(&env)).Methods("POST").Name("Authentication")
	api.Handle("/user", hfn.MustHaveUser(hfn.GetCurrentUser(&env).(hfn.AppHandler))).Methods("GET").Name("Get Current User")
	api.Handle("/user", hfn.MustHaveUser(hfn.UpdateUser(&env).(hfn.AppHandler))).Methods("PUT").Name("Update User")

	// Profile Routes
	api.Handle("/profiles/{username}", hfn.GetProfile(&env)).Methods("GET").Name("Get Profile")
	api.Handle("/profiles/{username}/follow", hfn.MustHaveUser(hfn.FollowUser(&env).(hfn.AppHandler))).Methods("POST").Name("Follow User")
	api.Handle("/profiles/{username}/follow", hfn.MustHaveUser(hfn.UnfollowUser(&env).(hfn.AppHandler))).Methods("DELETE").Name("Unfollow User")

	// Article Routes
	api.Handle("/articles", hfn.ListArticles(&env)).Methods("GET").Name("List Articles")
	api.Handle("/articles", hfn.MustHaveUser(hfn.CreateArticle(&env).(hfn.AppHandler))).Methods("POST").Name("Create Article")
	api.Handle("/articles/feed", hfn.MustHaveUser(hfn.FeedArticles(&env).(hfn.AppHandler))).Methods("GET").Name("Feed Articles")
	api.Handle("/articles/{slug}/favorite", hfn.MustHaveUser(hfn.FavouriteArticle(&env).(hfn.AppHandler))).Methods("POST").Name("Favourite Article")
	api.Handle("/articles/{slug}/favorite", hfn.MustHaveUser(hfn.UnfavouriteArticle(&env).(hfn.AppHandler))).Methods("DELETE").Name("Unavourite Article")
	api.Handle("/articles/{slug}/comments", hfn.GetComments(&env)).Methods("GET").Name("Get Comments from an Article")
	api.Handle("/articles/{slug}/comments", hfn.MustHaveUser(hfn.AddComment(&env).(hfn.AppHandler))).Methods("POST").Name("Add Comments to an Article")
	api.Handle("/articles/{slug}/comments/{id}", hfn.MustHaveUser(hfn.DeleteComment(&env).(hfn.AppHandler))).Methods("DELETE").Name("Delete Comment")
	api.Handle("/articles/{slug}", hfn.GetArticle(&env)).Methods("GET").Name("Get Article")
	api.Handle("/articles/{slug}", hfn.MustHaveUser(hfn.UpdateArticle(&env).(hfn.AppHandler))).Methods("PUT").Name("Update Article")
	api.Handle("/articles/{slug}", hfn.MustHaveUser(hfn.DeleteArticle(&env).(hfn.AppHandler))).Methods("DELETE").Name("Delete Article")

	// Tags routes
	api.Handle("/tags", hfn.GetTags(&env)).Methods("GET").Name("Get Tags")

	// Router Handler Chaining (inner most first)
	jwtRouter := hfn.Jwt2Ctx{Env: &env, Fn: r}
	// CORS
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methodsOk := handlers.AllowedMethods([]string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete})

	cfgOrigins := strings.Split(origins, ",")
	originsOk := handlers.AllowedOrigins(append([]string{fullyqualifiedlisteningaddress}, cfgOrigins...))

	corsRouter := handlers.CORS(originsOk, headersOk, methodsOk)(jwtRouter)
	logRouter := handlers.CombinedLoggingHandler(os.Stdout, corsRouter)
	return handlers.RecoveryHandler()(logRouter)
}

/*
type notImplemented struct{}

func (notImplemented) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "NOT IMPLEMENTED")
}
*/
