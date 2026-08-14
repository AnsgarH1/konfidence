package landscape

import (
	"context"
	"fmt"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Repository provides read access to Landscape CRs.
type Repository interface {
	ListForProject(ctx context.Context, projectNamespace string) ([]konfidence.Landscape, error)
}

type k8sRepository struct{ k8s client.Client }

// NewRepository creates a Repository backed by the given k8s client.
func NewRepository(k8s client.Client) Repository {
	return &k8sRepository{k8s: k8s}
}

func (r *k8sRepository) ListForProject(ctx context.Context, projectNamespace string) ([]konfidence.Landscape, error) {
	var list konfidence.LandscapeList
	if err := r.k8s.List(ctx, &list, client.InNamespace(projectNamespace)); err != nil {
		return nil, fmt.Errorf("listing landscapes in namespace %q: %w", projectNamespace, err)
	}
	return list.Items, nil
}

// ListForProject returns all Landscape CRs in the given project namespace using the given client.
func ListForProject(ctx context.Context, k8s client.Client, projectNamespace string) ([]konfidence.Landscape, error) {
	return NewRepository(k8s).ListForProject(ctx, projectNamespace)
}
