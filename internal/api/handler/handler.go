package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/konfidence-project/konfidence/internal/api/apierror"
	"github.com/konfidence-project/konfidence/internal/api/config"
	"github.com/konfidence-project/konfidence/internal/api/middleware"
	"github.com/konfidence-project/konfidence/internal/api/oidc"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
	"github.com/konfidence-project/konfidence/internal/api/session"
	landscapedomain "github.com/konfidence-project/konfidence/internal/landscape"
	projectdomain "github.com/konfidence-project/konfidence/internal/project"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type apiHandler struct {
	authHandler
	projectHandler
}

// APIHandler is the exported type for testing.
type APIHandler = apiHandler

var _ openapi.StrictServerInterface = (*apiHandler)(nil)

func NewAPIHandler(logger *slog.Logger, k8s client.Client, oidcClient oidc.Client,
	sessionStore session.Store, cfg config.Parsed) (http.Handler, error) {
	auth := newAuthHandler(logger, oidcClient, oidc.NewStateCacheStore(cfg), sessionStore, cfg)
	project := newProjectHandler(k8s)
	api := &apiHandler{
		authHandler:    *auth,
		projectHandler: *project,
	}
	return middleware.SessionAuthentication(logger, sessionStore, cfg, api.handler())
}

// NewAPIHandlerWithRepos creates an apiHandler with injected repositories — for testing only.
func NewAPIHandlerWithRepos(projects projectdomain.Repository, landscapes landscapedomain.Repository) *apiHandler {
	return &apiHandler{
		projectHandler: *newProjectHandlerWithRepos(projects, landscapes),
	}
}

func (s *apiHandler) handler() http.Handler {
	errHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		if apiErr := apierror.As(err); apiErr != nil {
			apierror.Write(w, apiErr)
			return
		}
		apierror.WriteInternal(w)
	}

	apiRouter := chi.NewRouter()
	apiRouter.Mount("/api",
		openapi.Handler(openapi.NewStrictHandlerWithOptions(s, nil, openapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  errHandler,
			ResponseErrorHandlerFunc: errHandler,
		})))
	return apiRouter
}
