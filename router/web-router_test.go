package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
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
	c := newIndexPageTestContext("https://api.other.test/openai-compatible-api")
	restore := setIndexPageBranding("OmniRouters", "/logo.png")
	defer restore()

	result := string(buildBrandedIndexPage(c, []byte(`<!doctype html><html><head></head><body></body></html>`)))

	assertContains(t, result, "<title>OpenAI-Compatible API for Multiple Models | OmniRouters</title>")
	assertContains(t, result, `<meta name="robots" content="index, follow" />`)
	assertContains(t, result, `<link rel="canonical" href="https://api.other.test/openai-compatible-api" />`)
	assertContains(t, result, "Use the OpenAI SDK with multiple AI models")
	assertContains(t, result, "https://api.other.test/v1")
	if strings.Contains(result, "$10") {
		t.Fatalf("SEO fallback should not assume a USD minimum top-up: %s", result)
	}
}

func TestPublicNoScriptBodyUsesConfiguredBillingAndSiteOrigin(t *testing.T) {
	billing := controller.PublicBillingInfo{
		CurrencyDisplayType:       "CNY",
		CurrencySymbol:            "¥",
		MinimumTopup:              10,
		MinimumTopupEnabled:       true,
		RegistrationCreditEnabled: true,
	}

	body := publicNoScriptBody(
		"Domestic Site",
		"Description",
		"/openai-compatible-api",
		"https://cn.example.test",
		billing,
	)

	assertContains(t, body, `base_url=&#34;https://cn.example.test/v1&#34;`)
	assertContains(t, body, "Registration includes a default initial account balance.")
	assertContains(t, body, "The minimum top-up is ¥10.")
	if strings.Contains(body, "$10") {
		t.Fatalf("CNY billing copy should not contain a USD amount: %s", body)
	}
}

func TestPublicNoScriptBodyOmitsDisabledMinimumTopup(t *testing.T) {
	billing := controller.PublicBillingInfo{
		CurrencyDisplayType:       "CNY",
		CurrencySymbol:            "¥",
		MinimumTopupEnabled:       false,
		RegistrationCreditEnabled: false,
	}

	body := publicNoScriptBody(
		"Domestic Site",
		"Description",
		"/multi-provider-ai-api",
		"https://cn.example.test",
		billing,
	)

	assertContains(t, body, "New accounts start without a default free balance.")
	if strings.Contains(body, "minimum top-up") || strings.Contains(body, "$10") {
		t.Fatalf("SEO fallback should omit a disabled minimum top-up: %s", body)
	}
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

func TestServeFaviconRedirectsToConfiguredLogo(t *testing.T) {
	router := gin.New()
	registerSEORoutes(router)
	restore := setIndexPageBranding("OmniRouters", "https://cdn.example.test/brand.png")
	defer restore()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://omnirouters.com/favicon.ico", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected favicon redirect status 302, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Location"); got != "https://cdn.example.test/brand.png" {
		t.Fatalf("expected configured logo redirect, got %q", got)
	}
}

func TestServeFaviconFallsBackToDefaultLogo(t *testing.T) {
	router := gin.New()
	registerSEORoutes(router)
	restore := setIndexPageBranding("OmniRouters", "")
	defer restore()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://omnirouters.com/favicon.ico", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected favicon redirect status 302, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Location"); got != "/logo.png" {
		t.Fatalf("expected default logo redirect, got %q", got)
	}
}

func TestDynamicFaviconMiddlewareRunsBeforeStaticAssets(t *testing.T) {
	router := gin.New()
	router.Use(dynamicFaviconMiddleware())
	router.Use(func(c *gin.Context) {
		c.String(http.StatusOK, "default static favicon")
	})
	restore := setIndexPageBranding("OmniRouters", "https://cdn.example.test/brand.png")
	defer restore()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://omnirouters.com/favicon.ico", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected dynamic favicon middleware to respond before static assets, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Location"); got != "https://cdn.example.test/brand.png" {
		t.Fatalf("expected configured logo redirect before static assets, got %q", got)
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
