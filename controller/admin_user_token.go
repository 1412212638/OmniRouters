package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type adminUserTokenCreateRequest struct {
	Name string `json:"name"`
}

type adminUserTokenUpdateRequest struct {
	Id                  int                  `json:"id"`
	Name                *string              `json:"name"`
	Status              *int                 `json:"status"`
	Group               *string              `json:"group"`
	ExpiredTime         *int64               `json:"expired_time"`
	RemainQuota         *int                 `json:"remain_quota"`
	UnlimitedQuota      *bool                `json:"unlimited_quota"`
	ModelLimitsEnabled  *bool                `json:"model_limits_enabled"`
	ModelLimits         *string              `json:"model_limits"`
	AllowIps            *string              `json:"allow_ips"`
	CrossGroupRetry     *bool                `json:"cross_group_retry"`
	AutoRouteStrategy   *string              `json:"auto_route_strategy"`
	AllowUserQuotaQuery *bool                `json:"allow_user_quota_query"`
	AutoGroups          tokenAutoGroupsInput `json:"auto_groups"`
}

func adminUserTokenTarget(c *gin.Context) (int, bool) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid user id"})
		return 0, false
	}
	return userID, true
}

func adminUserSelectableGroup(user *model.UserBase, group string) bool {
	if user == nil || group == "" {
		return false
	}
	_, ok := service.GetUserUsableGroupsWithExtras(user.Group, user.GetExtraGroups())[group]
	return ok && ratio_setting.ContainsGroupRatio(group)
}

