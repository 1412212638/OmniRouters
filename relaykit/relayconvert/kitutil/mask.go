package kitutil

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	maskURLPattern    = regexp.MustCompile(`(http|https)://[^\s/$.?#].[^\s]*`)
	maskDomainPattern = regexp.MustCompile(`\b(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}\b`)
	maskIPPattern     = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	// maskApiKeyPattern matches patterns like 'api_key:xxx' or "api_key:xxx" to mask the API key value
	maskApiKeyPattern = regexp.MustCompile(`(['"]?)api_key:([^\s'"]+)(['"]?)`)
)

// maskHostTail returns the tail parts of a domain/host that should be preserved.
// It keeps 2 parts for likely country-code TLDs (e.g., co.uk, com.cn), otherwise keeps only the TLD.
func maskHostTail(parts []string) []string {
	if len(parts) < 2 {
		return parts
	}
	lastPart := parts[len(parts)-1]
	secondLastPart := parts[len(parts)-2]
	if len(lastPart) == 2 && len(secondLastPart) <= 3 {
		// Likely country code TLD like co.uk, com.cn
		return []string{secondLastPart, lastPart}
	}
	return []string{lastPart}
}

// maskHostForURL collapses subdomains and keeps only masked prefix + preserved tail.
// Example: api.openai.com -> ***.com, sub.domain.co.uk -> ***.co.uk
func maskHostForURL(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return "***"
	}
	tail := maskHostTail(parts)
	return "***." + strings.Join(tail, ".")
}

// maskHostForPlainDomain masks a plain domain and reflects subdomain depth with multiple ***.
// Example: openai.com -> ***.com, api.openai.com -> ***.***.com, sub.domain.co.uk -> ***.***.co.uk
func maskHostForPlainDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return domain
	}
	tail := maskHostTail(parts)
	numStars := len(parts) - len(tail)
	if numStars < 1 {
		numStars = 1
	}
	stars := strings.TrimSuffix(strings.Repeat("***.", numStars), ".")
	return stars + "." + strings.Join(tail, ".")
}

// MaskSensitiveInfo masks sensitive information like URLs, IPs, and domain names in a string
// Example:
// http://example.com -> http://***.com
// https://api.test.org/v1/users/123?key=secret -> https://***.org/***/***/?key=***
// https://sub.domain.co.uk/path/to/resource -> https://***.co.uk/***/***
// 192.168.1.1 -> ***.***.***.***
// openai.com -> ***.com
// www.openai.com -> ***.***.com
// api.openai.com -> ***.***.com
func MaskSensitiveInfo(str string, upstreamURLs ...string) string {
	if len(upstreamURLs) > 0 {
		return maskConfiguredUpstream(str, upstreamURLs)
	}
	// Mask URLs
	str = maskURLPattern.ReplaceAllStringFunc(str, func(urlStr string) string {
		u, err := url.Parse(urlStr)
		if err != nil {
			return urlStr
		}

		host := u.Host
		if host == "" {
			return urlStr
		}

		// Mask host with unified logic
		maskedHost := maskHostForURL(host)

		result := u.Scheme + "://" + maskedHost

		// Mask path
		if u.Path != "" && u.Path != "/" {
			pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
			maskedPathParts := make([]string, len(pathParts))
			for i := range pathParts {
				if pathParts[i] != "" {
					maskedPathParts[i] = "***"
				}
			}
			if len(maskedPathParts) > 0 {
				result += "/" + strings.Join(maskedPathParts, "/")
			}
		} else if u.Path == "/" {
			result += "/"
		}

		// Mask query parameters
		if u.RawQuery != "" {
			values, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				// If can't parse query, just mask the whole query string
				result += "?***"
			} else {
				maskedParams := make([]string, 0, len(values))
				for key := range values {
					maskedParams = append(maskedParams, key+"=***")
				}
				if len(maskedParams) > 0 {
					result += "?" + strings.Join(maskedParams, "&")
				}
			}
		}

		return result
	})

	// Mask domain names without protocol (like openai.com, www.openai.com)
	str = maskDomainPattern.ReplaceAllStringFunc(str, func(domain string) string {
		return maskHostForPlainDomain(domain)
	})

	// Mask IP addresses
	str = maskIPPattern.ReplaceAllString(str, "***.***.***.***")

	// Mask API keys (e.g., "api_key:AIzaSyAAAaUooTUni8AdaOkSRMda30n_Q4vrV70" -> "api_key:***")
	str = maskApiKeyPattern.ReplaceAllString(str, "${1}api_key:***${3}")

	return str
}

func maskConfiguredUpstream(str string, upstreamURLs []string) string {
	hosts := make(map[string]struct{}, len(upstreamURLs))
	for _, raw := range upstreamURLs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if !strings.Contains(raw, "://") {
			raw = "https://" + raw
		}
		parsed, err := url.Parse(raw)
		if err == nil && parsed.Hostname() != "" {
			hosts[strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))] = struct{}{}
		}
	}
	str = maskApiKeyPattern.ReplaceAllString(str, "${1}api_key:***${3}")
	str = maskURLPattern.ReplaceAllStringFunc(str, func(raw string) string {
		u, err := url.Parse(raw)
		if err != nil {
			return raw
		}
		if _, configured := hosts[strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))]; configured {
			maskedHost := maskHostForURL(u.Hostname())
			if port := u.Port(); port != "" {
				maskedHost += ":" + port
			}
			path := ""
			if u.Path != "" && u.Path != "/" {
				parts := strings.Split(strings.Trim(u.Path, "/"), "/")
				for index := range parts {
					parts[index] = "***"
				}
				path = "/" + strings.Join(parts, "/")
			} else if u.Path == "/" {
				path = "/"
			}
			query := ""
			if u.RawQuery != "" {
				query = "?***"
			}
			return u.Scheme + "://" + maskedHost + path + query
		}
		// Keep public provider addresses readable, but do not expose credentials
		// embedded in URL userinfo or query values.
		if u.User != nil {
			u.User = url.User("***")
		}
		if u.RawQuery != "" {
			values, parseErr := url.ParseQuery(u.RawQuery)
			if parseErr != nil {
				u.RawQuery = "***"
			} else {
				for key := range values {
					values[key] = []string{"***"}
				}
				u.RawQuery = values.Encode()
			}
		}
		return u.String()
	})
	// An upstream hostname may be printed without a URL scheme in DNS errors.
	for host := range hosts {
		pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(host) + `\b`)
		str = pattern.ReplaceAllStringFunc(str, maskHostForPlainDomain)
	}
	str = maskBracketedConfiguredIPv6(str)
	str = maskIPPattern.ReplaceAllString(str, "***.***.***.***")
	return str
}

func maskBracketedConfiguredIPv6(str string) string {
	pattern := regexp.MustCompile(`\[([0-9A-Fa-f:.]*:[0-9A-Fa-f:.]*)(%[^\]\s]*)?\]`)
	return pattern.ReplaceAllStringFunc(str, func(match string) string {
		parts := pattern.FindStringSubmatch(match)
		if len(parts) < 2 || net.ParseIP(parts[1]) == nil {
			return match
		}
		return "[***]"
	})
}
