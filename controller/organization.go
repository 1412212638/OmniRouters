package controller

import (
	"errors"
	"net/mail"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type organizationMembershipResponse struct {
	Organization     model.Organization       `json:"organization"`
	Membership       model.OrganizationMember `json:"membership"`
	Period           string                   `json:"period"`
	MonthlyUsedQuota int                      `json:"monthly_used_quota"`
}

func organizationMember(c *gin.Context, organizationId int) (*model.OrganizationMember, error) {
	if organizationId <= 0 {
		return nil, errors.New("organization id is invalid")
	}
	member, err := model.GetOrganizationMember(organizationId, c.GetInt("id"))
	if err != nil || member.Status != model.OrganizationMemberActive {
		return nil, errors.New("organization membership is not active")
	}
	return member, nil
}

func organizationOverview(organizationId, userId int) (*organizationMembershipResponse, error) {
	organization, err := model.GetOrganizationById(organizationId)
	if err != nil {
		return nil, err
	}
	membership, err := model.GetOrganizationMember(organizationId, userId)
	if err != nil {
		return nil, err
	}
	period := model.CurrentOrganizationUsagePeriod()
	var usage model.OrganizationMemberUsage
	_ = model.DB.Where("organization_id = ? AND user_id = ? AND period = ?", organizationId, userId, period).First(&usage).Error
	if !model.OrganizationCanManage(membership.Role) {
		organization.Quota = 0
		organization.UsedQuota = 0
		organization.RequestCount = 0
	}
	return &organizationMembershipResponse{Organization: *organization, Membership: *membership, Period: period, MonthlyUsedQuota: usage.UsedQuota}, nil
}

func GetOrganizations(c *gin.Context) {
	var memberships []model.OrganizationMember
	if err := model.DB.Where("user_id = ? AND status = ?", c.GetInt("id"), model.OrganizationMemberActive).Find(&memberships).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	items := make([]organizationMembershipResponse, 0, len(memberships))
	for _, membership := range memberships {
		item, err := organizationOverview(membership.OrganizationId, c.GetInt("id"))
		if err != nil {
			common.ApiError(c, err)
			return
		}
		items = append(items, *item)
	}
	common.ApiSuccess(c, items)
}

type createOrganizationRequest struct {
	Name string `json:"name"`
}

func CreateOrganization(c *gin.Context) {
	var req createOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "invalid request")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 128 {
		common.ApiErrorMsg(c, "organization name must be between 1 and 128 characters")
		return
	}
	var organizationCount int64
	if err := model.DB.Model(&model.Organization{}).Where("owner_user_id = ?", c.GetInt("id")).Count(&organizationCount).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	if organizationCount >= 20 {
		common.ApiErrorMsg(c, "organization creation limit reached")
		return
	}
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	now := common.GetTimestamp()
	organization := &model.Organization{Name: req.Name, OwnerUserId: userId, Group: user.Group, CreatedAt: now, UpdatedAt: now}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(organization).Error; err != nil {
			return err
		}
		return tx.Create(&model.OrganizationMember{OrganizationId: organization.Id, UserId: userId, Role: model.OrganizationMemberOwner, Status: model.OrganizationMemberActive, CreatedAt: now, UpdatedAt: now}).Error
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	item, err := organizationOverview(organization.Id, userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, item)
}

func GetOrganization(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	if _, err := organizationMember(c, id); err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	item, err := organizationOverview(id, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, item)
}

func UpdateOrganization(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	member, err := organizationMember(c, id)
	if err != nil || !model.OrganizationCanManage(member.Role) {
		common.ApiErrorMsg(c, "organization manager permission required")
		return
	}
	var req createOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "invalid request")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 128 {
		common.ApiErrorMsg(c, "organization name must be between 1 and 128 characters")
		return
	}
	if err := model.DB.Model(&model.Organization{}).Where("id = ?", id).Updates(map[string]interface{}{"name": req.Name, "updated_at": common.GetTimestamp()}).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	item, err := organizationOverview(id, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, item)
}

type organizationMemberResponse struct {
	model.OrganizationMember
	Username         string `json:"username"`
	DisplayName      string `json:"display_name"`
	Email            string `json:"email"`
	MonthlyUsedQuota int    `json:"monthly_used_quota"`
	Period           string `json:"period"`
}

