package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const informerSyncTimeout = 30 * time.Second

type informerBackedCache struct {
	cache  ctrlcache.Cache
	cancel context.CancelFunc
	done   <-chan error
}

func newInformerBackedCache(ctx context.Context, config *rest.Config, scheme *runtime.Scheme,
	resources ...client.Object) (*informerBackedCache, error) {
	k8sCache, err := ctrlcache.New(config, ctrlcache.Options{
		Scheme:                      scheme,
		ReaderFailOnMissingInformer: true,
	})
	if err != nil {
		return nil, fmt.Errorf("creating informer cache: %w", err)
	}

	cacheCtx, cancel := context.WithCancel(ctx)
	for _, resource := range resources {
		if _, err := k8sCache.GetInformer(cacheCtx, resource); err != nil {
			cancel()
			return nil, fmt.Errorf("creating informer for %T: %w", resource, err)
		}
	}

	done := make(chan error, 1)
	go func() {
		err := k8sCache.Start(cacheCtx)
		done <- err
		if err != nil {
			cancel()
		}
	}()

	syncCtx, cancelSync := context.WithTimeout(cacheCtx, informerSyncTimeout)
	defer cancelSync()
	if !k8sCache.WaitForCacheSync(syncCtx) {
		cancel()
		startErr := <-done
		if startErr != nil && !errors.Is(startErr, context.Canceled) {
			return nil, fmt.Errorf("starting informer cache: %w", startErr)
		}
		if errors.Is(syncCtx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("informer cache failed to sync within %s", informerSyncTimeout)
		}
		return nil, fmt.Errorf("syncing informer cache: %w", syncCtx.Err())
	}

	return &informerBackedCache{cache: k8sCache, cancel: cancel, done: done}, nil
}

// CachedReader serves Get and List calls from the informers' local stores.
func (c *informerBackedCache) CachedReader() client.Reader {
	return c.cache
}

func (c *informerBackedCache) Close() error {
	c.cancel()
	err := <-c.done
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
