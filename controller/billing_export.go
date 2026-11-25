package controller

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billexport"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func billingExportScope(c *gin.Context, requestedUserID int) (model.BillingExport, bool) {
	viewerID := c.GetInt("id")
	scope := model.BillingExport{UserID: viewerID}
	if requestedUserID > 0 && requestedUserID != viewerID {
		if service.AuthorizeBillingExportUser(viewerID, requestedUserID) != nil {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "admin_required"})
			return scope, false
		}
		scope.UserID = requestedUserID
		scope.RequestedBy = viewerID
	}
	if id := c.Query("organization_id"); id != "" {
		var err error
		scope.OrganizationID, err = strconv.Atoi(id)
		if err != nil || scope.OrganizationID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid_scope"})
			return scope, false
		}
	}
	if requestedUserID > 0 && requestedUserID != viewerID && scope.OrganizationID > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid_scope"})
		return scope, false
	}
	if err := service.AuthorizeBillingExport(scope.UserID, scope.OrganizationID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "manager_required"})
		return scope, false
	}
	return scope, true
}

func ListBillingExports(c *gin.Context) {
	requestedUserID, err := requestedBillingUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid_user"})
		return
	}
	scope, ok := billingExportScope(c, requestedUserID)
	if !ok {
		return
	}
	reports, err := model.ListBillingExports(c.Request.Context(), scope.UserID, scope.OrganizationID, time.Now().Unix())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, reports)
}

func CreateBillingExport(c *gin.Context) {
	var request struct {
		StartMonth     string `json:"start_month"`
		EndMonth       string `json:"end_month"`
		Language       string `json:"language"`
		OrganizationID int    `json:"organization_id"`
		UserID         int    `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid_period"})
		return
	}
	if request.UserID < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid_user"})
		return
	}
	scope, ok := billingExportScope(c, request.UserID)
	if !ok {
		return
	}
	if request.UserID == 0 || request.UserID == c.GetInt("id") {
		scope.RequestedBy = scope.UserID
	}
	if request.OrganizationID > 0 && request.OrganizationID != scope.OrganizationID {
		if request.UserID > 0 && request.UserID != c.GetInt("id") {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid_scope"})
			return
		}
		scope.OrganizationID = request.OrganizationID
		if err := service.AuthorizeBillingExport(c.GetInt("id"), scope.OrganizationID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "manager_required"})
			return
		}
	}
	scope.StartMonth, scope.EndMonth = request.StartMonth, request.EndMonth
	scope.Language = request.Language
	switch err := service.EnqueueBillingExport(c.Request.Context(), scope); {
	case errors.Is(err, service.ErrBillingExportLanguage):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid_language"})
	case errors.Is(err, service.ErrBillingExportPeriod):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid_period"})
	case errors.Is(err, model.ErrBillingExportAlreadyQueued):
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "already_queued"})
	case err != nil:
		common.ApiError(c, err)
	default:
		common.ApiSuccess(c, nil)
	}
}

func DownloadBillingExport(c *gin.Context) {
	requestedUserID, err := requestedBillingUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid_user"})
		return
	}
	scope, ok := billingExportScope(c, requestedUserID)
	if !ok {
		return
	}
	level := c.Param("level")
	name, exists := billexport.FileNames[level]
	if !exists {
		c.Status(http.StatusNotFound)
		return
	}
	ctx := c.Request.Context()
	report, err := model.GetBillingExport(ctx, c.Param("task_id"))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		common.ApiError(c, err)
		return
	}
	if err != nil || report.UserID != scope.UserID || report.OrganizationID != scope.OrganizationID {
		c.Status(http.StatusNotFound)
		return
	}
	archive, err := model.ReadBillingExportArchive(ctx, report.TaskID, scope.UserID, scope.OrganizationID, time.Now().Unix())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}
	filename := fmt.Sprintf("OmniRouters-%s-%s-%s", report.StartMonth, report.EndMonth, name)
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Language", report.EffectiveLanguage())
	attachment := mime.FormatMediaType("attachment", map[string]string{"filename": filename})
	if level == "zip" {
		c.Header("Content-Disposition", attachment)
		c.Data(http.StatusOK, "application/zip", archive)
		return
	}
	zipped, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	file, err := zipped.Open(name)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	mediaTypes := map[string]string{"l0": "application/pdf", "l1": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "l2": "text/csv; charset=utf-8", "l3": "text/csv; charset=utf-8"}
	c.DataFromReader(http.StatusOK, info.Size(), mediaTypes[level], file, map[string]string{"Content-Disposition": attachment})
}

func requestedBillingUserID(c *gin.Context) (int, error) {
	value := c.Query("user_id")
	if value == "" {
		return 0, nil
	}
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid user id")
	}
	return id, nil
}