func GetOrganizationMembers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	actor, err := organizationMember(c, id)
	if err != nil || !model.OrganizationCanManage(actor.Role) {
		common.ApiErrorMsg(c, "organization manager permission required")
		return
	}
	var members []model.OrganizationMember
	if err := model.DB.Where("organization_id = ? AND status = ?", id, model.OrganizationMemberActive).Order("id asc").Find(&members).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	period := model.CurrentOrganizationUsagePeriod()
	result := make([]organizationMemberResponse, 0, len(members))
	for _, member := range members {
		var user model.User
		if err := model.DB.Select("username", "display_name", "email").First(&user, member.UserId).Error; err != nil {
			continue
		}
		var usage model.OrganizationMemberUsage
		_ = model.DB.Where("organization_id = ? AND user_id = ? AND period = ?", id, member.UserId, period).First(&usage).Error
		result = append(result, organizationMemberResponse{OrganizationMember: member, Username: user.Username, DisplayName: user.DisplayName, Email: user.Email, MonthlyUsedQuota: usage.UsedQuota, Period: period})
	}
	common.ApiSuccess(c, result)
}

type updateOrganizationMemberRequest struct {
	Role              string `json:"role"`
	MonthlyQuotaLimit *int   `json:"monthly_quota_limit"`
}

func UpdateOrganizationMember(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	targetId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid user id")
		return
	}
	var req updateOrganizationMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "invalid request")
		return
	}
	previousMember, _ := model.GetOrganizationMember(id, targetId)
	if err := model.UpdateOrganizationMember(id, c.GetInt("id"), targetId, req.Role, req.MonthlyQuotaLimit); err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	if req.Role != "" && (previousMember == nil || previousMember.Role != req.Role) {
		service.TryNotifyOrganization(service.OrganizationNotification{Kind: service.OrganizationNotificationRoleChanged, OrganizationId: id, UserId: targetId, Role: req.Role})
	}
	common.ApiSuccess(c, nil)
}

func RemoveOrganizationMember(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	targetId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid user id")
		return
	}
	actor, err := organizationMember(c, id)
	if err != nil || !model.OrganizationCanManage(actor.Role) {
		common.ApiErrorMsg(c, "organization manager permission required")
		return
	}
	target, err := model.GetOrganizationMember(id, targetId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if target.Role == model.OrganizationMemberOwner || (actor.Role != model.OrganizationMemberOwner && target.Role == model.OrganizationMemberAdmin) {
		common.ApiErrorMsg(c, "only owner can manage admins")
		return
	}
	if err := model.RemoveOrganizationMembership(id, c.GetInt("id"), targetId); err != nil {
		common.ApiError(c, err)
		return
	}
	service.TryNotifyOrganization(service.OrganizationNotification{Kind: service.OrganizationNotificationRemoved, OrganizationId: id, UserId: targetId})
	common.ApiSuccess(c, nil)
}

func LeaveOrganization(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	member, err := organizationMember(c, id)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	if member.Role == model.OrganizationMemberOwner {
		common.ApiErrorMsg(c, "organization owner cannot leave")
		return
	}
	if err := model.RemoveOrganizationMembership(id, c.GetInt("id"), c.GetInt("id")); err != nil {
		common.ApiError(c, err)
		return
	}
	service.TryNotifyOrganization(service.OrganizationNotification{Kind: service.OrganizationNotificationRemoved, OrganizationId: id, UserId: c.GetInt("id")})
	common.ApiSuccess(c, nil)
}

type organizationInvitationRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func CreateOrganizationInvitation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	var req organizationInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "invalid request")
		return
	}
	req.Email = model.NormalizeEmail(req.Email)
	if req.Email == "" {
		common.ApiErrorMsg(c, "email is required")
		return
	}
	if req.Role == "" {
		req.Role = model.OrganizationMemberMember
	}
	if req.Role != model.OrganizationMemberMember && req.Role != model.OrganizationMemberAdmin {
		common.ApiErrorMsg(c, "invalid organization role")
		return
	}
	if parsed, emailErr := mail.ParseAddress(req.Email); emailErr != nil || parsed.Address != req.Email {
		common.ApiErrorMsg(c, "invalid email")
		return
	}
	var invitation model.OrganizationInvitation
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var organization model.Organization
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&organization, id).Error; err != nil {
			return err
		}
		var actor model.OrganizationMember
		if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", id, c.GetInt("id"), model.OrganizationMemberActive).First(&actor).Error; err != nil {
			return err
		}
		if !model.OrganizationCanManage(actor.Role) {
			return errors.New("organization manager permission required")
		}
		if actor.Role != model.OrganizationMemberOwner && req.Role == model.OrganizationMemberAdmin {
			return errors.New("only owner can invite admins")
		}
		now := common.GetTimestamp()
		var existing model.OrganizationInvitation
		if err := tx.Where("organization_id = ? AND email = ? AND status = ? AND expires_at > ?", id, req.Email, model.OrganizationInvitationPending, now).First(&existing).Error; err == nil {
			return errors.New("an invitation is already pending")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var target model.User
		if err := tx.Where("LOWER(email) = ?", req.Email).First(&target).Error; err == nil {
			var membership model.OrganizationMember
			if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", id, target.Id, model.OrganizationMemberActive).First(&membership).Error; err == nil {
				return errors.New("user is already a member")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		invitation = model.OrganizationInvitation{OrganizationId: id, Email: req.Email, Role: req.Role, InvitedBy: c.GetInt("id"), Status: model.OrganizationInvitationPending, ExpiresAt: now + 7*24*60*60, CreatedAt: now}
		return tx.Create(&invitation).Error
	})
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	service.TryNotifyOrganization(service.OrganizationNotification{Kind: service.OrganizationNotificationInvitation, OrganizationId: id, Email: invitation.Email, Role: invitation.Role, ExpiresAt: invitation.ExpiresAt})
	common.ApiSuccess(c, invitation)
}

func ListOrganizationInvitations(c *gin.Context) {
	user, err := model.GetUserById(c.GetInt("id"), false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	invitations := []struct {
		model.OrganizationInvitation
		OrganizationName string `json:"organization_name"`
	}{}
	if user.Email == "" {
		common.ApiSuccess(c, invitations)
		return
	}
	if err := model.DB.Table("organization_invitations").Select("organization_invitations.*, organizations.name AS organization_name").Joins("JOIN organizations ON organizations.id = organization_invitations.organization_id").Where("organization_invitations.email = ? AND organization_invitations.status = ? AND organization_invitations.expires_at > ?", model.NormalizeEmail(user.Email), model.OrganizationInvitationPending, common.GetTimestamp()).Order("organization_invitations.id desc").Scan(&invitations).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, invitations)
}

func AcceptOrganizationInvitation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("invitation_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid invitation id")
		return
	}
	user, err := model.GetUserById(c.GetInt("id"), false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var invitation model.OrganizationInvitation
		if err := tx.Where("id = ? AND email = ? AND status = ? AND expires_at > ?", id, model.NormalizeEmail(user.Email), model.OrganizationInvitationPending, common.GetTimestamp()).First(&invitation).Error; err != nil {
			return err
		}
		var organization model.Organization
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&organization, invitation.OrganizationId).Error; err != nil {
			return err
		}
		var inviter model.OrganizationMember
		if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", invitation.OrganizationId, invitation.InvitedBy, model.OrganizationMemberActive).First(&inviter).Error; err != nil {
			return err
		}
		if !model.OrganizationCanManage(inviter.Role) || (invitation.Role == model.OrganizationMemberAdmin && inviter.Role != model.OrganizationMemberOwner) {
			return errors.New("inviter no longer has permission")
		}
		var member model.OrganizationMember
		if err := tx.Where("organization_id = ? AND user_id = ?", invitation.OrganizationId, user.Id).First(&member).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			member = model.OrganizationMember{OrganizationId: invitation.OrganizationId, UserId: user.Id, Role: invitation.Role, Status: model.OrganizationMemberActive, CreatedAt: common.GetTimestamp(), UpdatedAt: common.GetTimestamp()}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			if member.Status == model.OrganizationMemberActive {
				return errors.New("user is already a member")
			}
			if err := tx.Model(&member).Updates(map[string]interface{}{"role": invitation.Role, "status": model.OrganizationMemberActive, "updated_at": common.GetTimestamp()}).Error; err != nil {
				return err
			}
		}
		result := tx.Model(&model.OrganizationInvitation{}).Where("id = ? AND status = ? AND expires_at > ?", invitation.Id, model.OrganizationInvitationPending, common.GetTimestamp()).Update("status", model.OrganizationInvitationAccepted)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("invitation is no longer available")
		}
		return nil
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func RevokeOrganizationInvitation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("invitation_id"))
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "invalid invitation id")
		return
	}
	organizationId, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil || organizationId <= 0 {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		var organization model.Organization
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&organization, organizationId).Error; err != nil {
			return err
		}
		var actor model.OrganizationMember
		if err := tx.Where("organization_id = ? AND user_id = ? AND status = ?", organizationId, c.GetInt("id"), model.OrganizationMemberActive).First(&actor).Error; err != nil {
			return err
		}
		if !model.OrganizationCanManage(actor.Role) {
			return errors.New("organization manager permission required")
		}
		var invitation model.OrganizationInvitation
		if err := tx.Where("id = ? AND organization_id = ? AND status = ?", id, organizationId, model.OrganizationInvitationPending).First(&invitation).Error; err != nil {
			return err
		}
		if invitation.Role == model.OrganizationMemberAdmin && actor.Role != model.OrganizationMemberOwner {
			return errors.New("only owner can manage admin invitations")
		}
		result := tx.Model(&model.OrganizationInvitation{}).Where("id = ? AND organization_id = ? AND status = ?", id, organizationId, model.OrganizationInvitationPending).Update("status", model.OrganizationInvitationRevoked)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("invitation is no longer available")
		}
		return nil
	})
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	common.ApiSuccess(c, nil)
}

