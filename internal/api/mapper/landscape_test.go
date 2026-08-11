package mapper_test

import (
	"testing"

	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/mapper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestToLandscapeResponse_MapsFieldsCorrectly(t *testing.T) {
	l := konfidence.Landscape{
		ObjectMeta: metav1.ObjectMeta{Name: "dev"},
		Spec:       konfidence.LandscapeSpec{DisplayName: "Development"},
	}

	result := mapper.ToLandscapeResponse(l)

	if result.Id != "dev" {
		t.Errorf("expected Id %q, got %q", "dev", result.Id)
	}
	if result.Name != "Development" {
		t.Errorf("expected Name %q, got %q", "Development", result.Name)
	}
}

func TestToLandscapeResponse_EmptyDisplayName(t *testing.T) {
	l := konfidence.Landscape{
		ObjectMeta: metav1.ObjectMeta{Name: "prod"},
		Spec:       konfidence.LandscapeSpec{DisplayName: ""},
	}

	result := mapper.ToLandscapeResponse(l)

	if result.Id != "prod" {
		t.Errorf("expected Id %q, got %q", "prod", result.Id)
	}
	if result.Name != "" {
		t.Errorf("expected empty Name, got %q", result.Name)
	}
}
