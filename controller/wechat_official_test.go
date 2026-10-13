package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestVerifyWeChatOfficialSignature(t *testing.T) {
	token, timestamp, nonce := "secret-token", "1700000000", "nonce-1"
	signature := BuildWeChatOfficialSignature(token, timestamp, nonce)
	require.Equal(t, "1ef00fd4f4a432071c4514834970a7ccf6b825d3", signature)
	require.True(t, VerifyWeChatOfficialSignature(token, timestamp, nonce, signature))
	require.True(t, VerifyWeChatOfficialSignature(token, timestamp, nonce, strings.ToUpper(signature)))
	require.False(t, VerifyWeChatOfficialSignature(token, timestamp, nonce, "invalid"))
	require.False(t, VerifyWeChatOfficialSignature("wrong-token", timestamp, nonce, signature))
}

func TestWeChatOfficialCallbackUpdatesScannedLoginState(t *testing.T) {
	previousEnabled, previousName, previousAppID, previousSecret, previousToken, previousRedis, previousRDB := common.WeChatOfficialEnabled, common.WeChatOfficialName, common.WeChatOfficialAppID, common.WeChatOfficialAppSecret, common.WeChatOfficialToken, common.RedisEnabled, common.RDB
	t.Cleanup(func() {
		common.WeChatOfficialEnabled = previousEnabled
		common.WeChatOfficialName = previousName
		common.WeChatOfficialAppID = previousAppID
		common.WeChatOfficialAppSecret = previousSecret
		common.WeChatOfficialToken = previousToken
		common.RedisEnabled = previousRedis
		common.RDB = previousRDB
	})
	// The platform handshake must work before the administrator enables login.
	common.WeChatOfficialEnabled = false
	common.WeChatOfficialToken = "callback-token"
	common.WeChatOfficialAppID = "app-id"
	common.WeChatOfficialAppSecret = "app-secret"
	common.RedisEnabled = false
	common.RDB = nil

	code := "callback-state"
	require.NoError(t, saveWeChatOfficialState(wechatOfficialLoginState{
		Code: code, ExpiresAt: time.Now().Add(time.Minute).Unix(),
	}))

	gin.SetMode(gin.TestMode)
	timestamp, nonce := "1700000000", "nonce"
	body := `<xml><ToUserName><![CDATA[gh_account]]></ToUserName><FromUserName><![CDATA[openid-123]]></FromUserName><CreateTime>1700000000</CreateTime><MsgType><![CDATA[event]]></MsgType><Event><![CDATA[SCAN]]></Event><EventKey><![CDATA[or_callback-state]]></EventKey></xml>`
	req := httptest.NewRequest(http.MethodPost, "/api/oauth/wechat/official/callback?timestamp="+timestamp+"&nonce="+nonce+"&signature="+BuildWeChatOfficialSignature(common.WeChatOfficialToken, timestamp, nonce), strings.NewReader(body))
	res := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(res)
	ctx.Request = req
	WeChatOfficialCallback(ctx)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "扫码成功")
	state, err := loadWeChatOfficialState(code)
	require.NoError(t, err)
	require.Equal(t, "openid-123", state.OpenID)
}

func TestWeChatOfficialCallbackRejectsInvalidSignature(t *testing.T) {
	previousEnabled, previousAppID, previousSecret, previousToken, previousRedis, previousRDB := common.WeChatOfficialEnabled, common.WeChatOfficialAppID, common.WeChatOfficialAppSecret, common.WeChatOfficialToken, common.RedisEnabled, common.RDB
	t.Cleanup(func() {
		common.WeChatOfficialEnabled, common.WeChatOfficialAppID, common.WeChatOfficialAppSecret, common.WeChatOfficialToken = previousEnabled, previousAppID, previousSecret, previousToken
		common.RedisEnabled, common.RDB = previousRedis, previousRDB
	})
	common.WeChatOfficialEnabled = true
	common.WeChatOfficialToken = "callback-token"
	common.WeChatOfficialAppID = "app-id"
	common.WeChatOfficialAppSecret = "app-secret"
	common.RedisEnabled = false
	common.RDB = nil

	req := httptest.NewRequest(http.MethodGet, "/api/oauth/wechat/official/callback?timestamp=1&nonce=2&signature=bad&echostr=echo", nil)
	res := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(res)
	ctx.Request = req
	WeChatOfficialCallback(ctx)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Contains(t, res.Body.String(), "invalid signature")
}

func TestWeChatOfficialCallbackEchoesHandshakeChallenge(t *testing.T) {
	previousEnabled, previousAppID, previousSecret, previousToken := common.WeChatOfficialEnabled, common.WeChatOfficialAppID, common.WeChatOfficialAppSecret, common.WeChatOfficialToken
	t.Cleanup(func() {
		common.WeChatOfficialEnabled, common.WeChatOfficialAppID, common.WeChatOfficialAppSecret, common.WeChatOfficialToken = previousEnabled, previousAppID, previousSecret, previousToken
	})
	common.WeChatOfficialEnabled = true
	common.WeChatOfficialAppID = "app-id"
	common.WeChatOfficialAppSecret = "app-secret"
	common.WeChatOfficialToken = "callback-token"
	timestamp, nonce := "1700000000", "nonce"
	req := httptest.NewRequest(http.MethodGet, "/api/oauth/wechat/official/callback?timestamp="+timestamp+"&nonce="+nonce+"&signature="+BuildWeChatOfficialSignature(common.WeChatOfficialToken, timestamp, nonce)+"&echostr=challenge", nil)
	res := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(res)
	ctx.Request = req
	WeChatOfficialCallback(ctx)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "challenge", res.Body.String())
}
