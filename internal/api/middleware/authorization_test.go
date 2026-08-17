package middleware_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/middleware"
	"github.com/konfidence-project/konfidence/internal/api/session"
	"github.com/konfidence-project/konfidence/internal/project"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type authorizationProjectRepository struct {
	projects  []konfidence.Project
	err       error
	listCalls int
}

func (r *authorizationProjectRepository) Get(context.Context, string) (*konfidence.Project, error) {
	return nil, project.ErrNotFound
}

func (r *authorizationProjectRepository) List(context.Context) ([]konfidence.Project, error) {
	r.listCalls++
	return r.projects, r.err
}

func authorizationProject(name string, bindings map[string]konfidence.Subjects) konfidence.Project {
	return konfidence.Project{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       konfidence.ProjectSpec{RoleBindings: bindings},
	}
}

func sessionSubjects(groups ...string) konfidence.Subjects {
	return konfidence.Subjects{{Session: &konfidence.SessionSubject{MemberOf: groups}}}
}

func TestProjectAuthorizationAddsProjectRoles(t *testing.T) {
	repository := &authorizationProjectRepository{projects: []konfidence.Project{
		authorizationProject("accessible", map[string]konfidence.Subjects{
			"viewer": sessionSubjects("all-users"),
			"admin":  sessionSubjects("platform-engineers"),
		}),
		authorizationProject("hidden", map[string]konfidence.Subjects{
			"admin": sessionSubjects("platform-managers"),
		}),
	}}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, err := session.FromContext(r.Context())
		if err != nil {
			t.Error(err)
		} else if got := identity.Roles["accessible"]; len(got) != 2 || got[0] != "admin" || got[1] != "viewer" {
			t.Errorf("unexpected roles: %v", got)
		}
		if _, ok := identity.Roles["hidden"]; ok {
			t.Error("unexpected access to hidden project")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	h := middleware.ProjectAuthorization(slog.New(slog.NewTextHandler(io.Discard, nil)), repository, next)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	request = request.WithContext(session.NewContext(request.Context(), &session.Session{
		Context: session.Context{Groups: []string{"all-users", "platform-engineers"}},
	}))
	response := httptest.NewRecorder()

	h.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if repository.listCalls != 1 {
		t.Fatalf("expected one project list, got %d", repository.listCalls)
	}
}

func TestProjectAuthorizationSkipsPublicRequests(t *testing.T) {
	repository := &authorizationProjectRepository{}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	h := middleware.ProjectAuthorization(slog.New(slog.NewTextHandler(io.Discard, nil)), repository, next)
	response := httptest.NewRecorder()

	h.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/login", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if repository.listCalls != 0 {
		t.Fatalf("expected no project list, got %d", repository.listCalls)
	}
}

func TestProjectAuthorizationUsesLatestRepositorySnapshot(t *testing.T) {
	repository := &authorizationProjectRepository{projects: []konfidence.Project{
		authorizationProject("project", map[string]konfidence.Subjects{
			"viewer": sessionSubjects("team"),
		}),
	}}
	var gotRoles []string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, err := session.FromContext(r.Context())
		if err != nil {
			t.Error(err)
		}
		gotRoles = identity.Roles["project"]
		w.WriteHeader(http.StatusNoContent)
	})
	h := middleware.ProjectAuthorization(slog.New(slog.NewTextHandler(io.Discard, nil)), repository, next)
	request := func() *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		return r.WithContext(session.NewContext(r.Context(), &session.Session{
			Context: session.Context{Groups: []string{"team"}},
		}))
	}

	h.ServeHTTP(httptest.NewRecorder(), request())
	if len(gotRoles) != 1 || gotRoles[0] != "viewer" {
		t.Fatalf("expected viewer role, got %v", gotRoles)
	}

	repository.projects[0].Spec.RoleBindings = map[string]konfidence.Subjects{
		"admin": sessionSubjects("team"),
	}
	h.ServeHTTP(httptest.NewRecorder(), request())
	if len(gotRoles) != 1 || gotRoles[0] != "admin" {
		t.Fatalf("expected updated admin role, got %v", gotRoles)
	}
}
