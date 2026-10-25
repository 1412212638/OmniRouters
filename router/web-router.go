package router

import (
	"embed"
	"encoding/xml"
	htmlpkg "html"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

type ThemeAssets struct {
	DefaultBuildFS   embed.FS
	DefaultIndexPage []byte
	ClassicBuildFS   embed.FS
	ClassicIndexPage []byte
}

func SetWebRouter(router *gin.Engine, assets ThemeAssets) {
	defaultFS := common.EmbedFolder(assets.DefaultBuildFS, "web/default/dist")
	classicFS := common.EmbedFolder(assets.ClassicBuildFS, "web/classic/dist")
	themeFS := common.NewThemeAwareFS(defaultFS, classicFS)

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.GlobalWebRateLimit())
	router.Use(middleware.Cache())
	router.Use(dynamicFaviconMiddleware())
	router.Use(static.Serve("/", themeFS))
	router.NoRoute(func(c *gin.Context) {
		c.Set(middleware.RouteTagKey, "web")
		if strings.HasPrefix(c.Request.RequestURI, "/v1") || strings.HasPrefix(c.Request.RequestURI, "/api") || strings.HasPrefix(c.Request.RequestURI, "/assets") {
			controller.RelayNotFound(c)
			return
		}
		c.Header("Cache-Control", "no-cache")
		if common.GetTheme() == "classic" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", buildBrandedIndexPage(c, assets.ClassicIndexPage))
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", buildBrandedIndexPage(c, assets.DefaultIndexPage))
	})
}

func dynamicFaviconMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request != nil && c.Request.URL != nil && c.Request.URL.Path == "/favicon.ico" {
			serveFavicon(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

func buildBrandedIndexPage(c *gin.Context, page []byte) []byte {
	document := string(page)

	common.OptionMapRWMutex.RLock()
	systemName := strings.TrimSpace(common.SystemName)
	logo := strings.TrimSpace(common.Logo)
	common.OptionMapRWMutex.RUnlock()

	if systemName == "" {
		systemName = "OmniRouters"
	}
	if logo == "" {
		logo = "/logo.png"
	}

	description := strings.TrimSpace(getMetaNameContentByLang(document, "description", "en"))
	if description == "" {
		description = strings.TrimSpace(getMetaNameContent(document, "description"))
	}
	if description == "" {
		description = "A unified AI model hub for aggregation & distribution."
	}

	pageSEO := getPageSEO(c, systemName, description)
	if pageSEO.description != "" {
		description = pageSEO.description
	}

	absoluteLogo := absoluteWebURL(c, logo)
	pageURL := canonicalWebURL(c)

	document = replaceTitle(document, pageSEO.title)
	document = upsertLinkRel(document, "icon", logo)
	document = upsertLinkRel(document, "apple-touch-icon", logo)
	document = upsertMetaName(document, "title", pageSEO.title)
	document = upsertMetaName(document, "description", description)
	document = upsertMetaName(document, "robots", pageSEO.robots)
	if pageURL != "" {
		document = upsertLinkRel(document, "canonical", pageURL)
	}
	document = upsertMetaProperty(document, "og:title", pageSEO.title)
	document = upsertMetaProperty(document, "og:description", description)
	document = upsertMetaProperty(document, "og:image", absoluteLogo)
	document = upsertMetaProperty(document, "og:type", "website")
	if pageURL != "" {
		document = upsertMetaProperty(document, "og:url", pageURL)
	}
	document = upsertMetaName(document, "twitter:card", "summary")
	document = upsertMetaName(document, "twitter:title", pageSEO.title)
	document = upsertMetaName(document, "twitter:description", description)
	document = upsertMetaName(document, "twitter:image", absoluteLogo)
	if pageSEO.robots == "index, follow" {
		path := "/"
		if c != nil && c.Request != nil && c.Request.URL != nil {
			path = strings.TrimSuffix(c.Request.URL.Path, "/")
			if path == "" {
				path = "/"
			}
		}
		document = upsertPublicNoScriptContent(
			document,
			systemName,
			pageSEO.description,
			path,
			webOrigin(c),
			controller.GetPublicBillingInfo(),
		)
	}

	return []byte(document)
}

type pageSEO struct {
	title       string
	description string
	robots      string
}

func getPageSEO(c *gin.Context, systemName string, defaultDescription string) pageSEO {
	path := "/"
	if c != nil && c.Request != nil && c.Request.URL != nil && c.Request.URL.Path != "" {
		path = strings.TrimSuffix(c.Request.URL.Path, "/")
		if path == "" {
			path = "/"
		}
	}

	result := pageSEO{
		title:       systemName,
		description: defaultDescription,
		robots:      "noindex, nofollow",
	}

	switch {
	case path == "/":
		result.title = "Unified AI API Gateway | " + systemName
		result.description = "Access multiple AI providers through one unified API. Compare models and pricing, route requests, and manage usage with " + systemName + "."
		result.robots = "index, follow"
	case path == "/pricing" || strings.HasPrefix(path, "/pricing/"):
		result.title = "AI Model Pricing | " + systemName
		result.description = "Compare supported AI models, endpoints, and usage pricing on " + systemName + "."
		result.robots = "index, follow"
	case path == "/docs" || strings.HasPrefix(path, "/docs/"):
		result.title = "Developer API Documentation | " + systemName
		result.description = "Connect to multiple AI providers through a unified API. Read the " + systemName + " developer documentation."
		result.robots = "index, follow"
	case path == "/about":
		result.title = "About " + systemName
		result.description = "Learn about " + systemName + ", a unified AI API gateway for model access and routing."
		result.robots = "index, follow"
	case path == "/openai-compatible-api":
		result.title = "OpenAI-Compatible API for Multiple Models | " + systemName
		result.description = "Use the OpenAI SDK with a compatible API endpoint to access supported AI models through " + systemName + ". Check model availability and pricing before you send requests."
		result.robots = "index, follow"
	case path == "/multi-provider-ai-api":
		result.title = "Multi-Provider AI API Gateway | " + systemName
		result.description = "Access supported OpenAI-compatible, Anthropic-compatible, and Gemini-compatible AI endpoints through one account. Compare available models and usage pricing on " + systemName + "."
		result.robots = "index, follow"
	case path == "/rankings":
		result.title = "AI Model Rankings | " + systemName
		result.description = "Compare AI model performance and availability across providers on " + systemName + "."
		result.robots = "index, follow"
	case path == "/user-agreement":
		result.title = "User Agreement | " + systemName
		result.robots = "index, follow"
	case path == "/privacy-policy":
		result.title = "Privacy Policy | " + systemName
		result.robots = "index, follow"
	}

	return result
}

type sitemapURLSet struct {
	XMLName xml.Name    `xml:"urlset"`
	XMLNS   string      `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc string `xml:"loc"`
}

func registerSEORoutes(router *gin.Engine) {
	router.GET("/robots.txt", serveRobots)
	router.GET("/sitemap.xml", serveSitemap)
	router.GET("/favicon.ico", serveFavicon)
}

func serveFavicon(c *gin.Context) {
	common.OptionMapRWMutex.RLock()
	logo := strings.TrimSpace(common.Logo)
	common.OptionMapRWMutex.RUnlock()
	if logo == "" {
		logo = "/logo.png"
	}

	c.Header("Cache-Control", "no-cache")
	c.Redirect(http.StatusFound, logo)
}

func serveRobots(c *gin.Context) {
	lines := []string{
		"User-agent: *",
		"Allow: /",
		"Disallow: /api/",
		"Disallow: /v1/",
	}
	if sitemapURL := absoluteWebURL(c, "/sitemap.xml"); sitemapURL != "" {
		lines = append(lines, "Sitemap: "+sitemapURL)
	}

	c.Header("Cache-Control", "public, max-age=3600")
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(strings.Join(lines, "\n")+"\n"))
}

func serveSitemap(c *gin.Context) {
	origin := webOrigin(c)
	if origin == "" {
		c.String(http.StatusInternalServerError, "site origin is unavailable")
		return
	}

	paths := []string{"/", "/pricing", "/rankings", "/about", "/openai-compatible-api", "/multi-provider-ai-api", "/user-agreement", "/privacy-policy"}
	urls := make([]sitemapURL, 0, len(paths))
	for _, path := range paths {
		urls = append(urls, sitemapURL{Loc: origin + path})
	}

	data, err := xml.Marshal(sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to build sitemap")
		return
	}

	body := append([]byte(xml.Header), data...)
	c.Header("Cache-Control", "public, max-age=3600")
	c.Data(http.StatusOK, "application/xml; charset=utf-8", body)
}

func replaceTitle(document string, title string) string {
	tag := "<title>" + htmlpkg.EscapeString(title) + "</title>"
	titleRE := regexp.MustCompile(`(?is)<title>.*?</title>`)
	if titleRE.MatchString(document) {
		return titleRE.ReplaceAllString(document, tag)
	}
	return insertBeforeHeadClose(document, tag)
}

func upsertLinkRel(document string, rel string, href string) string {
	tag := `<link rel="` + htmlpkg.EscapeString(rel) + `" href="` + htmlpkg.EscapeString(href) + `" />`
	if rel == "icon" {
		tag = `<link rel="icon" type="image/png" href="` + htmlpkg.EscapeString(href) + `" />`
	}
	linkRE := regexp.MustCompile(`(?is)<link\s+[^>]*rel=["']` + regexp.QuoteMeta(rel) + `["'][^>]*>`)
	if linkRE.MatchString(document) {
		return linkRE.ReplaceAllString(document, tag)
	}
	return insertBeforeHeadClose(document, tag)
}

func upsertMetaName(document string, name string, content string) string {
	tag := `<meta name="` + htmlpkg.EscapeString(name) + `" content="` + htmlpkg.EscapeString(content) + `" />`
	metaRE := regexp.MustCompile(`(?is)<meta\s+[^>]*name=["']` + regexp.QuoteMeta(name) + `["'][^>]*>`)
	if metaRE.MatchString(document) {
		return metaRE.ReplaceAllString(document, tag)
	}
	return insertBeforeHeadClose(document, tag)
}

func upsertMetaProperty(document string, property string, content string) string {
	tag := `<meta property="` + htmlpkg.EscapeString(property) + `" content="` + htmlpkg.EscapeString(content) + `" />`
	metaRE := regexp.MustCompile(`(?is)<meta\s+[^>]*property=["']` + regexp.QuoteMeta(property) + `["'][^>]*>`)
	if metaRE.MatchString(document) {
		return metaRE.ReplaceAllString(document, tag)
	}
	return insertBeforeHeadClose(document, tag)
}

func getMetaNameContent(document string, name string) string {
	for _, tag := range metaTags(document) {
		if strings.EqualFold(attrValue(tag, "name"), name) {
			return htmlpkg.UnescapeString(attrValue(tag, "content"))
		}
	}
	return ""
}

func getMetaNameContentByLang(document string, name string, lang string) string {
	for _, tag := range metaTags(document) {
		if strings.EqualFold(attrValue(tag, "name"), name) && strings.EqualFold(attrValue(tag, "lang"), lang) {
			return htmlpkg.UnescapeString(attrValue(tag, "content"))
		}
	}
	return ""
}

func metaTags(document string) []string {
	metaTagsRE := regexp.MustCompile(`(?is)<meta\s+[^>]*>`)
	return metaTagsRE.FindAllString(document, -1)
}

func attrValue(tag string, attr string) string {
	attrRE := regexp.MustCompile(`(?is)\s` + regexp.QuoteMeta(attr) + `=["']([^"']*)["']`)
	matches := attrRE.FindStringSubmatch(tag)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func insertBeforeHeadClose(document string, tag string) string {
	lowerDocument := strings.ToLower(document)
	index := strings.Index(lowerDocument, "</head>")
	if index < 0 {
		return document + "\n" + tag
	}
	return document[:index] + "    " + tag + "\n" + document[index:]
}

func upsertPublicNoScriptContent(document string, systemName string, description string, path string, origin string, billing controller.PublicBillingInfo) string {
	content := publicNoScriptBody(systemName, description, path, origin, billing)
	content = `<noscript>` + content + `</noscript>`
	noscriptRE := regexp.MustCompile(`(?is)<noscript\s+data-seo-fallback=["'][^"']*["'][^>]*>.*?</noscript>`)
	if noscriptRE.MatchString(document) {
		return noscriptRE.ReplaceAllStringFunc(document, func(string) string { return content })
	}
	content = strings.Replace(content, "<noscript>", `<noscript data-seo-fallback="true">`, 1)
	return insertBeforeBodyClose(document, content)
}

func publicNoScriptBody(systemName string, description string, path string, origin string, billing controller.PublicBillingInfo) string {
	name := htmlpkg.EscapeString(systemName)
	apiBaseURL := strings.TrimRight(origin, "/") + "/v1"
	if origin == "" {
		apiBaseURL = "/v1"
	}
	billingCopy := publicBillingCopy(billing)
	pricingLink := `<a href="/pricing">AI model pricing</a>`
	openAILink := `<a href="/openai-compatible-api">OpenAI-compatible API guide</a>`
	multiProviderLink := `<a href="/multi-provider-ai-api">Multi-provider AI API guide</a>`

	switch path {
	case "/openai-compatible-api":
		code := "from openai import OpenAI\n\nclient = OpenAI(\n    base_url=\"" + apiBaseURL + "\",\n    api_key=\"YOUR_API_KEY\",\n)\n\nresponse = client.chat.completions.create(\n    model=\"YOUR_ENABLED_MODEL\",\n    messages=[{\"role\": \"user\", \"content\": \"Hello\"}],\n)"
		return `<main><h1>Use the OpenAI SDK with multiple AI models</h1><p>Point an OpenAI SDK client at the ` + name + ` API endpoint and select an enabled model from the model catalog. The OpenAI-compatible chat endpoint is <code>POST /v1/chat/completions</code>; model IDs and supported features vary.</p><h2>Python quick start</h2><pre><code>` + htmlpkg.EscapeString(code) + `</code></pre><p>Create an API key in your account and replace the example model ID with one listed in the catalog. Check current pricing and model availability before sending production traffic.</p>` + billingCopy + `<nav>` + pricingLink + ` | ` + multiProviderLink + `</nav></main>`
	case "/multi-provider-ai-api":
		return `<main><h1>One API gateway for multiple AI providers</h1><p>` + name + ` provides access to supported models through a single account and API key. Use the model catalog as the source of truth for current availability, pricing, and compatible endpoints.</p><h2>Choose a compatible API format</h2><ul><li>OpenAI-compatible chat: <code>POST /v1/chat/completions</code></li><li>Anthropic-compatible messages: <code>POST /v1/messages</code>, for compatible models</li><li>Gemini-compatible generate content: <code>/v1beta/models/{model}:generateContent</code>, for compatible models</li></ul><p>API format compatibility does not mean every model supports every endpoint or request parameter. Match the endpoint and capabilities shown for the model you plan to use.</p>` + billingCopy + `<p>Usage is charged according to the selected model's current pricing.</p><nav>` + pricingLink + ` | ` + openAILink + `</nav></main>`
	default:
		return `<main><h1>` + name + `</h1><p>` + htmlpkg.EscapeString(description) + `</p><nav>` + pricingLink + ` | ` + openAILink + ` | ` + multiProviderLink + ` | <a href="/about">About</a> | <a href="/user-agreement">User agreement</a> | <a href="/privacy-policy">Privacy policy</a></nav></main>`
	}
}

func publicBillingCopy(billing controller.PublicBillingInfo) string {
	parts := make([]string, 0, 2)
	if billing.RegistrationCreditEnabled {
		parts = append(parts, "Registration includes a default initial account balance.")
	} else {
		parts = append(parts, "New accounts start without a default free balance.")
	}
	if billing.MinimumTopupEnabled && billing.MinimumTopup > 0 {
		amount := strconv.FormatFloat(billing.MinimumTopup, 'f', -1, 64)
		if billing.CurrencyDisplayType == "TOKENS" {
			parts = append(parts, "The minimum top-up is "+amount+" tokens.")
		} else {
			parts = append(parts, "The minimum top-up is "+billing.CurrencySymbol+amount+".")
		}
	}
	return `<p>` + htmlpkg.EscapeString(strings.Join(parts, " ")) + `</p>`
}

func insertBeforeBodyClose(document string, tag string) string {
	lowerDocument := strings.ToLower(document)
	index := strings.Index(lowerDocument, "</body>")
	if index < 0 {
		return document + "\n" + tag
	}
	return document[:index] + "    " + tag + "\n" + document[index:]
}

func absoluteWebURL(c *gin.Context, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}

	scheme := requestScheme(c)
	if strings.HasPrefix(value, "//") {
		return scheme + ":" + value
	}

	host := requestHost(c)
	if host == "" {
		return value
	}
	if strings.HasPrefix(value, "/") {
		return scheme + "://" + host + value
	}
	return scheme + "://" + host + "/" + value
}

func currentWebURL(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	host := requestHost(c)
	if host == "" {
		return ""
	}
	return requestScheme(c) + "://" + host + c.Request.URL.RequestURI()
}

func canonicalWebURL(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return ""
	}
	origin := webOrigin(c)
	if origin == "" {
		return ""
	}
	path := c.Request.URL.Path
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}
	return origin + path
}

func webOrigin(c *gin.Context) string {
	host := requestHost(c)
	if host == "" {
		return ""
	}
	return requestScheme(c) + "://" + host
}

func requestScheme(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return "https"
	}
	if forwardedProto := firstForwardedHeaderValue(c.GetHeader("X-Forwarded-Proto")); forwardedProto != "" {
		return forwardedProto
	}
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}

func requestHost(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if forwardedHost := firstForwardedHeaderValue(c.GetHeader("X-Forwarded-Host")); forwardedHost != "" {
		return forwardedHost
	}
	return c.Request.Host
}

func firstForwardedHeaderValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(value, ",")[0])
}
