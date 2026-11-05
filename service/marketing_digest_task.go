package service

import (
	"context"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

type marketingDigestGenerationHandler struct{}
type marketingDigestSendHandler struct{}

func init() {
	RegisterSystemTaskHandler(marketingDigestGenerationHandler{})
	RegisterSystemTaskHandler(marketingDigestSendHandler{})
}

func (marketingDigestGenerationHandler) Type() string { return model.SystemTaskTypeMarketingDigestGeneration }
func (marketingDigestGenerationHandler) Enabled() bool { return MarketingDigestScheduleEnabled() }
func (marketingDigestGenerationHandler) Interval() time.Duration {
	return time.Duration(common.MarketingDigestIntervalDays) * 24 * time.Hour
}
func (marketingDigestGenerationHandler) NewPayload() any { return MarketingDigestGenerationTaskPayload() }
func (marketingDigestGenerationHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	RunMarketingDigestGenerationTask(ctx, task, runnerID)
}

func (marketingDigestSendHandler) Type() string { return model.SystemTaskTypeMarketingDigestSend }
func (marketingDigestSendHandler) Run(ctx context.Context, task *model.SystemTask, runnerID string) {
	RunMarketingDigestSendTask(ctx, task, runnerID)
}
