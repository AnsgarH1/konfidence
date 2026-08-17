package handler_test

import (
	"context"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/handler"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
	landscapedomain "github.com/konfidence-project/konfidence/internal/landscape"
	projectdomain "github.com/konfidence-project/konfidence/internal/project"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newProjectTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(s)
	_ = konfidence.AddToScheme(s)
	return s
}

func fakeProjectK8s(objs ...client.Object) client.Client {
	return fake.NewClientBuilder().WithScheme(newProjectTestScheme()).WithObjects(objs...).WithStatusSubresource(&konfidence.Project{}).Build()
}

func projectFixture(name, namespace string) *konfidence.Project {
	return &konfidence.Project{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     konfidence.ProjectStatus{Namespace: namespace},
	}
}

func landscapeFixture(name, namespace, displayName string) *konfidence.Landscape {
	return &konfidence.Landscape{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       konfidence.LandscapeSpec{DisplayName: displayName},
	}
}

func apiHandlerWith(objs ...client.Object) *handler.APIHandler {
	k8s := fakeProjectK8s(objs...)
	return handler.NewAPIHandlerWithRepos(
		projectdomain.NewRepository(k8s),
		landscapedomain.NewRepository(k8s),
	)
}

var _ = Describe("ProjectHandler", func() {
	Describe("ListLandscapesV1", func() {
		It("returns all landscapes for a project", func() {
			project := projectFixture("my-project", "kden-p-my-project")
			l1 := landscapeFixture("dev", "kden-p-my-project", "Dev")
			l2 := landscapeFixture("staging", "kden-p-my-project", "Staging")
			h := apiHandlerWith(project, l1, l2)

			resp, err := h.ListLandscapesV1(context.Background(), openapi.ListLandscapesV1RequestObject{ProjectId: "my-project"})
			Expect(err).NotTo(HaveOccurred())

			ok, is200 := resp.(openapi.ListLandscapesV1200JSONResponse)
			Expect(is200).To(BeTrue())
			Expect(ok.Data).To(HaveLen(2))
			Expect([]string{ok.Data[0].Id, ok.Data[1].Id}).To(ConsistOf("dev", "staging"))
		})

		It("returns an empty list when project has no landscapes", func() {
			project := projectFixture("empty-project", "kden-p-empty-project")
			h := apiHandlerWith(project)

			resp, err := h.ListLandscapesV1(context.Background(), openapi.ListLandscapesV1RequestObject{ProjectId: "empty-project"})
			Expect(err).NotTo(HaveOccurred())

			ok := resp.(openapi.ListLandscapesV1200JSONResponse)
			Expect(ok.Data).To(BeEmpty())
		})

		It("maps landscape fields correctly", func() {
			project := projectFixture("my-project", "kden-p-my-project")
			l := landscapeFixture("dev", "kden-p-my-project", "Development")
			h := apiHandlerWith(project, l)

			resp, err := h.ListLandscapesV1(context.Background(), openapi.ListLandscapesV1RequestObject{ProjectId: "my-project"})
			Expect(err).NotTo(HaveOccurred())

			item := resp.(openapi.ListLandscapesV1200JSONResponse).Data[0]
			Expect(item.Id).To(Equal("dev"))
			Expect(item.Name).To(Equal("Development"))
		})

		It("returns 404 when project does not exist", func() {
			h := apiHandlerWith()

			resp, err := h.ListLandscapesV1(context.Background(), openapi.ListLandscapesV1RequestObject{ProjectId: "nonexistent"})
			Expect(err).NotTo(HaveOccurred())

			r, is404 := resp.(openapi.ListLandscapesV1404JSONResponse)
			Expect(is404).To(BeTrue())
			Expect(r.Error.Code).To(Equal("not_found"))
			Expect(r.Error.Message).To(ContainSubstring("nonexistent"))
		})

		It("only returns landscapes belonging to the requested project", func() {
			project := projectFixture("project-a", "kden-p-project-a")
			lA := landscapeFixture("dev", "kden-p-project-a", "Dev A")
			lB := landscapeFixture("dev", "kden-p-project-b", "Dev B")
			h := apiHandlerWith(project, lA, lB)

			resp, err := h.ListLandscapesV1(context.Background(), openapi.ListLandscapesV1RequestObject{ProjectId: "project-a"})
			Expect(err).NotTo(HaveOccurred())

			ok := resp.(openapi.ListLandscapesV1200JSONResponse)
			Expect(ok.Data).To(HaveLen(1))
			Expect(ok.Data[0].Id).To(Equal("dev"))
		})
	})
})
