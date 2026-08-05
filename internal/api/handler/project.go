package handler // nolint

import (
	"context"
	"fmt"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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

	var project konfidence.Project
	if err := k8s.Get(ctx, types.NamespacedName{Name: req.ProjectId}, &project); err != nil {
		if apierrors.IsNotFound(err) {
			return openapi.ListLandscapes403JSONResponse{}, nil
		}
		return nil, NewInternal(fmt.Errorf("getting project %q: %w", req.ProjectId, err))
	}

	projectNamespace := project.Status.Namespace
	if projectNamespace == "" {
		return nil, NewInternal(fmt.Errorf("project %q has no namespace yet", req.ProjectId))
	}

	var list konfidence.LandscapeList
	if err := k8s.List(ctx, &list, client.InNamespace(projectNamespace)); err != nil {
		return nil, NewInternal(fmt.Errorf("listing landscapes for project %q: %w", req.ProjectId, err))
	}

	data := make([]openapi.Landscape, len(list.Items))
	for i, l := range list.Items {
		data[i] = openapi.Landscape{
			Id:   l.Name,
			Name: l.Spec.DisplayName,
		}
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
