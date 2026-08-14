package router_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/konfidence-project/konfidence/internal/api/router"
)

var _ = Describe("Router", func() {
	var h http.Handler

	BeforeEach(func() {
		var err error
		h, err = router.New(context.Background(), slog.Default(), nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns 404 for unknown paths", func() {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/unknown", nil))
		Expect(rec.Code).To(Equal(http.StatusNotFound))
	})
})
