package relay

import (
	"fmt"
	"io"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	taskopenaiimage "github.com/QuantumNous/new-api/relay/channel/task/openai_image"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

func ImageTaskSubmit(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) *types.NewAPIError {
	if info.TaskRelayInfo == nil {
		info.TaskRelayInfo = &relaycommon.TaskRelayInfo{}
	}
	if info.PublicTaskID == "" {
		info.PublicTaskID = model.GenerateTaskID()
	}
	info.Action = constant.TaskActionGenerate

	adaptor := &taskopenaiimage.TaskAdaptor{}
	adaptor.Init(info)
	if taskErr := adaptor.ValidateRequestAndSetAction(c, info); taskErr != nil {
		return newAPIErrorFromTaskError(taskErr)
	}

	resp, err := adaptor.DoRequest(c, info, requestBody)
	if err != nil {
		return types.NewOpenAIError(err, types.ErrorCodeDoRequestFailed, http.StatusInternalServerError)
	}
	if resp == nil {
		return types.NewOpenAIError(fmt.Errorf("empty upstream response"), types.ErrorCodeDoRequestFailed, http.StatusInternalServerError)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return service.RelayErrorHandler(c.Request.Context(), resp, false)
	}

	upstreamTaskID, taskData, taskErr := adaptor.DoResponse(c, resp, info)
	if taskErr != nil {
		return newAPIErrorFromTaskError(taskErr)
	}

	actualQuota := info.PriceData.QuotaToPreConsume
	info.PriceData.Quota = actualQuota

	if settleErr := service.SettleBilling(c, info, actualQuota); settleErr != nil {
		common.SysError("settle image task billing error: " + settleErr.Error())
	}
	service.LogTaskConsumption(c, info)

	task := model.InitTask(constant.TaskPlatformOpenAIImage, info)
	task.PrivateData.UpstreamTaskID = upstreamTaskID
	task.PrivateData.BillingSource = info.BillingSource
	task.PrivateData.SubscriptionId = info.SubscriptionId
	task.PrivateData.TokenId = info.TokenId
	task.PrivateData.BillingContext = &model.TaskBillingContext{
		ModelPrice:      info.PriceData.ModelPrice,
		GroupRatio:      info.PriceData.GroupRatioInfo.GroupRatio,
		ModelRatio:      info.PriceData.ModelRatio,
		OtherRatios:     info.PriceData.OtherRatios,
		OriginModelName: info.OriginModelName,
		PerCallBilling:  true,
	}
	task.Quota = actualQuota
	task.Action = info.Action
	task.Data = taskData
	if insertErr := task.Insert(); insertErr != nil {
		common.SysError("insert image task error: " + insertErr.Error())
	}
	return nil
}

func newAPIErrorFromTaskError(taskErr *dto.TaskError) *types.NewAPIError {
	if taskErr == nil {
		return nil
	}
	if taskErr.Error != nil {
		return types.NewOpenAIError(taskErr.Error, types.ErrorCodeBadResponseStatusCode, taskErr.StatusCode)
	}
	return types.NewOpenAIError(fmt.Errorf(taskErr.Message), types.ErrorCodeBadResponseStatusCode, taskErr.StatusCode)
}
