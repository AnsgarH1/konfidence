package controller_test

import (
	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// The ArtifactDeployment CRD marks Status.DeploymentResults as a list-map keyed by (name, type). These specs prove
// the apiserver enforces that key, independent of any controller logic.
var _ = Describe("ArtifactDeployment deployment-result uniqueness", func() {
	newAD := func(name string) *konfidence.ArtifactDeployment {
		return &konfidence.ArtifactDeployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
			Spec: konfidence.ArtifactDeploymentSpec{
				Manifest:      konfidence.ArtifactManifest{Type: "cloud.konfidence.flux.helm"},
				TaskManifests: []konfidence.TaskManifest{},
				Component:     konfidence.OCMComponent{Name: "github.com/acme/svc", Version: "1.0.0"},
			},
		}
	}
	result := func(name, typ, k8sName string) konfidence.DeploymentResult {
		return konfidence.DeploymentResult{
			Name: name,
			Type: typ,
			Spec: runtime.RawExtension{Raw: []byte(`{"K8sName":"` + k8sName + `"}`)},
		}
	}

	It("rejects two results with the same (name, type)", func() {
		ad := newAD("ad-dup-nametype")
		Expect(k8sClient.Create(ctx, ad)).To(Succeed())
		ad.Status.DeploymentResults = []konfidence.DeploymentResult{
			result("candidates", "http-k8s-service", "candidates-a"),
			result("candidates", "http-k8s-service", "candidates-b"),
		}
		Expect(k8sClient.Status().Update(ctx, ad)).ToNot(Succeed())
	})

	It("allows the same name under different types", func() {
		ad := newAD("ad-same-name-diff-type")
		Expect(k8sClient.Create(ctx, ad)).To(Succeed())
		ad.Status.DeploymentResults = []konfidence.DeploymentResult{
			result("candidates", "http-k8s-service", "candidates-a"),
			result("candidates", "grpc-k8s-service", "candidates-b"),
		}
		Expect(k8sClient.Status().Update(ctx, ad)).To(Succeed())
	})
})
