package controller

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ShowMarketingEmailUnsubscribe(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil || !service.VerifyMarketingEmailUnsubscribeToken(userID, c.Query("token")) {
		c.String(http.StatusBadRequest, "This unsubscribe link is invalid.")
		return
	}
	if _, err := model.GetUserById(userID, true); err != nil {
		c.String(http.StatusNotFound, "This account could not be found.")
		return
	}
	page := `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Email preferences</title></head><body style="font-family:system-ui,sans-serif;max-width:520px;margin:12vh auto;padding:24px;color:#1f2937"><h1>Email preferences</h1><p>Stop receiving marketing emails and offers from this platform?</p><form method="post" action="/api/marketing_email/unsubscribe"><input type="hidden" name="user_id" value="` + strconv.Itoa(userID) + `"><input type="hidden" name="token" value="` + url.QueryEscape(c.Query("token")) + `"><button type="submit">Unsubscribe</button></form></body></html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

func ApplyMarketingEmailUnsubscribe(c *gin.Context) {
	userID, err := strconv.Atoi(c.PostForm("user_id"))
	if err != nil || !service.VerifyMarketingEmailUnsubscribeToken(userID, c.PostForm("token")) {
		c.String(http.StatusBadRequest, "This unsubscribe link is invalid.")
		return
	}
	if err := model.UpdateMarketingEmailPreference(userID, false); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "This account could not be found.")
			return
		}
		c.String(http.StatusInternalServerError, "Could not update email preferences.")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<!doctype html><html><head><meta charset=\"utf-8\"><title>Email preferences</title></head><body style=\"font-family:system-ui,sans-serif;max-width:520px;margin:12vh auto;padding:24px;color:#1f2937\"><h1>Unsubscribed</h1><p>You will no longer receive marketing emails and offers.</p></body></html>"))
}
