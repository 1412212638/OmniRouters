package controller

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"

	"github.com/gin-gonic/gin"
)

const (
	wechatOfficialMaxRootButtons = 3
	wechatOfficialMaxSubButtons  = 5
	wechatOfficialMaxMenuName    = 64
	wechatOfficialMaxMenuKey     = 128
	wechatOfficialMaxReply       = 2000
)

// WeChatOfficialMenuConfig is the locally stored menu draft. Reply is local
// metadata used for CLICK callbacks and is intentionally removed before the
// menu is sent to WeChat.
type WeChatOfficialMenuConfig struct {
	Button []WeChatOfficialMenuButton `json:"button"`
}

type WeChatOfficialMenuButton struct {
	Name      string                     `json:"name"`
	Type      string                     `json:"type,omitempty"`
	Key       string                     `json:"key,omitempty"`
	URL       string                     `json:"url,omitempty"`
	Reply     string                     `json:"reply,omitempty"`
	SubButton []WeChatOfficialMenuButton `json:"sub_button,omitempty"`
}

type wechatOfficialRemoteMenu struct {
	Button []wechatOfficialRemoteButton `json:"button"`
}

type wechatOfficialRemoteButton struct {
	Name      string                       `json:"name"`
	Type      string                       `json:"type,omitempty"`
	Key       string                       `json:"key,omitempty"`
	URL       string                       `json:"url,omitempty"`
	SubButton []wechatOfficialRemoteButton `json:"sub_button,omitempty"`
}

type wechatOfficialMenuAPIResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func parseWeChatOfficialMenu(raw string) (WeChatOfficialMenuConfig, error) {
	if strings.TrimSpace(raw) == "" {
		return WeChatOfficialMenuConfig{}, nil
	}
	var config WeChatOfficialMenuConfig
	if err := common.UnmarshalJsonStr(raw, &config); err != nil {
		return WeChatOfficialMenuConfig{}, fmt.Errorf("菜单配置不是有效的 JSON")
	}
	if err := validateWeChatOfficialMenu(config); err != nil {
		return WeChatOfficialMenuConfig{}, err
	}
	return config, nil
}

func validateWeChatOfficialMenu(config WeChatOfficialMenuConfig) error {
	if len(config.Button) > wechatOfficialMaxRootButtons {
		return fmt.Errorf("一级菜单最多配置 %d 个", wechatOfficialMaxRootButtons)
	}
	for _, button := range config.Button {
		if err := validateWeChatOfficialMenuButton(button, false); err != nil {
			return err
		}
	}
	return nil
}

func validateWeChatOfficialMenuButton(button WeChatOfficialMenuButton, child bool) error {
	name := strings.TrimSpace(button.Name)
	if name == "" {
		return fmt.Errorf("菜单名称不能为空")
	}
	if utf8.RuneCountInString(name) > wechatOfficialMaxMenuName {
		return fmt.Errorf("菜单名称不能超过 %d 个字符", wechatOfficialMaxMenuName)
	}
	if utf8.RuneCountInString(strings.TrimSpace(button.Reply)) > wechatOfficialMaxReply {
		return fmt.Errorf("菜单回复不能超过 %d 个字符", wechatOfficialMaxReply)
	}
	if child && len(button.SubButton) > 0 {
		return fmt.Errorf("二级菜单不能继续嵌套")
	}
	if len(button.SubButton) > 0 {
		if len(button.SubButton) > wechatOfficialMaxSubButtons {
			return fmt.Errorf("二级菜单最多配置 %d 个", wechatOfficialMaxSubButtons)
		}
		if strings.TrimSpace(button.Type) != "" || strings.TrimSpace(button.Key) != "" || strings.TrimSpace(button.URL) != "" || strings.TrimSpace(button.Reply) != "" {
			return fmt.Errorf("包含二级菜单的一级菜单不能设置动作或回复")
		}
		for _, subButton := range button.SubButton {
			if err := validateWeChatOfficialMenuButton(subButton, true); err != nil {
				return err
			}
		}
		return nil
	}

	buttonType := strings.ToLower(strings.TrimSpace(button.Type))
	switch buttonType {
	case "click":
		if strings.TrimSpace(button.Key) == "" {
			return fmt.Errorf("点击菜单必须填写 Key")
		}
		if utf8.RuneCountInString(strings.TrimSpace(button.Key)) > wechatOfficialMaxMenuKey {
			return fmt.Errorf("菜单 Key 不能超过 %d 个字符", wechatOfficialMaxMenuKey)
		}
	case "view":
		parsedURL, err := url.Parse(strings.TrimSpace(button.URL))
		if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			return fmt.Errorf("跳转菜单必须填写有效的 http:// 或 https:// URL")
		}
	case "":
		return fmt.Errorf("菜单必须选择点击回复或跳转网页")
	default:
		return fmt.Errorf("不支持的菜单类型: %s", button.Type)
	}
	return nil
}

