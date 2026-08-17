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
	List(ctx context.Context) ([]konfidence.Project, error)
}

type k8sRepository struct{ reader client.Reader }

// NewRepository creates a Repository backed by the given reader. When reader is
// an informer cache, all Get and List calls are served from its local store.
func NewRepository(reader client.Reader) Repository {
	return &k8sRepository{reader: reader}
}

func (r *k8sRepository) Get(ctx context.Context, name string) (*konfidence.Project, error) {
	var project konfidence.Project
	if err := r.reader.Get(ctx, types.NamespacedName{Name: name}, &project); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting project %q: %w", name, err)
	}
	return &project, nil
}

func (r *k8sRepository) List(ctx context.Context) ([]konfidence.Project, error) {
	var projects konfidence.ProjectList
	if err := r.reader.List(ctx, &projects); err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	return projects.Items, nil
}

// Get fetches a Project CR by name using the given client.
func Get(ctx context.Context, k8s client.Reader, name string) (*konfidence.Project, error) {
	return NewRepository(k8s).Get(ctx, name)
}
