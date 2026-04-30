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
	c := newIndexPageTestContext("https://omnirouters.com/")
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