func GetOrganizationInvitations(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	actor, err := organizationMember(c, id)
	if err != nil || !model.OrganizationCanManage(actor.Role) {
		common.ApiErrorMsg(c, "organization manager permission required")
		return
	}
	var invitations []model.OrganizationInvitation
	if err := model.DB.Where("organization_id = ?", id).Order("id desc").Find(&invitations).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, invitations)
}

func GetOrganizationUsage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	actor, err := organizationMember(c, id)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	organization, err := model.GetOrganizationById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	period := model.CurrentOrganizationUsagePeriod()
	if !model.OrganizationCanManage(actor.Role) {
		var own model.OrganizationMemberUsage
		_ = model.DB.Where("organization_id = ? AND user_id = ? AND period = ?", id, c.GetInt("id"), period).First(&own).Error
		common.ApiSuccess(c, gin.H{"organization_id": id, "period": period, "monthly_used_quota": own.UsedQuota, "monthly_quota_limit": actor.MonthlyQuotaLimit})
		return
	}
	var usage []model.OrganizationMemberUsage
	if err := model.DB.Where("organization_id = ? AND period = ?", id, period).Find(&usage).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"organization_id": id, "period": period, "quota": organization.Quota, "used_quota": organization.UsedQuota, "request_count": organization.RequestCount, "members": usage})
}

