package handler_test

import (
	"context"
	"fmt"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/handler"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
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

func fakeProjectK8s(objs ...client.Object) func() (client.Client, error) {
	c := fake.NewClientBuilder().WithScheme(newProjectTestScheme()).WithObjects(objs...).WithStatusSubresource(&konfidence.Project{}).Build()
	return func() (client.Client, error) { return c, nil }
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

var _ = Describe("ProjectHandler", func() {
	Describe("ListLandscapes", func() {
		It("returns all landscapes for a project", func() {
			project := projectFixture("my-project", "kden-p-my-project")
			l1 := landscapeFixture("dev", "kden-p-my-project", "Dev")
			l2 := landscapeFixture("staging", "kden-p-my-project", "Staging")
			h := handler.NewServerHandler(fakeProjectK8s(project, l1, l2))

			resp, err := h.ListLandscapes(context.Background(), openapi.ListLandscapesRequestObject{ProjectId: "my-project"})
			Expect(err).NotTo(HaveOccurred())

			ok, is200 := resp.(openapi.ListLandscapes200JSONResponse)
			Expect(is200).To(BeTrue())
			Expect(ok.Data).To(HaveLen(2))

			ids := []string{ok.Data[0].Id, ok.Data[1].Id}
			Expect(ids).To(ConsistOf("dev", "staging"))
		})

		It("returns an empty list when project has no landscapes", func() {
			project := projectFixture("empty-project", "kden-p-empty-project")
			h := handler.NewServerHandler(fakeProjectK8s(project))

			resp, err := h.ListLandscapes(context.Background(), openapi.ListLandscapesRequestObject{ProjectId: "empty-project"})
			Expect(err).NotTo(HaveOccurred())

			ok := resp.(openapi.ListLandscapes200JSONResponse)
			Expect(ok.Data).To(BeEmpty())
		})

		It("maps landscape fields correctly", func() {
			project := projectFixture("my-project", "kden-p-my-project")
			l := landscapeFixture("dev", "kden-p-my-project", "Development")
			h := handler.NewServerHandler(fakeProjectK8s(project, l))

			resp, err := h.ListLandscapes(context.Background(), openapi.ListLandscapesRequestObject{ProjectId: "my-project"})
			Expect(err).NotTo(HaveOccurred())

			item := resp.(openapi.ListLandscapes200JSONResponse).Data[0]
			Expect(item.Id).To(Equal("dev"))
			Expect(item.Name).To(Equal("Development"))
		})

		It("returns 404 when project does not exist", func() {
			h := handler.NewServerHandler(fakeProjectK8s())

			resp, err := h.ListLandscapes(context.Background(), openapi.ListLandscapesRequestObject{ProjectId: "nonexistent"})
			Expect(err).NotTo(HaveOccurred())

			_, is404 := resp.(openapi.ListLandscapes404JSONResponse)
			Expect(is404).To(BeTrue())
		})

		It("only returns landscapes belonging to the requested project", func() {
			project := projectFixture("project-a", "kden-p-project-a")
			lA := landscapeFixture("dev", "kden-p-project-a", "Dev A")
			lB := landscapeFixture("dev", "kden-p-project-b", "Dev B")
			h := handler.NewServerHandler(fakeProjectK8s(project, lA, lB))

			resp, err := h.ListLandscapes(context.Background(), openapi.ListLandscapesRequestObject{ProjectId: "project-a"})
			Expect(err).NotTo(HaveOccurred())

			ok := resp.(openapi.ListLandscapes200JSONResponse)
			Expect(ok.Data).To(HaveLen(1))
			Expect(ok.Data[0].Id).To(Equal("dev"))
		})

		It("returns internal error when k8s client fails", func() {
			h := handler.NewServerHandler(func() (client.Client, error) {
				return nil, fmt.Errorf("no kubeconfig")
			})

			_, err := h.ListLandscapes(context.Background(), openapi.ListLandscapesRequestObject{ProjectId: "my-project"})
			Expect(err).To(HaveOccurred())
		})
	})
})
