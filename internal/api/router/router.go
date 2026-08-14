package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/konfidence-project/konfidence/internal/api/middleware"
)

// MountFunc registers a domain's routes onto the root router.
// Wiring happens in cmd/api/cmd/root.go, following the same explicit pattern
// used for controller registration in cmd/konfidence.
type MountFunc func(r chi.Router, logger *slog.Logger, k8s client.Client)

// New returns the root chi.Router with all routes and middleware registered.
// It builds a cache-backed k8s client, starts the cache, and waits for it
// to sync before returning — requests are only served once the cache is warm.
func New(ctx context.Context, logger *slog.Logger, scheme *runtime.Scheme, mounts ...MountFunc) (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logging(logger))

	if len(mounts) > 0 {
		k8s, err := buildCachedClient(ctx, scheme)
		if err != nil {
			return nil, err
		}
		r.Group(func(domain chi.Router) {
			for _, mount := range mounts {
				mount(domain, logger, k8s)
			}
		})
	}

	return r, nil
}

// buildCachedClient creates a cache-backed k8s client. The cache is started
// in the background and the call blocks until the initial sync completes.
func buildCachedClient(ctx context.Context, scheme *runtime.Scheme) (client.Client, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s config (set KUBECONFIG for local dev): %w", err)
	}

	c, err := cache.New(cfg, cache.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s cache: %w", err)
	}

	go func() {
		if err := c.Start(ctx); err != nil {
			ctrl.Log.WithName("cache").Error(err, "k8s cache stopped unexpectedly")
		}
	}()

	if !c.WaitForCacheSync(ctx) {
		return nil, fmt.Errorf("timed out waiting for k8s cache to sync")
	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Cache:  &client.CacheOptions{Reader: c},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build k8s client: %w", err)
	}

	return k8sClient, nil
}
