package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
	landscapedomain "github.com/konfidence-project/konfidence/internal/landscape"
	projectdomain "github.com/konfidence-project/konfidence/internal/project"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Mount(r chi.Router, _ *slog.Logger, k8s client.Client) {
	h := NewServerHandler(k8s)
	errHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		if apiErr := AsAPIError(err); apiErr != nil {
			WriteAPIError(w, apiErr)
			return
		}
		WriteInternalError(w)
	}
	openapi.HandlerWithOptions(openapi.NewStrictHandlerWithOptions(h, nil, openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  errHandler,
		ResponseErrorHandlerFunc: errHandler,
	}), openapi.ChiServerOptions{
		BaseURL:    "/api/v1",
		BaseRouter: r,
	})
}

func NewServerHandler(k8s client.Client) *ServerHandler {
	return &ServerHandler{
		InfoHandler{},
		AuthHandler{},
		newProjectHandler(k8s),
	}
}

func NewServerHandlerWithRepos(projects projectdomain.Repository, landscapes landscapedomain.Repository) *ServerHandler {
	return &ServerHandler{
		InfoHandler{},
		AuthHandler{},
		ProjectHandler{projects: projects, landscapes: landscapes},
	}
}

type ServerHandler struct {
	InfoHandler
	AuthHandler
	ProjectHandler
}

var _ openapi.StrictServerInterface = (*ServerHandler)(nil)
