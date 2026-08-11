package handler // nolint

import (
	"context"
	"errors"
	"fmt"

	"github.com/konfidence-project/konfidence/internal/api/mapper"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
	landscapedomain "github.com/konfidence-project/konfidence/internal/landscape"
	projectdomain "github.com/konfidence-project/konfidence/internal/project"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectHandler struct{ k8s func() (client.Client, error) }

func (h *ProjectHandler) ListProjects(_ context.Context, _ openapi.ListProjectsRequestObject) (openapi.ListProjectsResponseObject, error) {
	return nil, nil
}

func (h *ProjectHandler) ListLandscapes(ctx context.Context, req openapi.ListLandscapesRequestObject) (openapi.ListLandscapesResponseObject, error) {
	k8s, err := h.k8s()
	if err != nil {
		return nil, NewInternal(err)
	}

	project, err := projectdomain.Get(ctx, k8s, req.ProjectId)
	if err != nil {
		if errors.Is(err, projectdomain.ErrNotFound) {
			return openapi.ListLandscapes404JSONResponse{}, nil
		}
		return nil, NewInternal(fmt.Errorf("getting project %q: %w", req.ProjectId, err))
	}

	if project.Status.Namespace == "" {
		return nil, NewInternal(fmt.Errorf("project %q has no namespace yet", req.ProjectId))
	}

	landscapes, err := landscapedomain.ListForProject(ctx, k8s, project.Status.Namespace)
	if err != nil {
		return nil, NewInternal(err)
	}

	data := make([]openapi.Landscape, len(landscapes))
	for i, l := range landscapes {
		data[i] = mapper.ToLandscapeResponse(l)
	}

	return openapi.ListLandscapes200JSONResponse{Data: data}, nil
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
