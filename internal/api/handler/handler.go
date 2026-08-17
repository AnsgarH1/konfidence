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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type apiHandler struct {
	authHandler
	projectHandler
}

type KubernetesAccess struct {
	Client       client.Client
	CachedReader client.Reader
}

var _ openapi.StrictServerInterface = (*apiHandler)(nil)

func NewAPIHandler(logger *slog.Logger, kubernetes KubernetesAccess, oidcClient oidc.Client,
	sessionStore session.Store, cfg config.Parsed) (http.Handler, error) {
	auth := newAuthHandler(logger, oidcClient, oidc.NewStateCacheStore(cfg), sessionStore, cfg)
	project := newProjectHandler(kubernetes)
	api := &apiHandler{
		authHandler:    *auth,
		projectHandler: *project,
	}
	authorized := middleware.ProjectAuthorization(logger, project.projects, api.handler())
	return middleware.SessionAuthentication(logger, sessionStore, cfg, authorized)
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