func GetAdminUserTokens(c *gin.Context) {
	userID, ok := adminUserTokenTarget(c)
	if !ok {
		return
	}
	page := common.GetPageQuery(c)
	tokens, err := model.GetAllUserTokens(userID, page.GetStartIdx(), page.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	total, err := model.CountUserTokens(userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	page.SetTotal(int(total))
	page.SetItems(buildMaskedTokenResponses(tokens))
	common.ApiSuccess(c, page)
}

func GetAdminUserTokenEditorConfig(c *gin.Context) {
	userID, ok := adminUserTokenTarget(c)
	if !ok {
		return
	}
	user, err := model.GetUserCache(userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	usableGroups := service.GetUserUsableGroupsWithExtras(user.Group, user.GetExtraGroups())
	groups := make(map[string]map[string]interface{}, len(usableGroups))
	groupNames := make([]string, 0, len(usableGroups))
	for groupName, description := range usableGroups {
		if groupName == "auto" {
			continue
		}
		groupNames = append(groupNames, groupName)
		groups[groupName] = map[string]interface{}{
			"ratio": service.GetUserGroupRatio(user.Group, groupName),
			"desc":  description,
		}
	}
	if _, exists := usableGroups["auto"]; exists {
		groups["auto"] = map[string]interface{}{"ratio": "自动", "desc": setting.GetUsableGroupDescription("auto")}
	}
	common.ApiSuccess(c, gin.H{
		"groups":     groups,
		"models":     service.GetGroupsEnabledModels(groupNames),
		"auto_groups": service.GetUserAutoGroupWithExtras(user.Group, user.GetExtraGroups()),
		"max_auto_groups": setting.GetMaxTokenAutoGroups(),
	})
}

func CreateAdminUserToken(c *gin.Context) {
	userID, ok := adminUserTokenTarget(c)
	if !ok {
		return
	}
	user, err := model.GetUserById(userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "user not found"})
			return
		}
		common.ApiError(c, err)
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "user not found"})
		return
	}
	var request adminUserTokenCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	name := strings.TrimSpace(request.Name)
	if name == "" || len(name) > 50 {
		common.ApiErrorMsg(c, "API key name must be between 1 and 50 characters")
		return
	}
	count, err := model.CountUserTokens(userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if int(count) >= operation_setting.GetMaxUserTokens() {
		common.ApiErrorMsg(c, "user has reached the maximum API key limit")
		return
	}
	key, err := common.GenerateKey()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	group, err := model.GetUserGroup(userID, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	token := &model.Token{
		UserId: userID, Key: key, Name: name, Status: common.TokenStatusEnabled,
		CreatedTime: common.GetTimestamp(), AccessedTime: common.GetTimestamp(),
		ExpiredTime: -1, UnlimitedQuota: true, Group: group,
		KeyType: model.TokenKeyTypeGeneral, AllowUserQuotaQuery: true,
	}
	if err := token.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAuditFor(c, userID, "user.token.create", map[string]interface{}{"id": token.Id, "name": token.Name, "target_user_id": userID})
	common.ApiSuccess(c, gin.H{"token": buildMaskedTokenResponse(token), "key": "sk-" + token.GetFullKey()})
}

func UpdateAdminUserToken(c *gin.Context) {
	userID, ok := adminUserTokenTarget(c)
	if !ok {
		return
	}
	var request adminUserTokenUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	tokenID, err := strconv.Atoi(c.Param("token_id"))
	if err != nil || tokenID <= 0 || request.Id != tokenID {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid API key id"})
		return
	}
	token, err := model.GetTokenByIds(tokenID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "API key not found"})
			return
		}
		common.ApiError(c, err)
		return
	}
	if token.KeyType != "" && token.KeyType != model.TokenKeyTypeGeneral {
		common.ApiErrorMsg(c, "only general API keys can be managed here")
		return
	}
	if request.Name != nil {
		name := strings.TrimSpace(*request.Name)
		if name == "" || len(name) > 50 {
			common.ApiErrorMsg(c, "API key name must be between 1 and 50 characters")
			return
		}
		token.Name = name
	}
	if request.Status != nil {
		if *request.Status != common.TokenStatusEnabled && *request.Status != common.TokenStatusDisabled {
			common.ApiErrorMsg(c, "invalid API key status")
			return
		}
		if *request.Status == common.TokenStatusEnabled {
			if token.ExpiredTime != -1 && token.ExpiredTime <= common.GetTimestamp() {
				common.ApiErrorMsg(c, "expired API key cannot be enabled")
				return
			}
			if token.Status == common.TokenStatusExhausted && token.RemainQuota <= 0 && !token.UnlimitedQuota {
				common.ApiErrorMsg(c, "exhausted API key cannot be enabled")
				return
			}
		}
		token.Status = *request.Status
	}
	if request.Group != nil || request.AutoGroups.Set || request.AutoRouteStrategy != nil || request.CrossGroupRetry != nil {
		if request.Group != nil {
			token.Group = strings.TrimSpace(*request.Group)
		}
		user, userErr := model.GetUserCache(userID)
		if userErr != nil {
			common.ApiError(c, userErr)
			return
		}
		if token.Group != "auto" {
			if !adminUserSelectableGroup(user, token.Group) {
				common.ApiErrorMsg(c, "invalid API key group")
				return
			}
			token.CrossGroupRetry = false
			token.AutoRouteStrategy = ""
			if err := token.SetAutoGroups(nil); err != nil {
				common.ApiError(c, err)
				return
			}
		} else {
			if request.AutoRouteStrategy != nil {
				token.AutoRouteStrategy = *request.AutoRouteStrategy
			}
			if token.AutoRouteStrategy != "" && token.AutoRouteStrategy != autoRouteStrategyPrice {
				common.ApiErrorMsg(c, "invalid Auto routing strategy")
				return
			}
			if request.CrossGroupRetry != nil {
				token.CrossGroupRetry = *request.CrossGroupRetry
			}
			if request.AutoGroups.Set {
				groups := request.AutoGroups.Groups
				if len(groups) > setting.GetMaxTokenAutoGroups() {
					common.ApiErrorMsg(c, "too many Auto groups")
					return
				}
				seen := make(map[string]struct{}, len(groups))
				for _, group := range groups {
					if _, duplicate := seen[group]; duplicate || group == "auto" || !adminUserSelectableGroup(user, group) {
						common.ApiErrorMsg(c, "invalid Auto group selection")
						return
					}
					seen[group] = struct{}{}
				}
				if err := token.SetAutoGroups(groups); err != nil {
					common.ApiError(c, err)
					return
				}
			}
		}
	}
	if request.ExpiredTime != nil {
		token.ExpiredTime = *request.ExpiredTime
	}
	if request.RemainQuota != nil {
		token.RemainQuota = *request.RemainQuota
	}
	if request.UnlimitedQuota != nil {
		token.UnlimitedQuota = *request.UnlimitedQuota
	}
	if request.ModelLimitsEnabled != nil {
		token.ModelLimitsEnabled = *request.ModelLimitsEnabled
	}
	if request.ModelLimits != nil {
		token.ModelLimits = *request.ModelLimits
	}
	if request.AllowIps != nil {
		token.AllowIps = request.AllowIps
	}
	if request.AllowUserQuotaQuery != nil {
		token.AllowUserQuotaQuery = *request.AllowUserQuotaQuery
	}
	if !token.UnlimitedQuota && (token.RemainQuota < 0 || token.RemainQuota > maxTokenQuota()) {
		common.ApiErrorMsg(c, "API key quota is outside the allowed range")
		return
	}
	if err := token.Update(); err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAuditFor(c, userID, "user.token.update", map[string]interface{}{"id": token.Id, "name": token.Name, "status": token.Status, "group": token.Group, "target_user_id": userID})
	common.ApiSuccess(c, buildMaskedTokenResponse(token))
}