func toWeChatOfficialRemoteButton(button WeChatOfficialMenuButton) wechatOfficialRemoteButton {
	remote := wechatOfficialRemoteButton{
		Name: strings.TrimSpace(button.Name),
		Type: strings.ToLower(strings.TrimSpace(button.Type)),
		Key:  strings.TrimSpace(button.Key),
		URL:  strings.TrimSpace(button.URL),
	}
	for _, subButton := range button.SubButton {
		remote.SubButton = append(remote.SubButton, toWeChatOfficialRemoteButton(subButton))
	}
	return remote
}

func buildWeChatOfficialRemoteMenu(config WeChatOfficialMenuConfig) wechatOfficialRemoteMenu {
	remote := wechatOfficialRemoteMenu{}
	for _, button := range config.Button {
		remote.Button = append(remote.Button, toWeChatOfficialRemoteButton(button))
	}
	return remote
}

func findWeChatOfficialMenuReply(buttons []WeChatOfficialMenuButton, key string) string {
	for _, button := range buttons {
		if strings.EqualFold(strings.TrimSpace(button.Type), "click") && strings.TrimSpace(button.Key) == key {
			return strings.TrimSpace(button.Reply)
		}
		if reply := findWeChatOfficialMenuReply(button.SubButton, key); reply != "" {
			return reply
		}
	}
	return ""
}

func replyWeChatOfficialText(c *gin.Context, message wechatOfficialCallbackMessage, content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		c.String(http.StatusOK, "success")
		return
	}
	response := struct {
		XMLName    xml.Name `xml:"xml"`
		ToUserName string   `xml:"ToUserName"`
		FromUserName string `xml:"FromUserName"`
		CreateTime int64    `xml:"CreateTime"`
		MsgType    string   `xml:"MsgType"`
		Content    string   `xml:"Content"`
	}{
		ToUserName:   message.FromUserName,
		FromUserName: message.ToUserName,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      content,
	}
	c.XML(http.StatusOK, response)
}

// SyncWeChatOfficialMenu publishes the saved local menu draft to the
// configured official account. An empty draft deletes the remote menu.
func SyncWeChatOfficialMenu(c *gin.Context) {
	if !wechatOfficialCredentialsConfigured() {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请先完成官方微信公众号基础配置"})
		return
	}
	config, err := parseWeChatOfficialMenu(common.WeChatOfficialMenu)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	token, err := getWeChatOfficialAccessToken()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	action := "create"
	var body []byte
	if len(config.Button) == 0 {
		action = "delete"
	} else if body, err = common.Marshal(buildWeChatOfficialRemoteMenu(config)); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "菜单配置序列化失败"})
		return
	}
	endpoint := "https://api.weixin.qq.com/cgi-bin/menu/" + action + "?access_token=" + url.QueryEscape(token)
	var requestBody *strings.Reader
	if body == nil {
		requestBody = strings.NewReader("")
	} else {
		requestBody = strings.NewReader(string(body))
	}
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, endpoint, requestBody)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 8 * time.Second}).Do(request)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer response.Body.Close()
	responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if readErr != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": readErr.Error()})
		return
	}
	var result wechatOfficialMenuAPIResponse
	if err := common.Unmarshal(responseBody, &result); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "微信菜单接口返回格式无效"})
		return
	}
	if response.StatusCode/100 != 2 || result.ErrCode != 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": fmt.Sprintf("微信菜单同步失败: %s (%d)", result.ErrMsg, result.ErrCode)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": map[string]string{"create": "微信公众号菜单已同步", "delete": "微信公众号菜单已删除"}[action]})
}

// ValidateWeChatOfficialMenu is kept small and pure so the admin editor and
// backend regression tests share the same constraints.
func ValidateWeChatOfficialMenu(raw string) error {
	_, err := parseWeChatOfficialMenu(raw)
	return err
}
