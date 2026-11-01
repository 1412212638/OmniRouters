package controller

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	identityNamePattern = regexp.MustCompile(`^[\p{L}\p{M}·.]{2,64}$`)
	identityIDPattern   = regexp.MustCompile(`^[0-9Xx]{15,18}$`)
	identityCompanyPattern = regexp.MustCompile(`^.{2,128}$`)
	identityRegistrationPattern = regexp.MustCompile(`^[0-9A-Za-z]{8,32}$`)
)

type personalVerificationRequest struct {
	Name   string `json:"name"`
	IDCard string `json:"id_card"`
}

type enterpriseVerificationRequest struct {
	Company            string `json:"company_name"`
	RegistrationNumber string `json:"registration_number"`
	Name               string `json:"name"`
	IDCard             string `json:"id_card"`
}

func identityMethods(cfg service.RealNameConfig) []string {
	methods := make([]string, 0, 2)
	if cfg.PersonalAvailable() { methods = append(methods, service.RealNameTypePersonal) }
	if cfg.EnterpriseAvailable() { methods = append(methods, service.RealNameTypeEnterprise) }
	return methods
}

func maskedIdentity(value string) string {
	runes := []rune(value)
	if len(runes) <= 1 { return "*" }
	if len(runes) == 2 { return string(runes[:1]) + "*" }
	return string(runes[:1]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1:])
}

func identitySuffix(value string) string {
	runes := []rune(value)
	if len(runes) <= 4 { return string(runes) }
	return string(runes[len(runes)-4:])
}

func publicVerification(record *model.RealNameVerification) gin.H {
	if record == nil { return nil }
	return gin.H{
		"type": record.Type, "status": record.Status, "display_name": record.DisplayName,
		"id_card_suffix": record.IDCardSuffix, "company_name_masked": record.CompanyNameMasked,
		"registration_suffix": record.RegistrationSuffix, "provider": record.Provider,
		"verified_at": record.VerifiedAt,
	}
}

func GetIdentityVerification(c *gin.Context) {
	cfg := service.GetRealNameConfig()
	record, err := model.GetRealNameVerification(c.GetInt("id"))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		common.ApiError(c, err); return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": gin.H{
		"enabled": cfg.Enabled, "required": cfg.Required, "encryption_configured": cfg.EncryptionConfigured, "available_methods": identityMethods(cfg),
		"verification": publicVerification(record),
	}})
}

func requireVerificationMethod(c *gin.Context, method string) bool {
	cfg := service.GetRealNameConfig()
	if !cfg.Enabled { common.ApiErrorMsg(c, "实名认证功能未启用"); return false }
	if method == service.RealNameTypePersonal && !cfg.PersonalAvailable() { common.ApiErrorMsg(c, "个人实名认证暂不可用"); return false }
	if method == service.RealNameTypeEnterprise && !cfg.EnterpriseAvailable() { common.ApiErrorMsg(c, "企业实名认证暂不可用"); return false }
	if method != service.RealNameTypePersonal && method != service.RealNameTypeEnterprise { common.ApiErrorMsg(c, "不支持的认证类型"); return false }
	record, err := model.GetRealNameVerification(c.GetInt("id"))
	if err == nil && record.Status == service.RealNameStatusVerified { common.ApiErrorMsg(c, "认证成功后不能修改认证类型或认证资料"); return false }
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) { common.ApiError(c, err); return false }
	return true
}

func SavePersonalVerification(c *gin.Context) {
	if !requireVerificationMethod(c, service.RealNameTypePersonal) { return }
	var req personalVerificationRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil { common.ApiErrorI18n(c, i18n.MsgInvalidParams); return }
	req.Name, req.IDCard = strings.TrimSpace(req.Name), strings.TrimSpace(req.IDCard)
	if !identityNamePattern.MatchString(req.Name) || !identityIDPattern.MatchString(req.IDCard) { common.ApiErrorMsg(c, "姓名或身份证号格式不正确"); return }
	cfg := service.GetRealNameConfig()
	ok, orderNo, err := service.VerifyPersonal(c.Request.Context(), cfg.PersonalAppCode, req.Name, req.IDCard)
	if err != nil { c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "实名认证服务暂时不可用"}); return }
	if !ok { common.ApiErrorMsg(c, "姓名和身份证号不匹配"); return }
	nameCipher, err := service.EncryptRealName(req.Name); if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "认证信息保存失败"}); return }
	idCipher, err := service.EncryptRealName(req.IDCard); if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "认证信息保存失败"}); return }
	record := &model.RealNameVerification{UserID: c.GetInt("id"), Type: service.RealNameTypePersonal, Status: service.RealNameStatusVerified, PersonalNameCipher: nameCipher, PersonalIDCardCipher: idCipher, DisplayName: maskedIdentity(req.Name), IDCardSuffix: identitySuffix(req.IDCard), Provider: "aliyun", UpstreamReference: orderNo, VerifiedAt: common.GetTimestamp()}
	if err := model.SaveRealNameVerification(record); err != nil { common.ApiError(c, err); return }
	common.ApiSuccess(c, publicVerification(record))
}

