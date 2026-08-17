package ui_test

import (
	"net/http"
	"net/http/httptest"
	"testing/fstest"

	"github.com/konfidence-project/konfidence/internal/api/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UI handler", func() {
	var handler http.Handler

	BeforeEach(func() {
		var err error
		handler, err = ui.New(fstest.MapFS{
			"index.html":                       {Data: []byte("<h1>Konfidence</h1>")},
			"favicon.ico":                      {Data: []byte("icon")},
			"_app/immutable/chunks/example.js": {Data: []byte("export {}")},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable("serves the SPA index",
		func(target string) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(recorder.Body.String()).To(ContainSubstring("Konfidence"))
			Expect(recorder.Header().Get("Cache-Control")).To(Equal("no-cache, must-revalidate"))
			Expect(recorder.Header().Get("ETag")).NotTo(BeEmpty())
		},
		Entry("at the root", "/"),
		Entry("for a deep link", "/projects/example/landscape"),
	)

	It("serves static files", func() {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/favicon.ico", nil))
		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.String()).To(Equal("icon"))
		Expect(recorder.Header().Get("Cache-Control")).To(Equal("public, max-age=3600"))
	})

	It("caches immutable assets", func() {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/_app/immutable/chunks/example.js", nil))
		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Header().Get("Cache-Control")).To(Equal("public, max-age=31536000, immutable"))
	})

	It("returns not modified when the SPA index ETag matches", func() {
		initial := httptest.NewRecorder()
		handler.ServeHTTP(initial, httptest.NewRequest(http.MethodGet, "/projects/example", nil))

		request := httptest.NewRequest(http.MethodGet, "/projects/example", nil)
		request.Header.Set("If-None-Match", initial.Header().Get("ETag"))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusNotModified))
		Expect(recorder.Body.Len()).To(BeZero())
	})

	DescribeTable("does not use the SPA fallback",
		func(method, target string) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(method, target, nil))
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		},
		Entry("for the API root", http.MethodGet, "/api"),
		Entry("for versioned API paths", http.MethodGet, "/api/v1/unknown"),
		Entry("for unversioned API paths", http.MethodGet, "/api/unknown"),
		Entry("for missing files", http.MethodGet, "/missing.js"),
		Entry("for mutating requests", http.MethodPost, "/projects"),
	)
})
