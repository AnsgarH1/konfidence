package mapper

import (
	konfidence "github.com/konfidence-project/konfidence/api/v1alpha1"
	"github.com/konfidence-project/konfidence/internal/api/openapi"
)

func ToLandscapeResponse(l konfidence.Landscape) openapi.Landscape {
	return openapi.Landscape{
		Id:   l.Name,
		Name: l.Spec.DisplayName,
	}
}
