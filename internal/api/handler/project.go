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

type ProjectHandler struct {
	projects   projectdomain.Repository
	landscapes landscapedomain.Repository
}

func newProjectHandler(k8s client.Client) ProjectHandler {
	return ProjectHandler{
		projects:   projectdomain.NewRepository(k8s),
		landscapes: landscapedomain.NewRepository(k8s),
	}
}

func (h *ProjectHandler) ListProjects(_ context.Context, _ openapi.ListProjectsRequestObject) (openapi.ListProjectsResponseObject, error) {
	return nil, nil
}

func (h *ProjectHandler) ListLandscapes(ctx context.Context, req openapi.ListLandscapesRequestObject) (openapi.ListLandscapesResponseObject, error) {
	project, err := h.projects.Get(ctx, req.ProjectId)
	if err != nil {
		if errors.Is(err, projectdomain.ErrNotFound) {
			return openapi.ListLandscapes404JSONResponse{
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
		return openapi.ListLandscapes500JSONResponse{}, nil
	}

	if project.Status.Namespace == "" {
		return openapi.ListLandscapes500JSONResponse{}, nil
	}

	landscapes, err := h.landscapes.ListForProject(ctx, project.Status.Namespace)
	if err != nil {
		return openapi.ListLandscapes500JSONResponse{}, nil
	}

	data := make([]openapi.Landscape, len(landscapes))
	for i, l := range landscapes {
		data[i] = toLandscapeResponse(l)
	}

	return openapi.ListLandscapes200JSONResponse{Data: data}, nil
}

func toLandscapeResponse(l konfidence.Landscape) openapi.Landscape {
	return openapi.Landscape{
		Id:   l.Name,
		Name: l.Spec.DisplayName,
	}
}

func (h *ProjectHandler) ListStages(_ context.Context, _ openapi.ListStagesRequestObject) (openapi.ListStagesResponseObject, error) {
	return nil, nil
}

func (h *ProjectHandler) ListVectorDeployments(_ context.Context,
	_ openapi.ListVectorDeploymentsRequestObject) (openapi.ListVectorDeploymentsResponseObject, error) {
	return nil, nil
}

func (h *ProjectHandler) ListArtifactDeployments(_ context.Context,
	_ openapi.ListArtifactDeploymentsRequestObject) (openapi.ListArtifactDeploymentsResponseObject, error) {
	return nil, nil
}
