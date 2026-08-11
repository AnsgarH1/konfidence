package project_test

import (
	"context"
	"errors"
	"testing"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/project"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(s)
	_ = konfidence.AddToScheme(s)
	return s
}

func fakeClient(objs ...client.Object) client.Client {
	return fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(objs...).WithStatusSubresource(&konfidence.Project{}).Build()
}

func projectFixture(name, namespace string) *konfidence.Project {
	return &konfidence.Project{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     konfidence.ProjectStatus{Namespace: namespace},
	}
}

func TestGet_ReturnsProject(t *testing.T) {
	p := projectFixture("my-project", "kden-p-my-project")
	k8s := fakeClient(p)

	result, err := project.Get(context.Background(), k8s, "my-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "my-project" {
		t.Errorf("expected Name %q, got %q", "my-project", result.Name)
	}
}

func TestGet_ReturnsErrNotFound(t *testing.T) {
	k8s := fakeClient()

	_, err := project.Get(context.Background(), k8s, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, project.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
