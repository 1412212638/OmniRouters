package controller

import (
	"crypto/sha1"
	"crypto/subtle"
	"context"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	wechatOfficialStateTTL = 5 * time.Minute
	wechatOfficialRedisKey = "new-api:wechat:official:login:"
	wechatOfficialClaimKey = "new-api:wechat:official:claim:"
	wechatOfficialScanKey  = "new-api:wechat:official:scan:"
)

type wechatOfficialLoginState struct {
	Code      string `json:"code"`
	OpenID    string `json:"openid,omitempty"`
	ExpiresAt int64  `json:"expires_at"`
	Consumed  bool   `json:"consumed,omitempty"`
}

var wechatOfficialStateCache = struct {
	sync.Mutex
	items map[string]wechatOfficialLoginState
}{items: make(map[string]wechatOfficialLoginState)}

var wechatOfficialAccessToken = struct {
	sync.Mutex
	token     string
	appID     string
	secret    string
	expiresAt time.Time
}{}

// VerifyWeChatOfficialSignature validates the four values used by the
// official-account server callback handshake. It is exported for focused
// controller tests and deliberately uses constant-time comparison semantics
// through the digest representation.
func VerifyWeChatOfficialSignature(token, timestamp, nonce, signature string) bool {
	expected := BuildWeChatOfficialSignature(token, timestamp, nonce)
	provided := strings.ToLower(strings.TrimSpace(signature))
	return len(provided) == len(expected) && subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

// BuildWeChatOfficialSignature returns the sorted SHA-1 callback signature.
// It is useful when configuring or testing the WeChat server handshake.
func BuildWeChatOfficialSignature(token, timestamp, nonce string) string {
	parts := []string{token, timestamp, nonce}
	sort.Strings(parts)
	digest := sha1.Sum([]byte(strings.Join(parts, "")))
	return hex.EncodeToString(digest[:])
}

func wechatOfficialConfigured() bool {
	return common.WeChatOfficialEnabled &&
		strings.TrimSpace(common.WeChatOfficialAppID) != "" &&
		strings.TrimSpace(common.WeChatOfficialAppSecret) != "" &&
		strings.TrimSpace(common.WeChatOfficialToken) != ""
}

func wechatOfficialCallbackConfigured() bool {
	return strings.TrimSpace(common.WeChatOfficialAppID) != "" &&
		strings.TrimSpace(common.WeChatOfficialToken) != ""
}

func wechatOfficialStateKey(code string) string { return wechatOfficialRedisKey + code }

func saveWeChatOfficialState(state wechatOfficialLoginState) error {
	data, err := common.Marshal(state)
	if err != nil {
		return err
	}
	if common.RedisEnabled && common.RDB != nil {
		if err := common.RedisSet(wechatOfficialStateKey(state.Code), string(data), wechatOfficialStateTTL); err != nil {
			return err
		}
		return nil
	}
	wechatOfficialStateCache.Lock()
	wechatOfficialStateCache.items[state.Code] = state
	wechatOfficialStateCache.Unlock()
	return nil
}

func loadWeChatOfficialState(code string) (wechatOfficialLoginState, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return wechatOfficialLoginState{}, errors.New("微信登录状态无效")
	}
	var state wechatOfficialLoginState
	if common.RedisEnabled && common.RDB != nil {
		if raw, err := common.RedisGet(wechatOfficialStateKey(code)); err == nil {
			if err := common.UnmarshalJsonStr(raw, &state); err == nil {
				if state.ExpiresAt > time.Now().Unix() {
					return state, nil
				}
			}
		}
	}
	wechatOfficialStateCache.Lock()
	state, ok := wechatOfficialStateCache.items[code]
	if ok && state.ExpiresAt <= time.Now().Unix() {
		delete(wechatOfficialStateCache.items, code)
		ok = false
	}
	wechatOfficialStateCache.Unlock()
	if !ok {
		return wechatOfficialLoginState{}, errors.New("微信登录二维码已过期")
	}
	return state, nil
}

