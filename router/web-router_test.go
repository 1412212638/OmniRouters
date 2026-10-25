package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

func TestBuildBrandedIndexPageUsesConfiguredLogoForShareMetadata(t *testing.T) {
	c := newIndexPageTestContext("https://omnirouters.com/share")
	restore := setIndexPageBranding("OmniRouters", "https://cos.frostai.cn/omnirouters/logo/logo.png")
	defer restore()

	document := `<!doctype html><html><head>
<link rel="icon" type="image/png" href="/logo.png" />
<title>New API</title>
<meta name="title" content="New API" />
<meta name="description" content="Unified AI API gateway and admin dashboard." />
</head><body></body></html>`

	result := string(buildBrandedIndexPage(c, []byte(document)))

	assertContains(t, result, "<title>OmniRouters</title>")
	assertContains(t, result, `<link rel="icon" type="image/png" href="https://cos.frostai.cn/omnirouters/logo/logo.png" />`)
	assertContains(t, result, `<meta name="title" content="OmniRouters" />`)
	assertContains(t, result, `<meta property="og:title" content="OmniRouters" />`)
	assertContains(t, result, `<meta property="og:image" content="https://cos.frostai.cn/omnirouters/logo/logo.png" />`)
	assertContains(t, result, `<meta name="twitter:image" content="https://cos.frostai.cn/omnirouters/logo/logo.png" />`)
}

func TestBuildBrandedIndexPageMakesRelativeLogoAbsoluteForShareMetadata(t *testing.T) {
	c := newIndexPageTestContext("http://127.0.0.1/dashboard")
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	c.Request.Header.Set("X-Forwarded-Host", "omnirouters.com")
	restore := setIndexPageBranding("OmniRouters", "/brand.png")
	defer restore()

	document := `<!doctype html><html><head><title>New API</title></head><body></body></html>`

	result := string(buildBrandedIndexPage(c, []byte(document)))

	assertContains(t, result, `<link rel="icon" type="image/png" href="/brand.png" />`)
	assertContains(t, result, `<meta property="og:image" content="https://omnirouters.com/brand.png" />`)
	assertContains(t, result, `<meta property="og:url" content="https://omnirouters.com/dashboard" />`)
}

func TestBuildBrandedIndexPagePrefersEnglishDescription(t *testing.T) {
	c := newIndexPageTestContext("https://omnirouters.com/share")
	restore := setIndexPageBranding("OmniRouters", "/logo.png")
	defer restore()

	document := `<!doctype html><html><head>
<title>OmniRouters</title>
<meta name="description" lang="zh" content="中文描述" />
<meta name="description" lang="en" content="A unified AI model hub for aggregation &amp; distribution." />
</head><body></body></html>`

	result := string(buildBrandedIndexPage(c, []byte(document)))

	assertContains(t, result, `<meta property="og:description" content="A unified AI model hub for aggregation &amp; distribution." />`)
	assertContains(t, result, `<meta name="twitter:description" content="A unified AI model hub for aggregation &amp; distribution." />`)
}

func TestBuildBrandedIndexPageAddsCanonicalAndPublicPageSEO(t *testing.T) {
	c := newIndexPageTestContext("http://127.0.0.1/pricing?source=ad")
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	c.Request.Header.Set("X-Forwarded-Host", "omnirouters.com")
	restore := setIndexPageBranding("OmniRouters", "/logo.png")
	defer restore()

	result := string(buildBrandedIndexPage(c, []byte(`<!doctype html><html><head><title>New API</title></head><body></body></html>`)))

	assertContains(t, result, "<title>AI Model Pricing | OmniRouters</title>")
	assertContains(t, result, `<meta name="robots" content="index, follow" />`)
	assertContains(t, result, `<link rel="canonical" href="https://omnirouters.com/pricing" />`)
	assertContains(t, result, `<meta property="og:url" content="https://omnirouters.com/pricing" />`)
	assertContains(t, result, `<noscript data-seo-fallback="true"><main><h1>OmniRouters</h1>`)
}

func TestBuildBrandedIndexPageUsesHomepageSearchMetadata(t *testing.T) {
	c := newIndexPageTestContext("https://omnirouters.com/")
	restore := setIndexPageBranding("OmniRouters", "/logo.png")
	defer restore()

	result := string(buildBrandedIndexPage(c, []byte(`<!doctype html><html><head><title>New API</title><meta name="description" content="Generic description" /></head><body></body></html>`)))

	assertContains(t, result, "<title>Unified AI API Gateway | OmniRouters</title>")
	assertContains(t, result, `<meta name="description" content="Access multiple AI providers through one unified API. Compare models and pricing, route requests, and manage usage with OmniRouters." />`)
	assertContains(t, result, `<meta property="og:title" content="Unified AI API Gateway | OmniRouters" />`)
}

