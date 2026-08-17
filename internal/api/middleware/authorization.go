package middleware

import (
	"log/slog"
	"net/http"
	"sort"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/apierror"
	"github.com/konfidence-project/konfidence/internal/api/session"
	projectdomain "github.com/konfidence-project/konfidence/internal/project"
)

// ProjectAuthorization adds the caller's roles for every accessible project to
// the request context. The repository is expected to use an informer-backed reader.
func ProjectAuthorization(logger *slog.Logger, projects projectdomain.Repository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		storedSession, err := session.FromContext(r.Context())
		if err != nil {
			// Public operations do not have an authenticated session.
			next.ServeHTTP(w, r)
			return
		}

		projectList, err := projects.List(r.Context())
		if err != nil {
			logger.Error("failed to determine project roles", "error", err)
			apierror.WriteInternal(w)
			return
		}

		roles := projectRoles(storedSession.Groups, projectList)
		storedSession.Roles = roles
		next.ServeHTTP(w, r.WithContext(session.NewContext(r.Context(), &session.Session{Context: *storedSession})))
	})
}

func projectRoles(groups []string, projects []konfidence.Project) session.ProjectRoles {
	groupSet := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		groupSet[group] = struct{}{}
	}

	access := make(session.ProjectRoles)
	for _, project := range projects {
		roles := make([]string, 0, len(project.Spec.RoleBindings))
		for role, subjects := range project.Spec.RoleBindings {
			if matchesSessionGroup(groupSet, subjects) {
				roles = append(roles, role)
			}
		}
		if len(roles) > 0 {
			sort.Strings(roles)
			access[project.Name] = roles
		}
	}
	return access
}

func matchesSessionGroup(groups map[string]struct{}, subjects konfidence.Subjects) bool {
	for _, subject := range subjects {
		if subject.Session == nil {
			continue
		}
		for _, group := range subject.Session.MemberOf {
			if _, ok := groups[group]; ok {
				return true
			}
		}
	}
	return false
}
