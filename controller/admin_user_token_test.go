package controller

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func TestAdminUserTokenListMasksKeyAndUpdateEnforcesOwnership(t *testing.T) {
	db := setupTokenControllerTestDB(t)
	token := seedToken(t, db, 42, "Production", "secret-admin-managed-token")

	listContext, listRecorder := newAuthenticatedContext(t, http.MethodGet, "/api/user/42/tokens?p=1&size=20", nil, 1)
	listContext.Params = gin.Params{{Key: "id", Value: "42"}}
	GetAdminUserTokens(listContext)
	listResponse := decodeAPIResponse(t, listRecorder)
	if !listResponse.Success {
		t.Fatalf("expected token list to succeed, got %q", listResponse.Message)
	}
	var page struct {
		Items []tokenResponseItem `json:"items"`
	}
	if err := common.Unmarshal(listResponse.Data, &page); err != nil {
		t.Fatalf("failed to decode token list: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].Key == token.Key || page.Items[0].Key == "" {
		t.Fatalf("expected a masked API key in the list, got %#v", page.Items)
	}

	updateContext, updateRecorder := newAuthenticatedContext(t, http.MethodPut, "/api/user/42/tokens/"+stringInt(token.Id), map[string]any{
		"id":     token.Id,
		"name":   "Renamed key",
		"status": common.TokenStatusDisabled,
	}, 1)
	updateContext.Params = gin.Params{{Key: "id", Value: "42"}, {Key: "token_id", Value: stringInt(token.Id)}}
	UpdateAdminUserToken(updateContext)
	if response := decodeAPIResponse(t, updateRecorder); !response.Success {
		t.Fatalf("expected token update to succeed, got %q", response.Message)
	}

	updated := &model.Token{}
	if err := db.First(updated, token.Id).Error; err != nil {
		t.Fatalf("failed to load updated token: %v", err)
	}
	if updated.Name != "Renamed key" || updated.Status != common.TokenStatusDisabled {
		t.Fatalf("expected name and status update, got name=%q status=%d", updated.Name, updated.Status)
	}

	otherUserContext, otherUserRecorder := newAuthenticatedContext(t, http.MethodPut, "/api/user/43/tokens/"+stringInt(token.Id), map[string]any{
		"id":     token.Id,
		"name":   "Unauthorized rename",
		"status": common.TokenStatusEnabled,
	}, 1)
	otherUserContext.Params = gin.Params{{Key: "id", Value: "43"}, {Key: "token_id", Value: stringInt(token.Id)}}
	UpdateAdminUserToken(otherUserContext)
	if otherUserRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected updating another user's token to return 404, got %d", otherUserRecorder.Code)
	}
}