func SaveEnterpriseVerification(c *gin.Context) {
	if !requireVerificationMethod(c, service.RealNameTypeEnterprise) { return }
	var req enterpriseVerificationRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil { common.ApiErrorI18n(c, i18n.MsgInvalidParams); return }
	req.Company, req.RegistrationNumber, req.Name, req.IDCard = strings.TrimSpace(req.Company), strings.TrimSpace(req.RegistrationNumber), strings.TrimSpace(req.Name), strings.TrimSpace(req.IDCard)
	if !identityCompanyPattern.MatchString(req.Company) || !identityRegistrationPattern.MatchString(req.RegistrationNumber) || !identityNamePattern.MatchString(req.Name) || !identityIDPattern.MatchString(req.IDCard) { common.ApiErrorMsg(c, "企业认证信息格式不正确"); return }
	cfg := service.GetRealNameConfig()
	ok, seqNo, err := service.VerifyEnterprise(c.Request.Context(), cfg.EnterpriseAppCode, req.Company, req.RegistrationNumber, req.Name, req.IDCard)
	if err != nil { c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "企业认证服务暂时不可用"}); return }
	if !ok { common.ApiErrorMsg(c, "企业信息核验不一致"); return }
	values := []string{req.Company, req.RegistrationNumber, req.Name, req.IDCard}
	ciphers := make([]string, len(values))
	for i, value := range values { ciphers[i], err = service.EncryptRealName(value); if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "认证信息保存失败"}); return } }
	record := &model.RealNameVerification{UserID: c.GetInt("id"), Type: service.RealNameTypeEnterprise, Status: service.RealNameStatusVerified, EnterpriseNameCipher: ciphers[0], RegistrationNumberCipher: ciphers[1], RepresentativeNameCipher: ciphers[2], RepresentativeIDCardCipher: ciphers[3], DisplayName: maskedIdentity(req.Name), IDCardSuffix: identitySuffix(req.IDCard), CompanyNameMasked: maskedIdentity(req.Company), RegistrationSuffix: identitySuffix(req.RegistrationNumber), Provider: "aliyun", UpstreamReference: seqNo, VerifiedAt: common.GetTimestamp()}
	if err := model.SaveRealNameVerification(record); err != nil { common.ApiError(c, err); return }
	common.ApiSuccess(c, publicVerification(record))
}

func GetIdentityVerificationByAdmin(c *gin.Context) {
	userID, err := parsePositiveID(c.Param("id")); if err != nil { common.ApiError(c, err); return }
	record, err := model.GetRealNameVerification(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) { common.ApiSuccess(c, nil); return }
	if err != nil { common.ApiError(c, err); return }
	common.ApiSuccess(c, publicVerification(record))
}

func ViewIdentityVerificationByAdmin(c *gin.Context) {
	userID, err := parsePositiveID(c.Param("id")); if err != nil { common.ApiError(c, err); return }
	record, err := model.GetRealNameVerification(userID); if err != nil { common.ApiError(c, err); return }
	result := gin.H{"type": record.Type, "status": record.Status, "provider": record.Provider, "upstream_reference": record.UpstreamReference, "verified_at": record.VerifiedAt}
	if record.Type == service.RealNameTypePersonal { name, e1 := service.DecryptRealName(record.PersonalNameCipher); id, e2 := service.DecryptRealName(record.PersonalIDCardCipher); if e1 != nil || e2 != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "认证信息解密失败"}); return }; result["name"], result["id_card"] = name, id }
	if record.Type == service.RealNameTypeEnterprise { company, e1 := service.DecryptRealName(record.EnterpriseNameCipher); reg, e2 := service.DecryptRealName(record.RegistrationNumberCipher); name, e3 := service.DecryptRealName(record.RepresentativeNameCipher); id, e4 := service.DecryptRealName(record.RepresentativeIDCardCipher); if e1 != nil || e2 != nil || e3 != nil || e4 != nil { c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "认证信息解密失败"}); return }; result["company_name"], result["registration_number"], result["name"], result["id_card"] = company, reg, name, id }
	common.ApiSuccess(c, result)
}

type resetIdentityVerificationRequest struct { Reason string `json:"reason"` }

func ResetIdentityVerificationByAdmin(c *gin.Context) {
	userID, err := parsePositiveID(c.Param("id")); if err != nil { common.ApiError(c, err); return }
	var req resetIdentityVerificationRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil || len([]rune(strings.TrimSpace(req.Reason))) < 5 { common.ApiErrorMsg(c, "请填写至少 5 个字符的重置原因"); return }
	record, err := model.GetRealNameVerification(userID); if err != nil { common.ApiError(c, err); return }
	if err := model.ResetRealNameVerification(userID); err != nil { common.ApiError(c, err); return }
	recordManageAuditFor(c, userID, "user.real_name_reset", map[string]interface{}{"type": record.Type, "reason": strings.TrimSpace(req.Reason)})
	common.ApiSuccess(c, nil)
}

func parsePositiveID(value string) (int, error) { var id int; _, err := fmt.Sscanf(value, "%d", &id); if err != nil || id <= 0 { return 0, errors.New("invalid user id") }; return id, nil }
