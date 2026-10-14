package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetInvoiceProfile(c *gin.Context) {
	user, err := model.GetUserById(c.GetInt("id"), false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": user.GetInvoiceProfile()})
}

func UpdateInvoiceProfile(c *gin.Context) {
	var profile model.InvoiceProfile
	if err := common.DecodeJson(c.Request.Body, &profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid invoice profile"})
		return
	}
	if err := profile.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := model.SetUserInvoiceProfile(c.GetInt("id"), &profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "success", "data": profile})
}

func DeleteInvoiceProfile(c *gin.Context) {
	if err := model.SetUserInvoiceProfile(c.GetInt("id"), nil); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "success"})
}