func TestAdminUserTokenUpdateEditsNormalKeySettings(t *testing.T) {
	db := setupTokenControllerTestDB(t)
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("failed to migrate users table: %v", err)
	}
	user := &model.User{Id: 42, Username: "key-owner", Password: "password123", Group: "default", Status: common.UserStatusEnabled}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create token owner: %v", err)
	}
	token := seedToken(t, db, user.Id, "Production", "secret-admin-edit-token")
	token.Group = "auto"
	token.CrossGroupRetry = true
	if err := token.SetAutoGroups([]string{"default"}); err != nil {
		t.Fatalf("failed to set initial Auto groups: %v", err)
	}
	if err := db.Save(token).Error; err != nil {
		t.Fatalf("failed to save initial Auto token: %v", err)
	}

	autoUpdateContext, autoUpdateRecorder := newAuthenticatedContext(t, http.MethodPut, "/api/user/42/tokens/"+stringInt(token.Id), map[string]any{
		"id":                 token.Id,
		"group":              "auto",
		"auto_groups":        []string{"vip"},
		"cross_group_retry":  false,
		"auto_route_strategy": "price",
	}, 1)
	autoUpdateContext.Params = gin.Params{{Key: "id", Value: "42"}, {Key: "token_id", Value: stringInt(token.Id)}}
	UpdateAdminUserToken(autoUpdateContext)
	if response := decodeAPIResponse(t, autoUpdateRecorder); !response.Success {
		t.Fatalf("expected Auto routing settings update to succeed, got %q", response.Message)
	}
	if err := db.First(token, token.Id).Error; err != nil {
		t.Fatalf("failed to load updated Auto token: %v", err)
	}
	autoGroups, err := token.GetAutoGroups()
	if err != nil {
		t.Fatalf("failed to read updated Auto groups: %v", err)
	}
	if token.Group != "auto" || token.CrossGroupRetry || token.AutoRouteStrategy != "price" || len(autoGroups) != 1 || autoGroups[0] != "vip" {
		t.Fatalf("expected Auto routing values to persist, got group=%q retry=%t strategy=%q groups=%v", token.Group, token.CrossGroupRetry, token.AutoRouteStrategy, autoGroups)
	}

	expiredTime := common.GetTimestamp() + 3600
	updateContext, updateRecorder := newAuthenticatedContext(t, http.MethodPut, "/api/user/42/tokens/"+stringInt(token.Id), map[string]any{
		"id":                       token.Id,
		"name":                     "  Updated key  ",
		"group":                    "vip",
		"expired_time":             expiredTime,
		"remain_quota":              12345,
		"unlimited_quota":           false,
		"model_limits_enabled":       false,
		"model_limits":              "gpt-test,claude-test",
		"allow_ips":                 "192.0.2.10\n2001:db8::1",
		"allow_user_quota_query":     false,
		"cross_group_retry":          false,
		"auto_route_strategy":        "",
		"auto_groups":               []string{},
	}, 1)
	updateContext.Params = gin.Params{{Key: "id", Value: "42"}, {Key: "token_id", Value: stringInt(token.Id)}}
	UpdateAdminUserToken(updateContext)
	if response := decodeAPIResponse(t, updateRecorder); !response.Success {
		t.Fatalf("expected full token settings update to succeed, got %q", response.Message)
	}

	updated := &model.Token{}
	if err := db.First(updated, token.Id).Error; err != nil {
		t.Fatalf("failed to load updated token: %v", err)
	}
	if updated.Name != "Updated key" || updated.Group != "vip" || updated.ExpiredTime != expiredTime {
		t.Fatalf("expected name, group, and expiry updates, got %#v", updated)
	}
	if updated.RemainQuota != 12345 || updated.UnlimitedQuota {
		t.Fatalf("expected finite quota update, got quota=%d unlimited=%t", updated.RemainQuota, updated.UnlimitedQuota)
	}
	if updated.ModelLimitsEnabled || updated.ModelLimits != "gpt-test,claude-test" {
		t.Fatalf("expected model-limit values including explicit false, got enabled=%t models=%q", updated.ModelLimitsEnabled, updated.ModelLimits)
	}
	if updated.AllowIps == nil || *updated.AllowIps != "192.0.2.10\n2001:db8::1" || updated.AllowUserQuotaQuery {
		t.Fatalf("expected IP allowlist and explicit quota-query false, got ips=%v quota_query=%t", updated.AllowIps, updated.AllowUserQuotaQuery)
	}
	groups, err := updated.GetAutoGroups()
	if err != nil {
		t.Fatalf("failed to read updated Auto groups: %v", err)
	}
	if updated.CrossGroupRetry || updated.AutoRouteStrategy != "" || len(groups) != 0 {
		t.Fatalf("expected non-Auto group to clear Auto settings, got retry=%t strategy=%q groups=%v", updated.CrossGroupRetry, updated.AutoRouteStrategy, groups)
	}

	invalidGroupContext, invalidGroupRecorder := newAuthenticatedContext(t, http.MethodPut, "/api/user/42/tokens/"+stringInt(token.Id), map[string]any{
		"id":    token.Id,
		"group": "not-a-user-group",
	}, 1)
	invalidGroupContext.Params = gin.Params{{Key: "id", Value: "42"}, {Key: "token_id", Value: stringInt(token.Id)}}
	UpdateAdminUserToken(invalidGroupContext)
	if response := decodeAPIResponse(t, invalidGroupRecorder); response.Success {
		t.Fatal("expected an unavailable group to be rejected")
	}
	if err := db.First(updated, token.Id).Error; err != nil {
		t.Fatalf("failed to reload token after rejected update: %v", err)
	}
	if updated.Group != "vip" {
		t.Fatalf("rejected group update changed token group to %q", updated.Group)
	}
}

func TestAdminUserTokenUpdateCannotReenableExhaustedKey(t *testing.T) {
	db := setupTokenControllerTestDB(t)
	token := seedToken(t, db, 42, "Production", "secret-admin-exhausted-token")
	token.Status = common.TokenStatusExhausted
	token.RemainQuota = 0
	token.UnlimitedQuota = false
	if err := db.Save(token).Error; err != nil {
		t.Fatalf("failed to save exhausted token: %v", err)
	}

	context, recorder := newAuthenticatedContext(t, http.MethodPut, "/api/user/42/tokens/"+stringInt(token.Id), map[string]any{
		"id":     token.Id,
		"status": common.TokenStatusEnabled,
	}, 1)
	context.Params = gin.Params{{Key: "id", Value: "42"}, {Key: "token_id", Value: stringInt(token.Id)}}
	UpdateAdminUserToken(context)
	if response := decodeAPIResponse(t, recorder); response.Success {
		t.Fatal("expected an exhausted API key to remain disabled")
	}
	var stored model.Token
	if err := db.First(&stored, token.Id).Error; err != nil {
		t.Fatalf("failed to reload token: %v", err)
	}
	if stored.Status != common.TokenStatusExhausted {
		t.Fatalf("expected exhausted status to remain unchanged, got %d", stored.Status)
	}
}
