package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func verifyMarketingUnsubscribe(id, emailHash, token string) (userID int, ok bool) {
	if emailHash != "" {
		return 0, id == "" && service.VerifyMarketingAddressUnsubscribeToken(emailHash, token)
	}
	userID, err := strconv.Atoi(id)
	return userID, err == nil && service.VerifyMarketingEmailUnsubscribeToken(userID, token)
}

func ShowMarketingEmailUnsubscribe(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	emailHash, token := c.Query("email_hash"), c.Query("token")
	userID, ok := verifyMarketingUnsubscribe(c.Query("user_id"), emailHash, token)
	if !ok {
		c.String(http.StatusBadRequest, "This unsubscribe link is invalid.")
		return
	}
	field, value := "email_hash", emailHash
	if userID != 0 {
		if _, err := model.GetUserById(userID, false); err != nil {
			c.String(http.StatusNotFound, "This account could not be found.")
			return
		}
		field, value = "user_id", strconv.Itoa(userID)
	}
	page := `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Email preferences</title></head><body style="font-family:system-ui,sans-serif;max-width:520px;margin:12vh auto;padding:24px;color:#1f2937"><h1>Email preferences</h1><p>Stop receiving marketing emails and offers from this platform?</p><form method="post" action="/api/marketing_email/unsubscribe"><input type="hidden" name="` + field + `" value="` + value + `"><input type="hidden" name="token" value="` + token + `"><button type="submit">Unsubscribe</button></form></body></html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
}

func ApplyMarketingEmailUnsubscribe(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8192)
	oneClick := c.PostForm("List-Unsubscribe") == "One-Click"
	id, emailHash, token := c.PostForm("user_id"), c.PostForm("email_hash"), c.PostForm("token")
	if oneClick {
		id, emailHash, token = c.Query("user_id"), c.Query("email_hash"), c.Query("token")
	}
	userID, ok := verifyMarketingUnsubscribe(id, emailHash, token)
	if !ok {
		c.String(http.StatusBadRequest, "This unsubscribe link is invalid.")
		return
	}
	var err error
	if userID == 0 {
		err = model.SuppressMarketingEmailHash(emailHash, "unsubscribe")
	} else {
		err = model.UpdateMarketingEmailPreference(userID, false)
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.String(http.StatusNotFound, "This account could not be found.")
			return
		}
		c.String(http.StatusInternalServerError, "Could not update email preferences.")
		return
	}
	if oneClick {
		c.Status(http.StatusOK)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<!doctype html><html><head><meta charset=\"utf-8\"><title>Email preferences</title></head><body style=\"font-family:system-ui,sans-serif;max-width:520px;margin:12vh auto;padding:24px;color:#1f2937\"><h1>Unsubscribed</h1><p>You will no longer receive marketing emails and offers.</p></body></html>"))
}