func updateWeChatOfficialState(state wechatOfficialLoginState) error {
	return saveWeChatOfficialState(state)
}

func deleteWeChatOfficialState(code string) {
	if common.RedisEnabled && common.RDB != nil {
		_ = common.RDB.Del(context.Background(), wechatOfficialStateKey(code), wechatOfficialClaimKey+code, wechatOfficialScanKey+code).Err()
	}
	wechatOfficialStateCache.Lock()
	delete(wechatOfficialStateCache.items, code)
	wechatOfficialStateCache.Unlock()
}

func consumeWeChatOfficialState(code string) (bool, error) {
	if common.RedisEnabled && common.RDB != nil {
		ok, err := common.RDB.SetNX(context.Background(), wechatOfficialClaimKey+code, "1", wechatOfficialStateTTL).Result()
		if err != nil {
			return false, err
		}
		return ok, nil
	}
	wechatOfficialStateCache.Lock()
	defer wechatOfficialStateCache.Unlock()
	state, ok := wechatOfficialStateCache.items[code]
	if !ok || state.Consumed {
		return false, nil
	}
	state.Consumed = true
	wechatOfficialStateCache.items[code] = state
	return true, nil
}

// claimWeChatOfficialScan makes the first callback for a QR code win. This
// prevents two nearly simultaneous WeChat callbacks from replacing the OpenID
// associated with the browser session.
func claimWeChatOfficialScan(code, openID string) (bool, error) {
	code, openID = strings.TrimSpace(code), strings.TrimSpace(openID)
	if code == "" || openID == "" {
		return false, errors.New("微信扫码状态无效")
	}
	if common.RedisEnabled && common.RDB != nil {
		ok, err := common.RDB.SetNX(context.Background(), wechatOfficialScanKey+code, openID, wechatOfficialStateTTL).Result()
		if err != nil || !ok {
			return ok, err
		}
		state, err := loadWeChatOfficialState(code)
		if err != nil {
			_ = common.RDB.Del(context.Background(), wechatOfficialScanKey+code).Err()
			return false, err
		}
		if state.OpenID != "" {
			return false, nil
		}
		state.OpenID = openID
		if err := updateWeChatOfficialState(state); err != nil {
			_ = common.RDB.Del(context.Background(), wechatOfficialScanKey+code).Err()
			return false, err
		}
		return true, nil
	}
	wechatOfficialStateCache.Lock()
	defer wechatOfficialStateCache.Unlock()
	state, ok := wechatOfficialStateCache.items[code]
	if !ok || state.ExpiresAt <= time.Now().Unix() {
		delete(wechatOfficialStateCache.items, code)
		return false, errors.New("微信登录二维码已过期")
	}
	if state.OpenID != "" {
		return false, nil
	}
	state.OpenID = openID
	wechatOfficialStateCache.items[code] = state
	return true, nil
}

type wechatOfficialAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

