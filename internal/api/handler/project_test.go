package handler_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/handler"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
	"github.com/konfidence-project/konfidence/internal/api/session"
)

type mockSessionStore struct {
	sessions map[string]*session.Session
}

func (m *mockSessionStore) Save(sess *session.Session) (string, error) {
	id := "test-session-id"
	m.sessions[id] = sess
	return id, nil
}

func (m *mockSessionStore) Get(id string) (*session.Session, error) {
	return m.sessions[id], nil
}

func (m *mockSessionStore) Delete(id string) error {
	delete(m.sessions, id)
	return nil
}

func createTestProject(name string, groups []string) *konfidence.Project {
	return &konfidence.Project{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: konfidence.ProjectSpec{
			DisplayName: "Test Project",
			RoleBindings: map[string]konfidence.Subjects{
				"admin": {
					{
						Session: &konfidence.SessionSubject{
							MemberOf: groups,
						},
					},
				},
			},
		},
	}
}

var _ = Describe("Project", func() {
	Describe("ListProjects", func() {
		var (
			k8sClient      client.Client
			sessionStore   session.SessionStore
			projectHandler *handler.ProjectHandler
			ctx            context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()
			scheme := runtime.NewScheme()
			Expect(konfidence.AddToScheme(scheme)).To(Succeed())
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()

			sessionStore = &mockSessionStore{
				sessions: make(map[string]*session.Session),
			}

			projectHandler = handler.NewProjectHandler(k8sClient, sessionStore)
		})

		It("returns 200 with json response", func() {
			testSession := session.Session{
				PreferredUsername: "alice",
				Groups:            []string{"platform-engineers"},
			}
			sessionID, _ := sessionStore.Save(&testSession)

			project := createTestProject("test-project-1", []string{"platform-engineers"})
			Expect(k8sClient.Create(context.Background(), project)).To(Succeed())

			ctx := context.WithValue(context.Background(), session.ContextId, sessionID)
			response, err := projectHandler.ListProjects(ctx, openapi.ListProjectsRequestObject{})

			listResponse := response.(openapi.ListProjects200JSONResponse)
			Expect(err).NotTo(HaveOccurred())
			Expect(listResponse.Data).To(HaveLen(1))
			Expect(listResponse.Data[0].Name).To(Equal("Test Project"))
		})

		It("returns empty list for non-matching user groups", func() {
			testSession := session.Session{
				PreferredUsername: "alice-2",
				Groups:            []string{"platform-managers"},
			}
			sessionID, _ := sessionStore.Save(&testSession)

			project := createTestProject("test-project-2", []string{"platform-engineers"})
			Expect(k8sClient.Create(ctx, project)).To(Succeed())

			ctx := context.WithValue(context.Background(), session.ContextId, sessionID)
			response, err := projectHandler.ListProjects(ctx, openapi.ListProjectsRequestObject{})

			listResponse := response.(openapi.ListProjects200JSONResponse)
			Expect(err).NotTo(HaveOccurred())
			Expect(listResponse.Data).To(BeEmpty())
		})

		It("returns 401 for incorrect session", func() {
			ctx := context.WithValue(context.Background(), session.ContextId, "invalid-session")

			_, err := projectHandler.ListProjects(ctx, openapi.ListProjectsRequestObject{})

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
