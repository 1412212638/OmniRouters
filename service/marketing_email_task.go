package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const marketingEmailBatchSize = 100

type marketingEmailSendHandler struct{}

func init() {
	RegisterSystemTaskHandler(marketingEmailSendHandler{})
}

func (marketingEmailSendHandler) Type() string { return model.SystemTaskTypeMarketingEmailSend }

func (marketingEmailSendHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	payload := MarketingEmailSendPayload{}
	if err := task.DecodePayload(&payload); err != nil {
		failSystemTask(task, runnerID, err)
		return
	}
	if len(payload.UserIDs) == 0 {
		failSystemTask(task, runnerID, errors.New("marketing email audience is empty"))
		return
	}

	result := &MarketingEmailResult{Total: len(payload.UserIDs)}
	for start := 0; start < len(payload.UserIDs); start += marketingEmailBatchSize {
		if err := ctx.Err(); err != nil {
			failSystemTask(task, runnerID, err)
			return
		}
		end := start + marketingEmailBatchSize
		if end > len(payload.UserIDs) {
			end = len(payload.UserIDs)
		}
		users, err := model.GetUsersByIds(payload.UserIDs[start:end])
		if err != nil {
			failSystemTask(task, runnerID, err)
			return
		}
		userByID := make(map[int]model.User, len(users))
		for _, user := range users {
			userByID[user.Id] = user
		}
		for _, userID := range payload.UserIDs[start:end] {
			user, ok := userByID[userID]
			if !ok {
				result.Skipped++
				continue
			}
			if !model.IsMarketingEmailUserEligible(&user) || !marketingEmailUserHasDestination(&user) {
				result.Skipped++
				continue
			}
			if err := SendMarketingEmailToUser(&user, payload.SubjectTemplate, payload.ContentTemplate, common.SendEmail); err != nil {
				result.Failed++
				continue
			}
			result.Sent++
		}

		state := SystemTaskProgress{Total: result.Total, Processed: end, Progress: end * 100 / result.Total}
		if err := model.UpdateSystemTaskState(task.TaskID, runnerID, state); err != nil {
			logSystemTaskLockError(ctx, task, err)
			return
		}
	}
	common.SysLog(fmt.Sprintf("marketing email campaign completed: total=%d sent=%d skipped=%d failed=%d", result.Total, result.Sent, result.Skipped, result.Failed))
	if err := model.FinishSystemTask(task.TaskID, runnerID, model.SystemTaskStatusSucceeded, result, ""); err != nil {
		logSystemTaskLockError(ctx, task, err)
	}
}
