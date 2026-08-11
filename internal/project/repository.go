package project

import (
	"context"
	"fmt"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ErrNotFound is returned when a Project CR does not exist.
var ErrNotFound = fmt.Errorf("project not found")

// Get fetches a Project CR by name. Returns ErrNotFound if it does not exist.
func Get(ctx context.Context, k8s client.Client, name string) (*konfidence.Project, error) {
	var project konfidence.Project
	if err := k8s.Get(ctx, types.NamespacedName{Name: name}, &project); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting project %q: %w", name, err)
	}
	return &project, nil
}
