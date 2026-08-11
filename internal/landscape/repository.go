package landscape

import (
	"context"
	"fmt"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListForProject returns all Landscape CRs in the given project namespace.
func ListForProject(ctx context.Context, k8s client.Client, projectNamespace string) ([]konfidence.Landscape, error) {
	var list konfidence.LandscapeList
	if err := k8s.List(ctx, &list, client.InNamespace(projectNamespace)); err != nil {
		return nil, fmt.Errorf("listing landscapes in namespace %q: %w", projectNamespace, err)
	}
	return list.Items, nil
}