func getWeChatOfficialAccessToken() (string, error) {
	wechatOfficialAccessToken.Lock()
	defer wechatOfficialAccessToken.Unlock()
	if wechatOfficialAccessToken.token != "" && wechatOfficialAccessToken.appID == common.WeChatOfficialAppID && wechatOfficialAccessToken.secret == common.WeChatOfficialAppSecret && time.Now().Before(wechatOfficialAccessToken.expiresAt) {
		return wechatOfficialAccessToken.token, nil
	}
	query := url.Values{}
	query.Set("grant_type", "client_credential")
	query.Set("appid", common.WeChatOfficialAppID)
	query.Set("secret", common.WeChatOfficialAppSecret)
	request, err := http.NewRequest(http.MethodGet, "https://api.weixin.qq.com/cgi-bin/token?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}
	client := http.Client{Timeout: 8 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", err
	}
	var result wechatOfficialAccessTokenResponse
	if err := common.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if response.StatusCode/100 != 2 || result.ErrCode != 0 || result.AccessToken == "" {
		return "", fmt.Errorf("微信 access token 获取失败: %s (%d)", result.ErrMsg, result.ErrCode)
	}
	wechatOfficialAccessToken.token = result.AccessToken
	wechatOfficialAccessToken.appID = common.WeChatOfficialAppID
	wechatOfficialAccessToken.secret = common.WeChatOfficialAppSecret
	seconds := result.ExpiresIn
	if seconds <= 0 || seconds > 7200 {
		seconds = 7200
	}
	wechatOfficialAccessToken.expiresAt = time.Now().Add(time.Duration(seconds-60) * time.Second)
	return result.AccessToken, nil
}

type wechatOfficialQRCodeResponse struct {
	Ticket     string `json:"ticket"`
	ExpireSec  int    `json:"expire_seconds"`
	URL        string `json:"url"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WeChatOfficialQR creates a short-lived QR code whose scene value is tied to
// a browser login state. The frontend can display image_url and poll code.
func WeChatOfficialQR(c *gin.Context) {
	if !wechatOfficialConfigured() {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "官方微信公众号登录未配置"})
		return
	}
	code := common.GetRandomString(32)
	state := wechatOfficialLoginState{Code: code, ExpiresAt: time.Now().Add(wechatOfficialStateTTL).Unix()}
	if err := saveWeChatOfficialState(state); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	qrCreated := false
	defer func() {
		if !qrCreated {
			deleteWeChatOfficialState(code)
		}
	}()
	token, err := getWeChatOfficialAccessToken()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	payload := struct {
	ExpireSeconds int `json:"expire_seconds"`
	ActionName    string `json:"action_name"`
	ActionInfo    struct {
		Scene struct {
			SceneStr string `json:"scene_str"`
		} `json:"scene"`
	} `json:"action_info"`
	}{ExpireSeconds: int(wechatOfficialStateTTL.Seconds()), ActionName: "QR_STR_SCENE"}
	payload.ActionInfo.Scene.SceneStr = "or_" + code
	body, err := common.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	request, err := http.NewRequest(http.MethodPost, "https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token="+url.QueryEscape(token), strings.NewReader(string(body)))
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
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	var qr wechatOfficialQRCodeResponse
	if err := common.Unmarshal(responseBody, &qr); err != nil || qr.Ticket == "" || qr.ErrCode != 0 {
		if err == nil {
			err = fmt.Errorf("微信二维码生成失败: %s (%d)", qr.ErrMsg, qr.ErrCode)
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	imageURL := "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=" + url.QueryEscape(qr.Ticket)
	qrCreated = true
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"code": code, "image_url": imageURL, "expires_in": int(wechatOfficialStateTTL.Seconds()), "name": common.WeChatOfficialName,
		"callback_path": "/api/oauth/wechat/official/callback",
	}})
}

func WeChatOfficialPoll(c *gin.Context) {
	if !wechatOfficialConfigured() {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "官方微信公众号登录未配置"})
		return
	}
	state, err := loadWeChatOfficialState(c.Query("code"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	consumed := false
	if common.RedisEnabled && common.RDB != nil {
		consumed = common.RDB.Exists(context.Background(), wechatOfficialClaimKey+state.Code).Val() > 0
	} else {
		consumed = state.Consumed
	}
	status := "pending"
	if consumed {
		status = "consumed"
	} else if state.OpenID != "" {
		status = "ready"
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"status": status, "expires_at": state.ExpiresAt}})
}

// WeChatOfficialLogin exchanges a ready browser state for the normal gateway
// login session. Registration uses the same affiliate-code handling as legacy
// WeChat login.
func WeChatOfficialLogin(c *gin.Context) {
	if !wechatOfficialConfigured() {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "官方微信公众号登录未配置"})
		return
	}
	state, err := loadWeChatOfficialState(c.Query("code"))
	if err != nil || state.OpenID == "" {
		if err == nil {
			err = errors.New("请先使用微信扫码")
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if state.Consumed {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "微信登录二维码已使用"})
		return
	}
	consumed, consumeErr := consumeWeChatOfficialState(state.Code)
	if consumeErr != nil || !consumed {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "微信登录二维码已使用"})
		return
	}
	loginWithWeChatID(c, "official:"+common.WeChatOfficialAppID+":"+state.OpenID, c.Query("aff"))
}

type wechatOfficialCallbackMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Event        string   `xml:"Event"`
	EventKey     string   `xml:"EventKey"`
}

func WeChatOfficialCallback(c *gin.Context) {
	if !wechatOfficialCallbackConfigured() {
		c.String(http.StatusForbidden, "wechat official login is not configured")
		return
	}
	if !VerifyWeChatOfficialSignature(common.WeChatOfficialToken, c.Query("timestamp"), c.Query("nonce"), c.Query("signature")) {
		c.String(http.StatusForbidden, "invalid signature")
		return
	}
	if c.Request.Method == http.MethodGet {
		c.String(http.StatusOK, c.Query("echostr"))
		return
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		c.String(http.StatusOK, "success")
		return
	}
	var message wechatOfficialCallbackMessage
	if err := xml.Unmarshal(body, &message); err != nil {
		c.String(http.StatusOK, "success")
		return
	}
	scene := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(message.EventKey), "qrscene_"), "or_")
	validLoginScan := strings.EqualFold(message.Event, "SCAN") || strings.EqualFold(message.Event, "subscribe")
	isLoginScene := strings.HasPrefix(strings.TrimSpace(message.EventKey), "or_") || strings.HasPrefix(strings.TrimSpace(message.EventKey), "qrscene_or_")
	if validLoginScan && isLoginScene {
		if message.FromUserName != "" {
			if _, err := claimWeChatOfficialScan(scene, message.FromUserName); err != nil {
				c.String(http.StatusInternalServerError, "success")
				return
			}
		}
	}
	// Only acknowledge the login QR scan with a user-visible reply. Ordinary
	// account messages and unrelated QR events are acknowledged with "success".
	if validLoginScan && isLoginScene && message.FromUserName != "" {
		response := struct {
			XMLName      xml.Name `xml:"xml"`
			ToUserName   string   `xml:"ToUserName"`
			FromUserName string   `xml:"FromUserName"`
			CreateTime   int64    `xml:"CreateTime"`
			MsgType      string   `xml:"MsgType"`
			Content      string   `xml:"Content"`
		}{ToUserName: message.FromUserName, FromUserName: message.ToUserName, CreateTime: time.Now().Unix(), MsgType: "text", Content: fmt.Sprintf("扫码成功。你正在登录 %s，请确认这是你本人打开的网页后返回完成登录。", common.SystemName)}
		c.XML(http.StatusOK, response)
		return
	}
	c.String(http.StatusOK, "success")
}

func loginWithWeChatID(c *gin.Context, wechatID, affiliateCode string) {
	var user model.User
	created := false
	inviterID := 0
	if affiliateCode != "" {
		inviterID, _ = model.GetUserIdByAffCode(strings.TrimSpace(affiliateCode))
	}
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		lookup := model.User{}
		if err := tx.Where("wechat_id = ?", wechatID).First(&lookup).Error; err == nil {
			user = lookup
			return model.ClaimExternalIdentityWithTx(tx, model.ExternalIdentityProviderWeChatOfficial, wechatID, user.Id)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if !common.RegisterEnabled {
			return errors.New("管理员关闭了新用户注册")
		}
		user = model.User{
			WeChatId:  wechatID,
			Username:  "wechat_" + common.GetRandomString(10),
			DisplayName: "WeChat User",
			Role:      common.RoleCommonUser,
			Status:    common.UserStatusEnabled,
			InviterId: inviterID,
		}
		if err := user.InsertWithTx(tx, inviterID); err != nil {
			return err
		}
		created = true
		return model.ClaimExternalIdentityWithTx(tx, model.ExternalIdentityProviderWeChatOfficial, wechatID, user.Id)
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if created {
		user.FinalizeOAuthUserCreation(inviterID)
	}
	if user.Status != common.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户已被封禁"})
		return
	}
	setupLogin(&user, c)
}
