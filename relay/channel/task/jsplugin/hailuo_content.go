package jsplugin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/service"
)

// resolveHailuoFileContent performs only MiniMax's file metadata lookup, on
// the configured channel origin. It never fetches arbitrary URLs for plugin
// JavaScript and never carries the channel credential into the video request.
func (a *TaskAdaptor) resolveHailuoFileContent(artifactKey string, clientRequest channel.TaskArtifactClientRequest, descriptor requestDescriptor) (*channel.TaskContentRequest, error) {
	if artifactKey != "video" || (clientRequest.Method != http.MethodGet && clientRequest.Method != http.MethodHead) ||
		descriptor.Method != http.MethodGet || descriptor.Body != nil || descriptor.Credentialless {
		return nil, fmt.Errorf("invalid Hailuo file lookup")
	}
	base, baseErr := url.Parse(a.info.ChannelBaseUrl)
	lookup, lookupErr := url.Parse(descriptor.URL)
	if baseErr != nil || lookupErr != nil || lookup.Host == "" || lookup.User != nil || lookup.Fragment != "" ||
		(lookup.Scheme != "https" && lookup.Scheme != "http") ||
		!strings.EqualFold(lookup.Scheme, base.Scheme) || !strings.EqualFold(lookup.Host, base.Host) ||
		lookup.Path != strings.TrimRight(base.Path, "/")+"/v1/files/retrieve" {
		return nil, fmt.Errorf("invalid Hailuo file lookup URL")
	}
	query, err := url.ParseQuery(lookup.RawQuery)
	if err != nil || len(query) != 1 || len(query["file_id"]) != 1 || strings.TrimSpace(query.Get("file_id")) == "" {
		return nil, fmt.Errorf("invalid Hailuo file id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, descriptor.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid Hailuo file lookup request")
	}
	for name, value := range descriptor.Headers {
		req.Header.Set(name, value)
	}
	client, err := service.GetHttpClientWithProxySettings(a.info.ChannelSetting.Proxy, a.info.ChannelSetting)
	if err != nil {
		return nil, fmt.Errorf("Hailuo file lookup client is unavailable")
	}
	// A metadata redirect must not carry a vendor credential to another URL.
	lookupClient := *client
	lookupClient.Jar = nil
	lookupClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := lookupClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Hailuo file lookup failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Hailuo file lookup returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxTaskPluginPersistedJSONBytes+1))
	if err != nil || len(body) > maxTaskPluginPersistedJSONBytes {
		return nil, fmt.Errorf("invalid Hailuo file lookup response size")
	}
	var result struct {
		BaseResp *struct {
			StatusCode *int `json:"status_code"`
		} `json:"base_resp"`
		File *struct {
			DownloadURL string `json:"download_url"`
		} `json:"file"`
	}
	if err = common.Unmarshal(body, &result); err != nil || result.BaseResp == nil || result.BaseResp.StatusCode == nil || *result.BaseResp.StatusCode != 0 || result.File == nil {
		return nil, fmt.Errorf("Hailuo file lookup did not succeed")
	}
	downloadURL := strings.TrimSpace(result.File.DownloadURL)
	video, err := url.Parse(downloadURL)
	if err != nil || video.Host == "" || video.User != nil || video.Fragment != "" || (video.Scheme != "https" && video.Scheme != "http") {
		return nil, fmt.Errorf("invalid Hailuo video download URL")
	}
	// The media proxy applies its existing SSRF checks to this URL and every
	// redirect. No metadata headers, body or credentials may reach the CDN.
	return &channel.TaskContentRequest{URL: downloadURL, Method: http.MethodGet, Credentialless: true}, nil
}
