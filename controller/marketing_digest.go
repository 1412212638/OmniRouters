package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type marketingDigestGenerateRequest struct {
	PeriodDays int `json:"period_days"`
}

type marketingDigestUpdateRequest struct {
	Subject       string `json:"subject"`
	Content       string `json:"content"`
	AudienceType  string `json:"audience_type"`
	AudienceGroup string `json:"audience_group"`
	UserIDs       []int  `json:"user_ids"`
}

func ListMarketingDigestPromotions(c *gin.Context) {
	items, err := model.ListMarketingDigestPromotions()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, items)
}

func SaveMarketingDigestPromotion(c *gin.Context) {
	var req service.MarketingDigestPromotionInput
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "invalid promotion")
		return
	}
	req.ModelName = strings.TrimSpace(req.ModelName)
	req.Discount = strings.TrimSpace(req.Discount)
	req.Description = strings.TrimSpace(req.Description)
	req.URL = strings.TrimSpace(req.URL)
	if req.ModelName == "" || len(req.ModelName) > 128 {
		common.ApiErrorMsg(c, "model name is required")
		return
	}
	if req.StartsAt > 0 && req.EndsAt > 0 && req.EndsAt < req.StartsAt {
		common.ApiErrorMsg(c, "promotion end date must not precede its start date")
		return
	}
	if req.URL != "" {
		parsed, err := url.ParseRequestURI(req.URL)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			common.ApiErrorMsg(c, "promotion URL must use HTTP or HTTPS")
			return
		}
	}
	promotion := &model.MarketingDigestPromotion{
		ID: req.ID, ModelName: req.ModelName, Description: req.Description,
		Discount: req.Discount, StartsAt: req.StartsAt, EndsAt: req.EndsAt,
		URL: req.URL, Enabled: req.Enabled,
	}
	if err := model.SaveMarketingDigestPromotion(promotion); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, promotion)
}

func DeleteMarketingDigestPromotion(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "invalid promotion id")
		return
	}
	if err := model.DeleteMarketingDigestPromotion(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"id": id})
}

func ListMarketingDigests(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	digests, err := model.ListMarketingDigests(limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	for i := range digests {
		counts, countErr := model.GetMarketingDigestRecipientCounts(digests[i].ID)
		if countErr != nil {
			common.ApiError(c, countErr)
			return
		}
		digests[i].RecipientCount = counts.Total
		digests[i].SentCount = counts.Sent
		digests[i].SkippedCount = counts.Skipped
		digests[i].FailedCount = counts.Failed
	}
	common.ApiSuccess(c, digests)
}

func GetMarketingDigest(c *gin.Context) {
	digest, ok := marketingDigestFromPath(c)
	if !ok {
		return
	}
	counts, err := model.GetMarketingDigestRecipientCounts(digest.ID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	digest.RecipientCount = counts.Total
	digest.SentCount = counts.Sent
	digest.SkippedCount = counts.Skipped
	digest.FailedCount = counts.Failed
	common.ApiSuccess(c, gin.H{"digest": digest, "counts": counts})
}

func GenerateMarketingDigest(c *gin.Context) {
	var req marketingDigestGenerateRequest
	if c.Request.ContentLength > 0 {
		if err := common.DecodeJson(c.Request.Body, &req); err != nil {
			common.ApiErrorMsg(c, "invalid generation request")
			return
		}
	}
	digest, created, err := service.GenerateManualMarketingDigest(req.PeriodDays)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	message := "existing draft returned"
	if created {
		message = "draft generated and reviewer notified"
	}
	common.ApiSuccess(c, gin.H{"digest": digest, "created": created, "message": message})
}

func UpdateMarketingDigest(c *gin.Context) {
	digest, ok := marketingDigestFromPath(c)
	if !ok {
		return
	}
	var req marketingDigestUpdateRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		common.ApiErrorMsg(c, "invalid newsletter draft")
		return
	}
	req.Subject = strings.TrimSpace(req.Subject)
	req.Content = strings.TrimSpace(req.Content)
	req.AudienceType = strings.TrimSpace(req.AudienceType)
	req.AudienceGroup = strings.TrimSpace(req.AudienceGroup)
	if req.Subject == "" || req.Content == "" {
		common.ApiErrorMsg(c, "subject and content are required")
		return
	}
	if len(req.Subject) > 998 || len(req.Content) > 2*1024*1024 {
		common.ApiErrorMsg(c, "newsletter subject or content is too long")
		return
	}
	if req.AudienceType != "all" && req.AudienceType != "topup" && req.AudienceType != "active" && req.AudienceType != "group" && req.AudienceType != "manual" {
		common.ApiErrorMsg(c, "invalid audience type")
		return
	}
	if req.AudienceType == "group" && req.AudienceGroup == "" {
		common.ApiErrorMsg(c, "audience group is required")
		return
	}
	if req.AudienceType == "manual" && len(req.UserIDs) == 0 {
		common.ApiErrorMsg(c, "select at least one recipient")
		return
	}
	userIDsJSON, err := common.Marshal(uniqueMarketingDigestIDs(req.UserIDs))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := model.UpdateMarketingDigestDraft(digest.ID, req.Subject, req.Content, req.AudienceType, req.AudienceGroup, string(userIDsJSON)); err != nil {
		common.ApiError(c, err)
		return
	}
	updated, err := model.GetMarketingDigest(digest.ID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, updated)
}

func SendMarketingDigest(c *gin.Context) {
	digest, ok := marketingDigestFromPath(c)
	if !ok {
		return
	}
	task, err := service.StartMarketingDigestSend(digest.ID, c.GetInt("id"), false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"task": task.ToResponse(), "digest_id": digest.ID})
}

