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

// Repository provides read access to Project CRs.
type Repository interface {
	Get(ctx context.Context, name string) (*konfidence.Project, error)
}

type k8sRepository struct{ k8s client.Client }

// NewRepository creates a Repository backed by the given k8s client.
func NewRepository(k8s client.Client) Repository {
	return &k8sRepository{k8s: k8s}
}

func (r *k8sRepository) Get(ctx context.Context, name string) (*konfidence.Project, error) {
	var project konfidence.Project
	if err := r.k8s.Get(ctx, types.NamespacedName{Name: name}, &project); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting project %q: %w", name, err)
	}
	return &project, nil
}

// Get fetches a Project CR by name using the given client. Returns ErrNotFound if it does not exist.
func Get(ctx context.Context, k8s client.Client, name string) (*konfidence.Project, error) {
	return NewRepository(k8s).Get(ctx, name)
}