func TransferOrganizationWallet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	var req struct {
		Quota int `json:"quota"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Quota <= 0 {
		common.ApiErrorMsg(c, "quota must be positive")
		return
	}
	if err := model.TransferUserQuotaToOrganization(id, c.GetInt("id"), req.Quota); err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	common.ApiSuccess(c, nil)
}

func GetOrganizationWalletTransactions(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	actor, err := organizationMember(c, id)
	if err != nil || !model.OrganizationCanManage(actor.Role) {
		common.ApiErrorMsg(c, "organization manager permission required")
		return
	}
	var transactions []model.OrganizationWalletTransaction
	if err := model.DB.Where("organization_id = ?", id).Order("id desc").Limit(100).Find(&transactions).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	userIDs := make([]int, 0, len(transactions))
	seenUsers := make(map[int]struct{}, len(transactions))
	for _, transaction := range transactions {
		if transaction.UserId > 0 {
			if _, ok := seenUsers[transaction.UserId]; !ok {
				seenUsers[transaction.UserId] = struct{}{}
				userIDs = append(userIDs, transaction.UserId)
			}
		}
	}
	users, err := model.GetUsersByIds(userIDs)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	userNames := make(map[int]string, len(users))
	for _, user := range users {
		name := strings.TrimSpace(user.DisplayName)
		if name == "" {
			name = user.Username
		}
		userNames[user.Id] = name
	}
	type organizationWalletTransactionResponse struct {
		model.OrganizationWalletTransaction
		UserName string `json:"user_name"`
	}
	result := make([]organizationWalletTransactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		result = append(result, organizationWalletTransactionResponse{
			OrganizationWalletTransaction: transaction,
			UserName:                      userNames[transaction.UserId],
		})
	}
	common.ApiSuccess(c, result)
}

type organizationTokenResponse struct {
	*tokenResponse
	CreatorName string `json:"creator_name"`
}

func buildOrganizationMaskedTokenResponses(tokens []*model.Token) []*organizationTokenResponse {
	masked := buildMaskedTokenResponses(tokens)
	userIDs := make([]int, 0, len(tokens))
	seenUsers := make(map[int]struct{}, len(tokens))
	for _, token := range tokens {
		if token == nil || token.UserId <= 0 {
			continue
		}
		if _, ok := seenUsers[token.UserId]; !ok {
			seenUsers[token.UserId] = struct{}{}
			userIDs = append(userIDs, token.UserId)
		}
	}
	users, err := model.GetUsersByIds(userIDs)
	if err != nil {
		common.SysError("failed to load organization token creators: " + err.Error())
	}
	userNames := make(map[int]string, len(users))
	for _, user := range users {
		name := strings.TrimSpace(user.DisplayName)
		if name == "" {
			name = user.Username
		}
		userNames[user.Id] = name
	}
	result := make([]*organizationTokenResponse, 0, len(masked))
	for index, token := range masked {
		creatorID := 0
		if index < len(tokens) && tokens[index] != nil {
			creatorID = tokens[index].UserId
		}
		result = append(result, &organizationTokenResponse{tokenResponse: token, CreatorName: userNames[creatorID]})
	}
	return result
}

func GetOrganizationTokens(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	actor, err := organizationMember(c, id)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	query := model.DB.Where("organization_id = ?", id)
	if !model.OrganizationCanManage(actor.Role) {
		query = query.Where("user_id = ?", c.GetInt("id"))
	}
	var tokens []*model.Token
	if err := query.Order("id desc").Limit(1000).Find(&tokens).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, buildOrganizationMaskedTokenResponses(tokens))
}

func CreateOrganizationToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" || len(req.Name) > 50 {
		common.ApiErrorMsg(c, "token name must be between 1 and 50 characters")
		return
	}
	token, err := model.CreateOrganizationToken(id, c.GetInt("id"), req.Name)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	common.ApiSuccess(c, token)
}

func UpdateOrganizationToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	actor, err := organizationMember(c, id)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	tokenId, _ := strconv.Atoi(c.Param("token_id"))
	token, err := model.GetOrganizationToken(id, tokenId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if token.UserId != c.GetInt("id") && !model.OrganizationCanManage(actor.Role) {
		common.ApiErrorMsg(c, "organization token permission required")
		return
	}
	var req struct {
		Name   *string `json:"name"`
		Status *int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "invalid request")
		return
	}
	if req.Name != nil && (strings.TrimSpace(*req.Name) == "" || len(*req.Name) > 50) {
		common.ApiErrorMsg(c, "invalid token name")
		return
	}
	if req.Status != nil && *req.Status != common.TokenStatusEnabled && *req.Status != common.TokenStatusDisabled {
		common.ApiErrorMsg(c, "invalid token status")
		return
	}
	token, err = model.UpdateOrganizationToken(id, tokenId, req.Name, req.Status)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	response := buildOrganizationMaskedTokenResponses([]*model.Token{token})
	if len(response) == 1 {
		common.ApiSuccess(c, response[0])
		return
	}
	common.ApiSuccess(c, buildMaskedTokenResponse(token))
}

func DeleteOrganizationToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	actor, err := organizationMember(c, id)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	tokenId, _ := strconv.Atoi(c.Param("token_id"))
	token, err := model.GetOrganizationToken(id, tokenId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if token.UserId != c.GetInt("id") && !model.OrganizationCanManage(actor.Role) {
		common.ApiErrorMsg(c, "organization token permission required")
		return
	}
	if err := model.DeleteOrganizationToken(id, tokenId); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func GetOrganizationTokenKey(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("organization_id"))
	if err != nil {
		common.ApiErrorMsg(c, "invalid organization id")
		return
	}
	member, err := organizationMember(c, id)
	if err != nil {
		common.ApiErrorMsg(c, "organization token permission required")
		return
	}
	tokenId, _ := strconv.Atoi(c.Param("token_id"))
	token, err := model.GetOrganizationToken(id, tokenId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if token.UserId != c.GetInt("id") && !model.OrganizationCanManage(member.Role) {
		common.ApiErrorMsg(c, "organization token permission required")
		return
	}
	common.ApiSuccess(c, gin.H{"key": token.GetFullKey()})
}
