package handler

import (
	"context"

	"github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
	"github.com/konfidence-project/konfidence/internal/api/session"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type projectHandler struct{ k8s client.Client }

func newProjectHandler(k8s client.Client) *projectHandler {
	return &projectHandler{k8s: k8s}
}

func (h *projectHandler) ListProjectsV1(_ context.Context, _ openapi.ListProjectsV1RequestObject) (openapi.ListProjectsV1ResponseObject, error) {
	sessionId, err := session.GetSessionIdFromContext(ctx)
	if err != nil {
		return openapi.ListProjects401JSONResponse{}, nil
	}

	storedSession, err := h.sessionStore.Get(sessionId)
	if err != nil || storedSession == nil {
		return openapi.ListProjects401JSONResponse{}, nil
	}

	// Fetch all Project CRs from the cluster
	var projects v1alpha1.ProjectList
	if err := h.k8s.List(ctx, &projects); err != nil {
		return nil, err
	}

	var result []openapi.Project
	for _, p := range projects.Items {
		if hasAccess(storedSession.Groups, p.Spec.RoleBindings) {
			result = append(result, openapi.Project{
				Id:   p.Name,
				Name: p.Spec.DisplayName,
			})
		}
	}

	return openapi.ListProjectsV1200JSONResponse{Data: result}, nil
}

func hasAccess(userGroups []string, roleBindings map[string]v1alpha1.Subjects) bool {
	groupSet := make(map[string]struct{}, len(userGroups))
	for _, g := range userGroups {
		groupSet[g] = struct{}{}
	}

	for _, subjects := range roleBindings {
		for _, s := range subjects {
			if s.Session != nil {
				for _, bindGroup := range s.Session.MemberOf {
					if _, ok := groupSet[bindGroup]; ok {
						return true
					}
				}
			}
		}
	}

	return false
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