func RetryMarketingDigest(c *gin.Context) {
	digest, ok := marketingDigestFromPath(c)
	if !ok {
		return
	}
	task, err := service.StartMarketingDigestSend(digest.ID, c.GetInt("id"), true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"task": task.ToResponse(), "digest_id": digest.ID})
}

func ListMarketingDigestRecipients(c *gin.Context) {
	digest, ok := marketingDigestFromPath(c)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}
	recipients, total, err := model.ListMarketingDigestRecipients(digest.ID, c.Query("status"), limit, offset)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"items": recipients, "total": total, "limit": limit, "offset": offset})
}

func marketingDigestFromPath(c *gin.Context) (*model.MarketingDigest, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "invalid newsletter id")
		return nil, false
	}
	digest, err := model.GetMarketingDigest(id)
	if err != nil {
		common.ApiError(c, err)
		return nil, false
	}
	return digest, true
}

func uniqueMarketingDigestIDs(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func ShowMarketingEmailUnsubscribe(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil || !service.VerifyMarketingDigestUnsubscribeToken(userID, c.Query("token")) {
		c.String(http.StatusBadRequest, "This unsubscribe link is invalid.")
		return
	}
	_, err = model.GetUserById(userID, true)
	if err != nil {
		c.String(http.StatusNotFound, "This account could not be found.")
		return
	}
	page := `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Email preferences</title></head><body style="font-family:system-ui,sans-serif;max-width:520px;margin:12vh auto;padding:24px;color:#1f2937"><h1>Email preferences</h1><p>Stop receiving product updates and offers from this platform?</p><form method="post" action="/api/marketing_email/unsubscribe"><input type="hidden" name="user_id" value="` + strconv.Itoa(userID) + `"><input type="hidden" name="token" value="` + url.QueryEscape(c.Query("token")) + `"><button type="submit">Unsubscribe</button></form></body></html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

func ApplyMarketingEmailUnsubscribe(c *gin.Context) {
	userID, err := strconv.Atoi(c.PostForm("user_id"))
	if err != nil || !service.VerifyMarketingDigestUnsubscribeToken(userID, c.PostForm("token")) {
		c.String(http.StatusBadRequest, "This unsubscribe link is invalid.")
		return
	}
	if err := model.UpdateMarketingDigestEmailPreference(userID, false); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "This account could not be found.")
			return
		}
		c.String(http.StatusInternalServerError, "Could not update email preferences.")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<!doctype html><html><head><meta charset=\"utf-8\"><title>Email preferences</title></head><body style=\"font-family:system-ui,sans-serif;max-width:520px;margin:12vh auto;padding:24px;color:#1f2937\"><h1>Unsubscribed</h1><p>You will no longer receive product updates and offers.</p></body></html>"))
}

func logMarketingDigestAudit(c *gin.Context, digestID int64, action string) {
	model.RecordLog(c.GetInt("id"), model.LogTypeSystem, fmt.Sprintf("营销简报 #%d %s", digestID, action))
}
