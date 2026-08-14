package handler // nolint

import (
	"context"

	"github.com/konfidence-project/konfidence/internal/api/openapi"
)

type InfoHandler struct{}

func (h *InfoHandler) GetHealthStatus(_ context.Context, _ openapi.GetHealthStatusRequestObject) (openapi.GetHealthStatusResponseObject, error) {
	return openapi.GetHealthStatus200JSONResponse{Status: "ok"}, nil
}

func (h *InfoHandler) GetReadinessStatus(_ context.Context, _ openapi.GetReadinessStatusRequestObject) (openapi.GetReadinessStatusResponseObject, error) {
	return openapi.GetReadinessStatus200JSONResponse{Status: "ok"}, nil
}
