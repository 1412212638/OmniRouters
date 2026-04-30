package router

import (
	"embed"
	htmlpkg "html"
	"net/http"
	"regexp"
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

	absoluteLogo := absoluteWebURL(c, logo)
	pageURL := currentWebURL(c)

	document = replaceTitle(document, systemName)
	document = upsertLinkRel(document, "icon", logo)
	document = upsertLinkRel(document, "apple-touch-icon", logo)
	document = upsertMetaName(document, "title", systemName)
	document = upsertMetaProperty(document, "og:title", systemName)
	document = upsertMetaProperty(document, "og:description", description)
	document = upsertMetaProperty(document, "og:image", absoluteLogo)
	document = upsertMetaProperty(document, "og:type", "website")
	if pageURL != "" {
		document = upsertMetaProperty(document, "og:url", pageURL)
	}
	document = upsertMetaName(document, "twitter:card", "summary")
	document = upsertMetaName(document, "twitter:title", systemName)
	document = upsertMetaName(document, "twitter:description", description)
	document = upsertMetaName(document, "twitter:image", absoluteLogo)

	return []byte(document)
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