func TestBuildBrandedIndexPageUsesLongTailPageSEO(t *testing.T) {
	c := newIndexPageTestContext("https://omnirouters.com/openai-compatible-api")
	restore := setIndexPageBranding("OmniRouters", "/logo.png")
	defer restore()

	result := string(buildBrandedIndexPage(c, []byte(`<!doctype html><html><head></head><body></body></html>`)))

	assertContains(t, result, "<title>OpenAI-Compatible API for Multiple Models | OmniRouters</title>")
	assertContains(t, result, `<meta name="robots" content="index, follow" />`)
	assertContains(t, result, `<link rel="canonical" href="https://omnirouters.com/openai-compatible-api" />`)
	assertContains(t, result, "Use the OpenAI SDK with multiple AI models")
	assertContains(t, result, "https://omnirouters.com/v1")
}

func TestBuildBrandedIndexPageUsesMultiProviderLongTailFallback(t *testing.T) {
	c := newIndexPageTestContext("https://omnirouters.com/multi-provider-ai-api")
	restore := setIndexPageBranding("OmniRouters", "/logo.png")
	defer restore()

	result := string(buildBrandedIndexPage(c, []byte(`<!doctype html><html><head></head><body></body></html>`)))

	assertContains(t, result, "<title>Multi-Provider AI API Gateway | OmniRouters</title>")
	assertContains(t, result, "One API gateway for multiple AI providers")
	assertContains(t, result, "POST /v1/messages")
}

func TestBuildBrandedIndexPageMarksPrivatePagesNoindex(t *testing.T) {
	c := newIndexPageTestContext("https://omnirouters.com/console/topup")
	restore := setIndexPageBranding("OmniRouters", "/logo.png")
	defer restore()

	result := string(buildBrandedIndexPage(c, []byte(`<!doctype html><html><head></head><body></body></html>`)))

	assertContains(t, result, `<meta name="robots" content="noindex, nofollow" />`)
}

func TestBuildBrandedIndexPageNoindexesUnknownRoutes(t *testing.T) {
	c := newIndexPageTestContext("https://omnirouters.com/not-a-real-page")
	restore := setIndexPageBranding("OmniRouters", "/logo.png")
	defer restore()

	result := string(buildBrandedIndexPage(c, []byte(`<!doctype html><html><head></head><body></body></html>`)))

	assertContains(t, result, `<meta name="robots" content="noindex, nofollow" />`)
}

func TestServeRobotsUsesForwardedPublicOrigin(t *testing.T) {
	router := gin.New()
	registerSEORoutes(router)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/robots.txt", nil)
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Header.Set("X-Forwarded-Host", "omnirouters.com")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected robots status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Fatalf("expected text/plain content type, got %q", got)
	}
	assertContains(t, recorder.Body.String(), "User-agent: *\n")
	assertContains(t, recorder.Body.String(), "Sitemap: https://omnirouters.com/sitemap.xml\n")
	if strings.Contains(recorder.Body.String(), "Disallow: /console") {
		t.Fatalf("robots should allow private pages to be crawled so their noindex metadata can be seen")
	}
}

func TestServeSitemapContainsOnlyPublicCanonicalURLs(t *testing.T) {
	router := gin.New()
	registerSEORoutes(router)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/sitemap.xml", nil)
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Header.Set("X-Forwarded-Host", "omnirouters.com")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected sitemap status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/xml") {
		t.Fatalf("expected application/xml content type, got %q", got)
	}
	assertContains(t, recorder.Body.String(), "<loc>https://omnirouters.com/</loc>")
	assertContains(t, recorder.Body.String(), "<loc>https://omnirouters.com/pricing</loc>")
	assertContains(t, recorder.Body.String(), "<loc>https://omnirouters.com/openai-compatible-api</loc>")
	assertContains(t, recorder.Body.String(), "<loc>https://omnirouters.com/multi-provider-ai-api</loc>")
	if strings.Contains(recorder.Body.String(), "/console") || strings.Contains(recorder.Body.String(), "/keys") {
		t.Fatalf("sitemap contains a private route: %s", recorder.Body.String())
	}
}

func newIndexPageTestContext(target string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return c
}

func setIndexPageBranding(systemName string, logo string) func() {
	common.OptionMapRWMutex.Lock()
	originalSystemName := common.SystemName
	originalLogo := common.Logo
	common.SystemName = systemName
	common.Logo = logo
	common.OptionMapRWMutex.Unlock()

	return func() {
		common.OptionMapRWMutex.Lock()
		common.SystemName = originalSystemName
		common.Logo = originalLogo
		common.OptionMapRWMutex.Unlock()
	}
}

func assertContains(t *testing.T, haystack string, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected result to contain %q\nresult:\n%s", needle, haystack)
	}
}
