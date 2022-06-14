package transport

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/klwxsrx/arch-course-project/pkg/auth/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/auth/infra/auth"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/transport"
	"net/http"
)

const healthEndpoint = "/healthz"

type credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type route struct {
	Name    string
	Method  string
	Pattern string
	Handler func(*service.UserService, *auth.SessionService, http.ResponseWriter, *http.Request)
}

func getRoutes() []route {
	return []route{
		{
			"auth",
			http.MethodGet,
			"/auth",
			authHandler,
		},
		{
			"login",
			http.MethodPost,
			"/auth/login",
			loginHandler,
		},
		{
			"logout",
			http.MethodPost,
			"/auth/logout",
			logoutHandler,
		},
		{
			"register",
			http.MethodPost,
			"/auth/register",
			registerHandler,
		},
		{
			"health",
			http.MethodGet,
			healthEndpoint,
			healthCheckHandler,
		},
	}
}

func authHandler(_ *service.UserService, sessionService *auth.SessionService, w http.ResponseWriter, r *http.Request) {
	sessionService.Auth(r, w)
}

func loginHandler(_ *service.UserService, sessionService *auth.SessionService, w http.ResponseWriter, r *http.Request) {
	var credentials credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionService.Login(credentials.Login, credentials.Password, w)
}

func logoutHandler(_ *service.UserService, sessionService *auth.SessionService, w http.ResponseWriter, r *http.Request) {
	sessionService.Logout(r, w)
}

func registerHandler(userService *service.UserService, _ *auth.SessionService, w http.ResponseWriter, r *http.Request) {
	var credentials credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := userService.Register(credentials.Login, credentials.Password)
	if errors.Is(err, service.ErrUserByLoginAlreadyExists) {
		w.WriteHeader(http.StatusConflict)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		_ = json.NewEncoder(w).Encode(userID.String())
	}
}

func healthCheckHandler(_ *service.UserService, _ *auth.SessionService, w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{"OK"})
}

func getHandlerFunc(
	userService *service.UserService,
	sessionService *auth.SessionService,
	f func(*service.UserService, *auth.SessionService, http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(userService, sessionService, w, r)
	}
}

func NewHTTPHandler(userService *service.UserService, sessionService *auth.SessionService, logger log.Logger) (http.Handler, error) {
	router := mux.NewRouter()

	for _, route := range getRoutes() {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(getHandlerFunc(userService, sessionService, route.Handler))
	}

	router.Use(transport.NewLoggingMiddleware(logger, []string{healthEndpoint}))
	return router, nil
}
