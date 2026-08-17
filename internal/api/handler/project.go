package handler // nolint

import (
	"context"
	"errors"
	"fmt"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
	landscapedomain "github.com/konfidence-project/konfidence/internal/landscape"
	projectdomain "github.com/konfidence-project/konfidence/internal/project"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type projectHandler struct {
	projects   projectdomain.Repository
	landscapes landscapedomain.Repository
}

func newProjectHandler(k8s client.Client) *projectHandler {
	return &projectHandler{
		projects:   projectdomain.NewRepository(k8s),
		landscapes: landscapedomain.NewRepository(k8s),
	}
}

func newProjectHandlerWithRepos(projects projectdomain.Repository, landscapes landscapedomain.Repository) *projectHandler {
	return &projectHandler{projects: projects, landscapes: landscapes}
}

func (h *projectHandler) ListProjectsV1(_ context.Context, _ openapi.ListProjectsV1RequestObject) (openapi.ListProjectsV1ResponseObject, error) {
	return openapi.ListProjectsV1200JSONResponse{
		Data: []openapi.Project{
			{Id: "sample-project", Name: "Sample Project"},
		},
	}, nil
}

func (h *projectHandler) ListLandscapesV1(ctx context.Context, req openapi.ListLandscapesV1RequestObject) (openapi.ListLandscapesV1ResponseObject, error) {
	project, err := h.projects.Get(ctx, req.ProjectId)
	if err != nil {
		if errors.Is(err, projectdomain.ErrNotFound) {
			return openapi.ListLandscapesV1404JSONResponse{
				NotFoundJSONResponse: openapi.NotFoundJSONResponse{
					Error: struct {
						Code    string `json:"code"`
						Message string `json:"message"`
					}{
						Code:    "not_found",
						Message: fmt.Sprintf("project %q not found", req.ProjectId),
					},
				},
			}, nil
		}
		return openapi.ListLandscapesV1500JSONResponse{}, nil
	}

	if project.Status.Namespace == "" {
		return openapi.ListLandscapesV1500JSONResponse{}, nil
	}

	landscapes, err := h.landscapes.ListForProject(ctx, project.Status.Namespace)
	if err != nil {
		return openapi.ListLandscapesV1500JSONResponse{}, nil
	}

	data := make([]openapi.Landscape, len(landscapes))
	for i, l := range landscapes {
		data[i] = toLandscapeResponse(l)
	}

	return openapi.ListLandscapesV1200JSONResponse{Data: data}, nil
}

func toLandscapeResponse(l konfidence.Landscape) openapi.Landscape {
	return openapi.Landscape{
		Id:   l.Name,
		Name: l.Spec.DisplayName,
	}
}

func (h *projectHandler) ListStagesV1(_ context.Context, _ openapi.ListStagesV1RequestObject) (openapi.ListStagesV1ResponseObject, error) {
	return nil, nil
}

func (h *projectHandler) ListVectorDeploymentsV1(_ context.Context,
	_ openapi.ListVectorDeploymentsV1RequestObject) (openapi.ListVectorDeploymentsV1ResponseObject, error) {
	return nil, nil
}

func (h *projectHandler) ListArtifactDeploymentsV1(_ context.Context,
	_ openapi.ListArtifactDeploymentsV1RequestObject) (openapi.ListArtifactDeploymentsV1ResponseObject, error) {
	return nil, nil
}
