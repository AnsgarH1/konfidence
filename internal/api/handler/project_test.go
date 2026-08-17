package handler

import (
	"context"
	"errors"
	"testing"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
	"github.com/konfidence-project/konfidence/internal/api/session"
	"github.com/konfidence-project/konfidence/internal/project"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type projectRepository struct {
	projects []konfidence.Project
	err      error
}

func (r *projectRepository) Get(context.Context, string) (*konfidence.Project, error) {
	return nil, project.ErrNotFound
}

func (r *projectRepository) List(context.Context) ([]konfidence.Project, error) {
	return r.projects, r.err
}

func projectFixture(name, displayName string, groups ...string) konfidence.Project {
	return konfidence.Project{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: konfidence.ProjectSpec{
			DisplayName: displayName,
			RoleBindings: map[string]konfidence.Subjects{
				"admin": {{Session: &konfidence.SessionSubject{MemberOf: groups}}},
			},
		},
	}
}

func authorizedContext(projectRoles session.ProjectRoles) context.Context {
	return session.NewContext(context.Background(), &session.Session{Context: session.Context{Roles: projectRoles}})
}

func TestListProjectsV1(t *testing.T) {
	t.Run("returns only matching projects", func(t *testing.T) {
		h := &projectHandler{projects: &projectRepository{projects: []konfidence.Project{
			projectFixture("visible", "Visible Project", "platform-engineers"),
			projectFixture("hidden", "Hidden Project", "platform-managers"),
		}}}

		response, err := h.ListProjectsV1(authorizedContext(session.ProjectRoles{
			"visible": {"admin"},
		}), openapi.ListProjectsV1RequestObject{})
		if err != nil {
			t.Fatal(err)
		}
		projects := response.(openapi.ListProjectsV1200JSONResponse).Data
		if len(projects) != 1 || projects[0].Id != "visible" || projects[0].Name != "Visible Project" {
			t.Fatalf("unexpected projects: %#v", projects)
		}
	})

	t.Run("returns an empty list when no project matches", func(t *testing.T) {
		h := &projectHandler{projects: &projectRepository{projects: []konfidence.Project{
			projectFixture("hidden", "Hidden Project", "platform-engineers"),
		}}}

		response, err := h.ListProjectsV1(authorizedContext(session.ProjectRoles{}), openapi.ListProjectsV1RequestObject{})
		if err != nil {
			t.Fatal(err)
		}
		if projects := response.(openapi.ListProjectsV1200JSONResponse).Data; len(projects) != 0 {
			t.Fatalf("expected no projects, got %#v", projects)
		}
	})

	t.Run("returns unauthorized without an identity", func(t *testing.T) {
		h := &projectHandler{projects: &projectRepository{}}

		response, err := h.ListProjectsV1(context.Background(), openapi.ListProjectsV1RequestObject{})
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := response.(openapi.ListProjectsV1401JSONResponse); !ok {
			t.Fatalf("expected unauthorized response, got %T", response)
		}
	})

	t.Run("returns repository errors", func(t *testing.T) {
		repositoryErr := errors.New("Kubernetes unavailable")
		h := &projectHandler{projects: &projectRepository{err: repositoryErr}}

		response, err := h.ListProjectsV1(authorizedContext(session.ProjectRoles{
			"visible": {"admin"},
		}), openapi.ListProjectsV1RequestObject{})
		if response != nil || !errors.Is(err, repositoryErr) {
			t.Fatalf("expected repository error, got response=%T err=%v", response, err)
		}
	})
}
