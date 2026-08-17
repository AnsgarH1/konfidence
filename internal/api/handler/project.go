package handler

import (
	"context"

	"github.com/konfidence-project/konfidence/internal/api/openapi"
	"github.com/konfidence-project/konfidence/internal/api/session"
	projectdomain "github.com/konfidence-project/konfidence/internal/project"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type projectHandler struct {
	k8s      client.Client
	projects projectdomain.Repository
}

func newProjectHandler(kubernetes KubernetesAccess) *projectHandler {
	return &projectHandler{
		k8s:      kubernetes.Client,
		projects: projectdomain.NewRepository(kubernetes.CachedReader),
	}
}

func (h *projectHandler) ListProjectsV1(ctx context.Context, _ openapi.ListProjectsV1RequestObject) (openapi.ListProjectsV1ResponseObject, error) {
	identity, err := session.FromContext(ctx)
	if err != nil {
		return openapi.ListProjectsV1401JSONResponse{}, nil
	}

	projects, err := h.projects.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]openapi.Project, 0, len(projects))
	for _, p := range projects {
		if len(identity.Roles[p.Name]) > 0 {
			result = append(result, openapi.Project{
				Id:   p.Name,
				Name: p.Spec.DisplayName,
			})
		}
	}

	return openapi.ListProjectsV1200JSONResponse{Data: result}, nil
}

func (h *projectHandler) ListLandscapesV1(_ context.Context, _ openapi.ListLandscapesV1RequestObject) (openapi.ListLandscapesV1ResponseObject, error) {
	return nil, nil
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
